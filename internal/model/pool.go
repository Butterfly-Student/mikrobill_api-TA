package model

type IPPool struct {
	ID       string `json:"id"`
	Name     string `json:"name"`
	Ranges   string `json:"ranges"`
	NextPool string `json:"next_pool,omitempty"`
}

type IPPoolInput struct {
	Name     string `json:"name" binding:"required"`
	Ranges   string `json:"ranges" binding:"required"`
	NextPool string `json:"next_pool"`
}

type IPPoolUpdateInput struct {
	Name     *string `json:"name"`
	Ranges   *string `json:"ranges"`
	NextPool *string `json:"next_pool"`
}
