package abft

import (
	"database/sql"
	"testing"

	_ "github.com/mattn/go-sqlite3"
)

func TestRegressionData_AtroposChainMatches(t *testing.T) {
	conn, err := sql.Open("sqlite3", "testdata/events-5577.db")
	if err != nil {
		t.Fatal(err)
	}
	defer conn.Close()

	epochMin, epochMax, err := GetEpochRange(conn)
	if err != nil {
		t.Fatal(err)
	}
	for epoch := epochMin; epoch <= epochMax; epoch++ {
		if err := CheckEpochAgainstDB(conn, epoch); err != nil {
			t.Fatal(err)
		}
	}
}
