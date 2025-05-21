package service

import (
	"context"
	"red-feed/internal/domain"
	"red-feed/internal/repository"
)

type ArticleService interface {
	Save(ctx context.Context, article domain.Article) (id int64, err error)
	Publish(ctx context.Context, article domain.Article) (int64, error)
	WithDraw(ctx context.Context, article domain.Article) error
	List(ctx context.Context, uid int64, offset int, limit int) ([]domain.Article, error)
	ListPub(ctx context.Context, offset int, limit int) ([]domain.Article, error)
	GetById(ctx context.Context, id int64) (domain.Article, error)
	GetPublishedById(ctx context.Context, id int64) (domain.Article, error)
}

type articleService struct {
	repo repository.ArticleRepository
}

func (s *articleService) ListPub(ctx context.Context, offset int, limit int) ([]domain.Article, error) {
	return s.repo.ListPub(ctx, offset, limit)
}

func (s *articleService) GetById(ctx context.Context, id int64) (domain.Article, error) {
	return s.repo.GetById(ctx, id)
}

func (s *articleService) GetPublishedById(ctx context.Context, id int64) (domain.Article, error) {
	return s.repo.GetPubById(ctx, id)
}

func (s *articleService) List(ctx context.Context, uid int64, offset int, limit int) ([]domain.Article, error) {
	return s.repo.List(ctx, uid, offset, limit)
}

func (s *articleService) WithDraw(ctx context.Context, article domain.Article) error {
	return s.repo.SyncStatus(ctx, article.Id, article.Author.Id, domain.ArticleStatusPrivate)
}

func (s *articleService) Publish(ctx context.Context, article domain.Article) (int64, error) {
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
