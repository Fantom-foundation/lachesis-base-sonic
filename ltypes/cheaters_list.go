package ltypes

import (
	"github.com/ethereum/go-ethereum/rlp"
)

// Cheaters is a slice type for storing cheaters list.
type Cheaters []ValidatorID

// Set returns map of cheaters
func (s Cheaters) Set() map[ValidatorID]struct{} {
	set := map[ValidatorID]struct{}{}
	for _, element := range s {
		set[element] = struct{}{}
	}
	return set
}

// Len returns the length of s.
func (s Cheaters) Len() int { return len(s) }

// Swap swaps the i'th and the j'th element in s.
func (s Cheaters) Swap(i, j int) { s[i], s[j] = s[j], s[i] }

// getRlp implements Rlpable and returns the i'th element of s in rlp.
func (s Cheaters) GetRlp(i int) []byte {
	enc, _ := rlp.EncodeToBytes(s[i])
	return enc
}