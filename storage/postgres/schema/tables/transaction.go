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

package tables

// Transaction ...
const Transaction = `
create table if not exists transaction
(
    hash         bytea       not null
        constraint transaction_pk
            primary key,
	height       integer,
	confirmed_at timestamptz not null,
    block_height integer,
    block_tx_idx integer,
    version      smallint    not null,
    sender       bytea       not null,
    recipient    bytea,
    value        bigint,
    fee_address  bytea,
    fee_value    bigint,
    struct       jsonb
);
`

// TransactionIdxSender ...
const TransactionIdxSender = `
create index if not exists transaction_idx_sender
    on transaction (sender, height desc);
`

// TransactionIdxRecipient ...
const TransactionIdxRecipient = `
create index if not exists transaction_idx_recipient
    on transaction (recipient, height desc)
    where recipient is not null;
`

// TransactionIdxFeeAddress ...
const TransactionIdxFeeAddress = `
create index if not exists transaction_idx_fee_address
    on transaction (fee_address, height desc)
    where fee_address is not null;
`
