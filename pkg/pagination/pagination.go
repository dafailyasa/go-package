package pagination

type PaginationRequest struct {
	Limit   int    `json:"limit,omitempty" query:"limit" example:"10"`
	Page    int    `json:"page,omitempty" query:"page" example:"1"`
	Sort    string `json:"sort,omitempty" query:"sort" example:"id DESC"`
	Status  *int   `json:"status,omitempty" query:"status" example:"3"`
	Keyword string `json:"keyword,omitempty" query:"keyword"`
	Field   string `json:"field,omitempty" query:"field"`
	Total   int64  `json:"total,omitempty" example:"100"`
}

func (p *PaginationRequest) HasStatus() bool {
	return p.Status != nil
}

func (pr PaginationRequest) Validate() (err error) {
	pr.Limit = pr.GetLimit()
	pr.Page = pr.GetPage()
	pr.Sort = pr.GetSort()

	return
}

func (p *PaginationRequest) GetOffset() int {
	return (p.GetPage() - 1) * p.GetLimit()
}

func (p *PaginationRequest) GetLimit() int {
	if p.Limit == 0 {
		p.Limit = 10
	}
	return p.Limit
}

func (p *PaginationRequest) GetPage() int {
	if p.Page == 0 {
		p.Page = 1
	}
	return p.Page
}

func (p *PaginationRequest) GetSort() string {
	if p.Sort == "" {
		p.Sort = "Id desc"
	}
	return p.Sort
}
