package model

type ProcessID = Digest

type Process struct {
	SymbolBase
	SymbolContent

	Input      []*BOMLine    `json:"input" yaml:"input"`
	Output     []*BOMLine    `json:"output" yaml:"output"`
	Operations []OperationID `json:"operations" yaml:"operations"`
}

func (p *Process) GetType() string {
	return "process"
}

type CoProcess struct {
	SymbolBase
	SymbolContent

	Input      []*BOMLine   `json:"input" yaml:"input"`
	Output     []*BOMLine   `json:"output" yaml:"output"`
	Operations []*Operation `json:"operations" yaml:"operations"`
}

func (cp *CoProcess) GetType() string {
	return "coprocess"
}
