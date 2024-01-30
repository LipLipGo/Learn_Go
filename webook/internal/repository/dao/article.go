package dao

import (
	"context"
	"errors"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
	"time"
)

type ArticleDAO interface {
	Create(ctx context.Context, art Article) (int64, error)
	UpdateById(ctx context.Context, entity Article) error
	Sync(ctx context.Context, entity Article) (int64, error)
}

type ArticleGORMDAO struct {
	db *gorm.DB
}

// Sync 闭包写法，在 Transaction 中会自动开启事务，回滚和提交等操作
func (a *ArticleGORMDAO) Sync(ctx context.Context, art Article) (int64, error) {
	var id = art.Id
	err := a.db.Transaction(func(tx *gorm.DB) error {
		var err error

		dao := NewArticleGORMDAO(tx)
		if id > 0 {
			err = dao.UpdateById(ctx, art)
		} else {
			id, err = dao.Create(ctx, art)
		}
		if err != nil {
			return err
		}

		art.Id = id
		// 这里操作线上库
		pubArt := PublishedArticle(art)
		now := time.Now().UnixMilli()
		pubArt.Ctime = now
		pubArt.Utime = now
		err = tx.Clauses(clause.OnConflict{
			// 这一配置对 mysql 不起效，但是可以兼容其它方言
			Columns: []clause.Column{{Name: "id"}},
			// 如果这里使用的是 mysql ，只需要设置 DoUpdates
			DoUpdates: clause.Assignments(map[string]interface{}{
				// 如果冲突了，那么就更新数据
				"title":   pubArt.Title,
				"content": pubArt.Content,
				"utime":   now,
			}),
			// 如果不冲突，就创建数据
		}).Create(&pubArt).Error
		return err
	})

	return id, err
}

// SyncV1 自己管理事务的写法
func (a *ArticleGORMDAO) SyncV1(ctx context.Context, art Article) (int64, error) {
	tx := a.db.WithContext(ctx).Begin()
	if tx.Error != nil {
		return 0, tx.Error
	}
	// 防止后面的业务 panic，占用连接，回滚
	defer tx.Rollback()
	var (
		id  = art.Id
		err error
	)
	dao := NewArticleGORMDAO(tx)
	if id > 0 {
		err = dao.UpdateById(ctx, art)
	} else {
		id, err = dao.Create(ctx, art)
	}
	if err != nil {
		return 0, err
	}
	art.Id = id
	// 这里操作线上库
	pubArt := PublishedArticle(art)
	now := time.Now().UnixMilli()
	pubArt.Ctime = now
	pubArt.Utime = now
	err = tx.Clauses(clause.OnConflict{
		// 这一配置对 mysql 不起效，但是可以兼容其它方言
		Columns: []clause.Column{{Name: "id"}},
		// 如果这里使用的是 mysql ，只需要设置 DoUpdates
		DoUpdates: clause.Assignments(map[string]interface{}{
			// 如果冲突了，那么就更新数据
			"title":   pubArt.Title,
			"content": pubArt.Content,
			"utime":   now,
		}),
		// 如果不冲突，就创建数据
	}).Create(&pubArt).Error
	if err != nil {
		return 0, err
	}
	// 在者之前返回错误后，都会直接执行回滚操作，但是一旦 commit 了，再尝试回滚就会返回错误，不会执行回滚
	tx.Commit()
	return id, nil
}

func NewArticleGORMDAO(db *gorm.DB) ArticleDAO {
	return &ArticleGORMDAO{
		db: db,
	}

}

func (a *ArticleGORMDAO) UpdateById(ctx context.Context, art Article) error {
	now := time.Now().UnixMilli()
	res := a.db.WithContext(ctx).Model(&art).Where("id = ? AND author_id = ?", art.Id, art.AuthorId).Updates(map[string]any{
		"title":   art.Title,
		"content": art.Content,
		"utime":   now,
	})
	if res.Error != nil {
		return res.Error
	}
	if res.RowsAffected == 0 {
		// 这里不知道是 Id 不对还是 Author 不对，也不需要进行判定，普通用户进不来这里
		return errors.New("更新失败，作者不对或者Id不对")
	}
	return nil
}

func (a *ArticleGORMDAO) Create(ctx context.Context, art Article) (int64, error) {
	now := time.Now().UnixMilli()

	art.Utime = now
	art.Ctime = now
	err := a.db.WithContext(ctx).Create(&art).Error
	return art.Id, err // 自动会将自增组件 Id 填回 art
}

type Article struct {
	Id int64 `gorm:"primaryKey,autoIncrement"`
	// 创建时间
	Ctime int64

	// 更新时间
	Utime    int64
	Title    string `gorm:"type=varchar(4096)"`
	Content  string `gorm:"type=BLOB"`
	AuthorId int64  `gorm:"index"` // 这个索引是普通的索引
}

// 同库不同表

type PublishedArticle Article
