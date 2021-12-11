package crypto

import "github.com/ebfe/keccak"

// Keccak256 returns the keccak256 hash of the data.
func Keccak256(data ...[]byte) (result [32]byte) {
	h := keccak.New256()
	for _, b := range data {
		h.Write(b)
	}
	r := h.Sum(nil)
	copy(result[:], r)
	return
}
