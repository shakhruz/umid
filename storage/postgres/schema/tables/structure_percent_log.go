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

// StructurePercentLog ...
const StructurePercentLog = `
create table if not exists structure_percent_log
(
    version         integer     not null,
    prefix          char(3)     not null,
    level           smallint    not null,
    percent         smallint    not null,
	dev_percent     smallint    not null,
    profit_percent  smallint    not null,
    deposit_percent smallint    not null,
    block_height    integer     not null,
    updated_at      timestamptz not null,
    comment         text,
    check (percent between 0 and 4100 and level between 0 and 11)
);
`

// StructurePercentLogIdx ...
const StructurePercentLogIdx = `
create index if not exists structure_percent_log_idx
    on structure_percent_log (version);
`
