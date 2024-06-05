package cronx

type Task interface {
	Name() string
	Run() error
}
