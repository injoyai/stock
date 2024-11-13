package data

type Holiday struct {
	Year   int   `json:"year"`
	Month  int   `json:"month"`
	Day    int   `json:"day"`
	InDate int64 `json:"inDate" xorm:"created"`
}
