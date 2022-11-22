package mcrypto

// ScReduce32 is the same as scReduce, but it takes a 32 byte input that we widen first
func ScReduce32(s [32]byte) [32]byte {
	var reduced [32]byte
	var wideBytes [64]byte
	copy(wideBytes[:], s[:]) // LE so we are adding 32 zeros as high order bytes
	scReduce(&reduced, &wideBytes)
	return reduced
}

// Unfortunately, both the standard ed25519 library and filippo.io/edwards25519 do not
// export their scReduce function. They are both ported from the public domain, “ref10”
// implementation of ed25519 from SUPERCOP, so using them here shouldn't create any
// license incompatibility. This is the filippo.io/edwards25519 version:
//    https://github.com/FiloSottile/edwards25519/blob/v1.0.0/scalar.go#L618-L624

func load3(in []byte) int64 {
	r := int64(in[0])
	r |= int64(in[1]) << 8
	r |= int64(in[2]) << 16
	return r
}

func load4(in []byte) int64 {
	r := int64(in[0])
	r |= int64(in[1]) << 8
	r |= int64(in[2]) << 16
	r |= int64(in[3]) << 24
	return r
}

// Input:
//
//	s[0]+256*s[1]+...+256^63*s[63] = s
//
// Output:
//
//	s[0]+256*s[1]+...+256^31*s[31] = s mod l
//	where l = 2^252 + 27742317777372353535851937790883648493.
//
//nolint:ineffassign
func scReduce(out *[32]byte, s *[64]byte) {
	s0 := 2097151 & load3(s[:])
	s1 := 2097151 & (load4(s[2:]) >> 5)
	s2 := 2097151 & (load3(s[5:]) >> 2)
	s3 := 2097151 & (load4(s[7:]) >> 7)
	s4 := 2097151 & (load4(s[10:]) >> 4)
	s5 := 2097151 & (load3(s[13:]) >> 1)
	s6 := 2097151 & (load4(s[15:]) >> 6)
	s7 := 2097151 & (load3(s[18:]) >> 3)
	s8 := 2097151 & load3(s[21:])
	s9 := 2097151 & (load4(s[23:]) >> 5)
	s10 := 2097151 & (load3(s[26:]) >> 2)
	s11 := 2097151 & (load4(s[28:]) >> 7)
	s12 := 2097151 & (load4(s[31:]) >> 4)
	s13 := 2097151 & (load3(s[34:]) >> 1)
	s14 := 2097151 & (load4(s[36:]) >> 6)
	s15 := 2097151 & (load3(s[39:]) >> 3)
	s16 := 2097151 & load3(s[42:])
	s17 := 2097151 & (load4(s[44:]) >> 5)
	s18 := 2097151 & (load3(s[47:]) >> 2)
	s19 := 2097151 & (load4(s[49:]) >> 7)
	s20 := 2097151 & (load4(s[52:]) >> 4)
	s21 := 2097151 & (load3(s[55:]) >> 1)
	s22 := 2097151 & (load4(s[57:]) >> 6)
	s23 := load4(s[60:]) >> 3

	s11 += s23 * 666643
	s12 += s23 * 470296
	s13 += s23 * 654183
	s14 -= s23 * 997805
	s15 += s23 * 136657
	s16 -= s23 * 683901
	s23 = 0

	s10 += s22 * 666643
	s11 += s22 * 470296
	s12 += s22 * 654183
	s13 -= s22 * 997805
	s14 += s22 * 136657
	s15 -= s22 * 683901
	s22 = 0

	s9 += s21 * 666643
	s10 += s21 * 470296
	s11 += s21 * 654183
	s12 -= s21 * 997805
	s13 += s21 * 136657
	s14 -= s21 * 683901
	s21 = 0

	s8 += s20 * 666643
	s9 += s20 * 470296
	s10 += s20 * 654183
	s11 -= s20 * 997805
	s12 += s20 * 136657
	s13 -= s20 * 683901
	s20 = 0

	s7 += s19 * 666643
	s8 += s19 * 470296
	s9 += s19 * 654183
	s10 -= s19 * 997805
	s11 += s19 * 136657
	s12 -= s19 * 683901
	s19 = 0

	s6 += s18 * 666643
	s7 += s18 * 470296
	s8 += s18 * 654183
	s9 -= s18 * 997805
	s10 += s18 * 136657
	s11 -= s18 * 683901
	s18 = 0

	var carry [17]int64

	carry[6] = (s6 + (1 << 20)) >> 21
	s7 += carry[6]
	s6 -= carry[6] << 21
	carry[8] = (s8 + (1 << 20)) >> 21
	s9 += carry[8]
	s8 -= carry[8] << 21
	carry[10] = (s10 + (1 << 20)) >> 21
	s11 += carry[10]
	s10 -= carry[10] << 21
	carry[12] = (s12 + (1 << 20)) >> 21
	s13 += carry[12]
	s12 -= carry[12] << 21
	carry[14] = (s14 + (1 << 20)) >> 21
	s15 += carry[14]
	s14 -= carry[14] << 21
	carry[16] = (s16 + (1 << 20)) >> 21
	s17 += carry[16]
	s16 -= carry[16] << 21

	carry[7] = (s7 + (1 << 20)) >> 21
	s8 += carry[7]
	s7 -= carry[7] << 21
	carry[9] = (s9 + (1 << 20)) >> 21
	s10 += carry[9]
	s9 -= carry[9] << 21
	carry[11] = (s11 + (1 << 20)) >> 21
	s12 += carry[11]
	s11 -= carry[11] << 21
	carry[13] = (s13 + (1 << 20)) >> 21
	s14 += carry[13]
	s13 -= carry[13] << 21
	carry[15] = (s15 + (1 << 20)) >> 21
	s16 += carry[15]
	s15 -= carry[15] << 21

	s5 += s17 * 666643
	s6 += s17 * 470296
	s7 += s17 * 654183
	s8 -= s17 * 997805
	s9 += s17 * 136657
	s10 -= s17 * 683901
	s17 = 0

	s4 += s16 * 666643
	s5 += s16 * 470296
	s6 += s16 * 654183
	s7 -= s16 * 997805
	s8 += s16 * 136657
	s9 -= s16 * 683901
	s16 = 0

	s3 += s15 * 666643
	s4 += s15 * 470296
	s5 += s15 * 654183
	s6 -= s15 * 997805
	s7 += s15 * 136657
	s8 -= s15 * 683901
	s15 = 0

	s2 += s14 * 666643
	s3 += s14 * 470296
	s4 += s14 * 654183
	s5 -= s14 * 997805
	s6 += s14 * 136657
	s7 -= s14 * 683901
	s14 = 0

	s1 += s13 * 666643
	s2 += s13 * 470296
	s3 += s13 * 654183
	s4 -= s13 * 997805
	s5 += s13 * 136657
	s6 -= s13 * 683901
	s13 = 0

	s0 += s12 * 666643
	s1 += s12 * 470296
	s2 += s12 * 654183
	s3 -= s12 * 997805
	s4 += s12 * 136657
	s5 -= s12 * 683901
	s12 = 0

	carry[0] = (s0 + (1 << 20)) >> 21
	s1 += carry[0]
	s0 -= carry[0] << 21
	carry[2] = (s2 + (1 << 20)) >> 21
	s3 += carry[2]
	s2 -= carry[2] << 21
	carry[4] = (s4 + (1 << 20)) >> 21
	s5 += carry[4]
	s4 -= carry[4] << 21
	carry[6] = (s6 + (1 << 20)) >> 21
	s7 += carry[6]
	s6 -= carry[6] << 21
	carry[8] = (s8 + (1 << 20)) >> 21
	s9 += carry[8]
	s8 -= carry[8] << 21
	carry[10] = (s10 + (1 << 20)) >> 21
	s11 += carry[10]
	s10 -= carry[10] << 21

	carry[1] = (s1 + (1 << 20)) >> 21
	s2 += carry[1]
	s1 -= carry[1] << 21
	carry[3] = (s3 + (1 << 20)) >> 21
	s4 += carry[3]
	s3 -= carry[3] << 21
	carry[5] = (s5 + (1 << 20)) >> 21
	s6 += carry[5]
	s5 -= carry[5] << 21
	carry[7] = (s7 + (1 << 20)) >> 21
	s8 += carry[7]
	s7 -= carry[7] << 21
	carry[9] = (s9 + (1 << 20)) >> 21
	s10 += carry[9]
	s9 -= carry[9] << 21
	carry[11] = (s11 + (1 << 20)) >> 21
	s12 += carry[11]
	s11 -= carry[11] << 21

	s0 += s12 * 666643
	s1 += s12 * 470296
	s2 += s12 * 654183
	s3 -= s12 * 997805
	s4 += s12 * 136657
	s5 -= s12 * 683901
	s12 = 0

	carry[0] = s0 >> 21
	s1 += carry[0]
	s0 -= carry[0] << 21
	carry[1] = s1 >> 21
	s2 += carry[1]
	s1 -= carry[1] << 21
	carry[2] = s2 >> 21
	s3 += carry[2]
	s2 -= carry[2] << 21
	carry[3] = s3 >> 21
	s4 += carry[3]
	s3 -= carry[3] << 21
	carry[4] = s4 >> 21
	s5 += carry[4]
	s4 -= carry[4] << 21
	carry[5] = s5 >> 21
	s6 += carry[5]
	s5 -= carry[5] << 21
	carry[6] = s6 >> 21
	s7 += carry[6]
	s6 -= carry[6] << 21
	carry[7] = s7 >> 21
	s8 += carry[7]
	s7 -= carry[7] << 21
	carry[8] = s8 >> 21
	s9 += carry[8]
	s8 -= carry[8] << 21
	carry[9] = s9 >> 21
	s10 += carry[9]
	s9 -= carry[9] << 21
	carry[10] = s10 >> 21
	s11 += carry[10]
	s10 -= carry[10] << 21
	carry[11] = s11 >> 21
	s12 += carry[11]
	s11 -= carry[11] << 21

	s0 += s12 * 666643
	s1 += s12 * 470296
	s2 += s12 * 654183
	s3 -= s12 * 997805
	s4 += s12 * 136657
	s5 -= s12 * 683901
	s12 = 0

	carry[0] = s0 >> 21
	s1 += carry[0]
	s0 -= carry[0] << 21
	carry[1] = s1 >> 21
	s2 += carry[1]
	s1 -= carry[1] << 21
	carry[2] = s2 >> 21
	s3 += carry[2]
	s2 -= carry[2] << 21
	carry[3] = s3 >> 21
	s4 += carry[3]
	s3 -= carry[3] << 21
	carry[4] = s4 >> 21
	s5 += carry[4]
	s4 -= carry[4] << 21
	carry[5] = s5 >> 21
	s6 += carry[5]
	s5 -= carry[5] << 21
	carry[6] = s6 >> 21
	s7 += carry[6]
	s6 -= carry[6] << 21
	carry[7] = s7 >> 21
	s8 += carry[7]
	s7 -= carry[7] << 21
	carry[8] = s8 >> 21
	s9 += carry[8]
	s8 -= carry[8] << 21
	carry[9] = s9 >> 21
	s10 += carry[9]
	s9 -= carry[9] << 21
	carry[10] = s10 >> 21
	s11 += carry[10]
	s10 -= carry[10] << 21

	out[0] = byte(s0 >> 0)
	out[1] = byte(s0 >> 8)
	out[2] = byte((s0 >> 16) | (s1 << 5))
	out[3] = byte(s1 >> 3)
	out[4] = byte(s1 >> 11)
	out[5] = byte((s1 >> 19) | (s2 << 2))
	out[6] = byte(s2 >> 6)
	out[7] = byte((s2 >> 14) | (s3 << 7))
	out[8] = byte(s3 >> 1)
	out[9] = byte(s3 >> 9)
	out[10] = byte((s3 >> 17) | (s4 << 4))
	out[11] = byte(s4 >> 4)
	out[12] = byte(s4 >> 12)
	out[13] = byte((s4 >> 20) | (s5 << 1))
	out[14] = byte(s5 >> 7)
	out[15] = byte((s5 >> 15) | (s6 << 6))
	out[16] = byte(s6 >> 2)
	out[17] = byte(s6 >> 10)
	out[18] = byte((s6 >> 18) | (s7 << 3))
	out[19] = byte(s7 >> 5)
	out[20] = byte(s7 >> 13)
	out[21] = byte(s8 >> 0)
	out[22] = byte(s8 >> 8)
	out[23] = byte((s8 >> 16) | (s9 << 5))
	out[24] = byte(s9 >> 3)
	out[25] = byte(s9 >> 11)
	out[26] = byte((s9 >> 19) | (s10 << 2))
	out[27] = byte(s10 >> 6)
	out[28] = byte((s10 >> 14) | (s11 << 7))
	out[29] = byte(s11 >> 1)
	out[30] = byte(s11 >> 9)
	out[31] = byte(s11 >> 17)
}
