package dao

// Account is a placeholder for the real account type.
type Account struct {
	Id       int64 `gorm:"primaryKey,autoIncrement"`
	Uid      int64
	Account  int64 `gorm:"uniqueIndex:account_type"`
	Type     uint8 `gorm:"uniqueIndex:account_type"`
	Balance  int
	Currency string
	Ctime    int64
	Utime    int64
}

type AccountActivity struct {
	Id          int64 `gorm:"primaryKey,autoIncrement"`
	Uid         int64
	Biz         string `gorm:"uniqueIndex:biz_type_id"`
	BizId       string `gorm:"uniqueIndex:biz_type_id"`
	Account     int64  `gorm:"index:account_type"`
	AccountType uint8  `gorm:"index:account_type"`
	Amount      int64
	Currency    string
	Ctime       int64
	Utime       int64
}

func (AccountActivity) TableName() string {
	return "account_activities"
}
