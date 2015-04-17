package main

import (
	"math/rand"
	"sort"
	"time"
	"unsafe"
)

type Mutator struct {
	r *rand.Rand
}

func newMutator() *Mutator {
	return &Mutator{r: rand.New(rand.NewSource(time.Now().UnixNano()))}
}

func (m *Mutator) rand(n int) int {
	return m.r.Intn(n)
}

func (m *Mutator) generate(corpus []Input) ([]byte, int) {
	if len(corpus) == 0 {
		return []byte{byte(m.rand(256))}, 0
	}
	scoreSum := corpus[len(corpus)-1].runningScoreSum
	weightedIdx := m.rand(scoreSum)
	idx := sort.Search(len(corpus), func(i int) bool {
		return corpus[i].runningScoreSum > weightedIdx
	})
	input := &corpus[idx]
	return m.mutate(input.data, corpus), input.depth + 1
}

func (m *Mutator) mutate(data []byte, corpus []Input) []byte {
	res := make([]byte, len(data))
	copy(res, data)
	nm := 1
	for m.rand(2) == 0 {
		nm++
	}
	for iter := 0; iter < nm; iter++ {
		switch m.rand(15) {
		case 0:
			// Remove a range of bytes.
			if len(res) <= 1 {
				iter--
				continue
			}
			pos0 := m.rand(len(res))
			pos1 := pos0 + m.chooseLen(len(res)-pos0)
			copy(res[pos0:], res[pos1:])
			res = res[:len(res)-(pos1-pos0)]
		case 1:
			// Insert a range of random bytes.
			pos := m.rand(len(res) + 1)
			n := m.chooseLen(10)
			for i := 0; i < n; i++ {
				res = append(res, 0)
			}
			copy(res[pos+n:], res[pos:])
			for i := 0; i < n; i++ {
				res[pos+i] = byte(m.rand(256))
			}
		case 2:
			// Duplicate a range of bytes.
			if len(res) <= 1 {
				iter--
				continue
			}
			src := m.rand(len(res))
			dst := m.rand(len(res))
			for dst == src {
				dst = m.rand(len(res))
			}
			n := m.chooseLen(len(res) - src)
			tmp := make([]byte, n)
			copy(tmp, res[src:])
			for i := 0; i < n; i++ {
				res = append(res, 0)
			}
			copy(res[dst+n:], res[dst:])
			for i := 0; i < n; i++ {
				res[dst+i] = tmp[i]
			}
		case 3:
			// Copy a range of bytes.
			if len(res) <= 1 {
				iter--
				continue
			}
			src := m.rand(len(res))
			dst := m.rand(len(res))
			for dst == src {
				dst = m.rand(len(res))
			}
			n := m.chooseLen(len(res) - src)
			if dst > len(res) || src+n > len(res) {
				println(len(res), dst, src, n)
			}
			copy(res[dst:], res[src:src+n])
		case 4:
			// Bit flip. Spooky!
			if len(res) == 0 {
				iter--
				continue
			}
			pos := m.rand(len(res))
			res[pos] ^= 1 << uint(m.rand(8))
		case 5:
			// Set a byte to a random value.
			if len(res) == 0 {
				iter--
				continue
			}
			pos := m.rand(len(res))
			res[pos] ^= byte(m.rand(255)) + 1
		case 6:
			// Swap 2 bytes.
			if len(res) <= 1 {
				iter--
				continue
			}
			src := m.rand(len(res))
			dst := m.rand(len(res))
			for dst == src {
				dst = m.rand(len(res))
			}
			res[src], res[dst] = res[dst], res[src]
		case 7:
			// Add/subtract from a byte.
			if len(res) == 0 {
				iter--
				continue
			}
			pos := m.rand(len(res))
			v := byte(m.rand(35) + 1)
			if m.rand(2) == 0 {
				res[pos] += v
			} else {
				res[pos] -= v
			}
		case 8:
			// Add/subtract from a uint16.
			if len(res) < 2 {
				iter--
				continue
			}
			pos := m.rand(len(res) - 1)
			v := uint16(m.rand(35) + 1)
			switch m.rand(4) {
			case 0:
				*(*uint16)(unsafe.Pointer(&res[pos])) += v
			case 1:
				*(*uint16)(unsafe.Pointer(&res[pos])) -= v
			case 2:
				x := *(*uint16)(unsafe.Pointer(&res[pos]))
				*(*uint16)(unsafe.Pointer(&res[pos])) = swap16(swap16(x) + v)
			case 3:
				x := *(*uint16)(unsafe.Pointer(&res[pos]))
				*(*uint16)(unsafe.Pointer(&res[pos])) = swap16(swap16(x) - v)
			}
		case 9:
			// Add/subtract from a uint32.
			if len(res) < 4 {
				iter--
				continue
			}
			pos := m.rand(len(res) - 3)
			v := uint32(m.rand(35) + 1)
			switch m.rand(4) {
			case 0:
				*(*uint32)(unsafe.Pointer(&res[pos])) += v
			case 1:
				*(*uint32)(unsafe.Pointer(&res[pos])) -= v
			case 2:
				x := *(*uint32)(unsafe.Pointer(&res[pos]))
				*(*uint32)(unsafe.Pointer(&res[pos])) = swap32(swap32(x) + v)
			case 3:
				x := *(*uint32)(unsafe.Pointer(&res[pos]))
				*(*uint32)(unsafe.Pointer(&res[pos])) = swap32(swap32(x) - v)
			}
		case 10:
			// Add/subtract from a uint64.
			if len(res) < 8 {
				iter--
				continue
			}
			pos := m.rand(len(res) - 7)
			v := uint64(m.rand(35) + 1)
			switch m.rand(4) {
			case 0:
				*(*uint64)(unsafe.Pointer(&res[pos])) += v
			case 1:
				*(*uint64)(unsafe.Pointer(&res[pos])) -= v
			case 2:
				x := *(*uint64)(unsafe.Pointer(&res[pos]))
				*(*uint64)(unsafe.Pointer(&res[pos])) = swap64(swap64(x) + v)
			case 3:
				x := *(*uint64)(unsafe.Pointer(&res[pos]))
				*(*uint64)(unsafe.Pointer(&res[pos])) = swap64(swap64(x) - v)
			}
		case 11:
			// Replace a byte with an interesting value.
			if len(res) == 0 {
				iter--
				continue
			}
			pos := m.rand(len(res))
			res[pos] = byte(interesting8[m.rand(len(interesting8))])
		case 12:
			// Replace an uint16 with an interesting value.
			if len(res) < 2 {
				iter--
				continue
			}
			pos := m.rand(len(res) - 1)
			v := uint16(interesting16[m.rand(len(interesting16))])
			if m.rand(2) == 0 {
				v = swap16(v)
			}
			*(*uint16)(unsafe.Pointer(&res[pos])) = v
		case 13:
			// Replace an uint32 with an interesting value.
			if len(res) < 4 {
				iter--
				continue
			}
			pos := m.rand(len(res) - 3)
			v := uint32(interesting32[m.rand(len(interesting32))])
			if m.rand(2) == 0 {
				v = swap32(v)
			}
			*(*uint32)(unsafe.Pointer(&res[pos])) = v
		case 14:
			// Splice another input.
			if len(res) < 4 || len(corpus) < 2 {
				iter--
				continue
			}
			other := corpus[m.rand(len(corpus))].data
			if len(other) < 4 || &data[0] == &other[0] {
				iter--
				continue
			}
			// Find common prefix and suffix.
			idx0 := 0
			for idx0 < len(res) && idx0 < len(other) && res[idx0] == other[idx0] {
				idx0++
			}
			idx1 := 0
			for idx1 < len(res) && idx1 < len(other) && res[len(res)-idx1-1] == other[len(other)-idx1-1] {
				idx1++
			}
			// If diffing parts are too small, there is no sense in splicing, rely on byte flipping.
			diff := min(len(res)-idx0-idx1, len(other)-idx0-idx1)
			if diff < 4 {
				iter--
				continue
			}
			copy(res[idx0:idx0+m.rand(diff-2)+1], other[idx0:])

			// TODO: replace letters 'a'..'z' with other letters.
			// TODO: replace digits '0'..'9' with other digits.
			// TODO: replace runs of digits with other runs of digits.
			// TODO: auto-detect magic bytes and strings in inputs and insert them more frequently.
		}
	}
	if len(res) > maxInputSize {
		res = res[:maxInputSize]
	}
	return res
}

var (
	interesting8  = []int8{-128, -1, 0, 1, 16, 32, 64, 100, 127}
	interesting16 = []int16{-32768, -129, 128, 255, 256, 512, 1000, 1024, 4096, 32767}
	interesting32 = []int32{-2147483648, -100663046, -32769, 32768, 65535, 65536, 100663045, 2147483647}
)

func init() {
	for _, v := range interesting8 {
		interesting16 = append(interesting16, int16(v))
	}
	for _, v := range interesting16 {
		interesting32 = append(interesting32, int32(v))
	}
}

// chooseLen chooses length of range mutation.
// It gives preference to shorter ranges.
func (m *Mutator) chooseLen(n int) int {
	switch m.rand(9) {
	case 0, 1, 2, 3, 4:
		return m.rand(min(10, n)) + 1
	case 5, 6, 7:
		return m.rand(min(50, n)) + 1
	default:
		return m.rand(n) + 1
	}
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}

func swap16(v uint16) uint16 {
	v0 := byte(v >> 0)
	v1 := byte(v >> 8)
	v = 0
	v |= uint16(v1) << 0
	v |= uint16(v0) << 8
	return v
}

func swap32(v uint32) uint32 {
	v0 := byte(v >> 0)
	v1 := byte(v >> 8)
	v2 := byte(v >> 16)
	v3 := byte(v >> 24)
	v = 0
	v |= uint32(v3) << 0
	v |= uint32(v2) << 8
	v |= uint32(v1) << 16
	v |= uint32(v0) << 24
	return v
}

func swap64(v uint64) uint64 {
	v0 := byte(v >> 0)
	v1 := byte(v >> 8)
	v2 := byte(v >> 16)
	v3 := byte(v >> 24)
	v4 := byte(v >> 32)
	v5 := byte(v >> 40)
	v6 := byte(v >> 48)
	v7 := byte(v >> 56)
	v = 0
	v |= uint64(v7) << 0
	v |= uint64(v6) << 8
	v |= uint64(v5) << 16
	v |= uint64(v4) << 24
	v |= uint64(v3) << 32
	v |= uint64(v2) << 40
	v |= uint64(v1) << 48
	v |= uint64(v0) << 56
	return v
}