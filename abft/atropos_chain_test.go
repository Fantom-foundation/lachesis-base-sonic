package abft

import (
	"database/sql"
	"encoding/hex"
	"log"
	"testing"

	"github.com/Fantom-foundation/lachesis-base/hash"
	"github.com/Fantom-foundation/lachesis-base/inter/dag/tdag"
	"github.com/Fantom-foundation/lachesis-base/inter/idx"
	"github.com/Fantom-foundation/lachesis-base/inter/pos"
	"github.com/Fantom-foundation/lachesis-base/lachesis"
	_ "github.com/mattn/go-sqlite3"
)

type event struct {
	hash        hash.Event
	validatorId idx.ValidatorID
	seq         idx.Event
	frame       idx.Frame
	lamportTs   idx.Lamport
	parents     []hash.Event
}

func TestRegressionData_AtroposChainMatches(t *testing.T) {
	conn, err := sql.Open("sqlite3", "testdata/events-5577.db")
	if err != nil {
		t.Fatal(err)
	}
	defer conn.Close()

	epochs := getEpochs(t, conn)
	for _, epoch := range epochs {
		testAtroposForEpoch(t, conn, epoch)
	}
}

func testAtroposForEpoch(t *testing.T, conn *sql.DB, epoch idx.Epoch) {
	validators, weights := getValidatorMeta(t, conn, epoch)
	testLachesis, _, eventStore, _ := FakeLachesis(validators, weights)
	// Plant the real epoch state for the sake of event hash calculation (epoch=1 by default)
	testLachesis.store.SetEpochState(&EpochState{Epoch: epoch, Validators: testLachesis.store.GetValidators()})

	recalculatedAtropoi := make([]hash.Event, 0)
	// Capture the elected atropoi by planting the `applyBlock` callback (nil by default)
	testLachesis.applyBlock = func(block *lachesis.Block) *pos.Validators {
		recalculatedAtropoi = append(recalculatedAtropoi, block.Atropos)
		return nil
	}

	eventsOrdered, eventMap := getEvents(t, conn, epoch)
	// Ingesting by lamport ts guarantees that all parents are already ingested
	for _, event := range eventsOrdered {
		ingestEvent(t, testLachesis, eventStore, event)
	}

	expectedAtropoi := getAtropoi(t, conn, epoch)
	if want, got := len(expectedAtropoi), len(recalculatedAtropoi); want != got {
		t.Fatalf("incorrect number of atropoi recalculated for epoch %d, expected: %d, got: %d.", epoch, want, got)
	}
	for idx := range recalculatedAtropoi {
		if want, got := expectedAtropoi[idx], recalculatedAtropoi[idx]; want != got {
			t.Fatalf("incorrect atropos for epoch %d on position %d, expected: %v got: %v.", epoch, idx, eventMap[want], eventMap[got])
		}
	}
}

func ingestEvent(t *testing.T, testLachesis *TestLachesis, eventStore *EventStore, event *event) *tdag.TestEvent {
	testEvent := &tdag.TestEvent{}
	testEvent.SetSeq(event.seq)
	testEvent.SetCreator(event.validatorId)
	testEvent.SetParents(event.parents)
	testEvent.SetLamport(event.lamportTs)
	testEvent.SetEpoch(testLachesis.store.GetEpoch())
	if err := testLachesis.Build(testEvent); err != nil {
		t.Fatalf("error while building event for validator: %d, seq: %d, err: %v", event.validatorId, event.seq, err)
	}
	testEvent.SetID([24]byte(event.hash[8:]))
	eventStore.SetEvent(testEvent)
	if err := testLachesis.Process(testEvent); err != nil {
		t.Fatalf("error while processing event for validator: %d, seq: %d, err: %v", event.validatorId, event.seq, err)
	}
	return testEvent
}

func getEpochs(t *testing.T, conn *sql.DB) []idx.Epoch {
	// Query the `Event` table as `Validator` table may include future (empty) epochs
	rows, err := conn.Query(`
		SELECT DISTINCT e.EpochId
		FROM Event e
	`)
	if err != nil {
		t.Fatal(err)
	}
	defer rows.Close()

	epochs := make([]idx.Epoch, 0)
	for rows.Next() {
		var epoch idx.Epoch
		err = rows.Scan(&epoch)
		if err != nil {
			log.Fatal(err)
		}
		epochs = append(epochs, epoch)
	}
	return epochs
}

func getValidatorMeta(t *testing.T, conn *sql.DB, epoch idx.Epoch) ([]idx.ValidatorID, []pos.Weight) {
	rows, err := conn.Query(`
		SELECT ValidatorId, Weight
		FROM Validator
		WHERE EpochId = ?
	`, epoch)
	if err != nil {
		t.Fatal(err)
	}
	defer rows.Close()

	validators := make([]idx.ValidatorID, 0)
	weights := make([]pos.Weight, 0)
	for rows.Next() {
		var validatorId idx.ValidatorID
		var weight pos.Weight

		err = rows.Scan(&validatorId, &weight)
		if err != nil {
			t.Fatal(err)
		}

		validators = append(validators, validatorId)
		weights = append(weights, weight)
	}
	return validators, weights
}

func getEvents(t *testing.T, conn *sql.DB, epoch idx.Epoch) ([]*event, map[hash.Event]*event) {
	rows, err := conn.Query(`
		SELECT e.EventHash, e.ValidatorId, e.SequenceNumber, e.FrameId, e.LamportNumber
		FROM Event e
		WHERE e.EpochId = ?
		ORDER BY e.LamportNumber ASC
	`, epoch)
	if err != nil {
		t.Fatal(err)
	}
	defer rows.Close()

	eventMap := make(map[hash.Event]*event)
	eventsOrdered := make([]*event, 0)
	for rows.Next() {
		var hashStr string
		var validatorId idx.ValidatorID
		var seq idx.Event
		var frame idx.Frame
		var lamportTs idx.Lamport
		err = rows.Scan(&hashStr, &validatorId, &seq, &frame, &lamportTs)
		if err != nil {
			t.Fatal(err)
		}

		eventHash := decodeHashStr(hashStr, t)
		event := &event{
			hash:        eventHash,
			validatorId: validatorId,
			seq:         seq,
			frame:       frame,
			lamportTs:   lamportTs,
			parents:     make([]hash.Event, 0),
		}
		eventsOrdered = append(eventsOrdered, event)
		eventMap[eventHash] = event
	}
	appointParents(t, conn, eventMap, epoch)
	return eventsOrdered, eventMap
}

func appointParents(t *testing.T, conn *sql.DB, eventMap map[hash.Event]*event, epoch idx.Epoch) {
	rows, err := conn.Query(`
		SELECT e.EventHash, eParent.EventHash
		FROM Event e JOIN Parent p ON e.EventId = p.EventId JOIN Event eParent ON eParent.EventId = p.ParentId
		WHERE e.EpochId = ?
	`, epoch)
	if err != nil {
		t.Fatal(err)
	}
	defer rows.Close()

	for rows.Next() {
		var eventHashStr string
		var parentHashStr string
		err = rows.Scan(&eventHashStr, &parentHashStr)
		if err != nil {
			t.Fatal(err)
		}

		eventHash := decodeHashStr(eventHashStr, t)
		parentHash := decodeHashStr(parentHashStr, t)
		event, ok := eventMap[eventHash]
		if !ok {
			t.Fatalf(
				"incomplete events.db - child event not found. epoch: %d, child event: %s, parent event: %s",
				epoch,
				eventHash,
				parentHash,
			)
		}
		if _, ok := eventMap[parentHash]; !ok {
			t.Fatalf(
				"incomplete events.db - parent event not found. epoch: %d, child event: %s, parent event: %s",
				epoch,
				eventHash,
				parentHash,
			)
		}
		event.parents = append(event.parents, parentHash)
	}
}

func getAtropoi(t *testing.T, conn *sql.DB, epoch idx.Epoch) []hash.Event {
	rows, err := conn.Query(`
		SELECT e.EventHash
		FROM Atropos a JOIN Event e ON a.AtroposId = e.EventId
		WHERE e.EpochId = ?
		ORDER BY a.AtroposId ASC
	`, epoch)
	if err != nil {
		t.Fatal(err)
	}
	defer rows.Close()

	atropoi := make([]hash.Event, 0)
	for rows.Next() {
		var atroposHashStr string
		err = rows.Scan(&atroposHashStr)
		if err != nil {
			t.Fatal(err)
		}

		atroposHash := decodeHashStr(atroposHashStr, t)
		atropoi = append(atropoi, atroposHash)
	}
	return atropoi
}

// hashStr is in hex format, i.e. 0x1a2b3c4d...
func decodeHashStr(hashStr string, t *testing.T) hash.Event {
	hashSlice, err := hex.DecodeString(hashStr[2:])
	if err != nil {
		t.Fatal(err)
	}
	return hash.Event(hashSlice)
}
