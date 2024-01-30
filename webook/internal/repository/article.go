package repository

import (
	"Learn_Go/webook/internal/domain"
	"Learn_Go/webook/internal/repository/dao"
	"context"
	"gorm.io/gorm"
)

type ArticleRepository interface {
	Create(ctx context.Context, art domain.Article) (int64, error)
	Update(ctx context.Context, art domain.Article) error
	Sync(ctx context.Context, art domain.Article) (int64, error)
}

type CachedArticleRepository struct {
	dao       dao.ArticleDAO
	readerDAO dao.ArticleReaderDAO
	authorDAO dao.ArticleAuthorDAO
	db        *gorm.DB
}

func (c *CachedArticleRepository) Sync(ctx context.Context, art domain.Article) (int64, error) {
	return c.dao.Sync(ctx, c.toEntity(art))
}

// SyncV2 基于事务的实现，在 repository 层引入事务，必须知道 DAO 层是使用关系型数据库，并且是同库不同表
func (c *CachedArticleRepository) SyncV2(ctx context.Context, art domain.Article) (int64, error) {
	// 开启一个事务
	tx := c.db.WithContext(ctx).Begin()
	if tx.Error != nil {
		return 0, tx.Error
	}
	// 防止后面的业务 panic，占用连接，回滚
	defer tx.Rollback()
	authorDAO := dao.NewArticleGORMAuthorDAO(tx)
	readerDAO := dao.NewArticleGORMReaderDAO(tx)

	artn := c.toEntity(art)
	var (
		id  = art.Id
		err error
	)
	if id > 0 {
		err = authorDAO.Update(ctx, artn)
	} else {
		id, err = authorDAO.Create(ctx, artn)
	}
	if err != nil {
		return 0, err
	}

	artn.Id = id
	err = readerDAO.UpsertV2(ctx, dao.PublishedArticle(artn))

	if err != nil {
		return 0, err
	}
	// 在者之前返回错误后，都会直接执行回滚操作，但是一旦 commit 了，再尝试回滚就会返回错误，不会执行回滚
	tx.Commit()
	return id, nil

}

// Sync 非事务实现
func (c *CachedArticleRepository) SyncV1(ctx context.Context, art domain.Article) (int64, error) {
	// 先操作制作库
	// 再操作线上库
	artn := c.toEntity(art)
	var (
		id  = art.Id
		err error
	)
	if id > 0 {
		err = c.authorDAO.Update(ctx, artn)
	} else {
		id, err = c.authorDAO.Create(ctx, artn)
	}
	if err != nil {
		return 0, err
	}

	artn.Id = id
	err = c.readerDAO.Upsert(ctx, artn)
	return id, err
}

func NewCachedArticleRepository(d dao.ArticleDAO) ArticleRepository {
	return &CachedArticleRepository{
		dao: d,
	}
}

func NewCachedArticleRepositoryV2(authorDAO dao.ArticleAuthorDAO, readerDAO dao.ArticleReaderDAO) *CachedArticleRepository {
	return &CachedArticleRepository{
		authorDAO: authorDAO,
		readerDAO: readerDAO,
	}
}

func (c *CachedArticleRepository) Update(ctx context.Context, art domain.Article) error {
	return c.dao.UpdateById(ctx, c.toEntity(art))
}

func (c *CachedArticleRepository) Create(ctx context.Context, art domain.Article) (int64, error) {
	return c.dao.Create(ctx, c.toEntity(art))
}

func (c *CachedArticleRepository) toEntity(art domain.Article) dao.Article {
	return dao.Article{
		Id:       art.Id,
		Title:    art.Title,
		Content:  art.Content,
		AuthorId: art.Author.Id,
	}

}
