package controller

import (
	"net/http"
	"time"
	"ylsz/hit/log"
	"ylsz/hit/route"
	"ylsz/hitake/front/model"

	"github.com/dgrijalva/jwt-go"
	"github.com/gin-gonic/gin"
)

type ErrorResponse struct {
	Message string `json:"message"`
}

// AccountRegister 用户注册
// ShowAccount godoc
// @Tags 默认
// @Summary 注册
// @Account json
// @Produce json
// @Param account body model.Members true "member"
// @Success 200 {object} ErrorResponse
// @Failure 400 {object} ErrorResponse
// @Router /register [post]
func AccountRegister(ctx *gin.Context) {
	member := model.NewMember()
	err := ctx.Bind(member)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, &ErrorResponse{err.Error()})
		return
	}

	err = member.Add()
	if err != nil {
		ctx.JSON(http.StatusBadRequest, &ErrorResponse{err.Error()})
		return
	}
	ctx.JSON(http.StatusOK, &ErrorResponse{"OK"})
}

type AccountLoginRes struct {
	Token   string `json:"token"`
	ID      string `json:"id"`
	Account string `json:"account"`
	Phone   string `json:"phone"`
	Email   string `json:"email"`
}

// AccountLogin 登录
// ShowAccount godoc
// @Tags 默认
// @Summary 登录
// @Account json
// @Produce json
// @Param account body model.Members true "member"
// @Success 200 {object} AccountLoginRes
// @Failure 400 {object} ErrorResponse
// @Router /login [post]
func AccountLogin(ctx *gin.Context) {
	member := model.NewMember()
	err := ctx.BindJSON(member)
	if err != nil {
		log.Error(err.Error())
		ctx.JSON(http.StatusBadRequest, ErrorResponse{err.Error()})
		return
	}
	_m, err := member.Login()
	if err != nil {
		log.Error(err.Error())
		ctx.JSON(http.StatusBadRequest, ErrorResponse{err.Error()})
		return
	}

	// 设置token
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, &UserInfo{
		ID: _m.ID,
		StandardClaims: jwt.StandardClaims{
			ExpiresAt: time.Now().Add(24 * 7 * time.Hour).Unix(),
		},
	})
	sign, err := token.SignedString([]byte(TokenKey))
	if err != nil {
		log.Error(err.Error())
		ctx.JSON(http.StatusBadRequest, ErrorResponse{err.Error()})
		return
	}

	ctx.Header(TokenHeader, sign)

	res := &AccountLoginRes{
		ID:      _m.ID,
		Account: _m.Account,
		Phone:   _m.Phone,
		Email:   _m.Email,
		Token:   sign,
	}
	ctx.JSON(http.StatusOK, res)
}

var TokenHeader = "Authorization"
var TokenKey = "ylsz"

type UserInfo struct {
	ID string `json:"id"`
	jwt.StandardClaims
}

// AccountUpdateParcel 用户修改存储库信息
func AccountUpdateParcel(ctx *route.Context) {
	param := struct {
		Parcel string `json:"parcel,omitempty"`
		KeyID  string `json:"key,omitempty"`
		Secret string `json:"secret,omitempty"`
	}{}
	err := ctx.GetBody(&param)
	if err != nil {
		ctx.EJSON(http.StatusBadRequest, err.Error())
		return
	}

	member := model.NewMember()
	member.ID = ctx.GetValue("token").(*UserInfo).ID
	member.Parcel = param.Parcel
	member.KeyID = param.KeyID
	member.Secret = param.Secret
	err = member.Update()
	if err != nil {
		ctx.EJSON(http.StatusBadRequest, &ErrorResponse{err.Error()})
		return
	}

	// 获取分发记录

	type instance struct {
		ID      string `json:"id"`
		Name    string `json:"name"`
		Message string `json:"message"`
		Status  string `json:"status"`
	}
	data := []struct {
		ID   string `json:"id"`
		Name string `json:"name"`

		URL         string     `json:"url"`
		AllNum      int        `json:"allNum"`
		InstallNum  int        `json:"installNum"`
		DownloadNum int        `json:"downloadNum"`
		RunNum      int        `json:"runNum"`
		Instances   []instance `json:"instances"`
		CreateTime  string     `json:"createTime"`
		Message     string     `json:"message"`
	}{
		{
			ID:          "910rt72m1",
			Name:        "agent.tar.gz",
			URL:         "hitake-0.1.10/agent/agent.tar.gz",
			AllNum:      3,
			InstallNum:  1,
			DownloadNum: 1,
			RunNum:      1,
			Instances: []instance{
				{"3o29d8d", "192.168.10.120", "", "distributing"},
			},
		},
		{
			ID:          "910rt73m1",
			Name:        "filebeat.tar.gz",
			URL:         "hitake-0.1.10/agent/agent.tar.gz",
			AllNum:      3,
			InstallNum:  1,
			DownloadNum: 1,
			RunNum:      1,
			Instances: []instance{
				{"3o29d8d", "192.168.10.120", "", "distributing"},
			},
		},
	}
	ctx.JSON(&data)
}

// AccountInfo 用户存储库信息
// ShowAccount godoc
// @Tags 默认
// @Summary 用户存储库信息
// @Produce json
// @Param Authorization header string true "Authorization"
// @Success 200 {object} model.Members
// @Failure 400 {object} ErrorResponse
// @Router /member [get]
func AccountInfo(ctx *gin.Context) {
	member := model.NewMember()
	val, _ := ctx.Get("token")
	uinfo := val.(*UserInfo)
	member.ID = uinfo.ID
	err := member.Read()
	if err != nil {
		ctx.JSON(http.StatusBadRequest, &ErrorResponse{err.Error()})
		return
	}
	ctx.JSON(http.StatusOK, member)
}
