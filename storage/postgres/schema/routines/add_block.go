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

// AddBlock ...
const AddBlock = `
create or replace function add_block(bytes bytea)
    returns integer
    language plpgsql
as
$$
declare
    ver_genesis constant integer := 0;
    --
    blk_height           integer;
    blk_synced           boolean;
    --
    blk_hash             bytea;
    blk_version          smallint;
    blk_prv_hash         bytea;
    blk_merkle           bytea;
    blk_time             timestamptz;
    blk_tx_cnt           integer;
    blk_pubkey           bytea;
    --
    lst_blk_hash         bytea;
    lst_blk_height       integer;
begin
    select hash, version, prev_block_hash, merkle_root_hash, created_at, tx_count, public_key
    into blk_hash, blk_version, blk_prv_hash, blk_merkle, blk_time, blk_tx_cnt, blk_pubkey
    from parse_block_header(substr(bytes, 1, 167));

    select height, synced into blk_height, blk_synced from block where hash = blk_hash limit 1;

    if blk_synced is true then -- блок есть в цепочке и уже добавлен
        return blk_height;
    end if;

    if blk_synced is false then -- блок есть в цепочке, но еще не добавлен
        update block set synced = true where hash = blk_hash;
        --
        return lo_from_bytea(blk_height, bytes);
    end if;

    if blk_version = ver_genesis then
        blk_height = 1;
    else
        -- смотрим на последний добавленный блок
        select hash, height into lst_blk_hash, lst_blk_height from block order by height desc limit 1;

        if blk_prv_hash = lst_blk_hash then -- новый блок ссылается на последний блок в цепочке
            blk_height := lst_blk_height + 1;
        else
            -- не добавляем блок
            return null;
        end if;
    end if;

    insert into block(hash, height, version, prev_block_hash, merkle_root_hash, created_at, tx_count, public_key,synced)
    values (blk_hash, blk_height, blk_version, blk_prv_hash, blk_merkle, blk_time, blk_tx_cnt, blk_pubkey, true);

    return lo_from_bytea(blk_height, bytes);
end
$$;
`
