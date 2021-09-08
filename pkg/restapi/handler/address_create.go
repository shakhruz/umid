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

package handler

import (
	"crypto/ed25519"
	"encoding/json"
	"net/http"

	"gitlab.com/umitop/umid/pkg/umi"
)

type CreateAddressRequest struct {
	Prefix *string `json:"prefix,omitempty"`
	Seed   *[]byte `json:"seed,omitempty"`
}

type CreateAddressResponse struct {
	Data  *string `json:"data,omitempty"`
	Error *Error  `json:"error,omitempty"`
}

func CreateAddress() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		setHeaders(w, r)

		response := new(CreateAddressResponse)
		response.Data, response.Error = processCreateAddress(r)

		_ = json.NewEncoder(w).Encode(response)
	}
}

func processCreateAddress(r *http.Request) (*string, *Error) {
	request := new(CreateAddressRequest)

	if err := json.NewDecoder(r.Body).Decode(request); err != nil {
		return nil, NewError(400, err.Error())
	}

	if request.Prefix == nil {
		return nil, NewError(-1, "Параметр 'prefix' является обязательным.")
	}

	if !umi.VerifyHrp(*request.Prefix) {
		return nil, NewError(-1, "Параметр 'prefix' может содержать только 3 символа латиницы в нижнем регистре.")
	}

	if request.Seed == nil {
		return nil, NewError(-1, "Параметр 'seed' является обязательным.")
	}

	if len(*request.Seed) != 32 {
		return nil, NewError(-1, "Длина параметра 'seed' должна быть 32 байта.")
	}

	secretKey := ed25519.NewKeyFromSeed(*request.Seed)
	publicKey := secretKey[ed25519.PublicKeySize:ed25519.PrivateKeySize]

	prefix := umi.ParsePrefix(*request.Prefix)

	address := umi.Address{}
	address.SetPrefix(prefix)
	address.SetPublicKey((umi.PublicKey)(publicKey))

	bech32 := address.String()

	return &bech32, nil
}
