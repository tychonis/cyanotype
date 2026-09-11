package model

type Object struct {
	Source    string     `json:"source"`
	Type      string     `json:"type"`
	Placement *Placement `json:"placement"`
}

type Scene struct {
	Objects []*Object `json:"objects"`
}
