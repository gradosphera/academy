package repository

var OperatorsMap = map[string]struct{}{
	"=":      {},
	"<":      {},
	">":      {},
	"<=":     {},
	">=":     {},
	"!=":     {},
	"@>":     {},
	"like":   {},
	"ilike":  {},
	"is":     {},
	"is not": {},
}

type Options struct {
	Pagination Pagination `json:"pagination"`
	Order      Order      `json:"order"`
	Filters    []Filter   `json:"filters"`
}

type Pagination struct {
	PageSize int `json:"page_size"`
	PageNum  int `json:"page_num"`
}

func (p Pagination) isValid() bool {
	return p.PageSize > 0 && p.PageNum > 0
}

type Order struct {
	OrderBy   string `json:"order_by"`
	OrderType string `json:"order_type"`
}

const (
	OrderDesc = "desc"
	OrderAsc  = "asc"
)

func (o Order) IsValid() bool {
	return (o.OrderType == OrderDesc || o.OrderType == OrderAsc) && o.OrderBy != ""
}

type Filter struct {
	Column   string `json:"column"`
	Operator string `json:"operator"`
	Value    string `json:"value"`
	WhereOr  bool   `json:"where_or"`
}

func (f Filter) isValid() bool {
	if f.Column == "" || f.Operator == "" || f.Value == "" {
		return false
	}

	if _, ok := OperatorsMap[f.Operator]; !ok {
		return false
	}

	return true
}
