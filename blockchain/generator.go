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

package blockchain

import (
	"log"
	"time"

	"github.com/umitop/libumi"
)

const newBlockDelaySec = 10

// Generator ...
func (bc *Blockchain) Generator() {
	bc.wg.Add(1)
	defer bc.wg.Done()

	for {
		time.Sleep(newBlockDelaySec * time.Second)

		select {
		case <-bc.ctx.Done():
			return
		default:
			bc.generateNewBlock()
		}
	}
}

func (bc *Blockchain) generateNewBlock() {
	m, err := bc.storage.Mempool()
	if err != nil {
		log.Println(err.Error())

		return
	}
	defer m.Close()

	// now := time.Now().Unix()

	for m.Next() {
		_ = (libumi.Transaction)(m.Value())

		return
	}
}
