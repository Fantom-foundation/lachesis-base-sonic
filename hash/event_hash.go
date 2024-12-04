package hash

import (
	"bytes"
	"crypto/sha256"
	"fmt"
	"math/big"
	"math/rand"
	"sort"
	"strings"

	"github.com/ethereum/go-ethereum/common"

	"github.com/Fantom-foundation/lachesis-base/common/bigendian"
	"github.com/Fantom-foundation/lachesis-base/inter/idx"
)

type (
	// EventHash is a unique identifier of event.
	// It is a hash of EventHash.
	EventHash Hash

	// OrderedEventHashes is a sortable slice of event hashes.
	OrderedEventHashes []EventHash

	// EventHashes is a slice of event hashes.
	EventHashes []EventHash

	// EventHashStack is a stack of event hashes.
	EventHashStack []EventHash

	// EventHashSet provides additional methods of event hash index.
	EventHashSet map[EventHash]struct{}
)

var (
	// ZeroEvent is a hash of virtual initial event.
	ZeroEvent = EventHash{}
)

/*
 * Event methods:
 */

// Bytes returns value as byte slice.
func (h EventHash) Bytes() []byte {
	return (Hash)(h).Bytes()
}

// Big converts a hash to a big integer.
func (h *EventHash) Big() *big.Int {
	return (*Hash)(h).Big()
}

// setBytes converts bytes to event hash.
// If b is larger than len(h), b will be cropped from the left.
func (h *EventHash) SetBytes(raw []byte) {
	(*Hash)(h).SetBytes(raw)
}

// BytesToEvent converts bytes to event hash.
// If b is larger than len(h), b will be cropped from the left.
func BytesToEvent(b []byte) EventHash {
	return EventHash(FromBytes(b))
}

// FromBytes converts bytes to hash.
// If b is larger than len(h), b will be cropped from the left.
func FromBytes(b []byte) Hash {
	var h Hash
	h.SetBytes(b)
	return h
}

// HexToEventHash sets byte representation of s to hash.
// If b is larger than len(h), b will be cropped from the left.
func HexToEventHash(s string) EventHash {
	return EventHash(HexToHash(s))
}

// Hex converts an event hash to a hex string.
func (h EventHash) Hex() string {
	return Hash(h).Hex()
}

// Lamport returns [4:8] bytes, which store event's Lamport.
func (h EventHash) Lamport() idx.Lamport {
	return idx.BytesToLamport(h[4:8])
}

// Epoch returns [0:4] bytes, which store event's Epoch.
func (h EventHash) Epoch() idx.EpochID {
	return idx.BytesToEpoch(h[0:4])
}

// String returns human readable string representation.
func (h EventHash) String() string {
	return h.ShortID(3)
}

// FullID returns human readable string representation with no information loss.
func (h EventHash) FullID() string {
	return h.ShortID(32 - 4 - 4)
}

// ShortID returns human readable ID representation, suitable for API calls.
func (h EventHash) ShortID(precision int) string {
	if name := GetEventName(h); len(name) > 0 {
		return name
	}
	// last bytes, because first are occupied by epoch and lamport
	return fmt.Sprintf("%d:%d:%s", h.Epoch(), h.Lamport(), common.Bytes2Hex(h[8:8+precision]))
}

// IsZero returns true if hash is empty.
func (h *EventHash) IsZero() bool {
	return *h == EventHash{}
}

/*
 * EventsSet methods:
 */

// NewEventsSet makes event hash index.
func NewEventsSet(h ...EventHash) EventHashSet {
	hh := EventHashSet{}
	hh.Add(h...)
	return hh
}

// Copy copies events to a new structure.
func (hh EventHashSet) Copy() EventHashSet {
	ee := make(EventHashSet, len(hh))
	for k, v := range hh {
		ee[k] = v
	}

	return ee
}

// String returns human readable string representation.
func (hh EventHashSet) String() string {
	ss := make([]string, 0, len(hh))
	for h := range hh {
		ss = append(ss, h.String())
	}
	return "[" + strings.Join(ss, ", ") + "]"
}

// Slice returns whole index as slice.
func (hh EventHashSet) Slice() EventHashes {
	arr := make(EventHashes, len(hh))
	i := 0
	for h := range hh {
		arr[i] = h
		i++
	}
	return arr
}

// Add appends hash to the index.
func (hh EventHashSet) Add(hash ...EventHash) {
	for _, h := range hash {
		hh[h] = struct{}{}
	}
}

// Erase erase hash from the index.
func (hh EventHashSet) Erase(hash ...EventHash) {
	for _, h := range hash {
		delete(hh, h)
	}
}

// Contains returns true if hash is in.
func (hh EventHashSet) Contains(hash EventHash) bool {
	_, ok := hh[hash]
	return ok
}

/*
 * Events methods:
 */

// NewEvents makes event hash slice.
func NewEvents(h ...EventHash) EventHashes {
	hh := EventHashes{}
	hh.Add(h...)
	return hh
}

// Copy copies events to a new structure.
func (hh EventHashes) Copy() EventHashes {
	return append(EventHashes(nil), hh...)
}

// String returns human readable string representation.
func (hh EventHashes) String() string {
	ss := make([]string, 0, len(hh))
	for _, h := range hh {
		ss = append(ss, h.String())
	}
	return "[" + strings.Join(ss, ", ") + "]"
}

// Set returns whole index as a EventsSet.
func (hh EventHashes) Set() EventHashSet {
	set := make(EventHashSet, len(hh))
	for _, h := range hh {
		set[h] = struct{}{}
	}
	return set
}

// Add appends hash to the slice.
func (hh *EventHashes) Add(hash ...EventHash) {
	*hh = append(*hh, hash...)
}

/*
 * EventsStack methods:
 */

// Push event ID on top
func (s *EventHashStack) Push(v EventHash) {
	*s = append(*s, v)
}

// PushAll event IDs on top
func (s *EventHashStack) PushAll(vv EventHashes) {
	*s = append(*s, vv...)
}

// Pop event ID from top. Erases element.
func (s *EventHashStack) Pop() *EventHash {
	l := len(*s)
	if l == 0 {
		return nil
	}

	res := &(*s)[l-1]
	*s = (*s)[:l-1]

	return res
}

/*
 * OrderedEvents methods:
 */

// String returns string representation.
func (hh OrderedEventHashes) String() string {
	buf := &strings.Builder{}

	out := func(s string) {
		if _, err := buf.WriteString(s); err != nil {
			panic(err)
		}
	}

	out("[")
	for _, h := range hh {
		out(h.String() + ", ")
	}
	out("]")

	return buf.String()
}

func (hh OrderedEventHashes) Len() int      { return len(hh) }
func (hh OrderedEventHashes) Swap(i, j int) { hh[i], hh[j] = hh[j], hh[i] }
func (hh OrderedEventHashes) Less(i, j int) bool {
	return bytes.Compare(hh[i].Bytes(), hh[j].Bytes()) < 0
}

// ByEpochAndLamport sorts events by epoch first, by lamport second, by ID third
func (hh OrderedEventHashes) ByEpochAndLamport() {
	sort.Sort(hh)
}

// Of returns hash of data
func Of(data ...[]byte) (hash Hash) {
	d := sha256.New()
	for _, b := range data {
		_, err := d.Write(b)
		if err != nil {
			panic(err)
		}
	}
	d.Sum(hash[:0])
	return hash
}

/*
 * Utils:
 */

// FakePeer generates random fake peer id for testing purpose.
func FakePeer() idx.ValidatorID {
	return idx.BytesToValidatorID(FakeHash().Bytes()[:4])
}

// FakeEpoch gives fixed value of fake epoch for testing purpose.
func FakeEpoch() idx.EpochID {
	return 123456
}

// FakeEvent generates random fake event hash with the same epoch for testing purpose.
func FakeEvent() (h EventHash) {
	_, err := rand.Read(h[:]) // nolint:gosec
	if err != nil {
		panic(err)
	}
	copy(h[0:4], bigendian.Uint32ToBytes(uint32(FakeEpoch())))
	return
}

// FakeEvents generates random hashes of fake event with the same epoch for testing purpose.
func FakeEvents(n int) EventHashes {
	res := EventHashes{}
	for i := 0; i < n; i++ {
		res.Add(FakeEvent())
	}
	return res
}
