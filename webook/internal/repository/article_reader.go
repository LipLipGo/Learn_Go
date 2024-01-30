package repository

import (
	"Learn_Go/webook/internal/domain"
	"context"
)

type ArticleReaderRepository interface {
	// Save 这里因为我们不确定线上库中是否已经有数据，那么无则插入，有则更新，所以将语义统一为 Save
	Save(ctx context.Context, art domain.Article) error // 这里其实不需要返回 Id ，因为我们要保证线上库的 Id 和制作库的 Id 保持一致，所以直接拿到制作库的 Id 就可以
}
