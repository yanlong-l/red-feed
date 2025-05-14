package dao

import (
	"context"
	"errors"
	"fmt"
	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
	"time"
)

type ArticleDao interface {
	Insert(ctx context.Context, art Article) (int64, error)
	Update(ctx context.Context, art Article) error
	Sync(ctx context.Context, art Article) (int64, error)
	Upsert(ctx context.Context, art PublishedArticle) error
	SyncStatus(ctx *gin.Context, artId int64, authorId int64, status uint8) error
}

func NewGORMArticleDao(db *gorm.DB) ArticleDao {
	return &GORMArticleDao{db: db}
}

type GORMArticleDao struct {
	db *gorm.DB
}

func (d *GORMArticleDao) SyncStatus(ctx *gin.Context, artId int64, authorId int64, status uint8) error {
	// 同步线上库和制作库对应帖子的status
	now := time.Now().UnixMilli()
	var art Article
	var pubArt PublishedArticle
	return d.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		res := d.db.WithContext(ctx).Model(&art).
			Where("id=? AND author_id = ?", artId, authorId).
			Updates(map[string]any{
				"utime":  now,
				"status": status,
			})
		if res.Error != nil {
			return res.Error
		}
		if res.RowsAffected == 0 {
			return errors.New("系统错误")
		}
		return d.db.WithContext(ctx).Model(&pubArt).
			Where("id=? AND author_id = ?", artId, authorId).
			Updates(map[string]any{
				"utime":  now,
				"status": art.Status,
			}).Error
	})
}

func (d *GORMArticleDao) Upsert(ctx context.Context, art PublishedArticle) error {
	now := time.Now().UnixMilli()
	art.Ctime = now
	art.Utime = now
	err := d.db.WithContext(ctx).Clauses(clause.OnConflict{
		Columns: []clause.Column{{Name: "id"}},
		DoUpdates: clause.Assignments(map[string]interface{}{
			"title":   art.Title,
			"content": art.Content,
			"utime":   art.Utime,
			"status":  art.Status,
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
		return d.Upsert(ctx, PublishedArticle{Article: art})
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
			"status":  art.Status,
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
	Status   uint8
	Ctime    int64
	Utime    int64
}

type PublishedArticle struct {
	Article
}
