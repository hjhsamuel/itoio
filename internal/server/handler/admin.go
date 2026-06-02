package handler

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/hjhsamuel/itoio/internal/server/handler/request"
	"github.com/hjhsamuel/itoio/internal/server/handler/schema"
)

func (a *api) GetInviteCodeList(ctx *gin.Context) {
	val, _ := ctx.Get(request.AuthCtxKey)
	user, ok := val.(*request.User)
	if !ok {
		ctx.JSON(http.StatusUnauthorized, &request.Response{
			Code:    http.StatusUnauthorized,
			Message: "unauthorized",
		})
		return
	}
	var req *schema.GetInviteCodeReq
	if err := ctx.ShouldBindQuery(&req); err != nil || req.Page < 0 || req.Limit < 0 || req.Limit > 20 {
		ctx.JSON(http.StatusBadRequest, &request.Response{
			Code:    http.StatusBadRequest,
			Message: "invalid request",
		})
		return
	}
	codes, err := a.srv.GetInviteCodeList(user.ID, (req.Page-1)*req.Limit, req.Limit)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, &request.Response{
			Code:    http.StatusInternalServerError,
			Message: err.Error(),
		})
		return
	}
	rsp := make([]*schema.InviteCode, len(codes))
	for i, code := range codes {
		rsp[i] = &schema.InviteCode{
			Code:   code.Code,
			Expire: code.Expire,
		}
	}
	ctx.JSON(http.StatusOK, &request.Response{
		Code: http.StatusOK,
		Data: rsp,
	})
}

func (a *api) CreateInviteCode(ctx *gin.Context) {
	val, _ := ctx.Get(request.AuthCtxKey)
	user, ok := val.(*request.User)
	if !ok {
		ctx.JSON(http.StatusUnauthorized, &request.Response{
			Code:    http.StatusUnauthorized,
			Message: "unauthorized",
		})
		return
	}
	code, expire, err := a.srv.CreateInviteCode(user.ID)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, &request.Response{
			Code:    http.StatusInternalServerError,
			Message: err.Error(),
		})
		return
	}
	ctx.JSON(http.StatusOK, &request.Response{
		Code: http.StatusOK,
		Data: &schema.InviteCode{
			Code:   code,
			Expire: expire,
		},
	})
}

func (a *api) UpdatePassword(ctx *gin.Context) {
	val, _ := ctx.Get(request.AuthCtxKey)
	user, ok := val.(*request.User)
	if !ok {
		ctx.JSON(http.StatusUnauthorized, &request.Response{
			Code:    http.StatusUnauthorized,
			Message: "unauthorized",
		})
		return
	}
	var req *schema.UpdatePasswdReq
	if err := ctx.ShouldBindJSON(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, &request.Response{
			Code:    http.StatusBadRequest,
			Message: "invalid request",
		})
		return
	}
	err := a.srv.UpdatePasswd(user.ID, req.Old, req.New)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, &request.Response{
			Code:    http.StatusInternalServerError,
			Message: err.Error(),
		})
		return
	}
	ctx.JSON(http.StatusOK, &request.Response{
		Code: http.StatusOK,
	})
}
