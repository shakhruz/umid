package handler

import (
	"encoding/hex"
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"strings"
	"time"

	"gitlab.com/umitop/umid/pkg/ledger"
	"gitlab.com/umitop/umid/pkg/nft"
	"gitlab.com/umitop/umid/pkg/umi"
)

type GetNftMetaResponse struct {
	Data  *json.RawMessage `json:"data,omitempty"`
	Error *Error           `json:"error,omitempty"`
}

type GetNftRawResponse struct {
	Data  *[]byte `json:"data,omitempty"`
	Error *Error  `json:"error,omitempty"`
}

type ListNftsByAddressResponse struct {
	Data  *[]string `json:"data,omitempty"`
	Error *Error    `json:"error,omitempty"`
}

type ListNftRawResponse struct {
	Data  *[][]byte `json:"data,omitempty"`
	Error *Error    `json:"error,omitempty"`
}

func ListNftsByAddress(ledger1 *ledger.Ledger) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		setHeaders(w, r)

		response := new(ListNftsByAddressResponse)
		response.Data, response.Error = processListNftsByAddress(r, ledger1)

		_ = json.NewEncoder(w).Encode(response)
	}
}

func processListNftsByAddress(r *http.Request, ledger1 *ledger.Ledger) (*[]string, *Error) {
	bech32 := strings.TrimPrefix(r.URL.Path, "/api/addresses/")
	bech32 = strings.TrimSuffix(bech32, "/nfts")

	address, err := umi.ParseAddress(bech32)
	if err != nil {
		return nil, NewError(400, err.Error())
	}

	nfts := ledger1.NftsByAddr(address)

	hashes := make([]string, 0, len(nfts))

	for _, hash := range nfts {
		hashes = append(hashes, hash.String())
	}

	return &hashes, nil
}

func ListNft(nftStorage *nft.Storage) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Query().Get("raw") {
		case ParamTrue:
			response := new(ListNftRawResponse)
			response.Data, response.Error = processListNft(r, nftStorage)

			_ = json.NewEncoder(w).Encode(response)

		default:
			http.Error(w, "Not found", http.StatusNotFound)
		}
	}
}

func processListNft(r *http.Request, nftStorage *nft.Storage) (*[][]byte, *Error) {
	var height int

	idx := r.URL.Query().Get("height")
	if idx != "" {
		nftHeight, err := strconv.ParseUint(idx, 10, 32)
		if err != nil {
			return nil, NewError(400, err.Error())
		}

		height = int(nftHeight)
	}

	data, err := nftStorage.DataByHeight(height)
	if err != nil {
		return nil, NewError(400, err.Error())
	}

	resp := make([][]byte, 0, 1)
	resp = append(resp, data)

	return &resp, nil
}

func GetNft(nftStorage *nft.Storage) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		hexHash := strings.TrimPrefix(r.URL.Path, "/api/nfts/")
		if len(hexHash) != 64 {
			http.Error(w, "Not found", http.StatusNotFound)

			return
		}

		hashSlice, err := hex.DecodeString(hexHash)
		if err != nil {
			http.Error(w, "Not found", http.StatusNotFound)

			return
		}

		var hash umi.Hash
		copy(hash[:], hashSlice)

		tx, err := nftStorage.Data(hash)
		if err != nil {
			http.Error(w, "Not found", http.StatusNotFound)

			return
		}

		switch r.URL.Query().Get("raw") {
		case ParamTrue:
			response := new(GetNftRawResponse)
			response.Data = (*[]byte)(&tx)

			_ = json.NewEncoder(w).Encode(response)

		default:
			m := struct {
				ContentType *string `json:"contentType,omitempty"`
			}{}

			contentType := "application/octet-stream"

			if err := json.Unmarshal(tx.Meta(), &m); err == nil {
				if m.ContentType != nil {
					contentType = *m.ContentType
				}
			}

			data := tx.Data()

			w.Header().Set("Content-Type", contentType)
			w.Header().Set("Content-Length", fmt.Sprintf("%d", len(data)))
			w.Header().Set("Last-Modified", time.Unix(int64(tx.Timestamp()), 0).UTC().Format(http.TimeFormat))
			w.Header().Set("ETag", fmt.Sprintf("%x", tx.Hash()))
			w.Header().Set("Cache-Control", "public, immutable, no-transform, max-age=86400")

			_, _ = w.Write(data)
		}
	}
}

func GetNftMeta(nftStorage *nft.Storage) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		setHeaders(w, r)

		response := new(GetNftMetaResponse)
		response.Data, response.Error = processGetNftMeta(r, nftStorage)

		_ = json.NewEncoder(w).Encode(response)
	}
}

func processGetNftMeta(r *http.Request, nftStorage *nft.Storage) (*json.RawMessage, *Error) {
	hexHash := strings.TrimPrefix(r.URL.Path, "/api/nfts/")
	hexHash = strings.TrimSuffix(hexHash, "/meta")

	if len(hexHash) != 64 {
		return nil, NewError(404, "Not found")
	}

	hashSlice, err := hex.DecodeString(hexHash)
	if err != nil {
		return nil, NewError(404, "Not found")
	}

	var hash umi.Hash
	copy(hash[:], hashSlice)

	meta, _, err := nftStorage.ParsedData(hash)
	if err != nil {
		return nil, NewError(404, "Not found")
	}

	if len(meta) == 0 {
		meta = json.RawMessage("null")
	}

	return &meta, nil
}
