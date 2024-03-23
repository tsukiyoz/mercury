package domain

import (
	"time"

	"github.com/robfig/cron/v3"
)

var parser = cron.NewParser(cron.Second | cron.Minute |
	cron.Hour | cron.Dom | cron.Month | cron.Dow |
	cron.Descriptor)

type Task struct {
	Id         int64
	Name       string
	Executor   string
	Cfg        string
	Expression string
	NextTime   time.Time

	CancelFunc func()
}

func (tsk *Task) GetNextTime(start time.Time) time.Time {
	schedule, _ := parser.Parse(tsk.Expression)
	return schedule.Next(start)
}
