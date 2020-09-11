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

package network

import (
	"crypto/tls"
	"fmt"
	"net/http"
	"net/http/cookiejar"
	"os"
	"time"
)

const (
	clientTimeoutSec = 10
)

type transport struct {
	tr http.RoundTripper
}

func (t *transport) RoundTrip(req *http.Request) (*http.Response, error) {
	req.Header.Set("User-Agent", "UMId/0.0.1")
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Connection", "keep-alive")

	return t.tr.RoundTrip(req)
}

func newTransport() http.RoundTripper {
	tr := http.DefaultTransport
	tr.(*http.Transport).TLSClientConfig = &tls.Config{InsecureSkipVerify: true} // #nosec

	return &transport{tr: tr}
}

func newClient() *http.Client {
	jar, _ := cookiejar.New(nil)

	client := &http.Client{
		Transport: newTransport(),
		Jar:       jar,
		Timeout:   clientTimeoutSec * time.Second,
	}

	return client
}

func peer() (url string) {
	switch os.Getenv("NETWORK") {
	case "testnet":
		url = "https://testnet.umi.top"
	default:
		url = "https://mainnet.umi.top"
	}

	if val, ok := os.LookupEnv("PEER"); ok {
		url = val
	}

	return fmt.Sprintf("%s/json-rpc", url)
}
