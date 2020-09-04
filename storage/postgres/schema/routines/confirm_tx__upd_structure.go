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

// ConfirmTxUpdStructure ...
const ConfirmTxUpdStructure = `
create or replace function confirm_tx__upd_structure(bytes bytea,
                                          tx_height_ integer,
                                          blk_height integer,
                                          blk_tx_idx integer,
                                          blk_timestamp timestamptz)
    returns void
    language plpgsql
as
$$
declare
    tx_hash       bytea;
    tx_version    smallint;
    tx_sender     bytea;
    tx_struct     jsonb := '{}'::jsonb;
    st_version    integer;
    st_prefix     text;
    st_name       text;
    st_profit_prc smallint;
    st_fee_prc    smallint;
    --
    upd_ver       integer;
    upd_level     smallint;
    upd_percent   smallint;
    --
    dev_adr       bytea;
    mst_adr       bytea;
    fee_adr       bytea;
    prof_adr      bytea;
begin
    select hash, version, sender, prefix, name, profit_percent, fee_percent
    into tx_hash, tx_version, tx_sender, st_prefix, st_name, st_profit_prc, st_fee_prc
    from parse_transaction(bytes);

    st_version := convert_prefix_to_version(st_prefix);
    tx_struct := jsonb_set(tx_struct, '{prefix}', to_jsonb(st_prefix));
    tx_struct := jsonb_set(tx_struct, '{name}', to_jsonb(st_name));
    tx_struct := jsonb_set(tx_struct, '{profit_percent}', to_jsonb(st_profit_prc));
    tx_struct := jsonb_set(tx_struct, '{fee_percent}', to_jsonb(st_fee_prc));

    insert into transaction (hash, height, confirmed_at, block_height, block_tx_idx, version, sender, struct)
    values (tx_hash, tx_height_, blk_timestamp, blk_height, blk_tx_idx, tx_version, tx_sender, tx_struct);

    update structure_settings
    set name           = st_name,
        profit_percent = st_profit_prc,
        fee_percent    = st_fee_prc,
        tx_height      = tx_height_,
        updated_at     = blk_timestamp
    where version = st_version
    returning dev_address, master_address, profit_address, fee_address
         into dev_adr, mst_adr, fee_adr, prof_adr;
    --
    insert into structure_settings_log (version, prefix, name, profit_percent, fee_percent, dev_address, master_address, profit_address, fee_address, created_at, tx_height, comment)
    values (st_version, st_prefix, st_name, st_profit_prc, st_fee_prc, dev_adr, mst_adr, prof_adr, fee_adr, blk_timestamp, tx_height_, 'update structure');
    --
    update structure_percent s
       set deposit_percent = s.percent - st_profit_prc
    where version = st_version
        and level <> 0
    returning version, level, percent into upd_ver, upd_level, upd_percent;
	--
	if upd_ver is not null
	then
		insert into structure_percent_log (version, prefix, level, percent, dev_percent, profit_percent, deposit_percent, block_height, updated_at, comment)
		values (
				st_version,
				st_prefix,
				upd_level, 
				upd_percent,  -- percent
				case upd_level when 0 then 0 else upd_percent + 200 end, -- dev_percent
				upd_percent,  -- profit_percent
				case upd_level when 0 then 0 else upd_percent - st_profit_prc end, -- deposit_percent
				confirm_tx__upd_structure.blk_height, 
				confirm_tx__upd_structure.blk_timestamp,
		        'structure update'
				);
	end if;
end
$$;
`
