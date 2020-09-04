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

// ConfirmNextBlock ...
const ConfirmNextBlock = `
create or replace function confirm_next_block()
    returns integer
    language plpgsql
as
$$
declare
    genesis         constant smallint := 0;
    basic           constant smallint := 1;
    add_struct      constant smallint := 2;
    upd_struct      constant smallint := 3;
    upd_profit_adr  constant smallint := 4;
    upd_fee_adr     constant smallint := 5;
    add_transit_adr constant smallint := 6;
    del_transit_adr constant smallint := 7;
    --
    hdr_length      constant integer  := 167;
    trx_length      constant integer  := 150;
    --
    blk_bytes                bytea;
    blk_height               integer;
    blk_time                 timestamptz;
    blk_tx_cnt               integer;
    --
    tx_height                integer;
    tx_bytes                 bytea;
begin
    select height, tx_count, created_at
    into blk_height, blk_tx_cnt, blk_time
    from block
    where synced is true
      and confirmed is false
    order by height
    limit 1;

    if blk_height is null then -- все блоки уже подтверждены
        return null;
    end if;

    blk_bytes := lo_get(blk_height);
    blk_bytes := substr(blk_bytes, hdr_length + 1);

    tx_height := setval('tx_height', nextval('tx_height'), false); -- высота последней подтвержденной транзакции

    for blk_tx_idx in 0..(blk_tx_cnt - 1)
        loop
            tx_height := tx_height + 1;
            tx_bytes := substr(blk_bytes, (1 + (blk_tx_idx * trx_length)), trx_length);

            case get_byte(tx_bytes, 0)
                when basic
                    then perform confirm_tx__basic(tx_bytes, tx_height, blk_height, blk_tx_idx, blk_time);
                when add_transit_adr
                    then perform confirm_tx__add_transit_address(tx_bytes, tx_height, blk_height, blk_tx_idx, blk_time);
                when del_transit_adr
                    then perform confirm_tx__del_transit_address(tx_bytes, tx_height, blk_height, blk_tx_idx, blk_time);
                when upd_profit_adr
                    then perform confirm_tx__upd_profit_address(tx_bytes, tx_height, blk_height, blk_tx_idx, blk_time);
                when upd_fee_adr
                    then perform confirm_tx__upd_fee_address(tx_bytes, tx_height, blk_height, blk_tx_idx, blk_time);
                when upd_struct
                    then perform confirm_tx__upd_structure(tx_bytes, tx_height, blk_height, blk_tx_idx, blk_time);
                when add_struct
                    then perform confirm_tx__add_structure(tx_bytes, tx_height, blk_height, blk_tx_idx, blk_time);
                when genesis
                    then perform confirm_tx__genesis(tx_bytes, tx_height, blk_height, blk_tx_idx, blk_time);
                else raise exception 'unknown transaction version';
                end case;
        end loop;

	perform upd_structure_level(blk_height, blk_time);

    update block set confirmed = true where height = blk_height;

    perform setval('tx_height', tx_height, false);

    return blk_height;
exception when others then
    perform lo_unlink(height) from block where confirmed is false;
    delete from block where confirmed is false;
    return null;
end
$$;
`
