package model

type ProcessID = Digest

type ProcessContent struct {
	Input      []*BOMLine    `json:"input" yaml:"input"`
	Output     []*BOMLine    `json:"output" yaml:"output"`
	Operations []OperationID `json:"operations" yaml:"operations"`
}

func (pc *ProcessContent) GetBomLine(name string) *BOMLine {
	for _, line := range pc.Input {
		if line.Name == name {
			return line
		}
	}
	for _, line := range pc.Output {
		if line.Name == name {
			return line
		}
	}
	return nil
}

type Process struct {
	SymbolBase
	SymbolContent

	ProcessContent
}

func (p *Process) GetType() string {
	return "process"
}

type CoProcess struct {
	SymbolBase
	SymbolContent

	ProcessContent
}

func (cp *CoProcess) GetType() string {
	return "coprocess"
}
