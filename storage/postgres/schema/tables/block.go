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

// Block ...
const Block = `
create table if not exists block
(
    hash             bytea                 not null
        constraint block_pkey
            primary key,
    height           integer               not null,

    version          smallint              not null,
    prev_block_hash  bytea
        constraint block_fk
            references block,
    merkle_root_hash bytea                 not null,
    created_at       timestamptz           not null,

    tx_count         integer               not null,
    public_key       bytea                 not null,

    synced           boolean default false not null,
    confirmed        boolean default false not null,
    constraint block_none_genesis
        check ((height > 0) and ((height = 1) or (prev_block_hash is not null)))
);
`

// BlockHeightUidx ...
const BlockHeightUidx = `
create unique index if not exists block_height_uidx
    on block (height);
`

// BlockPrevUidx ...
const BlockPrevUidx = `
create unique index if not exists block_prev_uidx
    on block (prev_block_hash);
`

// BlockConfIdx ...
const BlockConfIdx = `
create index if not exists block_conf_idx
    on block (synced, confirmed)
    where ((synced is true) and (confirmed is false));
`
