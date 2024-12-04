package types

import (
	"fmt"
)

type Event interface {
	Epoch() Epoch
	Seq() Event
	Frame() Frame
	Creator() ValidatorID
	Lamport() Lamport
	Parents() Events
	SelfParent() *Event
	IsSelfParent(hash Event) bool
	ID() Event
	String() string
	Size() int
}

type MutableEvent interface {
	Event
	SetEpoch(Epoch)
	SetSeq(Event)
	SetFrame(Frame)
	SetCreator(ValidatorID)
	SetLamport(Lamport)
	SetParents(Events)
	SetID(id [24]byte)
}

// BaseEvent is the consensus message in the Lachesis consensus algorithm
type BaseEvent struct {
	epoch   Epoch
	seq     Event
	frame   Frame
	creator ValidatorID
	parents Events // first event must be self-parent
	lamport Lamport
	id      Event
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
func (e *BaseEvent) SelfParent() *.Event {
	if e.seq <= 1 || len(e.parents) == 0 {
		return nil
	}
	return &e.parents[0]
}

// IsSelfParent is true if specified ID is event's self-parent
func (e *BaseEvent) IsSelfParent(hash Event) bool {
	if e.SelfParent() == nil {
		return false
	}
	return *e.SelfParent() == hash
}

func (e *BaseEvent) Epoch() Epoch {
	return e.epoch
}

func (e *BaseEvent) Seq() Event {
	return e.seq
}

func (e *BaseEvent) Frame() Frame {
	return e.frame
}

func (e *BaseEvent) Creator() ValidatorID {
	return e.creator
}

func (e *BaseEvent) Parents() Events {
	return e.parents
}

func (e *BaseEvent) Lamport() Lamport {
	return e.lamport
}

func (e *BaseEvent) ID() Event {
	return e.id
}

// TBD: That is messy!
func (e *BaseEvent) Size() int {
	return 4 + 4 + 4 + 4 + len(e.parents)*32 + 4 + 32
}

func (e *MutableBaseEvent) SetEpoch(v Epoch) {
	e.epoch = v
}

func (e *MutableBaseEvent) SetSeq(v Event) {
	e.seq = v
}

func (e *MutableBaseEvent) SetFrame(v Frame) {
	e.frame = v
}

func (e *MutableBaseEvent) SetCreator(v ValidatorID) {
	e.creator = v
}

func (e *MutableBaseEvent) SetParents(v Events) {
	e.parents = v
}

func (e *MutableBaseEvent) SetLamport(v Lamport) {
	e.lamport = v
}

func (e *MutableBaseEvent) SetID(rID [24]byte) {
	copy(e.id[0:4], e.epoch.Bytes())
	copy(e.id[4:8], e.lamport.Bytes())
	copy(e.id[8:], rID[:])
}
