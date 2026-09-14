package model

type Artifact struct {
	Name     string `json:"name" yaml:"name"`
	Filename string `json:"filename" yaml:"filename"`
	Tag      string `json:"tag" yaml:"tag"`
	Source   string `json:"source" yaml:"source"`
	Digest   string `json:"digest" yaml:"digest"`
}
