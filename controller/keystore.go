package controller

import (
	"net/http"
	"ylsz/hitake/front/model"

	"github.com/gin-gonic/gin"
)

// KeystoreList 密码对列表
// ShowAccount godoc
// @Tags user
// @Summary 密码对列表
// @Produce json
// @Param Authorization header string true "Authorization"
// @Success 200 {array} model.Keystore
// @Failure 400 {object} Response
// @Router /user/keystore [get]
func KeystoreList(ctx *gin.Context) {
	val, _ := ctx.Get("token")
	uinfo := val.(*UserInfo)
	list, err := model.KeystoreList(map[string]interface{}{
		"userId": uinfo.ID,
	})
	if err != nil {
		ctx.JSON(http.StatusBadRequest, &Response{err.Error()})
		return
	}
	ctx.JSON(http.StatusOK, &list)
}

// AddKeystore 密码对
// ShowAccount godoc
// @Tags user
// @Summary 密码对
// @Account json
// @Produce json
// @Param Authorization header string true "Authorization"
// @Param account body model.Keystore true "Keystore"
// @Success 200 {object} Response
// @Failure 400 {object} Response
// @Router /user/keystore [post]
func AddKeystore(ctx *gin.Context) {
	ks := new(model.Keystore)
	val, _ := ctx.Get("token")
	ks.UserID = val.(*UserInfo).ID
	err := ctx.BindJSON(ks)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, &Response{err.Error()})
		return
	}

	err = model.KeystoreAdd(ks)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, &Response{err.Error()})
		return
	}
	ctx.JSON(http.StatusOK, &Response{"OK"})
}

// ParamDeleteKeystore 删除密码对参数
type ParamDeleteKeystore struct {
	ID string `json:"id"`
}

// DeleteKeystore 删除密码对
// ShowAccount godoc
// @Tags user
// @Summary 删除密码对
// @Account json
// @Produce json
// @Param Authorization header string true "Authorization"
// @Param account body ParamDeleteKeystore true "ParamDeleteKeystore"
// @Success 200 {object} Response
// @Failure 400 {object} Response
// @Router /user/keystore [delete]
func DeleteKeystore(ctx *gin.Context) {
	p := new(ParamDeleteKeystore)
	err := ctx.BindJSON(p)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, &Response{err.Error()})
		return
	}

	err = model.DeleteKeystore(p.ID)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, &Response{err.Error()})
		return
	}
	ctx.JSON(http.StatusOK, &Response{"OK"})
}

// ShowAccount godoc
// @Tags user
// @Summary 修改密码对
// @Account json
// @Produce json
// @Param Authorization header string true "Authorization"
// @Success 200 {object} Response
// @Failure 400 {object} Response
// @Router /user/keystore [put]
func PutKeystore(c *gin.Context) {
	c.JSON(http.StatusOK, &struct {
		Message string `json:"message"`
	}{"OK"})
}