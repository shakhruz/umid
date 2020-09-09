// Copyright (c) 2020 UMI
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

package libumi

// TxAddress is ...
type TxAddress []byte

// NewTxUpdProfitAddr ...
func NewTxUpdProfitAddr() TxAddress {
	t := make([]byte, TxLength)
	setTxVersion(t, UpdateProfitAddress)

	return t
}

// NewTxUpdFeeAddr ...
func NewTxUpdFeeAddr() TxAddress {
	t := make([]byte, TxLength)
	setTxVersion(t, UpdateFeeAddress)

	return t
}

// NewTxCrtTransitAddr ...
func NewTxCrtTransitAddr() TxAddress {
	t := make([]byte, TxLength)
	setTxVersion(t, CreateTransitAddress)

	return t
}

// NewTxDelTransitAddr ...
func NewTxDelTransitAddr() TxAddress {
	t := make([]byte, TxLength)
	setTxVersion(t, DeleteTransitAddress)

	return t
}

// Sender ...
func (t TxAddress) Sender() Address {
	return Address(t[1:35])
}

// SetSender ...
func (t TxAddress) SetSender(a Address) {
	copy(t[1:35], a)
}

// Address ...
func (t TxAddress) Address() Address {
	return Address(t[35:69])
}

// SetAddress ...
func (t TxAddress) SetAddress(a Address) {
	copy(t[35:69], a)
}
