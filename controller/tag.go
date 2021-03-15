package controller

import (
	"net/http"

	"ylsz/hitake/front/model"

	"github.com/gin-gonic/gin"
)

type Response struct {
	Message string `json:"message"`
}

// TagGet 请求标签列表
// ShowAccount godoc
// @Tags tag
// @Summary 请求标签列表
// @Produce json
// @Param Authorization header string true "Authorization"
// @Param name query string false "name"
// @Param type query string false "type"
// @Param popular query string false "popular"
// @Success 200 {array} model.Tag
// @Failure 400 {object} Response
// @Router /tag [get]
func TagGet(ctx *gin.Context) {
	val, _ := ctx.Get("token")
	uinfo := val.(*UserInfo)
	param := make(map[string]interface{})
	param["userId"] = uinfo.ID
	if name := ctx.Query("name"); name != "" {
		param["name"] = name
	}
	if _type := ctx.Query("type"); _type != "" {
		param["type"] = _type
	}

	if ctx.Query("popular") != "" {
		param["popular"] = true
	}
	list, err := model.TagList(param)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, &Response{err.Error()})
		return
	}
	ctx.JSON(http.StatusOK, &list)
}

// ResponseAllTag 所有标签返回结果
type ResponseAllTag struct {
	Level    []string `json:"level"`
	Category []string `json:"category"`
	Group    []string `json:"group"`
}

// AllTag 所有标签
// ShowAccount godoc
// @Tags tag
// @Summary 所有标签
// @Produce json
// @Param Authorization header string true "Authorization"
// @Success 200 {object} ResponseAllTag
// @Failure 400 {object} Response
// @Router /tag/all [get]
func AllTag(ctx *gin.Context) {
	param := make(map[string]interface{})
	list, err := model.TagList(param)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, &Response{err.Error()})
		return
	}

	res := new(ResponseAllTag)
	res.Level = []string{}
	res.Category = []string{}
	res.Group = []string{}
	for _, tag := range list {
		switch tag.Type {
		case model.TagTypeCategory:
			res.Category = append(res.Category, tag.Name)
		case model.TagTypeLevel:
			res.Level = append(res.Level, tag.Name)
		case model.TagTypeGroup:
			res.Group = append(res.Group, tag.Name)
		}
	}

	ctx.JSON(http.StatusOK, res)
}

// TagAdd 添加标签
// ShowAccount godoc
// @Tags tag
// @Summary 添加标签
// @Account json
// @Produce json
// @Param Authorization header string true "Authorization"
// @Param account body model.Tag true "tag"
// @Success 200 {object} Response
// @Failure 400 {object} Response
// @Router /tag [post]
func TagAdd(ctx *gin.Context) {
	tag := new(model.Tag)
	val, _ := ctx.Get("token")
	tag.UserID = val.(*UserInfo).ID
	err := ctx.BindJSON(tag)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, &Response{err.Error()})
		return
	}

	err = model.TagAdd(tag)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, &Response{err.Error()})
		return
	}
	ctx.JSON(http.StatusOK, &Response{"OK"})
}
