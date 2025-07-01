package repository

import (
	"context"
	"github.com/ecodeclub/ekit/slice"
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
	SyncStatus(ctx context.Context, artId int64, authorId int64, status domain.ArticleStatus) error
	List(ctx context.Context, uid int64, offset int, limit int) ([]domain.Article, error)
	ListPub(ctx context.Context, offset int, limit int) ([]domain.Article, error)
	GetById(ctx context.Context, artId int64) (domain.Article, error)
	GetPubById(ctx context.Context, artId int64) (domain.Article, error)
	ListPubForRanking(ctx context.Context, start time.Time, offset int, limit int) ([]domain.Article, error)
}

type CachedArticleRepository struct {
	dao      dao.ArticleDao
	cache    cache.ArticleCache
	userRepo CachedUserRepository
	l        logger.Logger
}

func (r *CachedArticleRepository) ListPubForRanking(ctx context.Context, start time.Time, offset int, limit int) ([]domain.Article, error) {
	res, err := r.dao.ListPubForRanking(ctx, start, offset, limit)
	if err != nil {
		return nil, err
	}
	return slice.Map(res, func(idx int, src dao.PublishedArticle) domain.Article {
		return r.pubToDomain(src)
	}), nil
}

func (r *CachedArticleRepository) ListPub(ctx context.Context, offset int, limit int) ([]domain.Article, error) {
	res, err := r.dao.ListPub(ctx, offset, limit)
	if err != nil {
		return []domain.Article{}, err
	}
	return slice.Map[dao.PublishedArticle, domain.Article](res, func(idx int, src dao.PublishedArticle) domain.Article {
		return r.pubToDomain(src)
	}), nil
}

func (r *CachedArticleRepository) GetById(ctx context.Context, artId int64) (domain.Article, error) {
	cachedArt, err := r.cache.Get(ctx, artId)
	if err == nil {
		return cachedArt, nil
	}
	art, err := r.dao.GetById(ctx, artId)
	if err != nil {
		return domain.Article{}, err
	}
	return r.toDomain(art), nil
}

func (r *CachedArticleRepository) GetPubById(ctx context.Context, artId int64) (domain.Article, error) {
	res, err := r.cache.GetPub(ctx, artId)
	if err == nil {
		return res, err
	}
	art, err := r.dao.GetPubById(ctx, artId)
	if err != nil {
		return domain.Article{}, err
	}
	user, err := r.userRepo.FindById(ctx, art.AuthorId)
	if err != nil {
		return domain.Article{}, err
	}
	res = domain.Article{
		Id:      art.Id,
		Title:   art.Title,
		Status:  domain.ArticleStatus(art.Status),
		Content: art.Content,
		Author: domain.Author{
			Id:   user.Id,
			Name: user.Nickname,
		},
	}
	// 也可以同步
	go func() {
		if err = r.cache.SetPub(ctx, res); err != nil {
			r.l.Error("缓存已发表文章失败",
				logger.Error(err), logger.Int64("artId", res.Id))
		}
	}()
	return res, nil
}

func (r *CachedArticleRepository) List(ctx context.Context, uid int64, offset int, limit int) ([]domain.Article, error) {
	// 如果当前是第一页，则查询缓存
	if limit == 100 || offset == 0 {
		arts, err := r.cache.GetFirstPage(ctx, uid)
		if err == nil {
			// 预测用户大概率会访问列表第一个，所以直接提前缓存
			go func() {
				r.preCache(ctx, arts)
			}()
			return arts, nil
		}
		r.l.Error("设置第1页作者文章列表错误", logger.Error(err), logger.Int64("authorId", uid))
	}
	artsDAO, err := r.dao.List(ctx, uid, offset, limit)
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
	go func() {
		r.preCache(ctx, arts)
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

func (r *CachedArticleRepository) pubToDomain(art dao.PublishedArticle) domain.Article {
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

func (r *CachedArticleRepository) SyncStatus(ctx context.Context, artId int64, authorId int64, status domain.ArticleStatus) error {
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

// preCache 业务预加载
func (r *CachedArticleRepository) preCache(ctx context.Context,
	arts []domain.Article) {
	// 1MB
	const contentSizeThreshold = 1024 * 1024
	if len(arts) > 0 && len(arts[0].Content) <= contentSizeThreshold {
		// 你也可以记录日志
		if err := r.cache.Set(ctx, arts[0]); err != nil {
			r.l.Error("提前准备缓存失败", logger.Error(err))
		}
	}
}

func NewArticleRepository(dao dao.ArticleDao, cache cache.ArticleCache, l logger.Logger, userRepo CachedUserRepository) ArticleRepository {
	return &CachedArticleRepository{
		dao:      dao,
		cache:    cache,
		userRepo: userRepo,
		l:        l,
	}
}
