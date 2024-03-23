package task

type Job interface {
	Name() string
	Run() error
}
