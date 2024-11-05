package tdx

import "github.com/injoyai/tdx/protocol"

type TypeKline uint8

func (this TypeKline) Uint8() uint8 {
	return uint8(this)
}

func (this TypeKline) TableName() string {
	switch uint8(this) {
	case protocol.TypeKlineMinute:
		return "kline_minute"
	case protocol.TypeKline5Minute:
		return "kline_5minute"
	case protocol.TypeKline15Minute:
		return "kline_15minute"
	case protocol.TypeKline30Minute:
		return "kline_30minute"
	case protocol.TypeKlineHour:
		return "kline_hour"
	case protocol.TypeKlineDay:
		return "kline_day"
	case protocol.TypeKlineWeek:
		return "kline_week"
	case protocol.TypeKlineMonth:
		return "kline_month"
	case protocol.TypeKlineQuarter:
		return "kline_quarter"
	case protocol.TypeKlineYear:
		return "kline_year"
	default:
		return "kline_unknown"
	}
}
