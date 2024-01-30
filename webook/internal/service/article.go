package service

import (
	"Learn_Go/webook/internal/domain"
	"Learn_Go/webook/internal/repository"
	"Learn_Go/webook/pkg/logger"
	"context"
	"errors"
)

type ArticleService interface {
	Save(ctx context.Context, art domain.Article) (int64, error)
	Publish(ctx context.Context, art domain.Article) (int64, error)
}

type articleService struct {
	repo repository.ArticleRepository

	// V1写法专用
	readerRepo repository.ArticleReaderRepository
	authorRepo repository.ArticleAuthorRepository
	l          logger.LoggerV1
}

func NewArticleService(repo repository.ArticleRepository) ArticleService {
	return &articleService{
		repo: repo,
	}
}

func NewArticleServiceV1(authorRepo repository.ArticleAuthorRepository, readerRepo repository.ArticleReaderRepository, l logger.LoggerV1) *articleService {
	return &articleService{
		readerRepo: readerRepo,
		authorRepo: authorRepo,
		l:          l,
	}
}

func (a *articleService) Publish(ctx context.Context, art domain.Article) (int64, error) {
	return a.repo.Sync(ctx, art)

}

func (a *articleService) PublishV1(ctx context.Context, art domain.Article) (int64, error) {
	// 先操作制作库
	// 再操作线上库
	var (
		id  = art.Id
		err error
	)
	if art.Id > 0 {
		err = a.authorRepo.Update(ctx, art)

	} else {
		id, err = a.authorRepo.Create(ctx, art)
	}
	if err != nil {
		return 0, err
	}
	art.Id = id

	// 这里可以考虑失败重试
	for i := 0; i < 3; i++ {
		err = a.readerRepo.Save(ctx, art)
		if err != nil {
			a.l.Error("保存到线上库失败",
				logger.Field{Key: "art_id", Value: art.Id},
				logger.Error(err))
		} else {
			return id, nil
		}
	}
	a.l.Error("保存到线上库失败，重试次数耗尽",
		logger.Field{Key: "art_id", Value: art.Id},
		logger.Error(err))
	return id, errors.New("保存到线上库失败，重试次数耗尽")
}

func (a *articleService) Save(ctx context.Context, art domain.Article) (int64, error) {
	if art.Id > 0 {
		err := a.repo.Update(ctx, art)
		return art.Id, err
	} else {
		return a.repo.Create(ctx, art)

	}
}
