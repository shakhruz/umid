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
	"encoding/binary"
	"errors"
	"fmt"
	"hash/crc32"
	"io"
	"log"
	"path"
	"sync"

	"gitlab.com/umitop/umid/pkg/config"
	"gitlab.com/umitop/umid/pkg/umi"
)

const (
	minBlockSize  = umi.HdrLength + umi.TxConfirmedLength
	maxBlockSize  = umi.HdrLength + (umi.TxConfirmedLength * 65535)
	indexDataSize = 14
)

var (
	ErrNotFound      = errors.New("not found")
	ErrBlockSequence = errors.New("block sequence")
	errMalformed     = errors.New("malformed")
)

type IBlockchain interface {
	OpenOrCreate() error
	Close()
	Subscribe(chan umi.Block)
	Scan(confirmer iConfirmer) error
	AppendBlock(umi.Block) error
	Block(uint32) (umi.Block, error)
	Transaction(uint32, uint16) (umi.Transaction, bool)
	StreamBlocks(io.Writer, uint32, uint32)
	Height() int
}

type iConfirmer interface {
	ProcessBlock([]byte) error
	Commit() error
}

type Blockchain struct {
	sync.Mutex
	config *config.Config

	indexFile   IFile
	chunkFiles  map[uint16]IFile
	chunkIndex  uint16
	chunkOffset uint32

	lastBlockHeight uint32
	lastBlockHash   umi.Hash
	lastBlockTime   uint32

	subscriptions []chan umi.Block
}

func NewBlockchain(conf *config.Config) *Blockchain {
	return &Blockchain{
		config:        conf,
		chunkFiles:    make(map[uint16]IFile, 1),
		subscriptions: make([]chan umi.Block, 0),
	}
}

func (bc *Blockchain) OpenOrCreate() (err error) {
	cfg := bc.config
	dir := path.Join(cfg.DataDir, cfg.Network)

	if err = CheckOrCreateDir(NewFSx(), dir); err != nil {
		return err
	}

	indexName := path.Join(dir, "index")

	if bc.indexFile, err = OpenOrCreateFile(NewFSx(), indexName, cfg.IndexSize); err != nil {
		return err
	}

	return nil
}

func (bc *Blockchain) Close() {
	_ = bc.indexFile.Close()

	for _, file := range bc.chunkFiles {
		_ = file.Close()
	}
}

func (bc *Blockchain) Subscribe(ch chan umi.Block) {
	bc.subscriptions = append(bc.subscriptions, ch)
}

func (bc *Blockchain) Scan(confirmer iConfirmer) error {
	bc.Lock()
	defer bc.Unlock()

	indexData := make([]byte, indexDataSize)

	for {
		indexOffset := bc.lastBlockHeight * indexDataSize

		if _, err := bc.indexFile.ReadAt(indexData, int64(indexOffset)); err != nil {
			return fmt.Errorf("%w", err)
		}

		chunkIndex := binary.BigEndian.Uint16(indexData[0:2])
		chunkOffset := binary.BigEndian.Uint32(indexData[2:6])
		blockSize := binary.BigEndian.Uint32(indexData[6:10])
		blockChecksum := binary.BigEndian.Uint32(indexData[10:14])

		if chunkIndex == bc.chunkIndex && chunkOffset != bc.chunkOffset {
			return nil
		}

		if blockSize < minBlockSize || blockSize > maxBlockSize {
			return nil
		}

		block := make(umi.Block, blockSize)

		if err := bc.chunkReadAt(block, chunkIndex, chunkOffset); err != nil {
			return err
		}

		if blockChecksum != crc32.ChecksumIEEE(block) {
			return nil
		}

		if err := confirmer.ProcessBlock(block); err != nil {
			log.Printf("blockchain: блок %d (%x) не прошел проверку %s", bc.lastBlockHeight+1, block.Hash(), err.Error())

			return nil
		}

		_ = confirmer.Commit()

		bc.chunkIndex = chunkIndex
		bc.chunkOffset = chunkOffset + blockSize
		bc.lastBlockHash = block.Hash()
		bc.lastBlockTime = block.Timestamp()
		bc.lastBlockHeight++

		bc.notify(block)
	}
}

func (bc *Blockchain) Block(height uint32) (umi.Block, error) {
	if height == 0 || height > bc.lastBlockHeight {
		return nil, ErrNotFound
	}

	b := make([]byte, 10)
	off := 14 * (height - 1)

	if _, err := bc.indexFile.ReadAt(b, int64(off)); err != nil {
		err = fmt.Errorf("%w", err)

		return nil, err
	}

	chunk := binary.BigEndian.Uint16(b[0:2])
	offset := binary.BigEndian.Uint32(b[2:6])
	length := binary.BigEndian.Uint32(b[6:10])

	block := make([]byte, length)

	if err := bc.chunkReadAt(block, chunk, offset); err != nil {
		return nil, err
	}

	return block, nil
}

func (bc *Blockchain) Transaction(blockHeight uint32, txIndex uint16) (umi.Transaction, bool) {
	if blockHeight == 0 || blockHeight > bc.lastBlockHeight {
		return nil, false
	}

	b := make([]byte, 6)
	off := 14 * (blockHeight - 1)

	if _, err := bc.indexFile.ReadAt(b, int64(off)); err != nil {
		return nil, false
	}

	chunk := binary.BigEndian.Uint16(b[0:2])
	offset := binary.BigEndian.Uint32(b[2:6]) + umi.HdrLength + (umi.TxConfirmedLength * uint32(txIndex))

	transaction := make([]byte, umi.TxConfirmedLength)

	if err := bc.chunkReadAt(transaction, chunk, offset); err != nil {
		return nil, false
	}

	return transaction, true
}

func (bc *Blockchain) StreamBlocks(writer io.Writer, height, limit uint32) {
	_, _ = height, limit
	reader := io.NewSectionReader(bc.chunkFiles[0], 0, int64(bc.chunkOffset))
	_, _ = io.Copy(writer, reader)
}

func (bc *Blockchain) AppendBlock(block umi.Block) error {
	bc.Lock()
	defer bc.Unlock()

	if bc.lastBlockHash != block.PreviousBlockHash() {
		return ErrBlockSequence
	}

	blockSize := uint32(len(block))
	chunkIndex := bc.chunkIndex
	chunkOffset := bc.chunkOffset

	if blockSize > uint32(bc.config.ChunkSize)-bc.chunkOffset {
		chunkIndex++

		chunkOffset = 0
	}

	if err := bc.chunkWriteAt(block, chunkIndex, chunkOffset); err != nil {
		return err
	}

	if err := bc.appendIndex(block, chunkIndex, chunkOffset); err != nil {
		return err
	}

	bc.chunkIndex = chunkIndex
	bc.chunkOffset = chunkOffset + blockSize
	bc.lastBlockHeight++
	bc.lastBlockHash = block.Hash()
	bc.lastBlockTime = block.Timestamp()

	bc.notify(block)

	return nil
}

func (bc *Blockchain) Height() int {
	return int(bc.lastBlockHeight)
}

func (bc *Blockchain) appendIndex(block []byte, chunkIndex uint16, chunkOffset uint32) error {
	blockSize := uint32(len(block))
	blockChecksum := crc32.ChecksumIEEE(block)

	data := make([]byte, 14)

	binary.BigEndian.PutUint16(data[0:2], chunkIndex)
	binary.BigEndian.PutUint32(data[2:6], chunkOffset)
	binary.BigEndian.PutUint32(data[6:10], blockSize)
	binary.BigEndian.PutUint32(data[10:14], blockChecksum)

	offset := 14 * bc.lastBlockHeight

	if _, err := bc.indexFile.WriteAt(data, int64(offset)); err != nil {
		return fmt.Errorf("%w", err)
	}

	return nil
}

func (bc *Blockchain) chunkWriteAt(data []byte, index uint16, offset uint32) error {
	chunk, err := bc.chunk(index)
	if err != nil {
		return err
	}

	if _, err = chunk.WriteAt(data, int64(offset)); err != nil {
		err = fmt.Errorf("%w", err)
	}

	return err
}

func (bc *Blockchain) chunkReadAt(data []byte, index uint16, offset uint32) error {
	chunk, err := bc.chunk(index)
	if err != nil {
		return err
	}

	if _, err = chunk.ReadAt(data, int64(offset)); err != nil {
		err = fmt.Errorf("%w", err)
	}

	return err
}

func (bc *Blockchain) chunk(index uint16) (IFile, error) {
	if f, ok := bc.chunkFiles[index]; ok {
		return f, nil
	}

	cfg := bc.config
	fileName := fmt.Sprintf("blockchain%d", index)
	filePath := path.Join(cfg.DataDir, cfg.Network, fileName)

	file, err := OpenOrCreateFile(NewFSx(), filePath, cfg.ChunkSize)
	if err != nil {
		return nil, err
	}

	bc.chunkFiles[index] = file

	return file, nil
}

func (bc *Blockchain) notify(block umi.Block) {
	for _, ch := range bc.subscriptions {
		ch <- block
	}
}
