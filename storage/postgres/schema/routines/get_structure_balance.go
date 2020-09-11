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

// GetStructureBalance ...
const GetStructureBalance = `
create or replace function get_structure_balance(version integer,
                                                 epoch timestamptz default now()::timestamptz(0),
                                                 out value bigint,
                                                 out percent smallint)
    language plpgsql
as
$$
declare
    rec         record;
    --
    periods   constant double precision := 30 * 24 * 60 * 60; -- Число периодов наращения в месяц
    nominal_interest   double precision; -- Номинальная месячная процентная ставка выражается в виде десятичной дроби.
    effective_interest double precision; -- Эффективная месячная процентная ставка выражается в виде десятичной дроби.
    period             double precision; -- Число периодов наращения для которых нужно получить сумму
    --
    lst_value   bigint;
    lst_percent smallint;
    lst_time    timestamptz;
begin
    epoch := epoch::timestamptz(0);
    value := 0::bigint;
	percent := 0::smallint;
    --
    select b.value, b.percent, b.updated_at
    into lst_value, lst_percent, lst_time
    from structure_balance b
    where b.version = get_structure_balance.version
    limit 1;

	if lst_value is null
    then
        raise debug 'Структура не найдена';
		return;
	end if;

    if lst_time > epoch
    then
		raise debug 'Запрошенная временная метка [%] меньше чем последнее обновление [%]', epoch, lst_time;
		--
		select b.value, b.percent, b.updated_at
		into lst_value, lst_percent, lst_time
		from structure_balance_log b
		where b.version = get_structure_balance.version
		  and updated_at <= epoch
		order by updated_at desc
		limit 1;
		--
	end if;

    if lst_time is null
    then
        raise debug 'На момент времени [%] структуры не существует', epoch;
        return;
	end if;

    for rec in
        select l.*
        from structure_percent_log l
        where l.version = get_structure_balance.version
          and l.updated_at between lst_time and epoch
        loop
			raise debug '% - % [%] - изменение процентов',
			    lst_time, (lst_value::double precision/100), (lst_percent::double precision/100);
			--
        	effective_interest := (lst_percent::double precision / 10000::double precision);
        	nominal_interest := periods * (
        	    (1::double precision + effective_interest) ^ (1::double precision / periods) - 1::double precision);
            period := extract(epoch from (rec.updated_at - lst_time));
            lst_value := floor(lst_value::double precision *
                               (1::double precision + (nominal_interest / periods)) ^ period)::bigint;
            --
            lst_percent := rec.deposit_percent;
            lst_time := rec.updated_at;
        end loop;
    --
    effective_interest := (lst_percent::double precision / 10000::double precision);
	nominal_interest := periods * (
	    (1::double precision + effective_interest) ^ (1::double precision / periods) - 1::double precision);
	period := extract(epoch from (epoch - lst_time));
	lst_value := floor(lst_value::double precision *
	                   (1::double precision + (nominal_interest / periods)) ^ period)::bigint;
    --
    raise debug '% - % [%] - финальное значение',
        lst_time, (lst_value::double precision/100), (lst_percent::double precision/100);
    --
    value := lst_value;
    percent := lst_percent;
end
$$;
`
