package model

import "time"

type Reference struct {
	Tag string `json:"tag" yaml:"tag"`
	URI string `json:"uri" yaml:"uri"`

	Size         int64     `json:"size" yaml:"size"`
	LastModified time.Time `json:"last_modified" yaml:"last_modified"`
	Digest       string    `json:"digest" yaml:"digest"`
}
