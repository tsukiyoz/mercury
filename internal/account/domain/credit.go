package domain

type Credit struct {
	Biz   string
	BizId string
	Items []CreditItem
}

type CreditItem struct {
	Account     int64       // 对外暴露的账户ID
	AccountType AccountType // 账户类型
	Amount      int
	Currency    string
	Uid         int64 // 用户ID，非系统账号必填
}

type AccountType uint8

func (a AccountType) AsUint8() uint8 {
	return uint8(a)
}

const (
	AccountTypeUnknown AccountType = iota
	AccountTypeReward
	AccountTypeSystem
)
