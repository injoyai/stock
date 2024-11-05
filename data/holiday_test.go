package data

import "testing"

func TestIsHoliday(t *testing.T) {
	t.Log(Holiday.Is("20241105"))
	t.Log(Holiday.Is("20241005"))
	t.Log(Holiday.Is("20241103"))
	t.Log(Holiday.Is("20241102"))
	t.Log(Holiday.Is("20230101"))
}
