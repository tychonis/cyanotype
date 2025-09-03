// Supersession chain
package ssc

import (
	"github.com/google/uuid"

	"github.com/tychonis/cyanotype/model"
)

type ChainID = uuid.UUID

type Chain struct {
	ID   ChainID
	Head model.ItemID
}

type Link struct {
	Item  model.ItemID
	Chain ChainID
}
