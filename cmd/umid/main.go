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

package main

import (
	"context"
	"errors"
	"fmt"
	"log"
	"net/http"
	"os"
	"time"

	"gitlab.com/umitop/umid/pkg/config"
	"gitlab.com/umitop/umid/pkg/events"
	"gitlab.com/umitop/umid/pkg/generator"
	"gitlab.com/umitop/umid/pkg/ledger"
	"gitlab.com/umitop/umid/pkg/legacy"
	"gitlab.com/umitop/umid/pkg/restapi"
	"gitlab.com/umitop/umid/pkg/storage"
	"gitlab.com/umitop/umid/pkg/syncer"
)

var ErrStorage = errors.New("storage")

//nolint:funlen // ...
//revive:disable:function-length
func main() {
	log.SetFlags(log.LstdFlags /*| log.Lshortfile*/)

	ctx := context.Background()

	conf := config.DefaultConfig()
	conf.Parse()

	blockchain, err := initBlockchain(conf)
	if err != nil {
		log.Fatalf("%v", err)
	}
	defer blockchain.Close()

	ledger1 := ledger.NewLedger(conf)
	confirmer := ledger.NewConfirmerLegacy(ledger1)
	confirmer.SetBlockchain(blockchain)

	index := storage.NewIndex()
	index.SubscribeTo(blockchain)

	go index.Worker(ctx)

	mempool := storage.NewMempool()
	mempool.SetLedger(ledger1)

	event := events.NewEvents()

	go func() {
		currentTime := time.Now()

		log.Println("scanning blockchain...")

		if err := blockchain.Scan(confirmer); err != nil {
			log.Fatal(err)
		}

		log.Printf("found %d blocks, time: %v.", blockchain.Height(), time.Since(currentTime))

		if blockchain.Height() == 0 {
			if err := confirmer.AppendBlock(storage.GenesisBlock(conf.Network)); err != nil {
				log.Fatal(err)
			}
		}

		if _, ok := os.LookupEnv("UMI_MASTER_KEY"); ok {
			go generator.NewGenerator(confirmer, mempool).Worker(ctx)
		} else {
			fetcher := legacy.NewFetcher(conf)
			fetcher.SetConfirmer(confirmer)

			go fetcher.Worker(ctx)

			pusher := legacy.NewPusher(conf, mempool)

			go pusher.Worker(ctx)

			event.SubscribeTo2(mempool)
			event.SubscribeTo(blockchain)

			go event.Worker(ctx)
		}

		mempool.SubscribeTo(blockchain)

		go mempool.Worker(ctx)
	}()

	api := restapi.NewRestAPI()
	api.SetBlockchain(blockchain)
	api.SetLedger(ledger1)
	api.SetIndex(index)
	api.SetMempool(mempool)
	api.SetEvents(event)

	syncr := syncer.NewSyncer()
	syncr.SetMempool(mempool)

	mux := http.NewServeMux()
	mux.HandleFunc("/api/", api.Router)
	mux.HandleFunc("/events/", api.Router)
	mux.HandleFunc("/sync/", syncr.Router)
	mux.HandleFunc("/healthz", api.Healthz)
	mux.HandleFunc("/status", api.Status)

	if err := http.ListenAndServe(conf.ListenAddress, mux); err != nil {
		log.Println(err)
	}
}

func initBlockchain(conf *config.Config) (blockchain storage.IBlockchain, err error) {
	switch conf.StorageType {
	case "memory":
		blockchain = storage.NewBlockchainMemory(conf)

	case "file":
		blockchain = storage.NewBlockchain(conf)
		if err = blockchain.OpenOrCreate(); err != nil {
			return nil, fmt.Errorf("%w", err)
		}

	default:
		return nil, fmt.Errorf("%w: unknown storage type", ErrStorage)
	}

	return blockchain, nil
}
