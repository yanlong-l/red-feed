package dao

import (
	"context"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
	"time"
)

type InteractiveDAO interface {
	IncrReadCnt(ctx context.Context, biz string, bizId int64) error
	InsertLikeInfo(ctx context.Context, biz string, bizId, uId int64) error
	DeleteLikeInfo(ctx context.Context, biz string, bizId, uId int64) error
}

type GORMInteractiveDAO struct {
	db *gorm.DB
}

func NewInteractiveDAO(db *gorm.DB) InteractiveDAO {
	return &GORMInteractiveDAO{
		db: db,
	}

}

func (d *GORMInteractiveDAO) InsertLikeInfo(ctx context.Context, biz string, bizId, uId int64) error {
	return nil
	//now := time.Now().UnixMilli()
	//d.db.Transaction(func(tx *gorm.DB) error {
	//	// 开启事务，同时操作 Interactive表 和 UserLikeBiz 表
	//	err := tx.WithContext(ctx).Clauses(clause.OnConflict{
	//		Columns:
	//		DoUpdates: clause.Assignments(map[string]any{
	//			"read_cnt": gorm.Expr("read_cnt + 1"),
	//			"utime":    time.Now().UnixMilli(),
	//		}),
	//	}).Create(&Interactive{
	//		Biz:     biz,
	//		BizId:   bizId,
	//		ReadCnt: 1,
	//		Ctime:   now,
	//		Utime:   now,
	//	}).Error
	//})
}

func (d *GORMInteractiveDAO) DeleteLikeInfo(ctx context.Context, biz string, bizId, uId int64) error {
	//TODO implement me
	panic("implement me")
}

func (d *GORMInteractiveDAO) IncrReadCnt(ctx context.Context, biz string, bizId int64) error {
	now := time.Now().UnixMilli()
	return d.db.WithContext(ctx).Clauses(clause.OnConflict{
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

type UserLikeBiz struct {
	Id     int64  `gorm:"primaryKey,autoIncrement"`
	Biz    string `gorm:"uniqueIndex:uid_biz_id_type;type:varchar(128)"`
	BizId  int64  `gorm:"uniqueIndex:uid_biz_id_type"`
	Uid    int64  `gorm:"uniqueIndex:uid_biz_id_type"`
	Ctime  int64
	Utime  int64
	Status uint8
}

// Collection 收藏夹
type Collection struct {
	Id    int64  `gorm:"primaryKey,autoIncrement"`
	Name  string `gorm:"type=varchar(1024)"`
	Uid   int64  `gorm:""`
	Ctime int64
	Utime int64
}
