package catalog

type Rank struct {
	Sequence int   `json:"sequence"`
	WallTime int64 `json:"wall_time"`
}

type Metadata struct {
	Rank *Rank `json:"rank"`
}

func ZeroRank() *Rank {
	return &Rank{Sequence: 0, WallTime: 0}
}

func CmpRank(r1, r2 Rank) int {
	if r1.Sequence != r2.Sequence {
		return r1.Sequence - r2.Sequence
	}
	if r1.WallTime != r2.WallTime {
		if r1.WallTime < r2.WallTime {
			return -1
		}
		return 1
	}
	return 0
}
