package drive

import (
	"unicode/utf16"
	"unicode/utf8"
	"unsafe"
)

const (
	surr1    = 0xd800
	surr2    = 0xdc00
	surr3    = 0xe000
	tx       = 0b10000000
	t3       = 0b11100000
	maskx    = 0b00111111
	rune1Max = 1<<7 - 1
	rune2Max = 1<<11 - 1
)

// UTF16ToString is a copy of the standard syscall.UTF16ToString function, except
// it uses a buffer pool instead of creating a new one on each call.
func UTF16ToString(alloc Allocator, s []uint16) string {
	maxLen := uint32(0)
	for i, v := range s {
		if v == 0 {
			s = s[0:i]

			break
		}

		switch {
		case v <= rune1Max:
			maxLen++
		case v <= rune2Max:
			maxLen += 2
		default:
			maxLen += 3
		}
	}

	buf, err := alloc.Alloc(maxLen)
	if err != nil {
		panic(err)
	}

	decodeWTF16(s, buf[0:0:maxLen])

	return unsafe.String(unsafe.SliceData(buf), maxLen)
}

func decodeWTF16(s []uint16, buf []byte) {
	for i := 0; i < len(s); i++ {
		var ar rune

		switch r := s[i]; {
		case r < surr1, surr3 <= r:
			ar = rune(r)
		case r < surr2 && i+1 < len(s) && surr2 <= s[i+1] && s[i+1] < surr3:
			ar = utf16.DecodeRune(rune(r), rune(s[i+1]))
			i++
		default:
			ar = rune(r)
			if ar > utf8.MaxRune {
				ar = utf8.RuneError
			}
			buf = append(buf, t3|byte(ar>>12), tx|byte(ar>>6)&maskx, tx|byte(ar)&maskx)

			continue
		}

		buf = utf8.AppendRune(buf, ar)
	}
}
