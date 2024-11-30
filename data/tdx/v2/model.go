package tdx

// Workday 工作日
type Workday struct {
	ID   int64  `json:"id"`
	Unix int64  `json:"unix"`
	Date string `json:"date"`
	Is   bool   `json:"is"`
}
