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

package legacy

import (
	"crypto/tls"
	"fmt"
	"net/http"
	"net/http/cookiejar"
)

type transport struct {
	tr http.RoundTripper
}

func (t *transport) RoundTrip(req *http.Request) (*http.Response, error) {
	req.Header.Set("User-Agent", "umid/0.0.1")
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Connection", "keep-alive")

	resp, err := t.tr.RoundTrip(req)
	if err != nil {
		err = fmt.Errorf("%w", err)
	}

	return resp, err
}

func newTransport() http.RoundTripper {
	tr := http.DefaultTransport
	tr.(*http.Transport).TLSClientConfig = &tls.Config{InsecureSkipVerify: true} // #nosec

	return &transport{tr: tr}
}

func newClient() *http.Client {
	client := &http.Client{
		Transport: newTransport(),
	}

	client.Jar, _ = cookiejar.New(nil)

	return client
}
