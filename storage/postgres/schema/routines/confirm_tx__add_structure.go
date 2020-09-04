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

// ConfirmTxAddStructure ...
const ConfirmTxAddStructure = `
create or replace function confirm_tx__add_structure(bytes bytea,
                                          tx_height integer,
                                          blk_height integer,
                                          blk_tx_idx integer,
                                          blk_time timestamptz)
    returns void
    language plpgsql
as
$$
declare
    tx_value      constant bigint := 5000000;
    tx_hash       bytea;
    tx_ver        smallint;
    tx_sender     bytea;
    tx_struct     jsonb := '{}'::jsonb;
    st_version    integer;
    st_prefix     text;
    st_name       text;
    st_fee_prc    smallint;
    st_profit_prc smallint;
    st_profit_adr bytea;
    st_dev_adr    bytea;
begin
    select hash, version, sender, prefix, name, profit_percent, fee_percent
    into tx_hash, tx_ver, tx_sender, st_prefix, st_name, st_profit_prc, st_fee_prc
    from parse_transaction(bytes);

    st_profit_adr := substr(bytes, 36, 2) || substr(bytes, 4, 32);
    st_dev_adr := substr(bytes, 36, 2) || substr(get_dev_address(), 3, 32);

    st_version := convert_prefix_to_version(st_prefix);
    tx_struct := jsonb_set(tx_struct, '{prefix}', to_jsonb(st_prefix));
    tx_struct := jsonb_set(tx_struct, '{name}', to_jsonb(st_name));
    tx_struct := jsonb_set(tx_struct, '{profit_percent}', to_jsonb(st_profit_prc));
    tx_struct := jsonb_set(tx_struct, '{fee_percent}', to_jsonb(st_fee_prc));

    -- списываем 50К
    perform upd_address_balance(tx_sender, -tx_value, confirm_tx__add_structure.blk_time, confirm_tx__add_structure.tx_height, 'создание структуры ' || st_prefix );

    --

	-- лочим dev адрес
    insert into structure_address (version, prefix, address, type, created_at, created_tx_height)
    values (st_version, st_prefix, st_dev_adr, 'dev'::address_type, confirm_tx__add_structure.blk_time, confirm_tx__add_structure.tx_height);
    -- создаем баланс dev-кошелька и фиксируем тип
    perform upd_address_balance(st_dev_adr, 0::bigint, confirm_tx__add_structure.blk_time, confirm_tx__add_structure.tx_height, 'создание структуры ' || st_prefix, 'dev'::address_type);

    --

    -- мастер-адрес по умолчанию является профит-адресом
    insert into structure_address (version, prefix, address, type, created_at, created_tx_height)
    values (st_version, st_prefix, st_profit_adr, 'profit'::address_type, confirm_tx__add_structure.blk_time, confirm_tx__add_structure.tx_height);
    -- создаем баланс profit-кошелька и фиксируем тип
    perform upd_address_balance(st_profit_adr, 0::bigint, confirm_tx__add_structure.blk_time, confirm_tx__add_structure.tx_height, 'создание структуры ' || st_prefix, 'profit'::address_type);

    --

    -- проценты по умолчанию
    insert into structure_percent (version, prefix, level, percent, dev_percent, profit_percent, deposit_percent, block_height, updated_at)
    values (st_version, st_prefix, 0::smallint, 0::smallint, 0::smallint, 0::smallint, 0::smallint, confirm_tx__add_structure.blk_height, confirm_tx__add_structure.blk_time);

	insert into structure_percent_log (version, prefix, level, percent, dev_percent, profit_percent, deposit_percent, block_height, updated_at, comment)
	values (st_version, st_prefix, 0::smallint, 0::smallint, 0::smallint, 0::smallint, 0::smallint, confirm_tx__add_structure.blk_height, confirm_tx__add_structure.blk_time, 'создание структуры ' || st_prefix);

    --

    -- выставляем настройки. на профит-адрес (мастер по умолчанию) переводится комиссия
    insert into structure_settings (version, prefix, name, profit_percent, fee_percent, dev_address, profit_address, master_address, fee_address, created_at, tx_height, updated_at)
    values (st_version, st_prefix, st_name, st_profit_prc, st_fee_prc, st_dev_adr, st_profit_adr, tx_sender, st_profit_adr, confirm_tx__add_structure.blk_time, confirm_tx__add_structure.tx_height, confirm_tx__add_structure.blk_time);    
	--
    insert into structure_settings_log (version, prefix, name, profit_percent, fee_percent, dev_address, master_address, profit_address, fee_address, created_at, tx_height, comment)
    values (st_version, st_prefix, st_name, st_profit_prc, st_fee_prc, st_dev_adr, tx_sender, st_profit_adr, st_profit_adr, confirm_tx__add_structure.blk_time, confirm_tx__add_structure.tx_height, 'создание структуры ' || st_prefix);

    --
    perform upd_structure_balance(st_version, 0::bigint, confirm_tx__add_structure.blk_time, confirm_tx__add_structure.tx_height, 'создание структуры ' || st_prefix);

    --

    -- добавляем транзакцию
    insert into transaction (hash, height, confirmed_at, block_height, block_tx_idx, version, sender, value, struct)
    values (tx_hash, tx_height, blk_time, blk_height, blk_tx_idx, tx_ver, tx_sender, tx_value, tx_struct);
end
$$;
`
