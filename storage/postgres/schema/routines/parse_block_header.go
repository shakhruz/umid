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

// ParseBlockHeader ...
const ParseBlockHeader = `
create or replace function parse_block_header(bytes bytea,
                                              out hash bytea,
                                              out version smallint,
                                              out prev_block_hash bytea,
                                              out merkle_root_hash bytea,
                                              out created_at timestamptz,
                                              out public_key bytea,
                                              out signature bytea,
                                              out tx_count integer)
    language plpgsql
as
$$
declare
    hdr_length constant  smallint := 167;
    ver_genesis constant integer := 0;
    blk_unixtime         integer;
begin
    if length(bytes) != hdr_length then
        raise exception 'block header size should be % bytes', hdr_length;
    end if;

    blk_unixtime := (get_byte(bytes, 65) << 24) + (get_byte(bytes, 66) << 16) +
                    (get_byte(bytes, 67) << 8) + get_byte(bytes, 68);

    hash := sha256(bytes);
    version := get_byte(bytes, 0);
    if version != ver_genesis then
        prev_block_hash := substr(bytes, 2, 32);
    end if;
    merkle_root_hash := substr(bytes, 34, 32);
    created_at := to_timestamp(blk_unixtime);
    public_key := substr(bytes, 72, 32);
    signature := substr(bytes, 104, 64);
    tx_count := (get_byte(bytes, 69) << 8) + get_byte(bytes, 70);
end
$$;
`
