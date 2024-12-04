package ltypes

import (
	"fmt"

	"github.com/Fantom-foundation/lachesis-base/hash"
	"github.com/Fantom-foundation/lachesis-base/inter/idx"
)

type Event interface {
	Epoch() idx.EpochID
	Seq() idx.EventID
	Frame() idx.FrameID
	Creator() idx.ValidatorID
	Lamport() idx.Lamport
	Parents() hash.EventHashes
	SelfParent() *hash.EventHash
	IsSelfParent(hash hash.EventHash) bool
	ID() hash.EventHash
	String() string
	Size() int
}

type MutableEvent interface {
	Event
	SetEpoch(idx.EpochID)
	SetSeq(idx.EventID)
	SetFrame(idx.FrameID)
	SetCreator(idx.ValidatorID)
	SetLamport(idx.Lamport)
	SetParents(hash.EventHashes)
	SetID(id [24]byte)
}

// BaseEvent is the consensus message in the Lachesis consensus algorithm
// The structure isn't supposed to be used as-is:
// Doesn't contain payload, it should be extended by an app
// Doesn't contain event signature, it should be extended by an app
type BaseEvent struct {
	epoch   idx.EpochID
	seq     idx.EventID
	frame   idx.FrameID
	creator idx.ValidatorID
	parents hash.EventHashes
	lamport idx.Lamport
	id      hash.EventHash
}

type MutableBaseEvent struct {
	BaseEvent
}

// Build build immutable event
func (me *MutableBaseEvent) Build(rID [24]byte) *BaseEvent {
	e := me.BaseEvent
	copy(e.id[0:4], e.epoch.Bytes())
	copy(e.id[4:8], e.lamport.Bytes())
	copy(e.id[8:], rID[:])
	return &e
}

// String returns string representation.
func (e *BaseEvent) String() string {
	return fmt.Sprintf("{id=%s, p=%s, by=%d, frame=%d}", e.id.ShortID(3), e.parents.String(), e.creator, e.frame)
}

// SelfParent returns event's self-parent, if any
func (e *BaseEvent) SelfParent() *hash.EventHash {
	if e.seq <= 1 || len(e.parents) == 0 {
		return nil
	}
	return &e.parents[0]
}

// IsSelfParent is true if specified ID is event's self-parent
func (e *BaseEvent) IsSelfParent(hash hash.EventHash) bool {
	if e.SelfParent() == nil {
		return false
	}
	return *e.SelfParent() == hash
}

func (e *BaseEvent) Epoch() idx.EpochID {
	return e.epoch
}

func (e *BaseEvent) Seq() idx.EventID {
	return e.seq
}

func (e *BaseEvent) Frame() idx.FrameID {
	return e.frame
}

func (e *BaseEvent) Creator() idx.ValidatorID {
	return e.creator
}

func (e *BaseEvent) Parents() hash.EventHashes {
	return e.parents
}

func (e *BaseEvent) Lamport() idx.Lamport {
	return e.lamport
}

func (e *BaseEvent) ID() hash.EventHash {
	return e.id
}

func (e *BaseEvent) Size() int {
	return 4 + 4 + 4 + 4 + len(e.parents)*32 + 4 + 32
}

func (e *MutableBaseEvent) SetEpoch(v idx.EpochID) {
	e.epoch = v
}

func (e *MutableBaseEvent) SetSeq(v idx.EventID) {
	e.seq = v
}

func (e *MutableBaseEvent) SetFrame(v idx.FrameID) {
	e.frame = v
}

func (e *MutableBaseEvent) SetCreator(v idx.ValidatorID) {
	e.creator = v
}

func (e *MutableBaseEvent) SetParents(v hash.EventHashes) {
	e.parents = v
}

func (e *MutableBaseEvent) SetLamport(v idx.Lamport) {
	e.lamport = v
}

func (e *MutableBaseEvent) SetID(rID [24]byte) {
	copy(e.id[0:4], e.epoch.Bytes())
	copy(e.id[4:8], e.lamport.Bytes())
	copy(e.id[8:], rID[:])
}
