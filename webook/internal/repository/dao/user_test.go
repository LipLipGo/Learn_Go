package dao

import (
	"context"
	"database/sql"
	"errors"
	"github.com/DATA-DOG/go-sqlmock"
	mysqlDriver "github.com/go-sql-driver/mysql"
	"github.com/stretchr/testify/assert"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
	"testing"
)

func TestGORMUserDao_Insert(t *testing.T) {
	testCases := []struct {
		name    string
		mock    func(t *testing.T) *sql.DB
		ctx     context.Context
		user    User
		wantErr error
	}{
		{
			name: "插入成功",
			mock: func(t *testing.T) *sql.DB {
				db, mock, err := sqlmock.New()
				assert.NoError(t, err)
				mockRes := sqlmock.NewResult(123, 123)
				// 这边要求传入sql的正则表达式，并且需要返回一个结果集
				mock.ExpectExec("INSERT INTO .*").WillReturnResult(mockRes)

				return db
			},
			ctx: context.Background(),
			user: User{
				Nickname: "lip",
			},
			wantErr: nil,
		},
		{
			name: "邮箱冲突",
			mock: func(t *testing.T) *sql.DB {
				db, mock, err := sqlmock.New()
				assert.NoError(t, err)
				mock.ExpectExec("INSERT INTO .*").WillReturnError(&mysqlDriver.MySQLError{Number: 1062})

				return db
			},
			ctx: context.Background(),
			user: User{
				Nickname: "lip",
			},
			wantErr: ErrDuplicateEmail,
		},
		{
			name: "插入失败",
			mock: func(t *testing.T) *sql.DB {
				db, mock, err := sqlmock.New()
				assert.NoError(t, err)
				mock.ExpectExec("INSERT INTO .*").WillReturnError(errors.New("数据库错误"))

				return db
			},
			ctx: context.Background(),
			user: User{
				Nickname: "lip",
			},
			wantErr: errors.New("数据库错误"),
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			sqlDB := tc.mock(t)
			db, err := gorm.Open(mysql.New(mysql.Config{
				Conn:                      sqlDB,
				SkipInitializeWithVersion: true, // 不要检测mysql版本，因为mock的sql做不到
			}), &gorm.Config{
				DisableAutomaticPing:   true, // 不要自动发ping
				SkipDefaultTransaction: true, // 不要自动添加commit机制
			})
			assert.NoError(t, err)
			dao := NewGORMUserDao(db)
			err = dao.Insert(tc.ctx, tc.user)
			assert.Equal(t, tc.wantErr, err)
		})
	}
}
