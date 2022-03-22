package nft

import (
	"crypto/ed25519"
	"crypto/sha256"
	"encoding/binary"
	"encoding/json"

	"gitlab.com/umitop/umid/pkg/umi"
)

const hdrLen = 17

type Transaction []byte

func NewTransaction() *Transaction {
	tx := make(Transaction, hdrLen) // type (1 byte) + time (4 byte) + nonce (4 bytes) + meta (4 bytes) + data (4 bytes)
	tx[0] = 17                      // umi.TxV17MintNft

	return &tx
}

func (t *Transaction) Hash() [32]byte {
	return sha256.Sum256(*t)
}

func (t *Transaction) SetTimestamp(epoch uint32) {
	tx := *t
	binary.BigEndian.PutUint32(tx[1:5], epoch)
}

func (t *Transaction) Timestamp() uint32 {
	tx := *t

	return binary.BigEndian.Uint32(tx[1:5])
}

func (t *Transaction) SetNonce(nonce uint32) {
	tx := *t
	binary.BigEndian.PutUint32(tx[5:9], nonce)
}

func (t *Transaction) Nonce() uint32 {
	tx := *t

	return binary.BigEndian.Uint32(tx[5:9])
}

func (t *Transaction) SetMeta(meta json.RawMessage) {
	tx := *t

	binary.BigEndian.PutUint32(tx[9:13], uint32(len(meta)))

	tx = append(tx, meta...)

	*t = tx
}

func (t *Transaction) Meta() json.RawMessage {
	tx := *t

	metaLen := int(binary.BigEndian.Uint32(tx[9:13]))
	metaStart := hdrLen
	metaStop := metaStart + metaLen

	meta := make([]byte, metaLen)

	copy(meta[:], tx[metaStart:metaStop])

	return meta
}

func (t *Transaction) SetData(data []byte) {
	tx := *t

	binary.BigEndian.PutUint32(tx[13:17], uint32(len(data)))

	tx = append(tx, data...)

	*t = tx
}

func (t *Transaction) Data() []byte {
	tx := *t

	metaLen := int(binary.BigEndian.Uint32(tx[9:13]))
	dataLen := int(binary.BigEndian.Uint32(tx[13:17]))

	dataStart := hdrLen + metaLen
	dataStop := dataStart + dataLen

	return tx[dataStart:dataStop]
}

func (t *Transaction) SetSender(addr umi.Address) {
	tx := *t

	tx = append(tx, addr[:]...)

	*t = tx
}

func (t *Transaction) Sender() umi.Address {
	tx := *t

	metaLen := int(binary.BigEndian.Uint32(tx[9:13]))
	dataLen := int(binary.BigEndian.Uint32(tx[13:17]))

	senderStart := hdrLen + metaLen + dataLen
	senderStop := senderStart + umi.AddrLength

	var addr umi.Address

	copy(addr[:], tx[senderStart:senderStop])

	return addr
}

func (t *Transaction) Sign(sec ed25519.PrivateKey) {
	tx := *t

	signature := ed25519.Sign(sec, tx)

	tx = append(tx, signature...)

	*t = tx
}
