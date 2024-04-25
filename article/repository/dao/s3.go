package dao

import (
	"bytes"
	"context"
	"fmt"
	"time"

	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/tsukaychan/mercury/internal/domain"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

var _ ArticleDAO = (*S3DAO)(nil)

var statusPrivate = domain.ArticleStatusPrivate.ToUint8()

type S3DAO struct {
	oss *s3.Client
	GORMArticleDAO
	bucket *string
}

func (dao *S3DAO) Sync(ctx context.Context, atcl Article) (int64, error) {
	id := atcl.Id
	err := dao.db.Transaction(func(tx *gorm.DB) error {
		var err error
		now := time.Now().UnixMilli()
		txDAO := NewGORMArticleDAO(tx)
		if id == 0 {
			id, err = txDAO.Insert(ctx, atcl)
		} else {
			err = txDAO.UpdateById(ctx, atcl)
		}
		if err != nil {
			return err
		}
		atcl.Id = id
		publishArt := PublishedArticle(atcl)
		publishArt.Utime = now
		publishArt.Ctime = now
		publishArt.Content = ""
		return tx.Clauses(clause.OnConflict{
			// ID 冲突的时候。实际上，在 MYSQL 里面你写不写都可以
			Columns: []clause.Column{{Name: "id"}},
			DoUpdates: clause.Assignments(map[string]interface{}{
				"title": atcl.Title,
				//"content": atcl.Content,
				"utime":  now,
				"status": atcl.Status,
			}),
		}).Create(&publishArt).Error
	})
	if err != nil {
		return 0, err
	}
	key := fmt.Sprintf("mercury/%v", atcl.Id)
	contentType := "text/plain;charset=utf-8"
	_, err = dao.oss.PutObject(ctx, &s3.PutObjectInput{
		Bucket:      dao.bucket,
		Key:         &key,
		Body:        bytes.NewReader([]byte(atcl.Content)),
		ContentType: &contentType,
	})
	if err != nil {
		return 0, err
	}
	return id, err
}

func (dao *S3DAO) SyncStatus(ctx context.Context, id int64, authorId int64, status uint8) error {
	now := time.Now().UnixMilli()
	err := dao.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		res := tx.Model(&Article{}).Where("id = ? AND author_id = ?", id, authorId).Updates(map[string]any{
			"status": status,
			"utime":  now,
		})
		if res.Error != nil {
			return res.Error
		}
		if res.RowsAffected != 1 {
			return ErrPossibleIncorrectAuthor
		}

		res = tx.Model(&PublishedArticle{}).Where("id = ?", id).Updates(map[string]any{
			"status": status,
			"utime":  now,
		})
		if res.Error != nil {
			return res.Error
		}
		if res.RowsAffected != 1 {
			return ErrPossibleIncorrectAuthor
		}

		return nil
	})
	if err != nil {
		return err
	}
	key := fmt.Sprintf("mercury/%v", id)
	if status == statusPrivate {
		_, err = dao.oss.DeleteObject(ctx, &s3.DeleteObjectInput{
			Bucket: dao.bucket,
			Key:    &key,
		})
	}
	return err
}
