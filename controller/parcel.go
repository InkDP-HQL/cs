package controller

import (
	"context"
	"fmt"
	"net/http"
	"path/filepath"
	"strings"

	"ylsz/hit/log"
	"ylsz/hit/route"

	"ylsz/hitake/front/model"

	"github.com/gin-gonic/gin"
	"github.com/minio/minio-go/v7"
	"github.com/minio/minio-go/v7/pkg/credentials"
)

type ParcelInfo struct {
	ID       string      `json:"id"`
	Name     string      `json:"title"`
	URL      string      `json:"url"`
	Checked  bool        `json:"checked"`
	Children interface{} `json:"children,omitempty"`
	Spread   interface{} `json:"spread,omitempty"`
}

// ShowAccount godoc
// @Tags parcel
// @Summary 存储库文件信息
// @Produce json
// @Param Authorization header string true "Authorization"
// @Success 200 {array} ParcelInfo
// @Failure 400 {object} ErrorResponse
// @Router /parcel [get]
func ParcelGet(ctx *gin.Context) {
	endpoint := "192.168.10.13:9000"
	accessKeyID := "minioadmin"
	secretAccessKey := "minioadmin"
	useSSL := false

	val, _ := ctx.Get("token")
	uinfo := val.(*UserInfo)

	conf, err := model.OneConf(uinfo.ID)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, &Response{err.Error()})
		return
	}

	if conf != nil {
		if conf.Pacel.URL != "" {
			endpoint = conf.Pacel.URL
		}
		if conf.Pacel.Account != "" {
			accessKeyID = conf.Pacel.Account
		}
		if conf.Pacel.Password != "" {
			secretAccessKey = conf.Pacel.Password
		}
	}

	minioClient, err := minio.New(endpoint, &minio.Options{
		Creds:  credentials.NewStaticV4(accessKeyID, secretAccessKey, ""),
		Secure: useSSL,
	})
	if err != nil {
		ctx.JSON(http.StatusBadRequest, &ErrorResponse{err.Error()})
		return
	}

	url := ctx.Query("url")

	data := make([]ParcelInfo, 0)
	if url == "" {
		buckets, err := minioClient.ListBuckets(context.Background())
		if err != nil {
			ctx.JSON(http.StatusBadRequest, &ErrorResponse{err.Error()})
			return
		}

		for _, bucket := range buckets {
			pi := ParcelInfo{
				Name: bucket.Name,
				URL:  bucket.Name,
			}
			pi.ID = pi.URL
			pi.Children = []struct{}{}
			pi.Spread = false
			data = append(data, pi)
		}
		ctx.JSON(http.StatusOK, &data)
		return
	}

	strs := strings.SplitN(url, "/", 2)
	bucket := strs[0]
	prefix := ""
	if len(strs) == 2 {
		prefix = strs[1]
	}

	if prefix != "" && !strings.HasSuffix(prefix, "/") {
		ctx.JSON(http.StatusOK, &data)
		return
	}

	log.Info(fmt.Sprintf("bucket: %s, prefix: %s", bucket, prefix))

	objects := minioClient.ListObjects(context.Background(), bucket, minio.ListObjectsOptions{
		Prefix: prefix,
	})
	for object := range objects {
		if object.Err != nil {
			continue
		}

		log.Info(fmt.Sprintf("s3对象信息: %+v", object))

		pi := ParcelInfo{
			Name: filepath.Base(object.Key),
			URL:  fmt.Sprintf("%s/%s", bucket, object.Key),
		}
		pi.ID = pi.URL
		if strings.HasSuffix(object.Key, "/") {
			pi.Children = []struct{}{}
			pi.Spread = false
		}
		data = append(data, pi)
	}

	ctx.JSON(http.StatusOK, &data)
	return
}

type instance struct {
	ID      string `json:"id"`
	Name    string `json:"name"`
	Message string `json:"message"`
	Status  string `json:"status"`
}
type ResponseParcelUse struct {
	ID          string     `json:"id"`
	Name        string     `json:"name"`
	URL         string     `json:"url"`
	AllNum      int        `json:"allNum"`
	InstallNum  int        `json:"installNum"`
	DownloadNum int        `json:"downloadNum"`
	RunNum      int        `json:"runNum"`
	Instances   []instance `json:"instances"`
	CreateTime  string     `json:"createTime"`
	Message     string     `json:"message"`
}

// ParcelUse 存储库文件分发情况
// ShowAccount godoc
// @Tags parcel
// @Summary 存储库文件分发情况
// @Produce json
// @Param Authorization header string true "Authorization"
// @Success 200 {array} ResponseParcelUse
// @Failure 400 {object} ErrorResponse
// @Router /parcel/use [get]
func ParcelUse(ctx *gin.Context) {
	val, _ := ctx.Get("token")
	uinfo := val.(*UserInfo)

	// 获取分发记录
	disList, err := model.DistributeProgressList(map[string]interface{}{
		"userId": uinfo.ID,
	})
	if err != nil {
		ctx.JSON(http.StatusBadRequest, &ErrorResponse{err.Error()})
		return
	}

	mData := map[string]*ResponseParcelUse{}

	for _, val := range disList {
		if _, ok := mData[val.Md5]; !ok {
			mData[val.Md5] = &ResponseParcelUse{
				ID:         val.Md5,
				Name:       val.Name,
				URL:        val.Name,
				CreateTime: val.CreateTime,
			}
		}

		mData[val.Md5].AllNum++
		switch val.Status {
		case model.DistributeStatusFailed:
		case model.DistributeStatusSuccessful:
			mData[val.Md5].InstallNum++
			mData[val.Md5].DownloadNum++
			mData[val.Md5].DownloadNum++
		case model.DistributeStatusStatrt:
		}

		mData[val.Md5].Instances = append(mData[val.Md5].Instances, instance{
			ID:      val.ID,
			Name:    val.IP,
			Message: val.Message,
			Status:  val.Status,
		})
	}

	data := make([]ResponseParcelUse, 0)
	for _, val := range mData {
		data = append(data, *val)
	}

	ctx.JSON(http.StatusOK, &data)
	return
}

func ParcelDistribute(ctx *route.Context) {

}

type DistributeParam struct {
	Hook         string `json:"hook"`
	ProgressHook string `json:"progressHook"`
	ID           string `json:"id"`
	IP           string `json:"ip"`
	GroupID      string `json:"groupid"`
	Files        []File `json:"files"`

	AgentPort int `json:"agentPort"`

	SSHAccount  string `json:"sshAccount"`
	SSHPort     int    `json:"sshPort"`
	SSHPassword string `json:"sshPassword"`
}

type File struct {
	ID           string `json:"id"`
	Name         string `json:"name"`
	DownloadPath string `json:"downloadPath"`
	InstallPath  string `json:"installPath"`
	Path         string `json:"path"`
	Type         string `json:"type"`

	S3Key    string `json:"s3key"`
	S3Secret string `json:"s3secret"`
}

type Parcel struct {
	Endpoint string
}

func getfils() {

}
