package service

import (
	"context"
	"github.com/gin-gonic/gin"
	"red-feed/internal/domain"
	"red-feed/internal/repository"
)

type ArticleService interface {
	Save(ctx context.Context, article domain.Article) (id int64, err error)
	Publish(ctx *gin.Context, article domain.Article) (int64, error)
}

type articleService struct {
	repo repository.ArticleRepository
}

func (s *articleService) Publish(ctx *gin.Context, article domain.Article) (int64, error) {
	return 1, nil
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
