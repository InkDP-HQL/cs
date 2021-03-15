package controller

import (
	"fmt"
	"net/http"
	xhttp "ylsz/hit/http"
	"ylsz/hit/log"
	"ylsz/hitake/front/model"

	"github.com/gin-gonic/gin"

	"gopkg.in/yaml.v2"
)

// InstanceList 实例列表
// ShowAccount godoc
// @Tags object
// @Summary 实例列表
// @Produce json
// @Param Authorization header string true "Authorization"
// @Param objectId query string false "objectId"
// @Param name query string false "name"
// @Param popular query string false "popular"
// @Success 200 {array} model.Instance
// @Failure 400 {object} Response
// @Router /object/instance [get]
func InstanceList(ctx *gin.Context) {
	val, _ := ctx.Get("token")
	uinfo := val.(*UserInfo)

	param := make(map[string]interface{})
	param["userId"] = uinfo.ID

	if objid := ctx.Query("objectId"); objid != "" {
		param["objectId"] = objid
	}
	if name := ctx.Query("name"); name != "" {
		param["name"] = name
	}
	if _type := ctx.Query("type"); _type != "" {
		param["type"] = _type
	}

	list, err := model.InstanceList(param)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, &Response{err.Error()})
		return
	}
	ctx.JSON(http.StatusOK, &list)
}

// ParamInstanceAction 实例操作参数
type ParamInstanceAction struct {
	ID       string `json:"id"`
	ObjectID string `json:"objectId"`
	Action   string `json:"action"`
}

// InstanceAction 实例操作
// ShowAccount godoc
// @Tags object
// @Summary 实例操作
// @Accept json
// @Produce  json
// @Param Authorization header string true "Authorization"
// @Param account body ParamInstanceAction true "ParamInstanceAction"
// @Success 200 {object} Response
// @Failure 400 {object} ErrorResponse
// @Router /object/instance/action [post]
func InstanceAction(ctx *gin.Context) {
	val, _ := ctx.Get("token")
	uinfo := val.(*UserInfo)

	p := new(ParamInstanceAction)
	err := ctx.BindJSON(p)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, &Response{err.Error()})
		return
	}
	// 1.获取实例信息
	ins, err := model.OneInstance(p.ID)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, &Response{err.Error()})
		return
	}
	if ins == nil {
		ctx.JSON(http.StatusBadRequest, &Response{"无效的参数"})
		return
	}

	// 2.获取配置信息
	conf, err := model.OneConf(uinfo.ID)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, &Response{err.Error()})
		return
	}
	if conf == nil || conf.SSH.Account == "" || conf.SSH.Password == "" || conf.SSH.Port == 0 {
		ctx.JSON(http.StatusBadRequest, &Response{"请完善配置信息"})
		return
	}
	// 3.调动后端接口
	var (
		url string
		req interface{}
		res struct {
			Message string `json:"message"`
		}
	)
	switch ins.Type {
	case model.InstanceTypeAgent:
		url = "http://192.168.99.132:8001/v1/srsAgent"
		req = struct {
			Action  string `json:"action"`
			AgentID string `json:"agentId"`
			Address string `json:"address"`
			MD5     string `json:"md5"`
			//ServerPort  int    `json:"server_port"`
			SSHAccount  string `json:"account"`
			SSHPassword string `json:"password"`
			Port        int    `json:"port"`
		}{
			Action:  p.Action,
			AgentID: ins.ID,
			Address: ins.IP,
			//ServerPort:  conf.Agent.Port,
			MD5:         ins.MD5,
			SSHAccount:  conf.SSH.Account,
			SSHPassword: conf.SSH.Password,
			Port:        conf.Agent.Port,
		}
	case model.InstanceTypeBeat:
		url = "http://192.168.99.132:8001/v1/srsBeats"
		conf, err := model.OneConf(uinfo.ID)
		if err != nil {
			ctx.JSON(http.StatusBadRequest, &Response{err.Error()})
			return
		}
		config := struct {
			FilebeatInputs []struct {
				Type    string   `yaml:"type"`
				Enabled bool     `yaml:"enabled"`
				Paths   []string `yaml:"paths"`
			} `yaml:"filebeat.inputs"`
			FilebeatConfigModules struct {
				Path          string `yaml:"path"`
				ReloadEnabled bool   `yaml:"reload.enabled"`
			} `yaml:"filebeat.config.modules"`
			SetupTemplateSettings struct {
				IndexNumberOfShards int `yaml:"index.number_of_shards"`
			} `yaml:"setup.template.settings"`
			OutputElasticsearch struct {
				Hosts    []string `yaml:"hosts"`
				Protocol string   `yaml:"protocol"`
				//  Index    string   `yaml:"index"`
			} `yaml:"output.elasticsearch"`
		}{}
		config.FilebeatInputs = append(config.FilebeatInputs, struct {
			Type    string   `yaml:"type"`
			Enabled bool     `yaml:"enabled"`
			Paths   []string `yaml:"paths"`
		}{
			Type:    "log",
			Enabled: true,
			Paths:   conf.Filebeat.Inputs,
		})
		config.OutputElasticsearch.Hosts = []string{conf.Filebeat.Output}
		config.FilebeatConfigModules.Path = "${path.config}/modules.d/*.yml"
		config.OutputElasticsearch.Protocol = "http"
		config.SetupTemplateSettings.IndexNumberOfShards = 1
		// config.OutputElasticsearch.Index = "filebeat_log"
		out, err := yaml.Marshal(&config)
		if err != nil {
			ctx.JSON(http.StatusBadRequest, &Response{err.Error()})
			return
		}
		log.Info(fmt.Sprintf("配置信息: %s", string(out)))
		req = struct {
			Action string `json:"action"`
			IPPort string `json:"ipPort"`
			MD5    string `json:"md5"`
			Args   string `json:"args"`
			Config string `json:"config"`
			Name   string `json:"name"`
			Others string `json:"others"`
		}{
			Action: p.Action,
			IPPort: fmt.Sprintf("%s:%d", ins.IP, conf.Agent.Port),
			Name:   ins.Name,
			MD5:    ins.MD5,
			Config: string(out),
			Args:   "",
			Others: "",
		}

	}

	err = xhttp.Post(url, &req, &res)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, &Response{err.Error()})
		return
	}

	// 4.修改实例状态
	mod := new(model.Instance)
	mod.ID = p.ID
	switch p.Action {
	case "start", "restart":
		mod.Status = model.InstanceStatusRuning
	case "stop":
		mod.Status = model.InstanceStatusStoped
	}
	err = mod.Update()
	if err != nil {
		ctx.JSON(http.StatusBadRequest, &Response{err.Error()})
		return
	}

	ctx.JSON(http.StatusOK, &Response{"OK"})
}

type ObjectInfo struct {
	ID         string `json:"id"`
	ObjectID   string `json:"objectId"`
	Name       string `json:"name"`
	CreateTime string `json:"createTime"`
	Status     string `json:"status"`
	Message    string `json:"message"`
}


// ShowAccount godoc
// @Tags object
// @Summary 数据
// @Produce  json
// @Param Authorization header string true "Authorization"
// @Param highlight query string false "highlight"
// @Param startTime query string false "开始时间"
// @Param endTime query string false "结束时间"
// @Success 200 {array} LogInfo
// @Failure 400 {object} ErrorResponse
// @Router /object/log [get]
func InstanceLogList(ctx *gin.Context) {
	param := make(map[string]interface{})
	if highlight := ctx.Query("highlight"); highlight != "" {
		param["highlight"] = highlight
	}
	if startTime := ctx.Query("startTime"); startTime != "" {
		param["startTime"] = startTime
	}
	if endTime := ctx.Query("endTime"); endTime != "" {
		param["endTime"] = endTime
	}
	list, err := model.LogList(param)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, &ErrorResponse{err.Error()})
		return
	}

	data := []LogInfo{}
	for _, val := range list {
		data = append(data, LogInfo{
			ID:         model.Uuid(),
			Message:    val.Message,
			CreateTime: val.Timestamp,
		})
	}

	ctx.JSON(http.StatusOK, &data)
}

type LogInfo struct {
	ID         string `json:"id"` // id
	CreateTime string `json:"createTime"` // 创建时间
	Message    string `json:"message"` // 信息
}
