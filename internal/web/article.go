package web

import "github.com/gin-gonic/gin"

type ArticleHandler struct {
}

var _ Handler = (*ArticleHandler)(nil)

func NewArticleHandler() *ArticleHandler {
	return &ArticleHandler{}
}

func (a *ArticleHandler) RegisterRoutes(server *gin.Engine) {
	ag := server.Group("/articles")
	ag.POST("/edit", a.Edit)
}

func (a *ArticleHandler) Edit(ctx *gin.Context) {

}
