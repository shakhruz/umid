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

package generator

import (
	"context"
	"crypto/ed25519"
	"encoding/base64"
	"fmt"
	"log"
	"os"
	"time"

	"gitlab.com/umitop/umid/pkg/ledger"
	"gitlab.com/umitop/umid/pkg/nft"
	"gitlab.com/umitop/umid/pkg/umi"
)

type iMempool interface {
	Mempool() (txs []*umi.Transaction)
}

type iNftMempool interface {
	Mempool() (txs [][]byte)
}

type Generator struct {
	confirmer  *ledger.ConfirmerLegacy
	mempool    iMempool
	nftMempool iNftMempool
	nftStorage *nft.Storage
}

func NewGenerator(confirmer *ledger.ConfirmerLegacy, mempool iMempool, nftMempool iNftMempool) *Generator {
	return &Generator{
		confirmer:  confirmer,
		mempool:    mempool,
		nftMempool: nftMempool,
	}
}

func (generator *Generator) SetNftStorage(nftStorage *nft.Storage) *Generator {
	generator.nftStorage = nftStorage

	return generator
}

func (generator *Generator) Worker(ctx context.Context) {
	ticker := time.NewTicker(time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			generator.generateBlock()
		case <-ctx.Done():
			return
		}
	}
}

func (generator *Generator) generateBlock() {
	timestamp := uint32(time.Now().Unix())
	transactions := generator.mempool.Mempool()
	generator.confirmer.ResetState()
	generator.confirmer.BlockTimestamp = timestamp

	block := umi.NewBlock()
	block.SetVersion(1)
	block.SetPreviousBlockHash(generator.confirmer.PrevBlockHash)
	block.SetTimestamp(timestamp)

	txCount := 0

	for _, transactionRaw := range transactions {
		if txCount == 65535 {
			break
		}

		transaction := make(umi.Transaction, umi.TxConfirmedLength)
		copy(transaction[:umi.TxLength], *transactionRaw)

		processors := map[string]func(umi.Transaction, uint32) (bool, error){
			umi.TxSend:                generator.processSend,
			umi.TxCreateStructure:     generator.processCreateStructure,
			umi.TxUpdateStructure:     generator.processUpdateStructure,
			umi.TxChangeProfitAddress: generator.processChangeProfitAddress,
			umi.TxChangeFeeAddress:    generator.processChangeFeeAddress,
			umi.TxActivateTransit:     generator.processActivateTransit,
			umi.TxDeactivateTransit:   generator.processDeactivateTransit,
			umi.TxBurn:                generator.processBurn,
			umi.TxIssue:               generator.processIssue,
		}

		processor, ok := processors[transaction.Type()]
		if !ok {
			continue
		}

		ok, err := processor(transaction, timestamp)
		if err != nil {
			return
		}

		if !ok {
			continue
		}

		block = append(block, *transactionRaw...)
		txCount++
	}

	// NFT
	nftTokens := make([][]byte, 0)

	if txCount < 65535 {
		nftTransactions := generator.nftMempool.Mempool()

		for _, nftTransactionRaw := range nftTransactions {
			if txCount == 65535 {
				break
			}

			transaction := make(nft.Transaction, len(nftTransactionRaw))
			copy(transaction[:], nftTransactionRaw)

			txWitness := make(umi.Transaction, umi.TxConfirmedLength)
			txWitness.SetVersion(umi.TxV18MintNftWitness)
			txWitness.SetSender(transaction.Sender())
			txWitness.SetHash(transaction.Hash())
			txWitness.SetAmount(uint64(len(transaction)))
			txWitness.SetTimestamp(transaction.Timestamp())
			txWitness.SetNonce(transaction.Nonce())
			copy(txWitness[86:150], ed25519.Sign(secKey(), txWitness[0:86]))

			ok, err := generator.processMintNftWitness(txWitness, timestamp)
			if err != nil {
				return
			}

			if !ok {
				log.Println("not OK")
				continue
			}

			block = append(block, txWitness...)
			txCount++

			nftTokens = append(nftTokens, transaction)
		}
	}

	if txCount == 0 {
		return
	}

	block.SetTransactionCount(txCount)
	block.SetMerkleRootHash(umi.MerkleRoot(block[umi.HdrLength:]))
	signBlock(block)

	if err := generator.confirmer.AppendBlockLegacy(block); err != nil {
		log.Printf("AppendBlockLegacy error: %v", err)
	} else {
		if len(nftTokens) > 0 {
			for _, data := range nftTokens {
				if err := generator.nftStorage.AppendData(data); err != nil {
					log.Printf("AppendData error: %v", err)
				}
			}
		}
	}
}

func signBlock(block umi.Block) {
	secKey := secKey()
	pubKey := secKey[ed25519.PublicKeySize:ed25519.PrivateKeySize]

	copy(block[71:103], pubKey)
	copy(block[103:167], ed25519.Sign(secKey, block[0:103]))
}

func secKey() ed25519.PrivateKey {
	secKey, _ := base64.StdEncoding.DecodeString(os.Getenv("UMI_MASTER_KEY"))

	return secKey
}

func (generator *Generator) processSend(transaction umi.Transaction, _ uint32) (bool, error) {
	sender := transaction.Sender()

	senderAccount, ok := generator.confirmer.Account(sender)
	if !ok {
		return false, nil
	}

	availableBalance := generator.confirmer.AvailableBalance(sender, senderAccount)
	if availableBalance < transaction.Amount() {
		return false, nil
	}

	recipient := transaction.Recipient()
	if _, ok := generator.confirmer.Account(recipient); !ok {
		log.Printf("???????????????????????? ???????????????????? %s", recipient.String())

		return false, nil
	}

	if _, err := generator.confirmer.ProcessSendLegacy(transaction); err != nil {
		log.Printf("????????????: %v", err)

		return false, fmt.Errorf("%w", err)
	}

	return true, nil
}

func (generator *Generator) processCreateStructure(transaction umi.Transaction, timestamp uint32) (bool, error) {
	senderAccount, ok := generator.confirmer.Account(transaction.Sender())
	if !ok {
		return false, nil
	}

	if senderAccount.BalanceAt(timestamp) < transaction.Amount() {
		return false, nil
	}

	prefix := transaction.Prefix()
	if _, ok := generator.confirmer.Structure(prefix); ok {
		log.Printf("?????????????????? '%s' ?????? ????????????????????", prefix.String())

		return false, nil
	}

	if _, err := generator.confirmer.ProcessCreateStructureLegacy(transaction); err != nil {
		log.Printf("????????????: %v", err)

		return false, fmt.Errorf("%w", err)
	}

	return true, nil
}

func (generator *Generator) processUpdateStructure(transaction umi.Transaction, _ uint32) (bool, error) {
	prefix := transaction.Prefix()

	structure, ok := generator.confirmer.Structure(prefix)
	if !ok {
		log.Printf("?????????????????? '%s' ???? ????????????????????", prefix.String())

		return false, nil
	}

	sender := transaction.Sender()
	if !structure.IsOwner(sender) {
		log.Printf("?????????? %s ???? ???????????????? ?????????????????? '%s'", sender.String(), prefix.String())

		return false, nil
	}

	if _, err := generator.confirmer.ProcessUpdateStructureLegacy(transaction); err != nil {
		log.Printf("????????????: %v", err)

		return false, fmt.Errorf("%w", err)
	}

	return true, nil
}

func (generator *Generator) processChangeProfitAddress(transaction umi.Transaction, _ uint32) (bool, error) {
	prefix := transaction.Prefix()

	structure, ok := generator.confirmer.Structure(prefix)
	if !ok {
		log.Printf("?????????????????? '%s' ???? ????????????????????", prefix.String())

		return false, nil
	}

	sender := transaction.Sender()
	if !structure.IsOwner(sender) {
		log.Printf("?????????? %s ???? ???????????????? ?????????????????? '%s'", sender.String(), prefix.String())

		return false, nil
	}

	account, ok := generator.confirmer.Account(transaction.Recipient())
	if !ok {
		return false, nil
	}

	if account.Type != umi.Deposit {
		return false, nil
	}

	if _, err := generator.confirmer.ProcessChangeProfitAddressLegacy(transaction); err != nil {
		log.Printf("????????????: %v", err)

		return false, fmt.Errorf("%w", err)
	}

	return true, nil
}

func (generator *Generator) processChangeFeeAddress(transaction umi.Transaction, _ uint32) (bool, error) {
	prefix := transaction.Prefix()

	structure, ok := generator.confirmer.Structure(prefix)
	if !ok {
		log.Printf("?????????????????? '%s' ???? ????????????????????", prefix.String())

		return false, nil
	}

	sender := transaction.Sender()
	if !structure.IsOwner(sender) {
		log.Printf("?????????? %s ???? ???????????????? ?????????????????? '%s'", sender.String(), prefix.String())

		return false, nil
	}

	account, ok := generator.confirmer.Account(transaction.Recipient())
	if !ok {
		return false, nil
	}

	if account.Type != umi.Deposit {
		return false, nil
	}

	if _, err := generator.confirmer.ProcessChangeFeeAddressLegacy(transaction); err != nil {
		log.Printf("????????????: %v", err)

		return false, fmt.Errorf("%w", err)
	}

	return true, nil
}

func (generator *Generator) processActivateTransit(transaction umi.Transaction, _ uint32) (bool, error) {
	prefix := transaction.Prefix()

	structure, ok := generator.confirmer.Structure(prefix)
	if !ok {
		log.Printf("?????????????????? '%s' ???? ????????????????????", prefix.String())

		return false, nil
	}

	sender := transaction.Sender()
	if !structure.IsOwner(sender) {
		log.Printf("?????????? %s ???? ???????????????? ?????????????????? '%s'", sender.String(), prefix.String())

		return false, nil
	}

	account, ok := generator.confirmer.Account(transaction.Recipient())
	if !ok {
		return false, nil
	}

	if account.Type != umi.Deposit {
		return false, nil
	}

	if _, err := generator.confirmer.ProcessActivateTransitLegacy(transaction); err != nil {
		log.Printf("????????????: %v", err)

		return false, fmt.Errorf("%w", err)
	}

	return true, nil
}

func (generator *Generator) processDeactivateTransit(transaction umi.Transaction, _ uint32) (bool, error) {
	prefix := transaction.Prefix()

	structure, ok := generator.confirmer.Structure(prefix)
	if !ok {
		log.Printf("?????????????????? '%s' ???? ????????????????????", prefix.String())

		return false, nil
	}

	sender := transaction.Sender()
	if !structure.IsOwner(sender) {
		log.Printf("?????????? %s ???? ???????????????? ?????????????????? '%s'", sender.String(), prefix.String())

		return false, nil
	}

	account, ok := generator.confirmer.Account(transaction.Recipient())
	if !ok {
		return false, nil
	}

	if account.Type != umi.Transit {
		return false, nil
	}

	if _, err := generator.confirmer.ProcessDeactivateTransitLegacy(transaction); err != nil {
		log.Printf("????????????: %v", err)

		return false, fmt.Errorf("%w", err)
	}

	return true, nil
}

func (generator *Generator) processBurn(transaction umi.Transaction, _ uint32) (bool, error) {
	sender := transaction.Sender()

	senderAccount, ok := generator.confirmer.Account(sender)
	if !ok {
		return false, nil
	}

	availableBalance := generator.confirmer.AvailableBalance(sender, senderAccount)
	if availableBalance < transaction.Amount() {
		return false, nil
	}

	if _, err := generator.confirmer.ProcessBurnLegacy(transaction); err != nil {
		log.Printf("????????????: %v", err)

		return false, fmt.Errorf("%w", err)
	}

	return true, nil
}

func (generator *Generator) processIssue(transaction umi.Transaction, _ uint32) (bool, error) {
	prefix := transaction.Prefix()

	structure, ok := generator.confirmer.Structure(prefix)
	if !ok {
		log.Printf("?????????????????? '%s' ???? ????????????????????", prefix.String())

		return false, nil
	}

	sender := transaction.Sender()
	if !structure.IsOwner(sender) {
		log.Printf("?????????? %s ???? ???????????????? ?????????????????? '%s'", sender.String(), prefix.String())

		return false, nil
	}

	if _, err := generator.confirmer.ProcessIssueLegacy(transaction); err != nil {
		log.Printf("????????????: %v", err)

		return false, fmt.Errorf("%w", err)
	}

	return true, nil
}

func (generator *Generator) processMintNftWitness(transaction umi.Transaction, _ uint32) (bool, error) {
	sender := transaction.Sender()

	senderAccount, ok := generator.confirmer.Account(sender)
	if !ok {
		return false, nil
	}

	availableBalance := generator.confirmer.AvailableBalance(sender, senderAccount)
	if availableBalance < transaction.Amount() {
		return false, nil
	}

	if _, err := generator.confirmer.ProcessMintNftWitnessLegacy(transaction); err != nil {
		log.Printf("????????????: %v", err)

		return false, fmt.Errorf("%w", err)
	}

	return true, nil
}
