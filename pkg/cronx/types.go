package cron

type Task interface {
	Name() string
	Run() error
}
