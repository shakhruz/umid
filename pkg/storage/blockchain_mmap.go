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

// +build !windows

package storage

import (
	"encoding/binary"
	"fmt"
	"hash/crc32"
	"io"
	"log"
	"os"
	"sync"
	"syscall"
	"time"

	"gitlab.com/umitop/umid/pkg/config"
	"gitlab.com/umitop/umid/pkg/umi"
)

type BlockchainMmap struct {
	sync.Mutex
	config *config.Config

	index  []byte
	blocks []byte

	indexFile  *os.File
	blocksFile *os.File

	lastBlockHeight uint32
	lastBlockHash   [32]byte
	lastBlockTime   uint32

	subscriptions []chan umi.Block
}

func NewBlockchainMmap(config1 *config.Config) *BlockchainMmap {
	return &BlockchainMmap{
		config:        config1,
		subscriptions: make([]chan umi.Block, 0),
	}
}

func (bc *BlockchainMmap) OpenOrCreate() error {
	return bc.AppendBlock(GenesisBlock(bc.config.Network))
}

/*
func (bc *BlockchainMmap) Open(dir string) (err error) {
	subDir := fmt.Sprintf(subdirPath, dir)

	if err = ensureDirectory(subDir); err != nil {
		return err
	}

	idxPath := fmt.Sprintf(indexPathMb, subDir)
	blkPath := fmt.Sprintf(blockchainFile, subDir)

	if err = ensurerFile(idxPath, indexSizeByte); err != nil {
		return err
	}

	if err = ensurerFile(blkPath, blockchainByte); err != nil {
		return err
	}

	// mmap для идекса
	f1, err := os.Open(idxPath)
	if err != nil {
		return fmt.Errorf("%w", err)
	}

	if bc.index, err = syscall.Mmap(int(f1.Fd()), 0, indexSizeByte, syscall.PROT_READ, syscall.MAP_SHARED); err != nil {
		return fmt.Errorf("%w", err)
	}

	if err = f1.Close(); err != nil {
		return fmt.Errorf("%w", err)
	}

	// mmap для блоков
	f2, err := os.Open(blkPath)
	if err != nil {
		return fmt.Errorf("%w", err)
	}

	if bc.blocks, err = syscall.Mmap(int(f2.Fd()), 0, blockchainByte, syscall.PROT_READ, syscall.MAP_SHARED); err != nil {
		return fmt.Errorf("%w", err)
	}

	if err = f2.Close(); err != nil {
		return fmt.Errorf("%w", err)
	}

	// Открываем индекс для записи
	if bc.indexFile, err = os.OpenFile(idxPath, os.O_RDWR, 0644); err != nil {
		return fmt.Errorf("%w", err)
	}

	// Открываем файл блокчейна для записи
	if bc.dataFile, err = os.OpenFile(blkPath, os.O_RDWR, 0644); err != nil {
		return fmt.Errorf("%w", err)
	}

	return fmt.Errorf("%w", err)
}
*/

func (bc *BlockchainMmap) Close() {
	if err := syscall.Munmap(bc.blocks); err != nil {
		log.Println(err)
	}

	if err := syscall.Munmap(bc.index); err != nil {
		log.Println(err)
	}

	if err := bc.blocksFile.Close(); err != nil {
		log.Println(err)
	}

	if err := bc.indexFile.Close(); err != nil {
		log.Println(err)
	}
}

func (bc *BlockchainMmap) Subscribe(ch chan umi.Block) {
	bc.subscriptions = append(bc.subscriptions, ch)
}

func (bc *BlockchainMmap) notify(block []byte) {
	for _, ch := range bc.subscriptions {
		ch <- block
	}
}

/*
func (bc *BlockchainMmap) QuickScan(updater *ledger.Confirmer) {
	log.Println("blockchain: quick scan start")

	t := time.Now()

	for {
		idx := bc.lastBlockHeight * 8

		low := binary.BigEndian.Uint32(bc.index[idx : idx+4])        // начало блока
		checksum := binary.BigEndian.Uint32(bc.index[idx+4 : idx+8]) // котрольная сумма
		high := binary.BigEndian.Uint32(bc.index[idx+8 : idx+12])    // конец блока

		if high <= low || high-low > 9830417 {
			break
		}

		if checksum != crc32.ChecksumIEEE(bc.blocks[low:high]) {
			break
		}

		block := umi.Block(bc.blocks[low:high])

		if bc.lastBlockHash != block.PreviousBlockHash() {
			break
		}

		if bc.lastBlockTime > block.Timestamp() {
			break
		}

		if err := updater.ProcessBlock(block); err != nil {
			log.Printf("blockchain: блок %d (%x) не прошел проверку", bc.lastBlockHeight+1, block.Hash())

			break
		}

		if err := updater.Commit(); err != nil {
			log.Printf("blockchain: блок %d (%x) не закоммитился", bc.lastBlockHeight+1, block.Hash())

			break
		}

		bc.lastBlockHeight++
		bc.lastBlockHash = block.Hash()
		bc.lastBlockTime = block.Timestamp()
	}

	log.Println("blockchain: quick scan done", time.Since(t), bc.lastBlockHeight)
}
*/

func (bc *BlockchainMmap) Block(height uint32) (umi.Block, error) {
	if height > bc.lastBlockHeight || height == 0 {
		return nil, fmt.Errorf("%w: блок еще не добавлен в блокчейн", ErrNotFound)
	}

	idx := (height - 1) * 8

	low := binary.BigEndian.Uint32(bc.index[idx : idx+4])
	high := binary.BigEndian.Uint32(bc.index[idx+8 : idx+12])

	if low >= high {
		return nil, fmt.Errorf("%w: индекс содержит некорректные данные", errMalformed)
	}

	return bc.blocks[low:high], nil
}

func (bc *BlockchainMmap) StreamBlocks(writer io.Writer, height, limit uint32) {
	_ = bc
	_ = writer
	_ = height
	_ = limit
}

func (bc *BlockchainMmap) Transaction(blockHeight uint32, txIndex uint16) (umi.Transaction, bool) {
	_ = bc
	_, _ = blockHeight, txIndex

	return nil, false
}

func (bc *BlockchainMmap) AppendBlock(blk umi.Block) error {
	if bc.lastBlockHash != blk.PreviousBlockHash() {
		return fmt.Errorf("%w: новый блок не ссылается на последний", ErrBlockSequence)
	}

	height := bc.lastBlockHeight + 1

	if err := bc.insertBlock(height, blk); err != nil {
		return err
	}

	bc.lastBlockHeight = height
	bc.lastBlockHash = blk.Hash()

	bc.notify(blk)

	if height%100_000 == 0 {
		log.Printf("blockchain: блок %d добавлен", height)
	}

	return nil
}

func (bc *BlockchainMmap) Height() int {
	return int(bc.lastBlockHeight)
}

func (bc *BlockchainMmap) insertBlock(height uint32, block []byte) error {
	idxOffset := (height - 1) * 8
	blkOffset := binary.BigEndian.Uint32(bc.index[idxOffset : idxOffset+4])

	idxData := make([]byte, 8)
	binary.BigEndian.PutUint32(idxData[0:4], crc32.ChecksumIEEE(block))    // контроьная сумма
	binary.BigEndian.PutUint32(idxData[4:8], blkOffset+uint32(len(block))) // индекс следующего блока

	if _, err := bc.indexFile.WriteAt(idxData, int64(idxOffset+4)); err != nil {
		return fmt.Errorf("%w", err)
	}

	if _, err := bc.blocksFile.WriteAt(block, int64(blkOffset)); err != nil {
		return fmt.Errorf("%w", err)
	}

	return nil
}

func (bc *BlockchainMmap) FastScan() {
	log.Println("blockchain: fast scan start")

	t := time.Now()

	for {
		off := bc.lastBlockHeight * 8

		data := make([]byte, 12) // 4 bytes * 3

		if _, err := bc.indexFile.ReadAt(data, int64(off)); err != nil {
			log.Println(err)

			return
		}

		low := binary.BigEndian.Uint32(data[0:4])      // начало блока
		crc32sum := binary.BigEndian.Uint32(data[4:8]) // котрольная сумма
		high := binary.BigEndian.Uint32(data[8:12])    // конец блока

		if high <= low || high-low > 9830417 {
			break
		}

		blk := make([]byte, high-low)

		if _, err := bc.blocksFile.ReadAt(blk, int64(low)); err != nil {
			log.Println(err)

			return
		}

		if crc32sum != crc32.ChecksumIEEE(blk) {
			break
		}

		block := (umi.Block)(blk)

		if bc.lastBlockHash != block.PreviousBlockHash() {
			break
		}

		bc.lastBlockHeight++
		bc.lastBlockHash = block.Hash()
		bc.lastBlockTime = block.Timestamp()
	}

	log.Println("blockchain: fast scan done", time.Since(t), bc.lastBlockHeight)
}
