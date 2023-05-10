package background

import (
	"time"

	"github.com/hung0913208/go-algorithm/lib/algorithm"
)

type Cron interface {
}

type jobImpl struct {
	next, interval time.Duration
	handler        Handler
}

type cronImpl struct {
	jobSched algorithm.Heap
}

func NewCron() Cron {
	return &cronImpl{
		jobSched: algorithm.NewHeap(compareJob),
	}
}

func (self *cronImpl) Start() {
	self.Reschedule()
}

func (self *cronImpl) Reschedule() {
}

func compareJob(l, r interface{}) bool {
	return false
}
