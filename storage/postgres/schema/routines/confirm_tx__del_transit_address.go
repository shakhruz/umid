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

// ConfirmTxDelTransitAddress ...
const ConfirmTxDelTransitAddress = `
create or replace function confirm_tx__del_transit_address(bytes bytea,
                                                           tx_height integer,
														   blk_height integer,
														   blk_tx_idx integer,
														   blk_time timestamptz)
    returns void
    language plpgsql
as
$$
declare
    tx_hash      bytea;
    tx_version   smallint;
    tx_sender    bytea;
    tx_recipient bytea;
    tx_struct    jsonb := '{}'::jsonb;
    st_version   integer;
    st_prefix    text;
begin
    select hash, version, sender, recipient, prefix
    into tx_hash, tx_version, tx_sender, tx_recipient, st_prefix
    from parse_transaction(bytes);

    st_version := convert_prefix_to_version(st_prefix);
    tx_struct := jsonb_set(tx_struct, '{prefix}', to_jsonb(st_prefix));

    insert into transaction (hash, height, confirmed_at, block_height, block_tx_idx, version, sender, recipient, struct)
    values (tx_hash, tx_height, blk_time, blk_height, blk_tx_idx, tx_version, tx_sender, tx_recipient, tx_struct);

    update structure_address
    set deleted_at = blk_time,
        deleted_tx_height = tx_height
    where version = st_version
      and type = 'transit'
      and deleted_at is null;

    perform upd_address_balance(tx_recipient, 0::bigint, blk_time, tx_height, 'деактивация транзитного адреса', 'deposit'::address_type);
end
$$;
`
