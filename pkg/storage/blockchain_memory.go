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

package storage

import (
	"io"
	"sync"

	"gitlab.com/umitop/umid/pkg/config"
	"gitlab.com/umitop/umid/pkg/umi"
)

type BlockchainMemory struct {
	sync.Mutex
	config *config.Config

	blocks []umi.Block

	lastBlockHash umi.Hash
	lastBlockTime uint32

	subscriptions []chan umi.Block
}

func NewBlockchainMemory(conf *config.Config) *BlockchainMemory {
	return &BlockchainMemory{
		config:        conf,
		blocks:        make([]umi.Block, 0),
		subscriptions: make([]chan umi.Block, 0),
	}
}

func (bc *BlockchainMemory) OpenOrCreate() error {
	_ = bc

	return nil // bc.AppendBlock(GenesisBlock(bc.config.Network))
}

func (bc *BlockchainMemory) AppendBlock(block umi.Block) error {
	bc.Lock()
	defer bc.Unlock()

	if bc.lastBlockHash != block.PreviousBlockHash() {
		return ErrBlockSequence
	}

	bc.blocks = append(bc.blocks, block)
	bc.lastBlockHash = block.Hash()
	bc.lastBlockTime = block.Timestamp()

	bc.notify(block)

	return nil
}

func (bc *BlockchainMemory) StreamBlocks(writer io.Writer, height, limit uint32) {
	_, _ = height, limit

	for _, block := range bc.blocks {
		if _, err := writer.Write(block); err != nil {
			break
		}
	}
}

func (bc *BlockchainMemory) Block(height uint32) (umi.Block, error) {
	if height == 0 || int(height) >= len(bc.blocks) {
		return nil, ErrNotFound
	}

	index := height - 1
	block := bc.blocks[index]

	return block, nil
}

func (bc *BlockchainMemory) Transaction(blockHeight uint32, txIndex uint16) (umi.Transaction, bool) {
	if blockHeight == 0 || blockHeight >= uint32(len(bc.blocks)) {
		return nil, false
	}

	blockIndex := blockHeight - 1

	return bc.blocks[blockIndex].Transaction(int(txIndex)), true
}

func (bc *BlockchainMemory) Subscribe(ch chan umi.Block) {
	bc.subscriptions = append(bc.subscriptions, ch)
}

func (bc *BlockchainMemory) Height() int {
	return len(bc.blocks)
}

func (*BlockchainMemory) Scan(iConfirmer) error {
	return nil
}

func (bc *BlockchainMemory) notify(block umi.Block) {
	for _, ch := range bc.subscriptions {
		ch <- block
	}
}

func (*BlockchainMemory) Close() {}
