package model

type Process struct {
	Input          []*Component
	Output         []*Component
	Transformation func([]*Contract) *[]Contract
}
