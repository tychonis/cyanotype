package model

type RevisionID = Digest

type Revision struct {
	ID        RevisionID   `json:"id"`
	CreatedAt int64        `json:"created_at"`
	Parents   []RevisionID `json:"parents"`
}
