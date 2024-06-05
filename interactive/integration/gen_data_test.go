package integration

import (
	_ "embed"
	"math/rand"
	"os"
	"strconv"
	"testing"
	"time"

	"gorm.io/gorm"

	"github.com/lazywoo/mercury/interactive/integration/startup"
	"github.com/lazywoo/mercury/interactive/repository/dao"

	"github.com/stretchr/testify/require"
)

//go:embed init.sql
var initSQL string

func TestGenSQL(t *testing.T) {
	file, err := os.OpenFile("data.sql",
		os.O_RDWR|os.O_APPEND|os.O_CREATE|os.O_TRUNC, 0o666)
	require.NoError(t, err)

	defer file.Close()

	_, err = file.WriteString(initSQL)
	require.NoError(t, err)

	const prefix = "INSERT INTO `interactives`(`biz_id`, `biz`, `read_cnt`, `favorite_cnt`, `like_cnt`, `ctime`, `utime`)\nVALUES"
	const rowNum = 10

	now := time.Now().UnixMilli()
	_, err = file.WriteString(prefix)

	for i := 0; i < rowNum; i++ {
		if i > 0 {
			file.Write([]byte{',', '\n'})
		}
		file.Write([]byte{'('})

		// biz_id
		file.WriteString(strconv.Itoa(i + 1))
		file.Write([]byte{','})

		// biz
		file.WriteString(`"test"`)
		file.Write([]byte{','})

		// read_cnt
		file.WriteString(strconv.Itoa(int(rand.Int31n(10000))))
		file.Write([]byte{','})

		// favorite_cnt
		file.WriteString(strconv.Itoa(int(rand.Int31n(10000))))
		file.Write([]byte{','})

		// like_cnt
		file.WriteString(strconv.Itoa(int(rand.Int31n(10000))))
		file.Write([]byte{','})

		// ctime
		file.WriteString(strconv.FormatInt(now, 10))
		file.Write([]byte{','})

		// utime
		file.WriteString(strconv.FormatInt(now, 10))

		file.Write([]byte{')'})
	}
}

func TestGenData(t *testing.T) {
	db := startup.InitTestDB()
	for i := 0; i < 10; i++ {
		const batchSize = 100
		data := make([]dao.Interactive, 0, batchSize)
		now := time.Now().UnixMilli()
		for j := 0; j < batchSize; j++ {
			data = append(data, dao.Interactive{
				Biz:         "test",
				BizId:       int64(i*batchSize + j + 1),
				ReadCnt:     rand.Int63(),
				LikeCnt:     rand.Int63(),
				FavoriteCnt: rand.Int63(),
				Utime:       now,
				Ctime:       now,
			})
		}
		err := db.Transaction(func(tx *gorm.DB) error {
			return tx.Create(data).Error
		})
		require.NoError(t, err)
	}
}
