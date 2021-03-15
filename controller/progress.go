package controller

import (
	"net/http"

	"ylsz/hitake/front/model"

	"github.com/gin-gonic/gin"
)

type ProgressGetRes struct {
	Progress int `json:"progress"`
	List     []struct {
		DisableRegister bool           `json:"disableRegister"`
		Instance        model.Instance `json:"instance"`
	}
}

type progress struct {
	num  int
	list []struct {
		id    string
		title string
		list  []struct{}
	}
}

// ResponseGetDistributeProgress 获取分发进度结果
type ResponseGetDistributeProgress struct {
	Progress int              `json:"progress"`
	List     []distributeInfo `json:"data"`
}

type distributeInfo struct {
	ID       string `json:"id"`
	File     string `json:"file"`
	Progress int    `json:"progress"`
	Status   string `json:"status"`
	Message  string `json:"message"`
	IP       string `json:"ip"`
}

// GetDistributeProgress 获取分发进度
// ShowAccount godoc
// @Tags object
// @Summary 获取分发进度
// @Account json
// @Produce  json
// @Param Authorization header string true "Authorization"
// @Param id query string false "id"
// @Success 200 {object} ResponseGetDistributeProgress
// @Failure 400 {object} ErrorResponse
// @Router /object/distribute [get]
func GetDistributeProgress(ctx *gin.Context) {
	orderId := ctx.Query("id")
	list, err := model.DistributeProgressList(map[string]interface{}{"orderId": orderId})
	if err != nil {
		ctx.JSON(http.StatusBadRequest, &ErrorResponse{err.Error()})
		return
	}
	if len(list) == 0 {
		ctx.JSON(http.StatusBadRequest, &ErrorResponse{"invalid param"})
		return
	}

	res := new(ResponseGetDistributeProgress)
	current := 0
	for _, val := range list {
		pro := 0
		switch val.Status {
		case model.DistributeStatusFailed, model.DistributeStatusSuccessful:
			pro = 100
			current++
		case model.DistributeStatusStatrt:
			pro = 30
		}
		res.List = append(res.List, distributeInfo{
			ID:       val.ID,
			File:     val.Name,
			Status:   val.Status,
			Message:  val.Message,
			IP:       val.IP,
			Progress: pro,
		})
	}
	res.Progress = current * 100 / list[0].All
	ctx.JSON(http.StatusOK, res)
	return
}
