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

// Level ...
const Level = `
create table if not exists level
(
	id        smallint
		constraint level_pk
			primary key,
	min_value bigint   not null,
    max_value bigint   not null,
	percent   smallint not null
);
`

// LevelData ...
const LevelData = `
insert into level (id, min_value, max_value, percent)
values (0, 0, 4999999, 0),
       (1, 5000000, 9999999, 1000),
       (2, 10000000, 49999999, 1500),
       (3, 50000000, 99999999, 2000),
       (4, 100000000, 499999999, 2500),
       (6, 500000000, 999999999, 3000),
       (7, 1000000000, 4999999999, 3500),
       (8, 5000000000, 9999999999, 3600),
       (9, 10000000000, 49999999999, 3700),
       (10, 50000000000, 99999999999, 3900),
       (11, 100000000000, 9223372036854775807, 4100)
on conflict on constraint level_pk
    do nothing;
`
