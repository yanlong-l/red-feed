package repository

import (
	"context"
	"github.com/ecodeclub/ekit/slice"
	"github.com/gin-gonic/gin"
	"red-feed/internal/domain"
	"red-feed/internal/repository/cache"
	"red-feed/internal/repository/dao"
	"red-feed/pkg/logger"
	"time"
)

type ArticleRepository interface {
	Create(ctx context.Context, article domain.Article) (artId int64, err error)
	Update(ctx context.Context, article domain.Article) error
	Sync(ctx context.Context, article domain.Article) (int64, error)
	SyncStatus(ctx *gin.Context, artId int64, authorId int64, status domain.ArticleStatus) error
	List(ctx *gin.Context, uid int64, offset int, limit int) ([]domain.Article, error)
}

type CachedArticleRepository struct {
	dao   dao.ArticleDao
	cache cache.ArticleCache
	l     logger.Logger
}

func (r *CachedArticleRepository) List(ctx *gin.Context, uid int64, offset int, limit int) ([]domain.Article, error) {
	// 如果当前是第一页，则查询缓存
	if limit == 100 || offset == 0 {
		arts, err := r.cache.GetFirstPage(ctx, uid)
		if err == nil {
			return arts, nil
		}
		r.l.Error("设置第1页作者文章列表错误", logger.Error(err), logger.Int64("authorId", uid))
	}
	artsDAO, err := r.dao.GetListByAuthor(ctx, uid, offset, limit)
	if err != nil {
		return []domain.Article{}, err
	}
	arts := slice.Map[dao.Article, domain.Article](artsDAO, func(idx int, src dao.Article) domain.Article {
		return r.toDomain(src)
	})
	// 异步回写缓存
	go func() {
		setErr := r.cache.SetFirstPage(ctx, uid, arts)
		if setErr != nil {
			r.l.Error("回写缓存：第1页作者文章列表 失败", logger.Error(setErr))
		}
	}()
	return arts, nil
}

func (r *CachedArticleRepository) toDomain(art dao.Article) domain.Article {
	return domain.Article{
		Id:      art.Id,
		Title:   art.Title,
		Status:  domain.ArticleStatus(art.Status),
		Content: art.Content,
		Author: domain.Author{
			Id: art.AuthorId,
		},
		Ctime: time.UnixMilli(art.Ctime),
		Utime: time.UnixMilli(art.Utime),
	}
}

func (r *CachedArticleRepository) toEntity(art domain.Article) dao.Article {
	return dao.Article{
		Id:       art.Id,
		Title:    art.Title,
		Content:  art.Content,
		AuthorId: art.Author.Id,
		Status:   art.Status.ToUint8(),
	}
}

func (r *CachedArticleRepository) SyncStatus(ctx *gin.Context, artId int64, authorId int64, status domain.ArticleStatus) error {
	err := r.dao.SyncStatus(ctx, artId, authorId, status.ToUint8())
	if err == nil {
		delErr := r.cache.DelFirstPage(ctx, authorId)
		if delErr != nil {
			r.l.Error("删除缓存：作者的第一页文章 失败", logger.Error(err), logger.Int64("authorId", authorId))
		}
	}
	return err
}

func (r *CachedArticleRepository) Sync(ctx context.Context, article domain.Article) (int64, error) {
	artId, err := r.dao.Sync(ctx, r.toEntity(article))
	if err == nil {
		delErr := r.cache.DelFirstPage(ctx, article.Author.Id)
		if delErr != nil {
			r.l.Error("删除缓存：作者的第一页文章 失败", logger.Error(err), logger.Int64("authorId", article.Author.Id))
		}
	}
	return artId, err
}

func (r *CachedArticleRepository) Update(ctx context.Context, article domain.Article) error {
	err := r.dao.Update(ctx, r.toEntity(article))
	if err == nil {
		delErr := r.cache.DelFirstPage(ctx, article.Author.Id)
		if delErr != nil {
			r.l.Error("删除缓存：作者的第一页文章 失败", logger.Error(err), logger.Int64("authorId", article.Author.Id))
		}
	}
	return err
}

func (r *CachedArticleRepository) Create(ctx context.Context, article domain.Article) (artId int64, err error) {
	artId, err = r.dao.Insert(ctx, r.toEntity(article))
	if err == nil {
		delErr := r.cache.DelFirstPage(ctx, article.Author.Id)
		if delErr != nil {
			r.l.Error("删除缓存：作者的第一页文章 失败", logger.Error(err), logger.Int64("authorId", article.Author.Id))
		}
	}
	return
}

func NewArticleRepository(dao dao.ArticleDao, cache cache.ArticleCache, l logger.Logger) ArticleRepository {
	return &CachedArticleRepository{
		dao:   dao,
		cache: cache,
		l:     l,
	}
}
