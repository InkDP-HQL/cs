package controller

import (
	"encoding/json"
	"fmt"
	"net/http"
	"path/filepath"
	"strconv"
	"strings"

	xhttp "ylsz/hit/http"
	"ylsz/hit/log"

	"ylsz/hitake/front/model"

	"github.com/gin-gonic/gin"
)

// AddObject 添加对象
// ShowAccount godoc
// @Tags object
// @Summary 添加对象
// @Account json
// @Produce json
// @Param Authorization header string true "Authorization"
// @Param account body model.Object true "object"
// @Success 200 {object} ErrorResponse
// @Failure 400 {object} ErrorResponse
// @Router /object [post]
func AddObject(ctx *gin.Context) {
	obj := new(model.Object)
	err := ctx.Bind(obj)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, &ErrorResponse{err.Error()})
		return
	}

	log.Info(fmt.Sprintf("对象信息: %+v", obj))

	val, _ := ctx.Get("token")
	uInfo := val.(*UserInfo)
	obj.UserID = uInfo.ID

	// 添加tag
	tags := make([]*model.Tag, 0)
	tags = append(tags, &model.Tag{
		UserID: uInfo.ID,
		Name:   obj.Level,
		Type:   model.TagTypeLevel,
	})
	tags = append(tags, &model.Tag{
		UserID: uInfo.ID,
		Name:   obj.Category,
		Type:   model.TagTypeCategory,
	})
	tags = append(tags, &model.Tag{
		UserID: uInfo.ID,
		Name:   obj.Group,
		Type:   model.TagTypeGroup,
	})
	for _, tag := range tags {
		exsit, err := tag.IsExsit()
		if err != nil || exsit || tag.Name == "" {
			continue
		}
		model.TagAdd(tag)
	}

	err = model.ObjectAdd(obj)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, &ErrorResponse{err.Error()})
		return
	}
	ctx.JSON(http.StatusOK, &ErrorResponse{"OK"})
}

// ObjectList 对象列表
// ShowAccount godoc
// @Tags object
// @Summary 对象列表
// @Produce json
// @Param Authorization header string true "Authorization"
// @Param tags query string false "tags"
// @Param tags[] query string false "tags"
// @Param register query string false "tags"
// @Success 200 {array} model.Object
// @Failure 400 {object} Response
// @Router /object [get]
func ObjectList(ctx *gin.Context) {
	val, _ := ctx.Get("token")
	uinfo := val.(*UserInfo)

	param := make(map[string]interface{})
	param["userId"] = uinfo.ID
	if tags := ctx.QueryArray("tags"); len(tags) == 0 {
		if tags := ctx.QueryArray("tags[]"); len(tags) > 0 {
			param["tag"] = tags
		}
	} else {
		param["tag"] = tags
	}

	if register := ctx.Query("register"); register != "" {
		b, err := strconv.ParseBool(register)
		if err == nil {
			param["register"] = b
		}
	}

	list, err := model.ObjectList(param)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, &Response{err.Error()})
		return
	}
	ctx.JSON(http.StatusOK, &list)
}

// ParamRegisterObject 注册对象参数
type ParamRegisterObject struct {
	Ids []string `json:"ids"`
}

// RegisterObject 注册对象
// ShowAccount godoc
// @Tags object
// @Summary 注册对象
// @Account json
// @Produce json
// @Param Authorization header string true "Authorization"
// @Param account body ParamRegisterObject true "ParamRegisterObject"
// @Success 200 {object} Response
// @Failure 400 {object} Response
// @Router /object [put]
func RegisterObject(ctx *gin.Context) {
	p := new(ParamRegisterObject)
	err := ctx.BindJSON(p)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, &Response{err.Error()})
		return
	}

	for _, id := range p.Ids {
		obj := new(model.Object)
		obj.ID = id
		obj.Register = true
		obj.Update()
	}

	ctx.JSON(http.StatusOK, &Response{"OK"})
}

// DistributeInstanceParam 分发信息
type DistributeInstanceParam struct {
	IP    string   `json:"ip"`
	ID    string   `json:"id"`
	Files []string `json:"files"`
}

// Distribute 后端分发接口参数
type Distribute struct {
	ID             string `json:"id"`        // 文件ID 唯一确定
	FileName       string `json:"file_name"` // 文件名 包括后缀 如: hitake-0.1.10/agent/agent.tar.gz
	Md5            string `json:"md5"`       // 文件二进制md5值
	DistributeType string `json:"distribute_type"`
	ToType         string `json:"to_type"` // 目标地址类型 如: host, k8s 等等
	To             string // target addr and port,  such as 192.168.0.120:8002 for host,  servicesName:8003 for k8s
	FromType       string `json:"from_type"`     // resource type like S3
	From           string `json:"from"`          // resource address
	CallbackAddr   string `json:"callback_addr"` // callback address for back status async
	err            error
}

// 分发类型
const (
	DistributeTypeSSH   = "ssh"
	DistributeTypeAgent = "agent"
)

// 文件存储类型
const (
	DistributeFromTypeS3 = "s3"
)

// ObjectDistribute 分发文件
// ShowAccount godoc
// @Tags object
// @Summary 分发文件
// @Account json
// @Produce  json
// @Param Authorization header string true "Authorization"
// @Param account body DistributeInstanceParam true "分配实例参数"
// @Success 200 {object} HookData
// @Failure 400 {object} ErrorResponse
// @Router /object/distribute [post]
func ObjectDistribute(ctx *gin.Context) {
	p := new(DistributeInstanceParam)
	err := ctx.BindJSON(p)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, &Response{err.Error()})
		return
	}
	val, _ := ctx.Get("token")
	uInfo := val.(*UserInfo)

	// 1.获取配置信息
	conf, err := model.OneConf(uInfo.ID)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, &ErrorResponse{err.Error()})
		return
	}

	if conf == nil || conf.Pacel.Account == "" || conf.Pacel.Password == "" || conf.Pacel.URL == "" {
		ctx.JSON(http.StatusBadRequest, &ErrorResponse{"请完善配置信息"})
		return
	}

	// 2.获取对象信息
	obj, err := model.OneObject(p.ID)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, &Response{err.Error()})
		return
	}

	if obj == nil {
		ctx.JSON(http.StatusBadRequest, &Response{"分发对象不存在"})
		return
	}

	// 3.提交分发请求
	orderid := model.Uuid()
	url := "http://192.168.99.111:8001/distribute/distribute_file"
	for _, filename := range p.Files {
		fileBase := filepath.Base(filename)

		i := strings.IndexByte(fileBase, '_')
		if i != 32 {
			ctx.JSON(http.StatusBadRequest, &ErrorResponse{"无效的分发文件"})
			return
		}
		bp := new(Distribute)
		bp.FileName = filename
		bp.Md5 = fileBase[:32]

		var instanceType string
		if strings.HasSuffix(fileBase, "agent.tar.gz") {
			instanceType = model.InstanceTypeAgent
			bp.DistributeType = DistributeTypeSSH
			to := struct {
				Port     int    `json:"port"`
				UserName string `json:"user_name"`
				Password string `json:"password"`
				IP       string `json:"host"`
			}{
				Port:     conf.SSH.Port,
				UserName: conf.SSH.Account,
				Password: conf.SSH.Password,
				IP:       obj.IP,
			}
			bp.ToType = "host"
			toByte, err := json.Marshal(&to)
			if err != nil {
				ctx.JSON(http.StatusBadRequest, &ErrorResponse{err.Error()})
				return
			}
			bp.To = string(toByte)
		} else if strings.HasSuffix(fileBase, ".tar.gz") {
			instanceType = model.InstanceTypeBeat
			bp.DistributeType = DistributeTypeAgent
			bp.To = fmt.Sprintf("%s:%d", obj.IP, conf.Agent.Port)
		}
		bp.FromType = DistributeFromTypeS3
		bp.From = fmt.Sprintf("{\"end_point\":\"%s\",\"access_key_id\":\"%s\",\"secret_access_key\":\"%s\"}", conf.Pacel.URL, conf.Pacel.Account, conf.Pacel.Password)
		bp.CallbackAddr = fmt.Sprintf("http://192.168.99.254:8080/object/distribute/hook")

		// 添加进度任务
		status := model.DistributeStatusStatrt
		message := "开始分发"
		mod := new(model.DistributeProgress)
		mod.OrderID = orderid
		mod.Type = instanceType
		mod.UserID = uInfo.ID
		mod.ObjectID = obj.ID
		mod.IP = obj.IP
		mod.Md5 = bp.Md5
		mod.Name = fileBase[33:]
		mod.All = len(p.Files)
		mod.Progress = 0
		mod.Status = status
		mod.Message = message
		mod.ID, err = model.AddDistributeProgress(mod)
		if err != nil {
			log.Error(err.Error())
			continue
		}

		// 分发
		bp.ID = mod.ID
		err = xhttp.Post(url, &bp, nil)
		if err != nil {
			log.Error(err.Error())
			status = model.DistributeStatusFailed
			message = err.Error()
			mod = &model.DistributeProgress{
				ID:      mod.ID,
				Status:  status,
				Message: message,
			}
			mod.Update()
		}
	}

	ctx.JSON(http.StatusOK, &HookData {
		Hook: fmt.Sprintf("/object/distribute?id=%s", orderid),
	})
}

type HookData struct {
	Hook string `json:"hook"`
}

// ParamObjectDistributeHook 对象分发进度回调参数
type ParamObjectDistributeHook struct {
	ID          string          `json:"id"`
	Md5         string          `json:"md5"`
	Code        int             `json:"code"`
	// Data        json.RawMessage `json:"data"`
	Data        []byte `json:"data"`
	ElaspedTime int64           `json:"elasped_time"`
	Message     string          `json:"message"`
}

// ObjectDistributeHook 对象分发回调

// ShowAccount godoc
// @Tags object
// @Summary 对象分发回调
// @Accept  application/json
// @Account json
// @Produce json
// @Param accept body ParamObjectDistributeHook true "ParamObjectDistributeHook"
// @Success 200 {object} Response
// @Failure 400 {object} Response
// @Router /object/distribute/hook [post]
func ObjectDistributeHook(ctx *gin.Context) {
	p := new(ParamObjectDistributeHook)
	err := ctx.BindJSON(p)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, &Response{err.Error()})
		return
	}

	log.Info(fmt.Sprintf("参数信息: %+v", p))
	status, message := "", ""
	switch p.Code {
	case 0:
		status = model.DistributeStatusSuccessful
		message = "分发成功"
		// 添加实例
		progress, err := model.OneDisributeProgress(p.ID)
		if err != nil {
			ctx.JSON(http.StatusOK, &Response{fmt.Sprintf("%s 分发文件回调状态处理失败: %v", p.ID, err)})
			return
		}

		if progress == nil {
			ctx.JSON(http.StatusOK, &Response{fmt.Sprintf("%s 无效的id参数值", p.ID)})
			return
		}
		ins := new(model.Instance)
		ins.UserID = progress.UserID
		ins.ObjectID = progress.ObjectID
		ins.IP = progress.IP
		ins.Type = progress.Type
		ins.Name = progress.Name
		ins.Status = status
		ins.Message = message
		ins.MD5 = p.Md5
		err = model.InstanceAdd(ins)
		if err != nil {
			ctx.JSON(http.StatusOK, &Response{fmt.Sprintf("%s 分发文件回调状态处理失败: %v", p.ID, err)})
			return
		}
	case 1:
		status = model.DistributeStatusFailed
		message = p.Message
	}

	mod := new(model.DistributeProgress)
	mod.ID = p.ID
	mod.Status = status
	mod.Message = message

	mod.Update()
	ctx.JSON(http.StatusOK, &Response{fmt.Sprintf("%s 分发文件回调状态已处理", mod.ID)})

}

// ParamObjectHearHook 对象心跳回调参数
type ParamObjectHearHook struct {
	ID      string `json:"id"`
	Status  string `json:"status"`
	Message string `json:"message"`
}

// ObjectHeartHook 回调心跳信息
// ShowAccount godoc
// @Tags object
// @Summary 回调心跳信息
// @Account json
// @Produce json
// @Param account body ParamObjectHearHook true "ParamObjectHearHook"
// @Failure 400 {object} Response
// @Router /object/heart [post]
func ObjectHeartHook(ctx *gin.Context) {
	p := new(ParamObjectHearHook)
	err := ctx.BindJSON(p)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, &Response{err.Error()})
		return
	}

	mod := new(model.Object)
	mod.ID = p.ID
	mod.Status = p.Status
	mod.Message = p.Message
	mod.Update()
}
