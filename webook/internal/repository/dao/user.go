package dao

import (
	"context"
	"errors"
	"github.com/gin-gonic/gin"
	"github.com/go-sql-driver/mysql"
	"gorm.io/gorm"
	"time"
)

type UserDao struct {
	db *gorm.DB
}

var (
	ErrDuplicateEmail = errors.New("邮箱冲突")
	ErrRecordNotFound = gorm.ErrRecordNotFound
)

func (dao *UserDao) Insert(ctx context.Context, u User) error {
	now := time.Now().UnixMilli()
	u.Ctime = now
	u.Utime = now

	err := dao.db.WithContext(ctx).Create(&u).Error

	if me, ok := err.(*mysql.MySQLError); ok { // 判断是否是数据库错误，邮箱唯一索引冲突
		const duplicateErr uint16 = 1062
		if me.Number == duplicateErr {
			// 用户冲突，邮箱冲突
			return ErrDuplicateEmail
		}
	}
	return err
}

func (dao *UserDao) FindByEmail(ctx context.Context, email string) (User, error) {
	var u User
	err := dao.db.WithContext(ctx).Where("email=?", email).First(&u).Error
	return u, err
}

func (dao *UserDao) UpdateById(ctx context.Context, entity User) error {
	return dao.db.WithContext(ctx).Model(&entity).Where("id = ?", entity.Id).Updates(map[string]any{
		"utime":    time.Now().UnixMilli(),
		"nickname": entity.Nickname,
		"about_me": entity.AboutMe,
		"birthday": entity.Birthday,
	}).Error
}

func (dao *UserDao) FindById(ctx *gin.Context, uid int64) (User, error) {
	var u User
	err := dao.db.WithContext(ctx).Where("id = ?", uid).First(&u).Error
	return u, err
}

func NewUserDao(db *gorm.DB) *UserDao {
	return &UserDao{
		db: db,
	}
}

type User struct {
	Id       int64  `gorm:"primaryKey,autoIncrement"` // 设置唯一索引；自增主键的优点：树单向增长，不存在叶分裂，适合于范围查询，充分利用操作系统的预读特性
	Email    string `gorm:"unique"`
	Password string

	// 这里时间使用 int64 ，是为了防止时区不一致问题，统一使用 UTC 0 的毫秒数，当需要将数据传给前端时再做对应处理
	// 创建时间
	Ctime int64

	// 更新时间
	Utime    int64
	Birthday int64
	AboutMe  string `gorm:"type=varchar(4096)"`
	Nickname string `gorm:"type=varchar(128)"`
}
