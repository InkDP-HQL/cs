package controller

import (
	"fmt"
	"net/http"
	"ylsz/hit/log"
	"ylsz/hitake/front/model"

	"github.com/gin-gonic/gin"
)

// UpdateConf 更新配置
// ShowAccount godoc
// @Tags object
// @Summary 更新配置
// @Account json
// @Produce json
// @Param Authorization header string true "Authorization"
// @Param account body model.Conf true "conf"
// @Success 200 {object} Response
// @Failure 400 {object} Response
// @Router /object/conf [post]
func UpdateConf(ctx *gin.Context) {
	obj := new(model.Conf)
	err := ctx.Bind(obj)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, &ErrorResponse{err.Error()})
		return
	}

	log.Info(fmt.Sprintf("对象信息: %+v", obj))

	val, _ := ctx.Get("token")
	uinfo := val.(*UserInfo)
	obj.UserID = uinfo.ID

	err = model.UpdateConf(obj)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, &ErrorResponse{err.Error()})
		return
	}
	ctx.JSON(http.StatusOK, &ErrorResponse{"OK"})
}

// GetConf 获取配置信息
// ShowAccount godoc
// @Tags object
// @Summary 获取配置信息
// @Produce json
// @Param Authorization header string true "Authorization"
// @Success 200 {object} model.Conf
// @Failure 400 {object} ErrorResponse
// @Router /object/conf [get]
func GetConf(ctx *gin.Context) {
	val, _ := ctx.Get("token")
	uInfo := val.(*UserInfo)

	conf, err := model.OneConf(uInfo.ID)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, &ErrorResponse{err.Error()})
		return
	}
	ctx.JSON(http.StatusOK, conf)
}
