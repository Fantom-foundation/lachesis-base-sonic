package ltypes

// Block is a part of an ordered chain of batches of events.
type Block struct {
	Atropos  EventHash
	Cheaters Cheaters
}
