package model

type Object struct {
	Name      string     `json:"name"`
	Item      ItemID     `json:"item"`
	ItemName  string     `json:"item_name"`
	Source    string     `json:"source"`
	Type      string     `json:"type"`
	Placement *Placement `json:"placement"`
}

type Scene struct {
	Objects []*Object `json:"objects"`
}
