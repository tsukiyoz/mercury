package dao

import (
	"context"
	"fmt"
	"github.com/bwmarrin/snowflake"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"time"
)

var _ ArticleDAO = (*MongoDBDAO)(nil)

type MongoDBDAO struct {
	client *mongo.Client
	// mdb for webook
	//mdb *mongo.Database
	// Production Library
	col *mongo.Collection
	// OnLive Library
	liveCol *mongo.Collection
	node    *snowflake.Node
}

func NewMongoDBDAO(mdb *mongo.Database, node *snowflake.Node) ArticleDAO {
	return &MongoDBDAO{
		//mdb:     mdb,
		col:     mdb.Collection("articles"),
		liveCol: mdb.Collection("published_articles"),
		node:    node,
	}
}

func InitCollections(db *mongo.Database) error {
	ctx, cancel := context.WithTimeout(context.Background(), time.Second*3)
	defer cancel()
	index := []mongo.IndexModel{
		{
			Keys:    bson.D{bson.E{Key: "id", Value: 1}},
			Options: options.Index().SetUnique(true),
		},
		{
			Keys: bson.D{bson.E{Key: "author_id", Value: 1},
				bson.E{Key: "ctime", Value: 1},
			},
			Options: options.Index(),
		},
	}
	_, err := db.Collection("articles").Indexes().
		CreateMany(ctx, index)
	if err != nil {
		return err
	}
	_, err = db.Collection("published_articles").Indexes().
		CreateMany(ctx, index)
	return err
}

func (dao *MongoDBDAO) Insert(ctx context.Context, atcl Article) (int64, error) {
	id := dao.node.Generate().Int64()
	now := time.Now().UnixMilli()
	atcl.Id = id
	atcl.Ctime, atcl.Utime = now, now
	_, err := dao.col.InsertOne(ctx, atcl)
	if err != nil {
		return 0, err
	}
	return id, nil
}

func (dao *MongoDBDAO) UpdateById(ctx context.Context, atcl Article) error {
	filter := bson.M{"id": atcl.Id, "author_id": atcl.AuthorId}
	update := bson.M{"$set": bson.M{
		"title":   atcl.Title,
		"content": atcl.Content,
		"status":  atcl.Status,
		"utime":   time.Now().UnixMilli(),
	},
	}
	res, err := dao.col.UpdateOne(ctx, filter, update)
	if err != nil {
		return err
	}
	if res.ModifiedCount == 0 {
		return fmt.Errorf("update article failed, perhaps invalid id: id %d author_id %d", atcl.Id, atcl.AuthorId)
	}

	return nil
}

func (dao *MongoDBDAO) GetByAuthor(ctx context.Context, author int64, offset, limit int) ([]Article, error) {
	//TODO implement me
	panic("implement me")
}

func (dao *MongoDBDAO) GetById(ctx context.Context, id int64) (Article, error) {
	//TODO implement me
	panic("implement me")
}

func (dao *MongoDBDAO) GetPubById(ctx context.Context, id int64) (PublishedArticle, error) {
	//TODO implement me
	panic("implement me")
}

func (dao *MongoDBDAO) Sync(ctx context.Context, atcl Article) (int64, error) {
	var (
		id  = atcl.Id
		err error
	)

	if id > 0 {
		err = dao.UpdateById(ctx, atcl)
	} else {
		id, err = dao.Insert(ctx, atcl)
	}
	if err != nil {
		return 0, err
	}

	atcl.Id = id
	now := time.Now().UnixMilli()
	atcl.Utime = now
	filter := bson.M{
		"id":        atcl.Id,
		"author_id": atcl.AuthorId,
	}
	updates := bson.M{
		"$set": PublishedArticle(atcl),
		"$setOnInsert": bson.M{
			"ctime": now,
		},
	}
	_, err = dao.liveCol.UpdateOne(ctx, filter, updates, options.Update().SetUpsert(true))
	if err != nil {
		return 0, err
	}
	return id, nil
}

func (dao *MongoDBDAO) SyncStatus(ctx context.Context, id, author int64, status uint8) error {
	//TODO implement me
	panic("implement me")
}

func (dao *MongoDBDAO) ListPubByUtime(ctx context.Context, utime time.Time, offset int, limit int) ([]PublishedArticle, error) {
	//TODO implement me
	panic("implement me")
}
