package dao

import (
	"context"
	"gorm.io/gorm"
)

type ArticleReaderDAO interface {
	// Upsert Insert Or Update
	Upsert(ctx context.Context, art Article) error
	UpsertV2(ctx context.Context, art PublishedArticle) error // 同库不同表，这里需要换一张表
}

type ArticleGORMReaderDAO struct {
	db *gorm.DB
}

func (a *ArticleGORMReaderDAO) UpsertV2(ctx context.Context, art PublishedArticle) error {
	//TODO implement me
	panic("implement me")
}

func (a *ArticleGORMReaderDAO) Upsert(ctx context.Context, art Article) error {
	//TODO implement me
	panic("implement me")
}

func NewArticleGORMReaderDAO(db *gorm.DB) ArticleReaderDAO {
	return &ArticleGORMReaderDAO{
		db: db,
	}

}
