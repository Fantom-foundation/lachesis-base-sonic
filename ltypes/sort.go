package ltypes

type (
	validator struct {
		ID     ValidatorID
		Weight Weight
	}

	validators []validator
)

func (vv validators) Less(i, j int) bool {
	if vv[i].Weight != vv[j].Weight {
		return vv[i].Weight > vv[j].Weight
	}

	return vv[i].ID < vv[j].ID
}

func (vv validators) Len() int {
	return len(vv)
}

func (vv validators) Swap(i, j int) {
	vv[i], vv[j] = vv[j], vv[i]
}
