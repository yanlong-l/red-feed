package repository

import (
	"context"
	"github.com/gin-gonic/gin"
	"red-feed/internal/domain"
	"red-feed/internal/repository/dao"
)

type ArticleRepository interface {
	Create(ctx context.Context, article domain.Article) (id int64, err error)
	Update(ctx context.Context, article domain.Article) error
	Sync(ctx context.Context, article domain.Article) (int64, error)
	SyncStatus(ctx *gin.Context, artId int64, authorId int64, status domain.ArticleStatus) error
}

type CachedArticleRepository struct {
	dao dao.ArticleDao
}

func (r *CachedArticleRepository) SyncStatus(ctx *gin.Context, artId int64, authorId int64, status domain.ArticleStatus) error {
	return r.dao.SyncStatus(ctx, artId, authorId, status.ToUint8())
}

func (r *CachedArticleRepository) Sync(ctx context.Context, article domain.Article) (int64, error) {
	return r.dao.Sync(ctx, dao.Article{
		Id:       article.Id,
		Title:    article.Title,
		Content:  article.Content,
		AuthorId: article.Author.Id,
		Status:   domain.ArticleStatusPublished.ToUint8(),
	})
}

func (r *CachedArticleRepository) Update(ctx context.Context, article domain.Article) error {
	return r.dao.Update(ctx, dao.Article{
		Id:       article.Id,
		Title:    article.Title,
		Content:  article.Content,
		AuthorId: article.Author.Id,
		Status:   domain.ArticleStatusUnPublished.ToUint8(),
	})
}

func (r *CachedArticleRepository) Create(ctx context.Context, article domain.Article) (id int64, err error) {
	return r.dao.Insert(ctx, dao.Article{
		Title:    article.Title,
		Content:  article.Content,
		AuthorId: article.Author.Id,
		Status:   domain.ArticleStatusUnPublished.ToUint8(),
	})
}

func NewArticleRepository(dao dao.ArticleDao) ArticleRepository {
	return &CachedArticleRepository{
		dao: dao,
	}
}
