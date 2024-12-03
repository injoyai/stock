package main

import (
	"fmt"
	"time"
)

func NewPlan(total int) *Plan {
	return &Plan{
		Total:    uint32(total),
		Current:  0,
		LastTime: time.Now(),
	}
}

type Plan struct {
	Total    uint32
	Current  uint32
	LastTime time.Time
}

func (this *Plan) Done() bool {
	return this.Current >= this.Total
}

func (this *Plan) Add() {
	this.Current++
	this.LastTime = time.Now()
}

func (this *Plan) String() string {
	if this.Current == 0 {
		//开始更新
		return fmt.Sprintf("开始更新\n%s 进度:  %d%%", this.LastTime.Format(time.TimeOnly), this.Current*100/this.Total)
	} else if !this.Done() {
		//更新中
		return fmt.Sprintf("更新中...\n%s 进度:  %d%%", this.LastTime.Format(time.TimeOnly), this.Current*100/this.Total)
	} else {
		//更新结束
		return fmt.Sprintf("更新结束\n%s 进度:  %d%%", this.LastTime.Format(time.TimeOnly), this.Current*100/this.Total)
	}
}
