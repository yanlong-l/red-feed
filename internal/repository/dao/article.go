package dao

import (
	"context"
	"fmt"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
	"time"
)

type ArticleDao interface {
	Insert(ctx context.Context, art Article) (int64, error)
	Update(ctx context.Context, art Article) error
	Sync(ctx context.Context, art Article) (int64, error)
	Upsert(ctx context.Context, art PublishArticle) error
}

func NewGORMArticleDao(db *gorm.DB) ArticleDao {
	return &GORMArticleDao{db: db}
}

type GORMArticleDao struct {
	db *gorm.DB
}

func (d *GORMArticleDao) Upsert(ctx context.Context, art PublishArticle) error {
	now := time.Now().UnixMilli()
	art.Ctime = now
	art.Utime = now
	err := d.db.WithContext(ctx).Clauses(clause.OnConflict{
		Columns: []clause.Column{{Name: "id"}},
		DoUpdates: clause.Assignments(map[string]interface{}{
			"title":   art.Title,
			"content": art.Content,
			"utime":   art.Utime,
		}),
	}).Create(&art).Error
	return err
}

func (d *GORMArticleDao) Sync(ctx context.Context, art Article) (int64, error) {
	var (
		id = art.Id
	)
	// 同步制作库和线上库，需要开启事务
	err := d.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		txDao := NewGORMArticleDao(tx)
		var err error
		if id > 0 {
			err = txDao.Update(ctx, art)
		} else {
			id, err = txDao.Insert(ctx, art)
		}
		if err != nil {
			return err
		}
		// 继续操作制作库
		return d.Upsert(ctx, PublishArticle{Article: art})
	})
	return id, err
}

func (d *GORMArticleDao) Update(ctx context.Context, art Article) error {
	now := time.Now().UnixMilli()
	art.Utime = now
	// 依赖 gorm 忽略零值的特性，会用主键进行更新 可读性很差
	res := d.db.WithContext(ctx).Model(&art).
		Where("id=? AND author_id = ?", art.Id, art.AuthorId).
		Updates(map[string]any{
			"title":   art.Title,
			"content": art.Content,
			"utime":   art.Utime,
		})
	if res.Error != nil {
		return res.Error
	}
	if res.RowsAffected == 0 {
		// 补充一点日志
		return fmt.Errorf("更新失败，可能是创作者非法 id %d, author_id %d",
			art.Id, art.AuthorId)
	}
	return nil
}

func (d *GORMArticleDao) Insert(ctx context.Context, art Article) (int64, error) {
	now := time.Now().UnixMilli()
	art.Ctime = now
	art.Utime = now
	err := d.db.WithContext(ctx).Create(&art).Error
	return art.Id, err
}

type Article struct {
	Id       int64  `gorm:"primaryKey,autoIncrement"`
	Title    string `gorm:"type=varchar(1024)"`
	Content  string `gorm:"type=BLOB"`
	AuthorId int64  `gorm:"index"`
	Ctime    int64
	Utime    int64
}

type PublishArticle struct {
	Article
}
