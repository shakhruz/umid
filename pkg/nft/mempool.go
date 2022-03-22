package nft

import (
	"context"
	"errors"
	"fmt"
	"sync"
	"time"

	"gitlab.com/umitop/umid/pkg/ledger"
	"gitlab.com/umitop/umid/pkg/umi"
)

var ErrMempool = errors.New("mempool")

type iLedger interface {
	Account(address umi.Address) (account *ledger.Account, ok bool)
	Structure(prefix umi.Prefix) (structure *ledger.Structure, ok bool)
	HasTransaction(hash umi.Hash) bool
}

type iSubscriber interface {
	Subscribe(chan umi.Block)
}

type Mempool struct {
	sync.RWMutex
	ledger       iLedger
	blocks       chan umi.Block
	transactions map[umi.Hash][]byte
}

func NewMempool() *Mempool {
	return &Mempool{
		blocks:       make(chan umi.Block, 64),
		transactions: make(map[umi.Hash][]byte),
	}
}

func (mempool *Mempool) Mempool() (txs [][]byte) {
	mempool.RLock()
	defer mempool.RUnlock()

	txs = make([][]byte, 0, len(mempool.transactions))

	for _, transaction := range mempool.transactions {
		txs = append(txs, transaction)
	}

	return txs
}

func (mempool *Mempool) SubscribeTo(subscriber iSubscriber) {
	subscriber.Subscribe(mempool.blocks)
}

func (mempool *Mempool) Worker(ctx context.Context) {
	ticker := time.NewTicker(time.Second * 30)
	defer ticker.Stop()

	for {
		select {
		case block := <-mempool.blocks:
			mempool.ParseBlock(block)

		case <-ticker.C:
			mempool.cleanup()

		case <-ctx.Done():
			return
		}
	}
}

func (mempool *Mempool) SetLedger(ledger1 iLedger) {
	mempool.ledger = ledger1
}

func (mempool *Mempool) Push(transaction []byte) error {
	tx := (Transaction)(transaction)

	hash := tx.Hash()

	mempool.Lock()
	defer mempool.Unlock()

	if _, ok := mempool.transactions[hash]; ok {
		return fmt.Errorf("%w: tranasction in mempool", ErrMempool)
	}

	if mempool.ledger.HasTransaction(hash) {
		return fmt.Errorf("%w: tranasction confirmed", ErrMempool)
	}

	//senderAccount, ok := mempool.ledger.Account(tx.Sender())
	//if !ok {
	//return fmt.Errorf("%w: sender account not found", ErrMempool)
	//}

	//if senderAccount.BalanceAt(uint32(time.Now().Unix())) < uint64(len(transaction)) {
	//	return fmt.Errorf("%w: insufficient funds", ErrMempool)
	//}

	mempool.transactions[hash] = transaction

	return nil
}

func (mempool *Mempool) ParseBlock(block umi.Block) {
	mempool.Lock()
	defer mempool.Unlock()

	for i, txCount := 0, block.TransactionCount(); i < txCount; i++ {
		transaction := block.Transaction(i)
		if transaction.Version() == umi.TxV18MintNftWitness {
			hash := transaction.Hash()

			mempool.remove(hash)
		}
	}
}

func (mempool *Mempool) remove(hash umi.Hash) {
	_, ok := mempool.transactions[hash]
	if !ok {
		return
	}

	delete(mempool.transactions, hash)
}

func (mempool *Mempool) cleanup() {
	timestamp := uint32(time.Now().Unix())

	mempool.Lock()
	defer mempool.Unlock()

	for hash, tx := range mempool.transactions {
		transaction := (Transaction)(tx)
		txTimestamp := transaction.Timestamp()

		// Транзакция из будущего.
		if txTimestamp > timestamp {
			mempool.remove(hash)

			continue
		}

		// Просроченная транзакция.
		if timestamp-txTimestamp > 3600 {
			mempool.remove(hash)

			continue
		}

		// Баланс отправителя не существует.
		account, ok := mempool.ledger.Account(transaction.Sender())
		if !ok {
			mempool.remove(hash)

			continue
		}

		// На балансе недостаточно монет.
		if account.BalanceAt(timestamp) < uint64(len(transaction)) {
			mempool.remove(hash)

			continue
		}
	}
}
