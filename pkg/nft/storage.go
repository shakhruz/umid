package nft

import (
	"crypto/sha256"
	"encoding/binary"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"path"
	"sync"

	"gitlab.com/umitop/umid/pkg/config"
	storage1 "gitlab.com/umitop/umid/pkg/storage"
	"gitlab.com/umitop/umid/pkg/umi"
)

var (
	ErrNotFound = errors.New("not found")
	//errMalformed = errors.New("malformed")
)

type idx struct {
	Offset int64
	Length int
}

type Storage struct {
	sync.Mutex
	config     *config.Config
	tokens     map[umi.Hash]idx
	height     []umi.Hash
	indexFile  storage1.IFile
	lastOffset int64
}

func NewStorage(conf *config.Config) *Storage {
	return &Storage{
		config: conf,
		tokens: make(map[umi.Hash]idx, 1),
		height: make([]umi.Hash, 0),
	}
}

func (storage *Storage) OpenOrCreate() (err error) {
	cfg := storage.config
	dir := path.Join(cfg.DataDir, cfg.Network)

	if err = storage1.CheckOrCreateDir(storage1.NewFSx(), dir); err != nil {
		return err
	}

	indexName := path.Join(dir, "nft")

	if storage.indexFile, err = storage1.OpenOrCreateFile(storage1.NewFSx(), indexName, 0); err != nil {
		return err
	}

	return nil
}

func (storage *Storage) Close() {
	_ = storage.indexFile.Close()
}

func (storage *Storage) Scan() error {
	storage.Lock()
	defer storage.Unlock()

	const signLen = 98
	offset := storage.lastOffset

	for {
		hdrData := make([]byte, hdrLen)

		if _, err := storage.indexFile.ReadAt(hdrData, offset); err != nil {
			if errors.Is(err, io.EOF) {
				return nil
			}

			return fmt.Errorf("%w", err)
		}

		metaLength := int64(binary.BigEndian.Uint32(hdrData[9:13]))
		dataLength := int64(binary.BigEndian.Uint32(hdrData[13:17]))
		totalLength := metaLength + dataLength + signLen

		offset += int64(hdrLen)
		data := make([]byte, totalLength)

		if _, err := storage.indexFile.ReadAt(data, offset); err != nil {
			return fmt.Errorf("%w", err)
		}

		tx := make(Transaction, len(hdrData)+len(data))
		copy(tx[0:17], hdrData)
		copy(tx[17:], data)

		hash := tx.Hash()

		storage.tokens[hash] = idx{
			Offset: storage.lastOffset,
			Length: len(tx),
		}

		storage.height = append(storage.height, hash)

		offset += totalLength
		storage.lastOffset = offset
	}
}

func (storage *Storage) AppendData(data []byte) error {
	storage.Lock()
	defer storage.Unlock()

	n, err := storage.indexFile.WriteAt(data, storage.lastOffset)
	if err != nil {
		return fmt.Errorf("%w", err)
	}

	hash := sha256.Sum256(data)

	storage.tokens[hash] = idx{
		Offset: storage.lastOffset,
		Length: n,
	}

	storage.height = append(storage.height, hash)
	storage.lastOffset += int64(n)

	return nil
}

func (storage *Storage) Count() int {
	return len(storage.height)
}

func (storage *Storage) DataByHeight(height int) (data Transaction, err error) {
	if height >= len(storage.height) {
		return nil, ErrNotFound
	}

	hash := storage.height[height]

	return storage.Data(hash)
}

func (storage *Storage) Data(hash [32]byte) (data Transaction, err error) {
	storage.Lock()

	idx, ok := storage.tokens[hash]

	storage.Unlock()

	if !ok {
		return nil, ErrNotFound
	}

	tx := make(Transaction, idx.Length)

	if _, err := storage.indexFile.ReadAt(tx, idx.Offset); err != nil {
		return nil, fmt.Errorf("%w", err)
	}

	return tx, nil
}

func (storage *Storage) ParsedData(hash [32]byte) (meta json.RawMessage, data []byte, err error) {
	tx, err := storage.Data(hash)

	if err != nil {
		return nil, nil, err
	}

	return tx.Meta(), tx.Data(), nil
}
