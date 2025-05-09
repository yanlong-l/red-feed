package repository

import (
	"context"
	"red-feed/internal/domain"
	"red-feed/internal/repository/dao"
)

type ArticleRepository interface {
	Create(ctx context.Context, article domain.Article) (id int64, err error)
}

type CachedArticleRepository struct {
	dao dao.ArticleDao
}

func (r *CachedArticleRepository) Create(ctx context.Context, article domain.Article) (id int64, err error) {
	return r.dao.Insert(ctx, dao.Article{
		Title:    article.Title,
		Content:  article.Content,
		AuthorId: article.Author.Id,
	})
}

func NewArticleRepository(dao dao.ArticleDao) ArticleRepository {
	return &CachedArticleRepository{
		dao: dao,
	}
}
