package domain

type AsyncSms struct {
	Id       int64
	TplId    string
	Target   string
	Args     []string
	Values   []string
	RetryMax int
}
