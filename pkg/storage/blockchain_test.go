package storage_test

import (
	"testing"
)

// Должны успешно создаться все необходимые файлы
func TestA001(t *testing.T) {
	//mockFs := NewMockFs()

	//blockchain := NewBlockchain(config.DefaultConfig())

	//err := blockchain.OpenOrCreate()

	//if err != nil {
	//	t.Errorf("ожидаем 'nil', получили '%v'", err)
	//}

	//if _, ok := mockFs.files["/tmp/mainnet/index"]; !ok {
	//	t.Fatalf("должен был создаться файл '%s'", "/tmp/mainnet/index")
	//}

	//f1 := mockFs.files["/tmp/mainnet/index"]
	//size1 := f1.Size()

	//if size1 != int64(config.IndexSize) {
	//	t.Errorf("%d %d", config.IndexSize, size1)
	//}

	//if _, ok := mockFs.files["/tmp/mainnet/blockchain0"]; !ok {
	//	t.Fatalf("должен был создаться файл '%s'", "/tmp/mainnet/blockchain0")
	//}

	//f2 := mockFs.files["/tmp/mainnet/blockchain0"]
	//size2 := f2.Size()

	//if size2 != int64(config.ChunkSize) {
	//	t.Errorf("%d %d", config.ChunkSize, size2)
	//}

	//block, err := blockchain.Block(1)

	//if err != nil {
	//	t.Errorf("ожидаем 'nil', получили '%v'", err)
	//}

	//if !bytes.Equal(block, GenesisBlock(Mainnet)) {
	//	t.Errorf("ожидаем '%x', получили '%x'", sha256.Sum256(GenesisBlock(Mainnet)), sha256.Sum256(block))
	//}
}
