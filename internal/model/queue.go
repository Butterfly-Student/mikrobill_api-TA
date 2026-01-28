package model

type QueueSimple struct {
	ID       string `json:"id"`
	Name     string `json:"name"`
	Target   string `json:"target"`
	MaxLimit string `json:"max_limit"`
	LimitAt  string `json:"limit_at,omitempty"`
	Priority string `json:"priority,omitempty"`
	Disabled bool   `json:"disabled"`
}

type QueueSimpleInput struct {
	Name     string `json:"name" binding:"required"`
	Target   string `json:"target" binding:"required"`
	MaxLimit string `json:"max_limit" binding:"required"`
	LimitAt  string `json:"limit_at"`
	Priority string `json:"priority"`
}

type QueueSimpleUpdateInput struct {
	Name     *string `json:"name"`
	Target   *string `json:"target"`
	MaxLimit *string `json:"max_limit"`
	LimitAt  *string `json:"limit_at"`
	Priority *string `json:"priority"`
	Disabled *bool   `json:"disabled"`
}
