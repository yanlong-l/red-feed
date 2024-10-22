package dao

import (
	"context"
	"database/sql"
	"errors"
	"github.com/gin-gonic/gin"
	"time"

	"github.com/go-sql-driver/mysql"
	"gorm.io/gorm"
)

var (
	ErrUserDuplicate = errors.New("邮箱冲突")
	ErrUserNotFound  = gorm.ErrRecordNotFound
)

type GORMUserDAO interface {
	FindByEmail(ctx context.Context, email string) (User, error)
	FindById(ctx context.Context, id int64) (User, error)
	FindByPhone(ctx *gin.Context, phone string) (User, error)
	Insert(ctx context.Context, u User) error
}

type UserDAO struct {
	db *gorm.DB
}

func NewUserDAO(db *gorm.DB) GORMUserDAO {
	return &UserDAO{
		db: db,
	}
}

func (dao *UserDAO) FindByEmail(ctx context.Context, email string) (User, error) {
	var user User
	result := dao.db.WithContext(ctx).Where("email = ?", email).First(&user)
	return user, result.Error
}

func (dao *UserDAO) FindById(ctx context.Context, id int64) (User, error) {
	var user User
	result := dao.db.WithContext(ctx).Where("id = ?", id).First(&user)
	return user, result.Error
}

func (dao *UserDAO) FindByPhone(ctx *gin.Context, phone string) (User, error) {
	var user User
	result := dao.db.WithContext(ctx).Where("phone = ?", phone).First(&user)
	return user, result.Error
}

func (dao *UserDAO) Insert(ctx context.Context, u User) error {
	// 存毫秒数
	now := time.Now().UnixMilli()
	u.Utime = now
	u.Ctime = now
	err := dao.db.WithContext(ctx).Create(&u).Error
	if mysqlErr, ok := err.(*mysql.MySQLError); ok {
		const uniqueConflictsErrNo uint16 = 1062
		if mysqlErr.Number == uniqueConflictsErrNo {
			// 邮箱冲突 或 手机号码冲突
			return ErrUserDuplicate
		}
	}
	return err
}

// User 直接对应数据库表结构
// 有些人叫做 entity，有些人叫做 model，有些人叫做 PO(persistent object)
type User struct {
	Id int64 `gorm:"primaryKey,autoIncrement"`
	// 全部用户唯一
	Email    sql.NullString `gorm:"unique"`
	Phone    sql.NullString `gorm:"unique"`
	Password string
	// 创建时间，毫秒数
	Ctime int64
	// 更新时间，毫秒数
	Utime int64
}
