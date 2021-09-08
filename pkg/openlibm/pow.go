// Copyright (c) 2021 UMI
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

// @see https://github.com/JuliaMath/openlibm/blob/master/src/e_pow.c
// ====================================================
// Copyright (C) 2004 by Sun Microsystems, Inc. All rights reserved.
//
// Permission to use, copy, modify, and distribute this
// software is freely granted, provided that this notice
// is preserved.
// ====================================================

package openlibm

import "math"

const (
	zero  = 0.0
	one   = 1.0
	two   = 2.0
	two53 = 9007199254740992.0 /* 0x43400000, 0x00000000 */

	/* poly coefs for (3/2)*(log(x)-2s-2/3*s**3 */

	L1    = 5.99999999999994648725e-01  /* 0x3FE33333, 0x33333303 */
	L2    = 4.28571428578550184252e-01  /* 0x3FDB6DB6, 0xDB6FABFF */
	L3    = 3.33333329818377432918e-01  /* 0x3FD55555, 0x518F264D */
	L4    = 2.72728123808534006489e-01  /* 0x3FD17460, 0xA91D4101 */
	L5    = 2.30660745775561754067e-01  /* 0x3FCD864A, 0x93C9DB65 */
	L6    = 2.06975017800338417784e-01  /* 0x3FCA7E28, 0x4A454EEF */
	P1    = 1.66666666666666019037e-01  /* 0x3FC55555, 0x5555553E */
	P2    = -2.77777777770155933842e-03 /* 0xBF66C16C, 0x16BEBD93 */
	P3    = 6.61375632143793436117e-05  /* 0x3F11566A, 0xAF25DE2C */
	P4    = -1.65339022054652515390e-06 /* 0xBEBBBD41, 0xC5D26BF1 */
	P5    = 4.13813679705723846039e-08  /* 0x3E663769, 0x72BEA4D0 */
	lg2   = 6.93147180559945286227e-01  /* 0x3FE62E42, 0xFEFA39EF */
	lg2_h = 6.93147182464599609375e-01  /* 0x3FE62E43, 0x00000000 */
	lg2_l = -1.90465429995776804525e-09 /* 0xBE205C61, 0x0CA86C39 */
	cp    = 9.61796693925975554329e-01  /* 0x3FEEC709, 0xDC3A03FD =2/(3ln2) */
	cp_h  = 9.61796700954437255859e-01  /* 0x3FEEC709, 0xE0000000 =(float)cp */
	cp_l  = -7.02846165095275826516e-09 /* 0xBE3E2FE0, 0x145B01F5 =tail of cp_h*/
)

func Pow(x float64, y float64) float64 {
	var bp = []float64{1.0, 1.5}
	var dp_h = []float64{0.0, 5.84962487220764160156e-01} /* 0x3FE2B803, 0x40000000 */
	var dp_l = []float64{0.0, 1.35003920212974897128e-08} /* 0x3E4CFDEB, 0x43CFD006 */

	var z, ax, z_h, z_l, p_h, p_l float64
	var y1, t1, t2, r, t, u, v, w float64
	var i, j, k, n int32
	var hx, hy, ix, iy int32
	var lx, ly uint32

	EXTRACT_WORDS(&hx, &lx, x)
	EXTRACT_WORDS(&hy, &ly, y)
	ix = hx & 0x7fffffff
	iy = hy & 0x7fffffff

	/* y==zero: x**0 = 1 */
	if (uint32(iy) | ly) == 0 {
		return one
	}

	/* x==1: 1**y = 1, even if y is NaN */
	if hx == 0x3ff00000 && lx == 0 {
		return one
	}

	/* y!=zero: result is NaN if either arg is NaN */
	if ix > 0x7ff00000 || ((ix == 0x7ff00000) && (lx != 0)) || iy > 0x7ff00000 || ((iy == 0x7ff00000) && (ly != 0)) {
		return (x + 0.0) + (y + 0.0)
	}

	/* special value of y */
	if ly == 0 {
		if iy == 0x3ff00000 { /* y is  1 */
			return x
		}

		if hy == 0x40000000 { /* y is  2 */
			return x * x
		}

		if hy == 0x40080000 { /* y is  3 */
			return x * x * x
		}

		if hy == 0x40100000 { /* y is  4 */
			u = x * x
			return u * u
		}
	}

	ax = fabs(x)

	var ss, s2, s_h, s_l, t_h, t_l float64

	n = 0
	/* take care subnormal number */
	if ix < 0x00100000 {
		ax *= two53
		n -= 53
		GET_HIGH_WORD(&ix, ax)
	}

	n += ((ix) >> 20) - 0x3ff
	j = ix & 0x000fffff
	/* determine interval */
	ix = j | 0x3ff00000 /* normalize ix */
	if j <= 0x3988E {   /* |x|<sqrt(3/2) */
		k = 0
	} else {
		if j < 0xBB67A { /* |x|<sqrt(3)   */
			k = 1
		} else {
			k = 0
			n += 1
			ix -= 0x00100000
		}
	}

	SET_HIGH_WORD(&ax, ix)

	/* compute ss = s_h+s_l = (x-1)/(x+1) or (x-1.5)/(x+1.5) */
	u = ax - bp[k] /* bp[0]=1.0, bp[1]=1.5 */
	v = one / (ax + bp[k])
	ss = u * v
	s_h = ss
	SET_LOW_WORD(&s_h, 0)
	/* t_h=ax+bp[k] High */
	t_h = zero
	SET_HIGH_WORD(&t_h, ((ix>>1)|0x20000000)+0x00080000+(k<<18))
	t_l = ax - (t_h - bp[k])
	s_l = v * ((u - s_h*t_h) - s_h*t_l)
	/* compute log(ax) */
	s2 = ss * ss
	r = s2 * s2 * (L1 + s2*(L2+s2*(L3+s2*(L4+s2*(L5+s2*L6)))))
	r += s_l * (s_h + ss)
	s2 = s_h * s_h
	t_h = 3.0 + s2 + r
	SET_LOW_WORD(&t_h, 0)
	t_l = r - ((t_h - 3.0) - s2)
	/* u+v = ss*(1+...) */
	u = s_h * t_h
	v = s_l*t_h + t_l*ss
	/* 2/(3log2)*(ss+...) */
	p_h = u + v
	SET_LOW_WORD(&p_h, 0)
	p_l = v - (p_h - u)
	z_h = cp_h * p_h /* cp_h+cp_l = 2/(3*log2) */
	z_l = cp_l*p_h + p_l*cp + dp_l[k]
	/* log2(ax) = (ss+..)*2/(3*log2) = n + dp_h + z_h + z_l */
	t = float64(n)
	t1 = ((z_h + z_l) + dp_h[k]) + t
	SET_LOW_WORD(&t1, 0)
	t2 = z_l - (((t1 - t) - dp_h[k]) - z_h)

	/* split up y into y1+y2 and compute (y1+y2)*(t1+t2) */
	y1 = y
	SET_LOW_WORD(&y1, 0)
	p_l = (y-y1)*t1 + y*t2
	p_h = y1 * t1
	z = p_l + p_h
	EXTRACT_WORDS2(&j, &i, z)

	/*
	 * compute 2**(p_h+p_l)
	 */
	i = j & 0x7fffffff
	k = (i >> 20) - 0x3ff
	n = 0

	if i > 0x3fe00000 { /* if |z| > 0.5, set n = [z+0.5] */
		n = j + (0x00100000 >> (k + 1))
		k = ((n & 0x7fffffff) >> 20) - 0x3ff /* new k for n */
		t = zero
		SET_HIGH_WORD(&t, n & ^(0x000fffff>>k))
		n = ((n & 0x000fffff) | 0x00100000) >> (20 - k)
		if j < 0 {
			n = -n
		}
		p_h -= t
	}

	t = p_l + p_h
	SET_LOW_WORD(&t, 0)
	u = t * lg2_h
	v = (p_l-(t-p_h))*lg2 + t*lg2_l
	z = u + v
	w = v - (z - u)
	t = z * z
	t1 = z - t*(P1+t*(P2+t*(P3+t*(P4+t*P5))))
	r = (z*t1)/(t1-two) - (w + z*w)
	z = one - (r - z)
	GET_HIGH_WORD(&j, z)
	j += n << 20
	SET_HIGH_WORD(&z, j)

	return z
}

func fabs(x float64) float64 {
	var high int32
	GET_HIGH_WORD(&high, x)
	SET_HIGH_WORD(&x, high&0x7fffffff)
	return x
}

func EXTRACT_WORDS(h *int32, l *uint32, c float64) {
	c64 := math.Float64bits(c)
	*h = int32(c64 >> 32)
	*l = uint32(c64 & 0xFFFFFFFF)
}

func EXTRACT_WORDS2(h *int32, l *int32, c float64) {
	c64 := math.Float64bits(c)
	*h = int32(c64 >> 32)
	*l = int32(c64 & 0xFFFFFFFF)
}

func GET_HIGH_WORD(i *int32, d float64) {
	v := math.Float64bits(d)
	*i = int32(v >> 32)
}

func SET_HIGH_WORD(d *float64, v int32) {
	t := math.Float64bits(*d)
	m := (uint64(v) << 32) | (t & 0xffffffff)
	*d = math.Float64frombits(m)
}

func SET_LOW_WORD(d *float64, v uint32) {
	t := math.Float64bits(*d)
	m := (t & 0xffffffff00000000) | uint64(v)
	*d = math.Float64frombits(m)
}
