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
	WithDraw(ctx *gin.Context, article domain.Article) error
	List(ctx *gin.Context, uid int64, offset int, limit int) ([]domain.Article, error)
}

type articleService struct {
	repo repository.ArticleRepository
}

func (s *articleService) List(ctx *gin.Context, uid int64, offset int, limit int) ([]domain.Article, error) {
	return s.repo.List(ctx, uid, offset, limit)
}

func (s *articleService) WithDraw(ctx *gin.Context, article domain.Article) error {
	return s.repo.SyncStatus(ctx, article.Id, article.Author.Id, domain.ArticleStatusPrivate)
}

func (s *articleService) Publish(ctx *gin.Context, article domain.Article) (int64, error) {
	article.Status = domain.ArticleStatusPublished
	return s.repo.Sync(ctx, article)
}

func (s *articleService) Save(ctx context.Context, article domain.Article) (id int64, err error) {
	article.Status = domain.ArticleStatusUnPublished
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
