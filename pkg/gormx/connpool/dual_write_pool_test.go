package connpool

import (
	"github.com/ecodeclub/ekit/syncx/atomicx"
	"github.com/stretchr/testify/require"
	interactiveDao "github.com/tsukaychan/mercury/interactive/repository/dao"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
	"testing"
)

func TestConnPool(t *testing.T) {
	mercuryDB, err := gorm.Open(mysql.Open("root:for.nothing@tcp(localhost:3306)/mercury"))
	require.NoError(t, err)

	err = mercuryDB.AutoMigrate(&interactiveDao.Interactive{})
	require.NoError(t, err)

	interactiveDB, err := gorm.Open(mysql.Open("root:for.nothing@tcp(localhost:3306)/mercury_interactive"))
	require.NoError(t, err)

	err = interactiveDB.AutoMigrate(&interactiveDao.Interactive{})
	require.NoError(t, err)

	db, err := gorm.Open(mysql.New(mysql.Config{
		Conn: &DualWritePool{
			src:     mercuryDB.ConnPool,
			dst:     interactiveDB.ConnPool,
			pattern: atomicx.NewValueOf(PatternSrcFirst),
		},
	}))
	//t.Log(db)

	err = db.Create(&interactiveDao.Interactive{
		Biz:   "test",
		BizId: 2333,
	}).Error
	require.NoError(t, err)

	err = db.Transaction(func(tx *gorm.DB) error {
		return tx.Create(&interactiveDao.Interactive{
			Biz:   "test_tx",
			BizId: 2333,
		}).Error
	})
	require.NoError(t, err)

	db.Model(&interactiveDao.Interactive{}).Where("biz LIKE ?", "test%").Updates(map[string]any{
		"biz_id": 7355608,
	})
}
