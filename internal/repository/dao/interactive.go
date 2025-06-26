package dao

import (
	"context"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
	"time"
)

var ErrDataNotFound = gorm.ErrRecordNotFound

type InteractiveDAO interface {
	IncrReadCnt(ctx context.Context, biz string, bizId int64) error
	BatchIncrReadCnt(ctx context.Context, bizs []string, bizIds []int64) error

	GetLikeInfo(ctx context.Context, biz string, bizId, uid int64) (UserLikeBiz, error)
	InsertLikeInfo(ctx context.Context, biz string, bizId, uId int64) error
	DeleteLikeInfo(ctx context.Context, biz string, bizId, uId int64) error

	InsertCollectInfo(ctx context.Context, biz string, bizId int64, uId, cId int64) error
	GetCollectionInfo(ctx context.Context, biz string, bizId, uid int64) (UserCollectionBiz, error)
	DeleteCollectInfo(ctx context.Context, biz string, bizId int64, uId, cId int64) error

	Get(ctx context.Context, biz string, bizId int64) (Interactive, error)
}

type GORMInteractiveDAO struct {
	db *gorm.DB
}

func NewInteractiveDAO(db *gorm.DB) InteractiveDAO {
	return &GORMInteractiveDAO{
		db: db,
	}
}

func (d *GORMInteractiveDAO) Get(ctx context.Context, biz string, bizId int64) (Interactive, error) {
	return Interactive{}, nil
}

func (d *GORMInteractiveDAO) GetLikeInfo(ctx context.Context, biz string, bizId, uid int64) (UserLikeBiz, error) {
	var res UserLikeBiz
	err := d.db.WithContext(ctx).First(&res, "biz = ? AND biz_id = ? AND uid = ?", biz, bizId, uid).Error
	return res, err
}

func (d *GORMInteractiveDAO) GetCollectionInfo(ctx context.Context, biz string, bizId, uid int64) (UserCollectionBiz, error) {
	var res UserCollectionBiz
	err := d.db.WithContext(ctx).First(&res, "biz = ? AND biz_id = ? AND uid = ?", biz, bizId, uid).Error
	return res, err
}

func (d *GORMInteractiveDAO) DeleteCollectInfo(ctx context.Context, biz string, bizId int64, uId int64, cId int64) error {
	now := time.Now().UnixMilli()
	return d.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		// 删除收藏记录
		err := tx.Model(&UserCollectionBiz{}).
			Where("biz=? AND biz_id = ? AND uid = ? AND cid = ?", biz, bizId, uId, cId).
			Updates(map[string]any{
				"utime":  now,
				"status": 0,
			}).Error
		if err != nil {
			return err
		}
		// 减收藏数量
		return tx.Model(&Interactive{}).
			Where("biz=? AND biz_id = ?", biz, bizId).
			Updates(map[string]any{
				"utime":       now,
				"collect_cnt": gorm.Expr("collect_cnt-1"),
			}).Error
	})
}

func (d *GORMInteractiveDAO) InsertCollectInfo(ctx context.Context, biz string, bizId int64, uId int64, cId int64) error {
	now := time.Now().UnixMilli()
	return d.db.Transaction(func(tx *gorm.DB) error {
		// 开启事务，同时操作 Interactive表 和 UserCollectBiz 表
		err := tx.WithContext(ctx).Clauses(clause.OnConflict{
			DoUpdates: clause.Assignments(map[string]any{
				"collect_cnt": gorm.Expr("collect_cnt + 1"),
				"utime":       time.Now().UnixMilli(),
			}),
		}).Create(&Interactive{
			Biz:        biz,
			BizId:      bizId,
			CollectCnt: 1,
			Ctime:      now,
			Utime:      now,
		}).Error
		if err != nil {
			return err
		}
		return tx.Clauses(clause.OnConflict{
			DoUpdates: clause.Assignments(map[string]any{
				"utime":  now,
				"status": 1,
			}),
		}).Create(&UserCollectionBiz{
			Biz:    biz,
			BizId:  bizId,
			Uid:    uId,
			Cid:    cId,
			Status: 1,
			Ctime:  now,
			Utime:  now,
		}).Error
	})
}

func (d *GORMInteractiveDAO) InsertLikeInfo(ctx context.Context, biz string, bizId, uId int64) error {
	now := time.Now().UnixMilli()
	return d.db.Transaction(func(tx *gorm.DB) error {
		// 开启事务，同时操作 Interactive表 和 UserLikeBiz 表
		err := tx.WithContext(ctx).Clauses(clause.OnConflict{
			DoUpdates: clause.Assignments(map[string]any{
				"like_cnt": gorm.Expr("like_cnt + 1"),
				"utime":    time.Now().UnixMilli(),
			}),
		}).Create(&Interactive{
			Biz:     biz,
			BizId:   bizId,
			ReadCnt: 1,
			Ctime:   now,
			Utime:   now,
		}).Error
		if err != nil {
			return err
		}
		return tx.Clauses(clause.OnConflict{
			DoUpdates: clause.Assignments(map[string]any{
				"utime":  now,
				"status": 1,
			}),
		}).Create(&UserLikeBiz{
			Biz:    biz,
			BizId:  bizId,
			Uid:    uId,
			Status: 1,
			Ctime:  now,
			Utime:  now,
		}).Error
	})
}

func (d *GORMInteractiveDAO) DeleteLikeInfo(ctx context.Context, biz string, bizId, uId int64) error {
	now := time.Now().UnixMilli()
	// 控制事务超时
	return d.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		// 两个操作
		// 一个是软删除点赞记录
		// 一个是减点赞数量
		err := tx.Model(&UserLikeBiz{}).
			Where("biz=? AND biz_id = ? AND uid = ?", biz, bizId, uId).
			Updates(map[string]any{
				"utime":  now,
				"status": 0,
			}).Error
		if err != nil {
			return err
		}
		return tx.Model(&Interactive{}).
			// 这边命中了索引，然后没找到，所以不会加锁
			Where("biz=? AND biz_id = ?", biz, bizId).
			Updates(map[string]any{
				"utime":    now,
				"like_cnt": gorm.Expr("like_cnt-1"),
			}).Error
	})
}

func (d *GORMInteractiveDAO) IncrReadCnt(ctx context.Context, biz string, bizId int64) error {
	return d.incrReadCnt(d.db.WithContext(ctx), biz, bizId)
}

func (d *GORMInteractiveDAO) incrReadCnt(tx *gorm.DB, biz string, bizId int64) error {
	now := time.Now().UnixMilli()
	return tx.Clauses(clause.OnConflict{
		// MySQL 不写
		//Columns:
		DoUpdates: clause.Assignments(map[string]any{
			"read_cnt": gorm.Expr("read_cnt + 1"),
			"utime":    time.Now().UnixMilli(),
		}),
	}).Create(&Interactive{
		Biz:     biz,
		BizId:   bizId,
		ReadCnt: 1,
		Ctime:   now,
		Utime:   now,
	}).Error
}

func (d *GORMInteractiveDAO) BatchIncrReadCnt(ctx context.Context, bizs []string, bizIds []int64) error {
	return d.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		// 让调用者保证两者是相等的
		for i := 0; i < len(bizs); i++ {
			err := d.incrReadCnt(tx, bizs[i], bizIds[i])
			if err != nil {
				return err
			}
		}
		return nil
	})
}

// Interactive 互动信息表
type Interactive struct {
	Id         int64  `gorm:"primaryKey,autoIncrement"`
	BizId      int64  `gorm:"uniqueIndex:biz_id_type"`
	Biz        string `gorm:"uniqueIndex:biz_id_type;type:varchar(128)"`
	ReadCnt    int64
	LikeCnt    int64
	CollectCnt int64
	Ctime      int64
	Utime      int64
}

// UserLikeBiz 用户点赞
type UserLikeBiz struct {
	Id     int64  `gorm:"primaryKey,autoIncrement"`
	Biz    string `gorm:"uniqueIndex:uid_biz_id_type;type:varchar(128)"`
	BizId  int64  `gorm:"uniqueIndex:uid_biz_id_type"`
	Uid    int64  `gorm:"uniqueIndex:uid_biz_id_type"`
	Ctime  int64
	Utime  int64
	Status uint8 // 1 点赞 0 取消点赞
}

// UserCollectionBiz 用户收藏
type UserCollectionBiz struct {
	Id     int64  `gorm:"primaryKey,autoIncrement"`
	Cid    int64  `gorm:"index"` // 收藏夹 ID 作为关联关系中的外键，我们这里需要索引
	BizId  int64  `gorm:"uniqueIndex:biz_type_id_uid"`
	Biz    string `gorm:"type:varchar(128);uniqueIndex:biz_type_id_uid"`
	Uid    int64  `gorm:"uniqueIndex:biz_type_id_uid"` // 只需要在 这算是一个冗余，因为正常来说， Collection 中维持住 Uid 就可以
	Ctime  int64
	Utime  int64
	Status uint8 // 1 收藏 0 取消收藏
}

// Collection 收藏夹
type Collection struct {
	Id    int64  `gorm:"primaryKey,autoIncrement"`
	Name  string `gorm:"type=varchar(1024)"`
	Uid   int64  `gorm:""`
	Ctime int64
	Utime int64
}
