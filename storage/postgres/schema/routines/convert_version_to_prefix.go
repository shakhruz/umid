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

// ConvertVersionToPrefix ...
const ConvertVersionToPrefix = `
create or replace function convert_version_to_prefix(version integer)
    returns text
    language plpgsql
    immutable
as
$$
declare
    prefix bytea := E'\\x000000'::bytea;
begin
    if version = 0 then
        return 'genesis';
    end if;

    prefix := set_byte(prefix, 0, (((version & x'7C00'::integer) >> 10) + 96));
    prefix := set_byte(prefix, 1, (((version & x'03E0'::integer) >> 5) + 96));
    prefix := set_byte(prefix, 2, ((version & x'001F'::integer) + 96));

    return convert_from(prefix, 'LATIN1');
end
$$;
`
