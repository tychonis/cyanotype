// Supersession chain
package model

import (
	"github.com/google/uuid"
)

type ChainID = uuid.UUID
type LinkType string

const (
	LT_SUPPERSESSEION LinkType = "supersession"
	LT_VARIANT        LinkType = "variant"
	LT_CHANGE         LinkType = "change"
)

type Chain struct {
	ID   ChainID
	Head ItemID
}

type Link struct {
	Item   ItemID
	Chain  ChainID
	Parent ItemID
	Type   LinkType
}
