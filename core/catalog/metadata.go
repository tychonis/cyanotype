package catalog

import "github.com/tychonis/cyanotype/model"

type Metadata struct {
	IntroducedBy  model.RevisionID   `json:"introduced_by"`
	CommitHistory []model.RevisionID `json:"commit_history"`
}

func (m *Metadata) Commit(rev model.RevisionID) {
	if m.CommitHistory == nil {
		m.CommitHistory = make([]model.RevisionID, 0)
	}
	m.CommitHistory = append(m.CommitHistory, rev)
}

func (m *Metadata) LastCommited() model.RevisionID {
	if len(m.CommitHistory) == 0 {
		return m.IntroducedBy
	}
	return m.CommitHistory[len(m.CommitHistory)-1]
}
