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
	"fmt"
	"net/http"
	"strings"

	"gitlab.com/umitop/umid/pkg/umi"
)

type iEvents interface {
	Subscribe(address umi.Address, ch chan<- []byte)
	Unsubscribe(address umi.Address, ch chan<- []byte)
}

func SubscribeAddress(event iEvents) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		cors(w, r)
		w.Header().Set("Content-Type", "text/event-stream")
		w.Header().Set("Cache-Control", "no-cache")

		bech32 := strings.TrimPrefix(r.URL.Path, "/events/addresses/")

		address, err := umi.ParseAddress(bech32)
		if err != nil {
			return
		}

		flusher, _ := w.(http.Flusher)
		ch := make(chan []byte, 64)

		event.Subscribe(address, ch)
		defer event.Unsubscribe(address, ch)

		_, _ = fmt.Fprintf(w, ": %s\n\n", address.String())

		flusher.Flush()

		for {
			select {
			case data := <-ch:
				if _, err = w.Write(data); err != nil {
					return
				}

				flusher.Flush()

			case <-r.Context().Done():
				return
			}
		}
	}
}
