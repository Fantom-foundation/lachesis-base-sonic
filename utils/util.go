package utils

import (
	"fmt"

	"github.com/Fantom-foundation/lachesis-base/ltypes"
)

// NameOf returns human readable string representation.
func NameOf(p ltypes.ValidatorID) string {
	if name := ltypes.GetNodeName(p); len(name) > 0 {
		return name
	}

	return fmt.Sprintf("%d", p)
}
