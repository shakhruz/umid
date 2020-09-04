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

package routines

// ConfirmTxGenesis ...
const ConfirmTxGenesis = `
create or replace function confirm_tx__genesis(bytes bytea,
                                    tx_height integer,
                                    blk_height integer,
                                    blk_tx_idx integer,
                                    blk_time timestamptz)
    returns void
    language plpgsql
as
$$
declare
    tx_hash   bytea;
    tx_ver    smallint;
    tx_sender bytea;
    tx_recip  bytea;
    tx_value  bigint;
begin
    select hash, version, sender, recipient, value
    into tx_hash, tx_ver, tx_sender, tx_recip, tx_value
    from parse_transaction(bytes);

    perform upd_address_balance(tx_recip, tx_value, blk_time, tx_height, 'genesis', 'umi');

    -- добавляем транзакцию
    insert into transaction (hash, height, confirmed_at, block_height, block_tx_idx, version, sender, recipient, value)
    values (tx_hash, tx_height, blk_time, blk_height, blk_tx_idx, tx_ver, tx_sender, tx_recip, tx_value);
end
$$;
`
