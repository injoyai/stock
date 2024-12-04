package main

import (
	"fmt"
	"time"
)

func NewPlan(total int) *Plan {
	return &Plan{
		Total:        uint32(total),
		Current:      0,
		UpdateTime:   time.Now(),
		CompressTime: time.Now(),
	}
}

type Plan struct {
	Total        uint32
	Current      uint32
	UpdateTime   time.Time
	Compress     uint8 //0未压缩 1压缩中 2压缩完成
	CompressTime time.Time
}

func (this *Plan) Done() bool {
	return this.Current >= this.Total
}

func (this *Plan) Add() *Plan {
	this.Current++
	this.UpdateTime = time.Now()
	return this
}

func (this *Plan) CompressStart() *Plan {
	this.Compress = 1
	this.CompressTime = time.Now()
	return this
}

func (this *Plan) CompressEnd() *Plan {
	this.Compress = 2
	this.CompressTime = time.Now()
	return this
}

func (this *Plan) String() string {
	if this.Current == 0 {
		//开始更新
		return fmt.Sprintf(`开始更新,数量: %d
%s 更新: %d%%
%s 压缩: `,
			this.Total,
			this.UpdateTime.Format(time.TimeOnly), this.Current*100/this.Total,
			this.CompressTime.Format(time.TimeOnly))

	} else if !this.Done() {
		//更新中
		return fmt.Sprintf(`更新中...,数量: %d
%s 更新: %d%%
%s 压缩: `,
			this.Total,
			this.UpdateTime.Format(time.TimeOnly), this.Current*100/this.Total,
			this.CompressTime.Format(time.TimeOnly))

	} else {
		//更新结束

		switch this.Compress {

		case 1:
			return fmt.Sprintf(`更新中...,数量: %d
%s 更新: %d%%
%s 压缩: 进行中...`,
				this.Total,
				this.UpdateTime.Format(time.TimeOnly), this.Current*100/this.Total,
				this.CompressTime.Format(time.TimeOnly))

		case 2:
			return fmt.Sprintf(`更新结束,数量: %d
%s 更新: %d%%
%s 压缩: 完成`,
				this.Total,
				this.UpdateTime.Format(time.TimeOnly), this.Current*100/this.Total,
				this.CompressTime.Format(time.TimeOnly))

		default:
			return fmt.Sprintf(`更新结束,数量: %d
%s 更新: %d%%
%s 压缩: 无`,
				this.Total,
				this.UpdateTime.Format(time.TimeOnly), this.Current*100/this.Total,
				this.UpdateTime.Format(time.TimeOnly))

		}

	}
}
