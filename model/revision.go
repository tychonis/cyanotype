package model

type RevisionID = Digest

type Revision struct {
	Digest    RevisionID   `json:"id"`
	CreatedAt int64        `json:"created_at"`
	Parents   []RevisionID `json:"parents"`
}
