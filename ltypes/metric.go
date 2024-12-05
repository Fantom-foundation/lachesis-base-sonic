package ltypes

import (
	"fmt"
)

type Metric struct {
	Num  EventID
	Size uint64
}

func (m Metric) String() string {
	return fmt.Sprintf("{Num=%d,Size=%d}", m.Num, m.Size)
}
