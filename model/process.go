package model

type Process struct {
	Input  []*Process
	Output []*Process
}
