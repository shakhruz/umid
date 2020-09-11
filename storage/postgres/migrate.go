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

package postgres

import (
	"context"
	"log"
	"sync"
	"time"
	"umid/storage/postgres/schema"
	"umid/storage/postgres/schema/sequences"

	"github.com/jackc/pgx/v4"
	"github.com/jackc/pgx/v4/pgxpool"
)

// Migrate ...
func Migrate(ctx context.Context, wg *sync.WaitGroup, conn *pgxpool.Pool) {
	wg.Add(1)
	defer wg.Done()

	// retry if database is not available
	for {
		select {
		case <-ctx.Done():
			return
		default:
			if err := doMigrate(ctx, conn); err != nil {
				log.Println(err.Error())
				time.Sleep(time.Second)

				continue
			}

			return
		}
	}
}

func doMigrate(ctx context.Context, conn *pgxpool.Pool) error {
	tx, err := conn.Begin(ctx)
	if err != nil {
		return err
	}

	defer func() { _ = tx.Rollback(ctx) }()

	cur, err := currentVersion(ctx, tx)
	if err != nil {
		return err
	}

	if err = processMigrations(tx, cur); err != nil {
		return err
	}

	return tx.Commit(ctx)
}

func processMigrations(tx pgx.Tx, cur int) (err error) {
	for v, m := range schema.Migrations() {
		if v > cur {
			if err = runMigrations(tx, v, m); err != nil {
				return err
			}
		}
	}

	return err
}

func runMigrations(tx pgx.Tx, v int, m []string) error {
	ctx := context.Background()

	log.Printf("upgrade db v%d\n", v)

	for _, sql := range m {
		if _, err := tx.Exec(ctx, sql); err != nil {
			return err
		}
	}

	if _, err := tx.Exec(ctx, `select setval('db_version', $1, false)`, v); err != nil {
		return err
	}

	return nil
}

func currentVersion(ctx context.Context, tx pgx.Tx) (v int, err error) {
	if _, err = tx.Exec(ctx, sequences.DBVersion); err == nil {
		err = tx.QueryRow(ctx, `select setval('db_version', nextval('db_version'), false)`).Scan(&v)
	}

	return v, err
}
