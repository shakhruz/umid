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
	txHashes   []umi.Hash

	// Абсолютная высота транзакции в блокчейне. Удобно использовать для синхронизации.
	TransactionHeight uint64
	// Время создания блока. По этой временной метке считаются балансы.
	BlockTimestamp uint32
	// Высота блока в блокчейне.
	BlockHeight uint32
	// Хэш обрабатываемого блока. После коммита это будет LastBlockHash.
	BlockHash umi.Hash
	// Хэш последнего обработанного леджером блока. На случай если во время обаботки данные поменяются.
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
		// Клонируем.
		c := *account
		account = &c

		// Добавляем в снэпшот.
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
		// Клонируем.
		c := *structure
		structure = &c

		// Добавляем в снэпшот.
		confirmer.structures[prefix] = structure

		return structure, true
	}

	return nil, false
}

func (confirmer *Confirmer) verifyBlock(blockRaw []byte) error {
	block := (umi.Block)(blockRaw)

	if confirmer.BlockHash != block.PreviousBlockHash() {
		return fmt.Errorf("%w: блок не ссылается на предыдущий", errBlock)
	}

	if confirmer.BlockTimestamp > block.Timestamp() {
		return fmt.Errorf("%w: метка времени нового блока меньше чем у предыдущего", errBlock)
	}

	if confirmer.BlockHeight == 0 && block.Version() != 0 {
		return fmt.Errorf("%w: самым первым блоком может быть только GENESIS", errBlock)
	}

	if confirmer.BlockHeight > 0 && block.Version() == 0 {
		return fmt.Errorf("%w: GENESIS блок может быть только самым первым", errBlock)
	}

	return nil
}

func (confirmer *Confirmer) processGenesis(transaction umi.Transaction) error {
	recipient := transaction.Recipient()
	amount := transaction.Amount()

	// Увеличиваем баланс получателя и увеличиваем счетчик его транзакций.
	// В GENESIS-транзакции в качестве получателя может быть только UMI-адрес,
	// поэтому тут не может быть ошибки, но на всякий случай обработку ошибок оставим.
	return confirmer.increaseAccountBalance(recipient, amount)
}

func (confirmer *Confirmer) processSend(transaction umi.Transaction) error {
	sender := transaction.Sender()
	recipient := transaction.Recipient()
	amount := transaction.Amount()

	// Уменьшаем баланс отправителя и увеличиваем его счетчик транзакций.
	// Возвращаем ошибку в случае, если аккаунт не существует или баланс меньше чем сумма транзакции.
	if err := confirmer.decreaseAccountBalance(sender, amount); err != nil {
		return err
	}

	// Проверяем нужно ли списывать комиссию.
	// Комиссия зависит от типов аккаунтов отправителя и получателя.
	if feeAmount, feeAddress, ok := confirmer.calculateFee(sender, recipient, amount); ok {
		// Увеличиваем баланс получателя комиссий и увеличиваем его счетчик транзакций.
		// При расчете комиссии мы проверяем существование структуры, поэтому здесь не должно
		// быть ошибок, но обработку все равно оставим для порядка.
		if err := confirmer.increaseAccountBalance(feeAddress, feeAmount); err != nil {
			return err
		}

		// Уменьшаем сумму, которую зачислим на баланс получателя на размер комиссии.
		// В теории, комиссия всегда должна быть больше нуля, но из-за ошибки в обработке
		// транзакций в ранних версиях ноды в блокчейн были добавлены транзакции с нулевой
		// суммой, из-за чего комиссия тоже была равна нулю.
		amount -= feeAmount
	}

	// Увеличиваем баланс получателя на сумму отправления за вычетом комиссии
	// и увеличиваем его счетчик транзакций.
	// Ошибка возвращается в случае если структура, к которой принадлежит адрес получателя
	// не существует.
	return confirmer.increaseAccountBalance(recipient, amount)
}

func (confirmer *Confirmer) processCreateStructure(transaction umi.Transaction) error {
	prefix := transaction.Prefix()
	sender := transaction.Sender()
	amount := transaction.Amount()

	// Проверяем, что структура с таким префиксом НЕ существует.
	if _, ok := confirmer.Structure(prefix); ok {
		return fmt.Errorf("%s: %w", prefix.String(), errStructureExists)
	}

	// Уменьшаем баланс отправителя и увеличиваем его счетчик транзакций.
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

	// Проверяем, что структура существует.
	structure, ok := confirmer.Structure(prefix)
	if !ok {
		return fmt.Errorf("%s: %w", prefix.String(), errStructureDoesNotExist)
	}

	// Проверяем, что отправитель транзакции является владельцем этой структуры.
	if !structure.IsOwner(sender) {
		return fmt.Errorf("%s %s: %w", sender.String(), prefix.String(), errInsufficientPrivileges)
	}

	// Изменение структуры стоит 0 монет, но нам нужно обновить баланс
	// отправителя и увеличить его счетчик транзакций.
	if err := confirmer.decreaseAccountBalance(sender, amount); err != nil {
		return err
	}

	// Пересчитываем баланс структуры.
	structure.updateBalance(timestamp)

	// Обновляем настройки.
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

	// Проверяем, что структуры существует.
	structure, ok := confirmer.Structure(prefix)
	if !ok {
		return fmt.Errorf("%s: %w", prefix.String(), errStructureDoesNotExist)
	}

	// Проверяем, что отправитель транзакции является владельцем структуры.
	if !structure.IsOwner(sender) {
		return fmt.Errorf("%s %s: %w", sender.String(), prefix.String(), errInsufficientPrivileges)
	}

	// Смена profit-адреса стоит 0 монет, но нам нужно обновить баланс отправителя
	// и увеличить счетчик транзакций.
	if err := confirmer.decreaseAccountBalance(sender, amount); err != nil {
		return err
	}

	// При смене профит адреса происходит пересчет балансов и счетчик транзакций
	// нового адреса увеличивается.
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

	// Проверяем, что структура, к которой принадлежит активируемый адрес, существует.
	structure, ok := confirmer.Structure(prefix)
	if !ok {
		return fmt.Errorf("%s: %w", transaction.Prefix(), errStructureDoesNotExist)
	}

	// Проверяем, что отправитель транзакции является владельцем структуры.
	if !structure.IsOwner(sender) {
		return fmt.Errorf("%s %s: %w", sender.String(), prefix.String(), errInsufficientPrivileges)
	}

	// Уменьшаем баланс отправителя и увеличиваем счетчик транзакций.
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

	// Уменьшаем баланс отправителя и увеличиваем его счетчик транзакций.
	// Возвращаем ошибку в случае, если аккаунт не существует или баланс меньше чем сумма транзакции.
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

// increaseAccountBalance увеличивает баланс счета, связанного с адресом на указанную сумму
// и увеличивает счетчик транзакций. Перед операцией баланс аккаунта считается по временной
// метке обрабатываемого блока.
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

	// Profit и Fee не участвуют в балансе структуры, но баланс структуры включается в баланс Profit.
	structureBalance := structure.BalanceAt(timestamp)

	// Если адреса Profit и Fee одинаковые, значит старый адрес имеет тип Profit.
	// Нужно пересчитать баланс, поменять тип старого адреса на Fee, а на новом аккаунте поставить Profit,
	// изменить процентную ставку и обновить настройки структуры.
	if structure.ProfitAddress == structure.FeeAddress {
		feeAccount, _ := confirmer.Account(structure.FeeAddress)
		feeAccount.SetType(umi.Fee)
		feeAccount.SetInterestRate(structure.InterestRate(umi.Fee), timestamp)
		// Т.к. Dev включает в себя баланс Profit, мы должны вычесть старый баланс Profit из Dev.
		devAccount.DecreaseBalance(feeAccount.Balance, timestamp)
		// Т.к. Profit включает в себя баланс структуры, мы должны вычесть из нового Fee этот баланс.
		feeAccount.DecreaseBalance(structureBalance, timestamp)

		newProfitAccount, _ := confirmer.Account(newProfitAddress)
		newProfitAccount.SetType(umi.Profit)
		newProfitAccount.SetInterestRate(structure.InterestRate(umi.Profit), timestamp)
		// Т.к. баланс Profit не включается в баланс структуры, нужно его вычесть
		structure.DecreaseBalance(newProfitAccount.Balance, timestamp)
		// Добавляем баланс структуры к новому Profit
		newProfitAccount.IncreaseBalance(structure.Balance, timestamp)
		// Добавляем новый баланс Profit к балансу Dev
		devAccount.IncreaseBalance(newProfitAccount.Balance, timestamp)

		structure.ProfitAddress = newProfitAddress

		newProfitAccount.TransactionCount++

		return nil
	}

	oldProfitAccount, _ := confirmer.Account(structure.ProfitAddress)
	oldProfitAccount.SetType(umi.Deposit)
	oldProfitAccount.SetInterestRate(structure.InterestRate(umi.Deposit), timestamp)
	// Т.к. Dev включает в себя баланс Profit, то нужно его вычесть
	devAccount.DecreaseBalance(oldProfitAccount.Balance, timestamp)
	// Баланс Profit включает в себя баланс структуры, нужно вычесть
	oldProfitAccount.DecreaseBalance(structureBalance, timestamp)
	// Т.к. тип адреса теперь Deposit - его баланс должен включаться в баланс структуры
	structure.IncreaseBalance(oldProfitAccount.Balance, timestamp)

	newProfitAccount, _ := confirmer.Account(newProfitAddress)
	newProfitAccount.SetType(umi.Profit)
	newProfitAccount.SetInterestRate(structure.InterestRate(umi.Profit), timestamp)
	// Вычитаем из баланса структуры баланс нового Профит
	structure.DecreaseBalance(newProfitAccount.Balance, timestamp)
	// Добавляем к балансу нового Profit баланс структуры
	newProfitAccount.IncreaseBalance(structure.Balance, timestamp)
	// Добавляем в Dev аккаунт баланс нового Profit
	devAccount.IncreaseBalance(newProfitAccount.Balance, timestamp)

	structure.ProfitAddress = newProfitAddress

	newProfitAccount.TransactionCount++

	return nil
}

func (confirmer *Confirmer) changeFeeAddress(structure *Structure, newFeeAddress umi.Address) error {
	timestamp := confirmer.BlockTimestamp
	devAccount, _ := confirmer.Account(structure.DevAddress)
	profitAccount, _ := confirmer.Account(structure.ProfitAddress)

	// Если адреса Fee и Profit одинаковые, значит старый адрес имеет тип Profit.
	// Нужно установить на новом аккаунте тип Fee, поменять процентную ставку, обновить настройки структуры
	if structure.FeeAddress == structure.ProfitAddress {
		newFeeAccount, _ := confirmer.Account(newFeeAddress)
		newFeeAccount.SetType(umi.Fee)
		newFeeAccount.SetInterestRate(structure.InterestRate(umi.Fee), timestamp)

		structure.FeeAddress = newFeeAddress

		// Т.к. балансы Dev, Profit и Fee не учитываются в балансе структуры,
		// нужно вычесть баланс нового адреса из баланса стркутуры, баланса Profit и Dev
		structure.DecreaseBalance(newFeeAccount.Balance, timestamp)
		profitAccount.DecreaseBalance(newFeeAccount.Balance, timestamp)
		devAccount.DecreaseBalance(newFeeAccount.Balance, timestamp)

		newFeeAccount.TransactionCount++

		return nil
	}

	// Фиксируем баланс старого адреса, меняем его тип и процентную ставку
	oldFeetAccount, _ := confirmer.Account(structure.FeeAddress)
	oldFeetAccount.SetType(umi.Deposit)
	oldFeetAccount.SetInterestRate(structure.InterestRate(umi.Deposit), timestamp)

	// Т.к. тип сменился на Fee, мы должны учитывать этот баланс в балансе структуры, Dev и Profit
	structure.IncreaseBalance(oldFeetAccount.Balance, timestamp)
	profitAccount.IncreaseBalance(oldFeetAccount.Balance, timestamp)
	devAccount.IncreaseBalance(oldFeetAccount.Balance, timestamp)

	// Фиксируем баланс нового адреса, меняем его тип и процентную ставку
	newFeeAccount, _ := confirmer.Account(newFeeAddress)
	newFeeAccount.SetType(umi.Fee)
	newFeeAccount.SetInterestRate(structure.InterestRate(umi.Fee), timestamp)

	// Т.к. баланс Fee не учитывается в балансах структуры, Profit и Dev - вычитаем его.
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

	// При переводах внутри структуры привилегированные адреса не платят комиссию.
	switch senderAccount.Type {
	case umi.Fee, umi.Profit, umi.Transit, umi.Dev:
		return feeAmount, feeAddress, false
	}

	recipientAccount, ok := confirmer.Account(recipient)
	if !ok {
		return feeAmount, feeAddress, false
	}

	// Все остальные типы адресов платят только при переводах на депозиты.
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
		return fmt.Errorf("%w: !! новый блок не ссылается на последний блок", errBlock)
	}

	// Фиксируем изменения в аккаунтах
	for address, account := range confirmer.accounts {
		prefix := address.Prefix()

		accounts, ok := confirmer.ledger.accounts[prefix]
		if !ok {
			accounts = make(map[umi.Address]*Account)
			confirmer.ledger.accounts[prefix] = accounts
		}

		accounts[address] = account
	}

	// Фиксируем изменения в структурах
	for prefix, structure := range confirmer.structures {
		old, ok := confirmer.ledger.structures[prefix]

		structure.AddressCount = len(confirmer.ledger.accounts[prefix])

		confirmer.ledger.structures[prefix] = structure

		if ok && ((old.ProfitPercent != structure.ProfitPercent) || (old.FeePercent != structure.FeePercent)) {
			confirmer.updateLevelAddresses(prefix)
		}
	}

	// Добавляем хеши транзакций
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
