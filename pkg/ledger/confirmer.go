// Copyright (c) 2021 UMI
//
// Permission is hereby granted, free of charge, to any person obtaining a copy
// of this software and associated documentation files (the "Software"), to deal
// in the Software without restriction, including without limitation the rights
// to use, copy, modify, merge, publish, distribute, sublicense, and/or sell
// copies of the Software, and to permit persons to whom the Software is
// furnished to do so, subject to the following conditions:
//
// The above copyright notice and this permission notice shall be included in all
// copies or substantial portions of the Software.
//
// THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
// IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
// FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
// AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
// LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
// OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN THE
// SOFTWARE.

package ledger

import (
	"errors"
	"fmt"
	"math"
	"sync"

	"gitlab.com/umitop/umid/pkg/umi"
)

var (
	errUnsupportedTxType      = errors.New("unsupported transaction type")
	errNotFound               = errors.New("not found")
	errInsufficientFunds      = errors.New("insufficient funds")
	errAccountType            = errors.New("incorrect account type")
	errStructureExists        = errors.New("structure already exists")
	errStructureDoesNotExist  = errors.New("structure does not exist")
	errInsufficientPrivileges = errors.New("insufficient privileges")
	errBlock                  = errors.New("block")
	ErrTxConfirmed            = errors.New("transaction confirmed")
)

type iBlockchain interface {
	AppendBlock(umi.Block) error
}

type Confirmer struct {
	sync.RWMutex

	ledger     *Ledger
	blockchain iBlockchain

	accounts   map[umi.Address]*Account
	structures map[umi.Prefix]*Structure
	nfts       map[umi.Hash]umi.Address
	txHashes   []umi.Hash

	// ???????????????????? ???????????? ???????????????????? ?? ??????????????????. ???????????? ???????????????????????? ?????? ??????????????????????????.
	TransactionHeight uint64
	// ?????????? ???????????????? ??????????. ???? ???????? ?????????????????? ?????????? ?????????????????? ??????????????.
	BlockTimestamp uint32
	// ???????????? ?????????? ?? ??????????????????.
	BlockHeight uint32
	// ?????? ?????????????????????????????? ??????????. ?????????? ?????????????? ?????? ?????????? LastBlockHash.
	BlockHash umi.Hash
	// ?????? ???????????????????? ?????????????????????????? ???????????????? ??????????. ???? ???????????? ???????? ???? ?????????? ???????????????? ???????????? ????????????????????.
	PrevBlockHash umi.Hash
}

func NewConfirmer(ledger *Ledger) *Confirmer {
	return &Confirmer{
		ledger: ledger,
	}
}

func (confirmer *Confirmer) SetBlockchain(blockchain iBlockchain) {
	confirmer.blockchain = blockchain
}

func (confirmer *Confirmer) AppendBlock(block []byte) error {
	confirmer.Lock()
	defer confirmer.Unlock()

	if err := confirmer.ProcessBlock(block); err != nil {
		return err
	}

	if err := confirmer.blockchain.AppendBlock(block); err != nil {
		return fmt.Errorf("%w", err)
	}

	if err := confirmer.Commit(); err != nil {
		return fmt.Errorf("%w", err)
	}

	return nil
}

func (confirmer *Confirmer) ResetState() {
	confirmer.accounts = make(map[umi.Address]*Account)
	confirmer.structures = make(map[umi.Prefix]*Structure)
	confirmer.nfts = make(map[umi.Hash]umi.Address)
	confirmer.txHashes = make([]umi.Hash, 0)

	confirmer.ledger.RLock()
	confirmer.PrevBlockHash = confirmer.ledger.LastBlockHash
	confirmer.BlockTimestamp = confirmer.ledger.LastBlockTimestamp
	confirmer.BlockHeight = confirmer.ledger.LastBlockHeight
	confirmer.BlockHash = confirmer.ledger.LastBlockHash
	confirmer.TransactionHeight = confirmer.ledger.LastTransactionHeight
	confirmer.ledger.RUnlock()
}

func (confirmer *Confirmer) ProcessBlock(blockRaw []byte) error {
	confirmer.ResetState()

	if err := confirmer.verifyBlock(blockRaw); err != nil {
		return err
	}

	block := (umi.Block)(blockRaw)

	confirmer.BlockHash = block.Hash()
	confirmer.BlockTimestamp = block.Timestamp()
	confirmer.BlockHeight++

	handlers := map[string]func(umi.Transaction) error{
		umi.TxGenesis:             confirmer.processGenesis,
		umi.TxSend:                confirmer.processSend,
		umi.TxCreateStructure:     confirmer.processCreateStructure,
		umi.TxUpdateStructure:     confirmer.processUpdateStructure,
		umi.TxChangeProfitAddress: confirmer.processChangeProfitAddress,
		umi.TxChangeFeeAddress:    confirmer.processChangeFeeAddress,
		umi.TxActivateTransit:     confirmer.processActivateTransit,
		umi.TxDeactivateTransit:   confirmer.processDeactivateTransit,
		umi.TxBurn:                confirmer.processBurn,
		umi.TxIssue:               confirmer.processIssue,
		umi.TxMintNftWitness:      confirmer.processMintNftWitness,
	}

	for txIndex, txCount := 0, block.TransactionCount(); txIndex < txCount; txIndex++ {
		transaction := block.Transaction(txIndex)
		hash := transaction.Hash()

		if confirmer.ledger.HasTransaction(hash) {
			return ErrTxConfirmed
		}

		handler, ok := handlers[transaction.Type()]
		if !ok {
			return errUnsupportedTxType
		}

		if err := handler(transaction); err != nil {
			return err
		}

		confirmer.TransactionHeight++
		confirmer.txHashes = append(confirmer.txHashes, hash)
	}

	return nil
}

func (confirmer *Confirmer) Account(address umi.Address) (account *Account, ok bool) {
	if account, ok = confirmer.accounts[address]; ok {
		return account, true
	}

	confirmer.ledger.RLock()
	defer confirmer.ledger.RUnlock()

	if account, ok = confirmer.ledger.Account(address); ok {
		// ??????????????????.
		c := *account
		account = &c

		// ?????????????????? ?? ??????????????.
		confirmer.accounts[address] = account

		return account, true
	}

	structure, ok := confirmer.Structure(address.Prefix())
	if !ok {
		return nil, false
	}

	account = NewAccount(structure)

	confirmer.accounts[address] = account

	return account, true
}

func (confirmer *Confirmer) Structure(prefix umi.Prefix) (structure *Structure, ok bool) {
	if structure, ok = confirmer.structures[prefix]; ok {
		return structure, true
	}

	confirmer.ledger.RLock()
	defer confirmer.ledger.RUnlock()

	if structure, ok = confirmer.ledger.Structure(prefix); ok {
		// ??????????????????.
		c := *structure
		structure = &c

		// ?????????????????? ?? ??????????????.
		confirmer.structures[prefix] = structure

		return structure, true
	}

	return nil, false
}

func (confirmer *Confirmer) verifyBlock(blockRaw []byte) error {
	block := (umi.Block)(blockRaw)

	if confirmer.BlockHash != block.PreviousBlockHash() {
		return fmt.Errorf("%w: ???????? ???? ?????????????????? ???? ????????????????????", errBlock)
	}

	if confirmer.BlockTimestamp > block.Timestamp() {
		return fmt.Errorf("%w: ?????????? ?????????????? ???????????? ?????????? ???????????? ?????? ?? ??????????????????????", errBlock)
	}

	if confirmer.BlockHeight == 0 && block.Version() != 0 {
		return fmt.Errorf("%w: ?????????? ???????????? ???????????? ?????????? ???????? ???????????? GENESIS", errBlock)
	}

	if confirmer.BlockHeight > 0 && block.Version() == 0 {
		return fmt.Errorf("%w: GENESIS ???????? ?????????? ???????? ???????????? ?????????? ????????????", errBlock)
	}

	return nil
}

func (confirmer *Confirmer) processGenesis(transaction umi.Transaction) error {
	recipient := transaction.Recipient()
	amount := transaction.Amount()

	// ?????????????????????? ???????????? ???????????????????? ?? ?????????????????????? ?????????????? ?????? ????????????????????.
	// ?? GENESIS-???????????????????? ?? ???????????????? ???????????????????? ?????????? ???????? ???????????? UMI-??????????,
	// ?????????????? ?????? ???? ?????????? ???????? ????????????, ???? ???? ???????????? ???????????? ?????????????????? ???????????? ??????????????.
	return confirmer.increaseAccountBalance(recipient, amount)
}

func (confirmer *Confirmer) processSend(transaction umi.Transaction) error {
	sender := transaction.Sender()
	recipient := transaction.Recipient()
	amount := transaction.Amount()

	// ?????????????????? ???????????? ?????????????????????? ?? ?????????????????????? ?????? ?????????????? ????????????????????.
	// ???????????????????? ???????????? ?? ????????????, ???????? ?????????????? ???? ???????????????????? ?????? ???????????? ???????????? ?????? ?????????? ????????????????????.
	if err := confirmer.decreaseAccountBalance(sender, amount); err != nil {
		return err
	}

	// ?????????????????? ?????????? ???? ?????????????????? ????????????????.
	// ???????????????? ?????????????? ???? ?????????? ?????????????????? ?????????????????????? ?? ????????????????????.
	if feeAmount, feeAddress, ok := confirmer.calculateFee(sender, recipient, amount); ok {
		// ?????????????????????? ???????????? ???????????????????? ???????????????? ?? ?????????????????????? ?????? ?????????????? ????????????????????.
		// ?????? ?????????????? ???????????????? ???? ?????????????????? ?????????????????????????? ??????????????????, ?????????????? ?????????? ???? ????????????
		// ???????? ????????????, ???? ?????????????????? ?????? ?????????? ?????????????? ?????? ??????????????.
		if err := confirmer.increaseAccountBalance(feeAddress, feeAmount); err != nil {
			return err
		}

		// ?????????????????? ??????????, ?????????????? ???????????????? ???? ???????????? ???????????????????? ???? ???????????? ????????????????.
		// ?? ????????????, ???????????????? ???????????? ???????????? ???????? ???????????? ????????, ???? ????-???? ???????????? ?? ??????????????????
		// ???????????????????? ?? ???????????? ?????????????? ???????? ?? ???????????????? ???????? ?????????????????? ???????????????????? ?? ??????????????
		// ????????????, ????-???? ???????? ???????????????? ???????? ???????? ?????????? ????????.
		amount -= feeAmount
	}

	// ?????????????????????? ???????????? ???????????????????? ???? ?????????? ?????????????????????? ???? ?????????????? ????????????????
	// ?? ?????????????????????? ?????? ?????????????? ????????????????????.
	// ???????????? ???????????????????????? ?? ???????????? ???????? ??????????????????, ?? ?????????????? ?????????????????????? ?????????? ????????????????????
	// ???? ????????????????????.
	return confirmer.increaseAccountBalance(recipient, amount)
}

func (confirmer *Confirmer) processCreateStructure(transaction umi.Transaction) error {
	prefix := transaction.Prefix()
	sender := transaction.Sender()
	amount := transaction.Amount()

	// ??????????????????, ?????? ?????????????????? ?? ?????????? ?????????????????? ???? ????????????????????.
	if _, ok := confirmer.Structure(prefix); ok {
		return fmt.Errorf("%s: %w", prefix.String(), errStructureExists)
	}

	// ?????????????????? ???????????? ?????????????????????? ?? ?????????????????????? ?????? ?????????????? ????????????????????.
	if err := confirmer.decreaseAccountBalance(sender, amount); err != nil {
		return err
	}

	structure := NewStructure(confirmer.ledger.config.Network, prefix, sender)
	structure.CreatedAt = confirmer.BlockTimestamp
	structure.Description = transaction.Description()
	structure.ProfitPercent = transaction.ProfitPercent()
	structure.FeePercent = transaction.FeePercent()

	confirmer.structures[prefix] = structure

	profitAccount, _ := confirmer.Account(structure.ProfitAddress)
	profitAccount.SetType(umi.Profit)

	devAccount, _ := confirmer.Account(structure.DevAddress)
	devAccount.SetType(umi.Dev)

	return nil
}

func (confirmer *Confirmer) processUpdateStructure(transaction umi.Transaction) error {
	timestamp := confirmer.BlockTimestamp

	prefix := transaction.Prefix()
	sender := transaction.Sender()
	amount := transaction.Amount()

	// ??????????????????, ?????? ?????????????????? ????????????????????.
	structure, ok := confirmer.Structure(prefix)
	if !ok {
		return fmt.Errorf("%s: %w", prefix.String(), errStructureDoesNotExist)
	}

	// ??????????????????, ?????? ?????????????????????? ???????????????????? ???????????????? ???????????????????? ???????? ??????????????????.
	if !structure.IsOwner(sender) {
		return fmt.Errorf("%s %s: %w", sender.String(), prefix.String(), errInsufficientPrivileges)
	}

	// ?????????????????? ?????????????????? ?????????? 0 ??????????, ???? ?????? ?????????? ???????????????? ????????????
	// ?????????????????????? ?? ?????????????????? ?????? ?????????????? ????????????????????.
	if err := confirmer.decreaseAccountBalance(sender, amount); err != nil {
		return err
	}

	// ?????????????????????????? ???????????? ??????????????????.
	structure.updateBalance(timestamp)

	// ?????????????????? ??????????????????.
	structure.Description = transaction.Description()
	structure.ProfitPercent = transaction.ProfitPercent()
	structure.FeePercent = transaction.FeePercent()

	return nil
}

func (confirmer *Confirmer) processChangeProfitAddress(transaction umi.Transaction) error {
	prefix := transaction.Prefix()
	sender := transaction.Sender()
	recipient := transaction.Recipient()
	amount := transaction.Amount()

	// ??????????????????, ?????? ?????????????????? ????????????????????.
	structure, ok := confirmer.Structure(prefix)
	if !ok {
		return fmt.Errorf("%s: %w", prefix.String(), errStructureDoesNotExist)
	}

	// ??????????????????, ?????? ?????????????????????? ???????????????????? ???????????????? ???????????????????? ??????????????????.
	if !structure.IsOwner(sender) {
		return fmt.Errorf("%s %s: %w", sender.String(), prefix.String(), errInsufficientPrivileges)
	}

	// ?????????? profit-???????????? ?????????? 0 ??????????, ???? ?????? ?????????? ???????????????? ???????????? ??????????????????????
	// ?? ?????????????????? ?????????????? ????????????????????.
	if err := confirmer.decreaseAccountBalance(sender, amount); err != nil {
		return err
	}

	// ?????? ?????????? ???????????? ???????????? ???????????????????? ???????????????? ???????????????? ?? ?????????????? ????????????????????
	// ???????????? ???????????? ??????????????????????????.
	return confirmer.changeProfitAddress(structure, recipient)
}

func (confirmer *Confirmer) processChangeFeeAddress(transaction umi.Transaction) error {
	prefix := transaction.Prefix()
	sender := transaction.Sender()
	recipient := transaction.Recipient()
	amount := transaction.Amount()

	structure, ok := confirmer.Structure(prefix)
	if !ok {
		return fmt.Errorf("%s: %w", transaction.Prefix(), errStructureDoesNotExist)
	}

	if !structure.IsOwner(sender) {
		return fmt.Errorf("%s %s: %w", transaction.Sender().String(), transaction.Prefix(), errInsufficientPrivileges)
	}

	if err := confirmer.decreaseAccountBalance(sender, amount); err != nil {
		return err
	}

	return confirmer.changeFeeAddress(structure, recipient)
}

func (confirmer *Confirmer) processActivateTransit(transaction umi.Transaction) error {
	prefix := transaction.Prefix()
	sender := transaction.Sender()
	recipient := transaction.Recipient()
	amount := transaction.Amount()

	// ??????????????????, ?????? ??????????????????, ?? ?????????????? ?????????????????????? ???????????????????????? ??????????, ????????????????????.
	structure, ok := confirmer.Structure(prefix)
	if !ok {
		return fmt.Errorf("%s: %w", transaction.Prefix(), errStructureDoesNotExist)
	}

	// ??????????????????, ?????? ?????????????????????? ???????????????????? ???????????????? ???????????????????? ??????????????????.
	if !structure.IsOwner(sender) {
		return fmt.Errorf("%s %s: %w", sender.String(), prefix.String(), errInsufficientPrivileges)
	}

	// ?????????????????? ???????????? ?????????????????????? ?? ?????????????????????? ?????????????? ????????????????????.
	if err := confirmer.decreaseAccountBalance(sender, amount); err != nil {
		return err
	}

	err := confirmer.activateTransitAddress(recipient)

	return err
}

func (confirmer *Confirmer) processDeactivateTransit(transaction umi.Transaction) error {
	prefix := transaction.Prefix()
	sender := transaction.Sender()
	recipient := transaction.Recipient()
	amount := transaction.Amount()

	structure, ok := confirmer.Structure(prefix)
	if !ok {
		return fmt.Errorf("%s: %w", prefix.String(), errStructureDoesNotExist)
	}

	if !structure.IsOwner(sender) {
		return fmt.Errorf("%s %s: %w", sender.String(), prefix.String(), errInsufficientPrivileges)
	}

	if err := confirmer.decreaseAccountBalance(sender, amount); err != nil {
		return err
	}

	err := confirmer.deactivateTransitAddress(recipient)

	return err
}

func (confirmer *Confirmer) processBurn(transaction umi.Transaction) error {
	sender := transaction.Sender()
	amount := transaction.Amount()

	// ?????????????????? ???????????? ?????????????????????? ?? ?????????????????????? ?????? ?????????????? ????????????????????.
	// ???????????????????? ???????????? ?? ????????????, ???????? ?????????????? ???? ???????????????????? ?????? ???????????? ???????????? ?????? ?????????? ????????????????????.
	return confirmer.decreaseAccountBalance(sender, amount)
}

func (confirmer *Confirmer) processIssue(transaction umi.Transaction) error {
	prefix := transaction.Prefix()
	sender := transaction.Sender()
	recipient := transaction.Recipient()
	amount := transaction.Amount()

	structure, ok := confirmer.Structure(prefix)
	if !ok {
		return fmt.Errorf("%s: %w", prefix.String(), errStructureDoesNotExist)
	}

	if !structure.IsOwner(sender) {
		return fmt.Errorf("%s %s: %w", sender.String(), prefix.String(), errInsufficientPrivileges)
	}

	return confirmer.increaseAccountBalance(recipient, amount)
}

func (confirmer *Confirmer) processMintNftWitness(transaction umi.Transaction) error {
	sender := transaction.Sender()
	amount := transaction.Amount()

	confirmer.nfts[transaction.Hash()] = transaction.Sender()

	// ?????????????????? ???????????? ?????????????????????? ?? ?????????????????????? ?????? ?????????????? ????????????????????.
	// ???????????????????? ???????????? ?? ????????????, ???????? ?????????????? ???? ???????????????????? ?????? ???????????? ???????????? ?????? ?????????? ????????????????????.
	return confirmer.decreaseAccountBalance(sender, amount)
}

// increaseAccountBalance ?????????????????????? ???????????? ??????????, ???????????????????? ?? ?????????????? ???? ?????????????????? ??????????
// ?? ?????????????????????? ?????????????? ????????????????????. ?????????? ?????????????????? ???????????? ???????????????? ?????????????????? ???? ??????????????????
// ?????????? ?????????????????????????????? ??????????.
func (confirmer *Confirmer) increaseAccountBalance(address umi.Address, amount uint64) error {
	timestamp := confirmer.BlockTimestamp

	account, ok := confirmer.Account(address)
	if !ok {
		return fmt.Errorf("account %s: %w", address.String(), errNotFound)
	}

	account.TransactionCount++
	account.IncreaseBalance(amount, timestamp)

	switch account.Type {
	case umi.Umi:
		structure, _ := confirmer.Structure(address.Prefix())
		structure.IncreaseBalance(amount, timestamp)

	case umi.Deposit, umi.Transit:
		structure, _ := confirmer.Structure(address.Prefix())
		structure.IncreaseBalance(amount, timestamp)

		profitAccount, _ := confirmer.Account(structure.ProfitAddress)
		profitAccount.IncreaseBalance(amount, timestamp)

		devAccount, _ := confirmer.Account(structure.DevAddress)
		devAccount.IncreaseBalance(amount, timestamp)

	case umi.Profit:
		structure, _ := confirmer.Structure(address.Prefix())
		devAccount, _ := confirmer.Account(structure.DevAddress)
		devAccount.IncreaseBalance(amount, timestamp)
	}

	return nil
}

func (confirmer *Confirmer) decreaseAccountBalance(address umi.Address, amount uint64) error {
	timestamp := confirmer.BlockTimestamp

	account, ok := confirmer.Account(address)
	if !ok {
		return fmt.Errorf("account %s: %w", address.String(), errNotFound)
	}

	account.TransactionCount++

	availableBalance := confirmer.AvailableBalance(address, account)

	if amount > availableBalance {
		return fmt.Errorf("%s: %w", address.String(), errInsufficientFunds)
	}

	if !account.DecreaseBalance(amount, timestamp) {
		return fmt.Errorf("%s: %w", address.String(), errInsufficientFunds)
	}

	switch account.Type {
	case umi.Umi:
		structure, _ := confirmer.Structure(address.Prefix())
		structure.DecreaseBalance(amount, timestamp)

	case umi.Deposit, umi.Transit:
		structure, _ := confirmer.Structure(address.Prefix())
		structure.DecreaseBalance(amount, timestamp)

		profitAccount, _ := confirmer.Account(structure.ProfitAddress)
		profitAccount.DecreaseBalance(amount, timestamp)

		devAccount, _ := confirmer.Account(structure.DevAddress)
		devAccount.DecreaseBalance(amount, timestamp)

	case umi.Profit:
		structure, _ := confirmer.Structure(address.Prefix())
		devAccount, _ := confirmer.Account(structure.DevAddress)
		devAccount.DecreaseBalance(amount, timestamp)
	}

	return nil
}

func (confirmer *Confirmer) AvailableBalance(address umi.Address, account *Account) uint64 {
	switch account.Type {
	case umi.Profit:
		structure, _ := confirmer.Structure(address.Prefix())

		return account.BalanceAt(confirmer.BlockTimestamp) - structure.BalanceAt(confirmer.BlockTimestamp)

	case umi.Dev:
		structure, _ := confirmer.Structure(address.Prefix())
		profitAccount, _ := confirmer.Account(structure.ProfitAddress)

		return account.BalanceAt(confirmer.BlockTimestamp) - profitAccount.BalanceAt(confirmer.BlockTimestamp)
	}

	return account.BalanceAt(confirmer.BlockTimestamp)
}

func (confirmer *Confirmer) changeProfitAddress(structure *Structure, newProfitAddress umi.Address) error {
	timestamp := confirmer.BlockTimestamp
	devAccount, _ := confirmer.Account(structure.DevAddress)

	// Profit ?? Fee ???? ?????????????????? ?? ?????????????? ??????????????????, ???? ???????????? ?????????????????? ???????????????????? ?? ???????????? Profit.
	structureBalance := structure.BalanceAt(timestamp)

	// ???????? ???????????? Profit ?? Fee ????????????????????, ???????????? ???????????? ?????????? ?????????? ?????? Profit.
	// ?????????? ?????????????????????? ????????????, ???????????????? ?????? ?????????????? ???????????? ???? Fee, ?? ???? ?????????? ???????????????? ?????????????????? Profit,
	// ???????????????? ???????????????????? ???????????? ?? ???????????????? ?????????????????? ??????????????????.
	if structure.ProfitAddress == structure.FeeAddress {
		feeAccount, _ := confirmer.Account(structure.FeeAddress)
		feeAccount.SetType(umi.Fee)
		feeAccount.SetInterestRate(structure.InterestRate(umi.Fee), timestamp)
		// ??.??. Dev ???????????????? ?? ???????? ???????????? Profit, ???? ???????????? ?????????????? ???????????? ???????????? Profit ???? Dev.
		devAccount.DecreaseBalance(feeAccount.Balance, timestamp)
		// ??.??. Profit ???????????????? ?? ???????? ???????????? ??????????????????, ???? ???????????? ?????????????? ???? ???????????? Fee ???????? ????????????.
		feeAccount.DecreaseBalance(structureBalance, timestamp)

		newProfitAccount, _ := confirmer.Account(newProfitAddress)
		newProfitAccount.SetType(umi.Profit)
		newProfitAccount.SetInterestRate(structure.InterestRate(umi.Profit), timestamp)
		// ??.??. ???????????? Profit ???? ???????????????????? ?? ???????????? ??????????????????, ?????????? ?????? ??????????????
		structure.DecreaseBalance(newProfitAccount.Balance, timestamp)
		// ?????????????????? ???????????? ?????????????????? ?? ???????????? Profit
		newProfitAccount.IncreaseBalance(structure.Balance, timestamp)
		// ?????????????????? ?????????? ???????????? Profit ?? ?????????????? Dev
		devAccount.IncreaseBalance(newProfitAccount.Balance, timestamp)

		structure.ProfitAddress = newProfitAddress

		newProfitAccount.TransactionCount++

		return nil
	}

	oldProfitAccount, _ := confirmer.Account(structure.ProfitAddress)
	oldProfitAccount.SetType(umi.Deposit)
	oldProfitAccount.SetInterestRate(structure.InterestRate(umi.Deposit), timestamp)
	// ??.??. Dev ???????????????? ?? ???????? ???????????? Profit, ???? ?????????? ?????? ??????????????
	devAccount.DecreaseBalance(oldProfitAccount.Balance, timestamp)
	// ???????????? Profit ???????????????? ?? ???????? ???????????? ??????????????????, ?????????? ??????????????
	oldProfitAccount.DecreaseBalance(structureBalance, timestamp)
	// ??.??. ?????? ???????????? ???????????? Deposit - ?????? ???????????? ???????????? ???????????????????? ?? ???????????? ??????????????????
	structure.IncreaseBalance(oldProfitAccount.Balance, timestamp)

	newProfitAccount, _ := confirmer.Account(newProfitAddress)
	newProfitAccount.SetType(umi.Profit)
	newProfitAccount.SetInterestRate(structure.InterestRate(umi.Profit), timestamp)
	// ???????????????? ???? ?????????????? ?????????????????? ???????????? ???????????? ????????????
	structure.DecreaseBalance(newProfitAccount.Balance, timestamp)
	// ?????????????????? ?? ?????????????? ???????????? Profit ???????????? ??????????????????
	newProfitAccount.IncreaseBalance(structure.Balance, timestamp)
	// ?????????????????? ?? Dev ?????????????? ???????????? ???????????? Profit
	devAccount.IncreaseBalance(newProfitAccount.Balance, timestamp)

	structure.ProfitAddress = newProfitAddress

	newProfitAccount.TransactionCount++

	return nil
}

func (confirmer *Confirmer) changeFeeAddress(structure *Structure, newFeeAddress umi.Address) error {
	timestamp := confirmer.BlockTimestamp
	devAccount, _ := confirmer.Account(structure.DevAddress)
	profitAccount, _ := confirmer.Account(structure.ProfitAddress)

	// ???????? ???????????? Fee ?? Profit ????????????????????, ???????????? ???????????? ?????????? ?????????? ?????? Profit.
	// ?????????? ???????????????????? ???? ?????????? ???????????????? ?????? Fee, ???????????????? ???????????????????? ????????????, ???????????????? ?????????????????? ??????????????????
	if structure.FeeAddress == structure.ProfitAddress {
		newFeeAccount, _ := confirmer.Account(newFeeAddress)
		newFeeAccount.SetType(umi.Fee)
		newFeeAccount.SetInterestRate(structure.InterestRate(umi.Fee), timestamp)

		structure.FeeAddress = newFeeAddress

		// ??.??. ?????????????? Dev, Profit ?? Fee ???? ?????????????????????? ?? ?????????????? ??????????????????,
		// ?????????? ?????????????? ???????????? ???????????? ???????????? ???? ?????????????? ??????????????????, ?????????????? Profit ?? Dev
		structure.DecreaseBalance(newFeeAccount.Balance, timestamp)
		profitAccount.DecreaseBalance(newFeeAccount.Balance, timestamp)
		devAccount.DecreaseBalance(newFeeAccount.Balance, timestamp)

		newFeeAccount.TransactionCount++

		return nil
	}

	// ?????????????????? ???????????? ?????????????? ????????????, ???????????? ?????? ?????? ?? ???????????????????? ????????????
	oldFeetAccount, _ := confirmer.Account(structure.FeeAddress)
	oldFeetAccount.SetType(umi.Deposit)
	oldFeetAccount.SetInterestRate(structure.InterestRate(umi.Deposit), timestamp)

	// ??.??. ?????? ???????????????? ???? Fee, ???? ???????????? ?????????????????? ???????? ???????????? ?? ?????????????? ??????????????????, Dev ?? Profit
	structure.IncreaseBalance(oldFeetAccount.Balance, timestamp)
	profitAccount.IncreaseBalance(oldFeetAccount.Balance, timestamp)
	devAccount.IncreaseBalance(oldFeetAccount.Balance, timestamp)

	// ?????????????????? ???????????? ???????????? ????????????, ???????????? ?????? ?????? ?? ???????????????????? ????????????
	newFeeAccount, _ := confirmer.Account(newFeeAddress)
	newFeeAccount.SetType(umi.Fee)
	newFeeAccount.SetInterestRate(structure.InterestRate(umi.Fee), timestamp)

	// ??.??. ???????????? Fee ???? ?????????????????????? ?? ???????????????? ??????????????????, Profit ?? Dev - ???????????????? ??????.
	structure.DecreaseBalance(newFeeAccount.Balance, timestamp)
	profitAccount.DecreaseBalance(newFeeAccount.Balance, timestamp)
	devAccount.DecreaseBalance(newFeeAccount.Balance, timestamp)

	structure.FeeAddress = newFeeAddress

	newFeeAccount.TransactionCount++

	return nil
}

func (confirmer *Confirmer) activateTransitAddress(address umi.Address) error {
	timestamp := confirmer.BlockTimestamp
	transitAccount, _ := confirmer.Account(address)

	if transitAccount.Type != umi.Deposit {
		return errAccountType
	}

	transitAccount.SetType(umi.Transit)
	transitAccount.UpdateBalance(timestamp)

	transitAccount.TransactionCount++

	return nil
}

func (confirmer *Confirmer) deactivateTransitAddress(address umi.Address) error {
	timestamp := confirmer.BlockTimestamp
	transitAccount, _ := confirmer.Account(address)

	if transitAccount.Type != umi.Transit {
		return errAccountType
	}

	transitAccount.SetType(umi.Deposit)
	transitAccount.UpdateBalance(timestamp)

	transitAccount.TransactionCount++

	return nil
}

func (confirmer *Confirmer) calculateFee(
	sender, recipient umi.Address, amount uint64) (feeAmount uint64, feeAddress umi.Address, ok bool) {
	senderAccount, _ := confirmer.Account(sender)

	// ?????? ?????????????????? ???????????? ?????????????????? ?????????????????????????????????? ???????????? ???? ???????????? ????????????????.
	switch senderAccount.Type {
	case umi.Fee, umi.Profit, umi.Transit, umi.Dev:
		return feeAmount, feeAddress, false
	}

	recipientAccount, ok := confirmer.Account(recipient)
	if !ok {
		return feeAmount, feeAddress, false
	}

	// ?????? ?????????????????? ???????? ?????????????? ???????????? ???????????? ?????? ?????????????????? ???? ????????????????.
	if recipientAccount.Type != umi.Deposit {
		return feeAmount, feeAddress, false
	}

	structure, ok := confirmer.Structure(recipient.Prefix())
	if !ok {
		return feeAmount, feeAddress, false
	}

	if structure.FeePercent == 0 {
		return feeAmount, feeAddress, false
	}

	feeAmount = uint64(math.Ceil(float64(amount) * float64(structure.FeePercent) / float64(100_00)))
	feeAddress = structure.FeeAddress

	return feeAmount, feeAddress, true
}

func (confirmer *Confirmer) Commit() error {
	confirmer.ledger.Lock()
	defer confirmer.ledger.Unlock()

	if confirmer.ledger.LastBlockHash != confirmer.PrevBlockHash {
		return fmt.Errorf("%w: !! ?????????? ???????? ???? ?????????????????? ???? ?????????????????? ????????", errBlock)
	}

	// ?????????????????? ?????????????????? ?? ??????????????????
	for address, account := range confirmer.accounts {
		prefix := address.Prefix()

		accounts, ok := confirmer.ledger.accounts[prefix]
		if !ok {
			accounts = make(map[umi.Address]*Account)
			confirmer.ledger.accounts[prefix] = accounts
		}

		accounts[address] = account
	}

	// ?????????????????? ?????????????????? ?? ????????????????????
	for prefix, structure := range confirmer.structures {
		old, ok := confirmer.ledger.structures[prefix]

		structure.AddressCount = len(confirmer.ledger.accounts[prefix])

		confirmer.ledger.structures[prefix] = structure

		if ok && ((old.ProfitPercent != structure.ProfitPercent) || (old.FeePercent != structure.FeePercent)) {
			confirmer.updateLevelAddresses(prefix)
		}
	}

	// ?????????????????? ???????????????? ???????? ???? NFT
	for hash, addr := range confirmer.nfts {
		confirmer.ledger.nfts[hash] = addr
	}

	// ?????????????????? ???????? ????????????????????
	for _, hash := range confirmer.txHashes {
		confirmer.ledger.transactions[hash] = struct{}{}
	}

	confirmer.ledger.LastBlockTimestamp = confirmer.BlockTimestamp
	confirmer.ledger.LastBlockHeight = confirmer.BlockHeight
	confirmer.ledger.LastBlockHash = confirmer.BlockHash
	confirmer.ledger.LastTransactionHeight = confirmer.TransactionHeight

	confirmer.checkStaking()

	return nil
}

func (confirmer *Confirmer) checkStaking() {
	switch {
	case confirmer.ledger.LastBlockHeight < 5_393_000:
		confirmer.checkStructureLevel()

	// Stop staking.
	case confirmer.ledger.LastBlockHeight == 5_393_000:
		confirmer.stopStaking()
	}

	confirmer.checkGlize()
}
