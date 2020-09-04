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

// GetAddressBalance ...
const GetAddressBalance = `
create or replace function get_address_balance(address bytea,
                                               epoch timestamptz default now()::timestamptz(0),
                                               composite boolean default true,
                                               out confirmed_value bigint,
                                               out confirmed_percent smallint,
                                               out unconfirmed_value bigint,
                                               out composite_value bigint,
                                               out type address_type)
    language plpgsql
as
$$
declare
    umi_basic   constant integer := x'55A9'::integer;
    adr_version integer := (get_byte(get_address_balance.address, 0) << 8) + get_byte(get_address_balance.address, 1);
    --
    rec                record;
    --
    periods   constant double precision := 30 * 24 * 60 * 60; -- Число периодов наращения в месяц
    nominal_interest   double precision; -- Номинальная месячная процентная ставка выражается в виде десятичной дроби. 6 т.д .:% = 0,06
    effective_interest double precision; -- Эффективная месячная процентная ставка выражается в виде десятичной дроби. 6 т.д .:% = 0,06
    period             double precision; -- Число периодов наращения для которых нужно получить сумму
    --
    lst_value          bigint;
    lst_percent        smallint;
    lst_time           timestamptz;
    lst_type           address_type;
    --
    address_ver        integer;
    --
    lock_balance       bigint;
    profit_addr        bytea;
begin
    raise debug '% %', address, epoch;
    --
    epoch := epoch::timestamptz(0);
    confirmed_value := 0::bigint;
    confirmed_percent := 0::smallint;
    unconfirmed_value := 0::bigint;
    --
    select b.value, b.percent, b.updated_at, b.version, b.type
    into lst_value, lst_percent, lst_time, address_ver, lst_type
    from address_balance_confirmed b
    where b.address = get_address_balance.address
    limit 1;
	--
    if lst_value is null
    then
        raise debug 'Баланса на запрошенную дату не существовало';
        -- при создании структуры мы принудительно создаем балансы для dev, profit и fee,
        -- поэтому нет смысла проверять составной баланс
		type := case adr_version when umi_basic then 'umi'::address_type else 'deposit'::address_type end;
        --
        -- при первом пополнении структурных кошелей нужно проверить актуальный процент структуры
        if type <> 'umi'::address_type
        then
			select deposit_percent into confirmed_percent
			from structure_percent_log
			where version = adr_version
			  and updated_at <= epoch order by updated_at desc limit 1;
			-- в блоке не может быть транзакций в структуру, которой не сущестует, но можно запрасить баланс на момент до создания структуры
			confirmed_percent := coalesce(confirmed_percent, 0::smallint);
		end if;
		--        
		return;
	end if;


    if lst_time > epoch -- смотрим в прошлое
    then
        raise debug 'Запрошенная дата [%] меньше чем последнее обновление баланса [%], смотрим в лог', epoch, lst_time;
        --
        select b.value, b.percent, b.updated_at, b.version, b.type
        into lst_value, lst_percent, lst_time, address_ver, lst_type
        from address_balance_confirmed_log b
        where b.address = get_address_balance.address
          and updated_at <= epoch
        order by updated_at desc
        limit 1;
    end if;


    if lst_time is null then
        raise debug 'Баланса на запрошенную дату не существовало';
        type := case adr_version when umi_basic then 'umi'::address_type else 'deposit'::address_type end;
		return;
	else
        raise debug '% - % [%] - первое найденное обновление баланса', lst_time, (lst_value::double precision / 100), (lst_percent::double precision / 100);
	end if;
    --
    if address_ver <> umi_basic
    then
        for rec in
            select l.*
            from structure_percent_log l
            where l.version = address_ver
              and l.updated_at between lst_time and epoch
            loop
            	effective_interest := (lst_percent::double precision / 10000::double precision);
            	nominal_interest := periods * ((1::double precision + effective_interest) ^ (1::double precision / periods) - 1::double precision);
                period := extract(epoch from (rec.updated_at - lst_time));
            	
            	raise debug 'ef %, nom %, per %', effective_interest, nominal_interest, period;
            	
                --
                lst_value := floor(lst_value::double precision * (1::double precision + (nominal_interest / periods)) ^ period)::bigint;
                lst_percent := case lst_type
                               when 'dev'::address_type then rec.dev_percent
                               when 'profit'::address_type then rec.profit_percent
                               when 'fee'::address_type then rec.profit_percent
                               else rec.deposit_percent end;
                lst_time := rec.updated_at;
                --
                raise debug '% - % [%] - изменение процента', lst_time, (lst_value::double precision / 100), (lst_percent::double precision / 100);
            end loop;
        --
        period := extract(epoch from (epoch - lst_time));
        effective_interest := (lst_percent::double precision / 10000::double precision);
        nominal_interest := periods * ((1::double precision + effective_interest) ^ (1::double precision / periods) - 1::double precision);
        
        raise debug 'ef %, nom %, per %', effective_interest, nominal_interest, period;
        
        lst_value := floor(lst_value::double precision * (1::double precision + (nominal_interest / periods)) ^ period)::bigint;
        --
        raise debug '% - % [%] - итоговое значение', epoch, (lst_value::double precision / 100), (lst_percent::double precision / 100);
        --
        if composite
        then
            if lst_type = 'profit'::address_type
            then
                select value into lock_balance from get_structure_balance(address_ver, epoch);
                composite_value := lst_value;
                lst_value := composite_value - lock_balance;
            elseif lst_type = 'dev'::address_type
            then
                select profit_address into profit_addr from structure_settings_log
                where version = adr_version and created_at <= epoch order by created_at desc limit 1;
                --
                select b.confirmed_value into lock_balance from get_address_balance(profit_addr, epoch, false) as b;
                --
                raise debug '% % %', epoch, lock_balance, profit_addr;
                composite_value := lst_value;
                lst_value := composite_value - lock_balance;
            end if;
        end if;
    end if;
    --
    confirmed_value := lst_value;
    unconfirmed_value := confirmed_value;
    confirmed_percent := lst_percent;
    type := lst_type;
end
$$;
`
