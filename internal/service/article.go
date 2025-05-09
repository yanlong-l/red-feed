package service

import (
	"context"
	"red-feed/internal/domain"
	"red-feed/internal/repository"
)

type ArticleService interface {
	Save(ctx context.Context, article domain.Article) (id int64, err error)
}

type articleService struct {
	repo repository.ArticleRepository
}

func (s *articleService) Save(ctx context.Context, article domain.Article) (id int64, err error) {
	if article.Id != 0 {
		return article.Id, s.repo.Update(ctx, article)
	}
	return s.repo.Create(ctx, article)
}

func NewArticleService(repo repository.ArticleRepository) ArticleService {
	return &articleService{
		repo: repo,
	}
}
