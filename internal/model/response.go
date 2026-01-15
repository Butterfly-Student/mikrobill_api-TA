package model

type Metadata struct {
	Total  int64 `json:"total,omitempty"`
	Limit  int   `json:"limit,omitempty"`
	Offset int   `json:"offset,omitempty"`
}

type Response struct {
	Success  bool      `json:"success"`
	Data     any       `json:"data,omitempty"`
	Metadata *Metadata `json:"metadata,omitempty"`
	Error    any       `json:"error,omitempty"`
}
