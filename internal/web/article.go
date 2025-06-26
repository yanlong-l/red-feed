package web

import (
	"fmt"
	"github.com/ecodeclub/ekit/slice"
	"github.com/gin-gonic/gin"
	"golang.org/x/sync/errgroup"
	"net/http"
	"red-feed/internal/domain"
	"red-feed/internal/service"
	ijwt "red-feed/internal/web/jwt"
	"red-feed/pkg/logger"
	"strconv"
	"time"
)

var _ Handler = (*ArticleHandler)(nil)

type ArticleHandler struct {
	svc     service.ArticleService
	intrSvc service.InteractiveService
	l       logger.Logger
	biz     string
}

func NewArticleHandler(svc service.ArticleService, l logger.Logger, intrSvc service.InteractiveService) *ArticleHandler {
	return &ArticleHandler{
		svc:     svc,
		l:       l,
		intrSvc: intrSvc,
		biz:     "article",
	}
}

func (a *ArticleHandler) RegisterRoutes(server *gin.Engine) {
	ag := server.Group("/articles")
	ag.POST("/edit", a.Edit)         // 创作者保存文章
	ag.POST("/publish", a.Publish)   // 创作者发表文章
	ag.POST("/withdraw", a.WithDraw) // 创作撤销发表的文章
	ag.POST("/list", a.List)         // 创作者查看自己的文章列表
	ag.GET("/detail/:id", a.Detail)  // 创作者查看自己的文章详情

	pub := ag.Group("/pub")
	//pub.GET("/pub", a.ListPub)
	pub.GET("/:id", a.PubDetail) // 读者查看文章详情
	pub.POST("/list", a.PubList) // 读者查看文章列表

	pub.POST("/like", a.Like)       // 读者点赞 or 取消点赞
	pub.POST("/collect", a.Collect) // 读者收藏 or 取消收藏
}

func (a *ArticleHandler) Edit(ctx *gin.Context) {
	var req struct {
		Id      int64  `json:"id"`
		Title   string `json:"title"`
		Content string `json:"content"`
	}
	if err := ctx.ShouldBind(&req); err != nil {
		return
	}
	// 获取用户id
	claims := ctx.MustGet("claims")
	claimsVal, ok := claims.(*ijwt.UserClaims)
	if !ok {
		ctx.JSON(http.StatusOK, Result{
			Code: 5,
			Msg:  "系统错误",
		})
		a.l.Error("未发现用户的 session 信息")
		return
	}
	// 调用article service
	id, err := a.svc.Save(ctx, domain.Article{
		Id:      req.Id,
		Title:   req.Title,
		Content: req.Content,
		Author: domain.Author{
			Id: claimsVal.Uid,
		},
		Status: domain.ArticleStatusUnPublished,
	})
	if err != nil {
		ctx.JSON(http.StatusOK, Result{
			Code: 5,
			Msg:  "系统错误",
		})
		a.l.Error("保存帖子失败", logger.Error(err))
		return
	}
	ctx.JSON(http.StatusOK, Result{
		Msg:  "OK",
		Data: id,
	})
}

func (a *ArticleHandler) Publish(ctx *gin.Context) {
	var req struct {
		Id      int64  `json:"id"`
		Title   string `json:"title"`
		Content string `json:"content"`
	}
	if err := ctx.ShouldBind(&req); err != nil {
		return
	}
	// 获取用户id
	claims := ctx.MustGet("claims")
	claimsVal, ok := claims.(*ijwt.UserClaims)
	if !ok {
		ctx.JSON(http.StatusOK, Result{
			Code: 5,
			Msg:  "系统错误",
		})
		a.l.Error("未发现用户的 session 信息")
		return
	}
	// 调用article service
	id, err := a.svc.Publish(ctx, domain.Article{
		Id:      req.Id,
		Title:   req.Title,
		Content: req.Content,
		Author: domain.Author{
			Id: claimsVal.Uid,
		},
	})
	if err != nil {
		ctx.JSON(http.StatusOK, Result{
			Code: 5,
			Msg:  "系统错误",
		})
		a.l.Error("发表帖子失败", logger.Error(err))
		return
	}
	ctx.JSON(http.StatusOK, Result{
		Msg:  "OK",
		Data: id,
	})
}

func (a *ArticleHandler) WithDraw(ctx *gin.Context) {
	var req struct {
		Id int64 `json:"id"`
	}
	if err := ctx.ShouldBind(&req); err != nil {
		return
	}
	// 获取用户id
	claims := ctx.MustGet("claims")
	claimsVal, ok := claims.(*ijwt.UserClaims)
	if !ok {
		ctx.JSON(http.StatusOK, Result{
			Code: 5,
			Msg:  "系统错误",
		})
		a.l.Error("未发现用户的 session 信息")
		return
	}
	// 调用article service
	err := a.svc.WithDraw(ctx, domain.Article{
		Id: req.Id,
		Author: domain.Author{
			Id: claimsVal.Uid,
		},
	})
	if err != nil {
		ctx.JSON(http.StatusOK, Result{
			Code: 5,
			Msg:  "系统错误",
		})
		a.l.Error("撤回帖子失败", logger.Error(err))
		return
	}
	ctx.JSON(http.StatusOK, Result{
		Msg: "OK",
	})
}

func (a *ArticleHandler) List(ctx *gin.Context) {
	var req struct {
		Offset int `json:"offset"`
		Limit  int `json:"limit"`
	}
	if err := ctx.ShouldBind(&req); err != nil {
		return
	}
	// 获取用户id
	claims := ctx.MustGet("claims")
	claimsVal, ok := claims.(*ijwt.UserClaims)
	if !ok {
		ctx.JSON(http.StatusOK, Result{
			Code: 5,
			Msg:  "系统错误",
		})
		a.l.Error("未发现用户的 session 信息")
		return
	}
	res, err := a.svc.List(ctx, claimsVal.Uid, req.Offset, req.Limit)
	if err != nil {
		ctx.JSON(http.StatusOK, Result{
			Code: 5,
			Msg:  "系统错误",
		})
		return
	}
	ctx.JSON(http.StatusOK, Result{
		Data: slice.Map[domain.Article, ArticleVO](res,
			func(idx int, src domain.Article) ArticleVO {
				return ArticleVO{
					Id:       src.Id,
					Title:    src.Title,
					Abstract: src.Abstract(),
					Status:   src.Status.ToUint8(),
					// 这个列表请求，不需要返回内容
					//Content: src.Content,
					// 这个是创作者看自己的文章列表，也不需要这个字段
					//Author: src.Author
					Ctime: src.Ctime.Format(time.DateTime),
					Utime: src.Utime.Format(time.DateTime),
				}
			}),
	})
}

func (a *ArticleHandler) Detail(ctx *gin.Context) {
	idStr := ctx.Param("id")
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		ctx.JSON(http.StatusOK, Result{
			Code: 4,
			Msg:  "参数错误",
		})
		a.l.Error("前端输入的 ID 不对", logger.Error(err))
		return
	}

	usr, ok := ctx.MustGet("claims").(*ijwt.UserClaims)
	fmt.Println(usr)
	if !ok {
		ctx.JSON(http.StatusOK, Result{
			Code: 5,
			Msg:  "系统错误",
		})
		a.l.Error("获得用户会话信息失败")
		return
	}
	art, err := a.svc.GetById(ctx, id)
	if err != nil {
		ctx.JSON(http.StatusOK, Result{
			Code: 5,
			Msg:  "系统错误",
		})
		a.l.Error("获得文章信息失败", logger.Error(err))
		return
	}
	// 这是不借助数据库查询来判定的方法
	if art.Author.Id != usr.Uid {
		ctx.JSON(http.StatusOK, Result{
			Code: 4,
			// 也不需要告诉前端究竟发生了什么
			Msg: "输入有误",
		})
		// 如果公司有风控系统，这个时候就要上报这种非法访问的用户了。
		a.l.Error("非法访问文章，创作者 ID 不匹配",
			logger.Int64("uid", usr.Uid))
		return
	}
	ctx.JSON(http.StatusOK, Result{
		Data: ArticleVO{
			Id:      art.Id,
			Title:   art.Title,
			Status:  art.Status.ToUint8(),
			Content: art.Content,
			// 这个是创作者看自己的文章列表，也不需要这个字段
			Ctime: art.Ctime.Format(time.DateTime),
			Utime: art.Utime.Format(time.DateTime),
		},
	})
}

func (a *ArticleHandler) PubDetail(ctx *gin.Context) {
	idStr := ctx.Param("id")
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		a.l.Error("前端输入的 ID 不对", logger.Error(err))
		ctx.JSON(http.StatusOK, Result{
			Code: 4,
			Msg:  "参数错误",
		})
		return
	}
	uc, ok := ctx.MustGet("claims").(*ijwt.UserClaims)
	if !ok {
		ctx.JSON(http.StatusOK, Result{
			Code: 5,
			Msg:  "系统错误",
		})
		a.l.Error("获得用户会话信息失败")
		return
	}

	// 使用 error group 来同时查询数据
	var (
		eg   errgroup.Group
		art  domain.Article
		intr domain.Interactive
	)
	eg.Go(func() error {
		var er error
		art, er = a.svc.GetPublishedById(ctx, id, uc.Uid)
		return er
	})

	eg.Go(func() error {
		var er error
		intr, er = a.intrSvc.Get(ctx, a.biz, id, uc.Uid)
		return er
	})

	err = eg.Wait()
	if err != nil {
		a.l.Error("获得文章详情信息失败", logger.Error(err))
		ctx.JSON(http.StatusOK, Result{
			Code: 5,
			Msg:  "系统错误",
		})
		return
	}

	ctx.JSON(http.StatusOK, Result{
		Data: ArticleVO{
			Id:         art.Id,
			Title:      art.Title,
			Status:     art.Status.ToUint8(),
			Content:    art.Content,
			Author:     art.Author.Name, // 详情页 要把作者信息带出去
			CollectCnt: intr.CollectCnt,
			ReadCnt:    intr.ReadCnt,
			LikeCnt:    intr.LikeCnt,
			Collected:  intr.Collected,
			Liked:      intr.Liked,
			Ctime:      art.Ctime.Format(time.DateTime),
			Utime:      art.Utime.Format(time.DateTime),
		},
	})
}

func (a *ArticleHandler) PubList(ctx *gin.Context) {
	var req struct {
		Offset int `json:"offset"`
		Limit  int `json:"limit"`
	}
	if err := ctx.ShouldBind(&req); err != nil {
		return
	}
	res, err := a.svc.ListPub(ctx, req.Offset, req.Limit)
	if err != nil {
		ctx.JSON(http.StatusOK, Result{
			Code: 5,
			Msg:  "系统错误",
		})
		return
	}
	ctx.JSON(http.StatusOK, Result{
		Data: slice.Map[domain.Article, ArticleVO](res,
			func(idx int, src domain.Article) ArticleVO {
				return ArticleVO{
					Id:       src.Id,
					Title:    src.Title,
					Abstract: src.Abstract(),
					Status:   src.Status.ToUint8(),
					// 这个列表请求，不需要返回内容
					//Content: src.Content,
					// 这个是创作者看自己的文章列表，也不需要这个字段
					Author: src.Author.Name,
					Ctime:  src.Ctime.Format(time.DateTime),
					Utime:  src.Utime.Format(time.DateTime),
				}
			}),
	})
}

func (a *ArticleHandler) Like(ctx *gin.Context) {
	var req struct {
		Id   int64 `json:"id"`
		Like bool  `json:"like"`
	}
	if err := ctx.ShouldBind(&req); err != nil {
		return
	}
	uc, ok := ctx.MustGet("claims").(*ijwt.UserClaims)
	if !ok {
		ctx.JSON(http.StatusOK, Result{
			Code: 5,
			Msg:  "系统错误",
		})
		a.l.Error("获得用户会话信息失败")
		return
	}
	var err error
	if req.Like {
		err = a.intrSvc.Like(ctx, a.biz, req.Id, uc.Uid)
	} else {
		err = a.intrSvc.CancelLike(ctx, a.biz, req.Id, uc.Uid)
	}
	if err != nil {
		ctx.JSON(http.StatusOK, Result{
			Code: 5,
			Msg:  "系统错误",
		})
	}
	ctx.JSON(http.StatusOK, Result{
		Msg: "OK",
	})
}

func (a *ArticleHandler) Collect(ctx *gin.Context) {
	var req struct {
		Id      int64 `json:"id"`
		Cid     int64 `json:"cid"`
		Collect bool  `json:"collect"`
	}
	if err := ctx.ShouldBind(&req); err != nil {
		return
	}
	uc, ok := ctx.MustGet("claims").(*ijwt.UserClaims)
	if !ok {
		ctx.JSON(http.StatusOK, Result{
			Code: 5,
			Msg:  "系统错误",
		})
		a.l.Error("获得用户会话信息失败")
		return
	}
	var err error
	if req.Collect {
		err = a.intrSvc.Collect(ctx, a.biz, req.Id, uc.Uid, req.Cid)
	} else {
		err = a.intrSvc.CancelCollect(ctx, a.biz, req.Id, uc.Uid, req.Cid)
	}
	if err != nil {
		ctx.JSON(http.StatusOK, Result{
			Code: 5,
			Msg:  "系统错误",
		})
		return
	}
	ctx.JSON(http.StatusOK, Result{
		Msg: "OK",
	})
}
