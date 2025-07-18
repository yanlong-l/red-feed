package dao

import (
	"context"
	"errors"
	"fmt"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
	"red-feed/internal/domain"
	"time"
)

type ArticleDao interface {
	Insert(ctx context.Context, art Article) (int64, error)
	Update(ctx context.Context, art Article) error
	Sync(ctx context.Context, art Article) (int64, error)
	SyncStatus(ctx context.Context, artId int64, authorId int64, status uint8) error
	List(ctx context.Context, authorId int64, offset int, limit int) ([]Article, error)
	ListPub(ctx context.Context, offset int, limit int) ([]PublishedArticle, error)
	GetById(ctx context.Context, artId int64) (Article, error)
	GetPubById(ctx context.Context, artId int64) (PublishedArticle, error)
	ListPubForRanking(ctx context.Context, start time.Time, offset int, limit int) ([]PublishedArticle, error)
}

func NewGORMArticleDao(db *gorm.DB) ArticleDao {
	return &GORMArticleDao{db: db}
}

type GORMArticleDao struct {
	db *gorm.DB
}

func (d *GORMArticleDao) ListPubForRanking(ctx context.Context, start time.Time, offset int, limit int) ([]PublishedArticle, error) {
	var res []PublishedArticle
	err := d.db.WithContext(ctx).
		Where("utime > ?", start.UnixMilli()).
		Order("utime DESC").Offset(offset).Limit(limit).Find(&res).Error
	return res, err
}

func (d *GORMArticleDao) ListPub(ctx context.Context, offset int, limit int) ([]PublishedArticle, error) {
	var arts = make([]PublishedArticle, 0)
	err := d.db.WithContext(ctx).Model(&PublishedArticle{}).
		Where("status = ?", domain.ArticleStatusPublished.ToUint8()).
		Offset(offset).
		Limit(limit).
		Order("utime DESC").
		Find(&arts).Error
	return arts, err
}

func (d *GORMArticleDao) GetById(ctx context.Context, artId int64) (Article, error) {
	var art Article
	err := d.db.WithContext(ctx).Model(&Article{}).
		Where("id = ?", artId).
		First(&art).Error
	return art, err
}

func (d *GORMArticleDao) GetPubById(ctx context.Context, artId int64) (PublishedArticle, error) {
	var art PublishedArticle
	err := d.db.WithContext(ctx).Model(&Article{}).
		Where("id = ?", artId).
		First(&art).Error
	return art, err
}

func (d *GORMArticleDao) List(ctx context.Context, authorId int64, offset int, limit int) ([]Article, error) {
	var arts = make([]Article, 0)
	err := d.db.WithContext(ctx).Model(&Article{}).
		Where("author_id = ?", authorId).
		Offset(offset).
		Limit(limit).
		Order("utime DESC").
		Find(&arts).Error
	return arts, err
}

func (d *GORMArticleDao) SyncStatus(ctx context.Context, artId int64, authorId int64, status uint8) error {
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
