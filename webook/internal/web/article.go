package web

import (
	"Learn_Go/webook/internal/domain"
	"Learn_Go/webook/internal/service"
	"Learn_Go/webook/internal/web/jwt"
	"Learn_Go/webook/pkg/logger"
	"github.com/gin-gonic/gin"
	"net/http"
)

type ArticleHandler struct {
	svc service.ArticleService
	l   logger.LoggerV1
}

func NewArticleHandler(svc service.ArticleService, l logger.LoggerV1) *ArticleHandler {
	return &ArticleHandler{
		svc: svc,
		l:   l,
	}
}

func (h *ArticleHandler) RegisterRouters(server *gin.Engine) {

	g := server.Group("/articles")
	g.POST("/edit", h.Edit)
	g.POST("/publish", h.Publish)

}

// Edit 接受 Article 输入，返回一个文章的 ID
func (h *ArticleHandler) Edit(ctx *gin.Context) {
	type ArticleReq struct {
		Id      int64  `json:"id"`
		Title   string `json:"title"`
		Content string `json:"content"`
	}
	var req ArticleReq
	err := ctx.Bind(&req)
	if err != nil {
		ctx.JSON(http.StatusOK, Result{
			Code: 5,
			Msg:  "系统错误",
		})
		return
	}
	uc, ok := ctx.MustGet("user").(jwt.UserClaims)
	if !ok {
		ctx.AbortWithStatus(http.StatusUnauthorized)
		return
	}
	id, err := h.svc.Save(ctx, domain.Article{
		Id:      req.Id,
		Title:   req.Title,
		Content: req.Content,
		Author: domain.Author{
			Id: uc.Uid,
		},
	})
	if err != nil {
		ctx.JSON(http.StatusOK, Result{
			Code: 5,
			Msg:  "系统错误",
		})
		// 大日志
		h.l.Error("保存文章失败",
			logger.Int64("Uid", uc.Uid), // 打印下用户 id
			logger.Error(err))           // 打印下错误信息
	}
	ctx.JSON(http.StatusOK, Result{
		Msg:  "保存成功",
		Data: id,
	})

}

func (h *ArticleHandler) Publish(ctx *gin.Context) {
	type ArticleReq struct {
		Id      int64  `json:"id"`
		Title   string `json:"title"`
		Content string `json:"content"`
	}
	var req ArticleReq
	err := ctx.Bind(&req)
	if err != nil {
		ctx.JSON(http.StatusOK, Result{
			Code: 5,
			Msg:  "系统错误",
		})
		return
	}
	uc, ok := ctx.MustGet("user").(jwt.UserClaims)
	if !ok {
		ctx.AbortWithStatus(http.StatusUnauthorized)
		return
	}
	id, err := h.svc.Publish(ctx, domain.Article{
		Id:      req.Id,
		Title:   req.Title,
		Content: req.Content,
		Author: domain.Author{
			Id: uc.Uid,
		},
	})
	if err != nil {
		ctx.JSON(http.StatusOK, Result{
			Code: 5,
			Msg:  "系统错误",
		})
		// 大日志
		h.l.Error("发表文章失败",
			logger.Int64("Uid", uc.Uid), // 打印下用户 id
			logger.Error(err))           // 打印下错误信息
	}
	ctx.JSON(http.StatusOK, Result{
		Msg:  "发表成功",
		Data: id,
	})
}
