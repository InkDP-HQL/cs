package main

import (
	"fmt"
	"net/http"
	"os"

	"github.com/dgrijalva/jwt-go"
	"ylsz/hit/log"

	"ylsz/hitake/front/controller"
	"ylsz/hitake/front/model"

	"github.com/gin-gonic/gin"
	swaggerFiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"
	"github.com/urfave/cli/v2"

	_ "ylsz/hitake/front/docs"
)

var App = &cli.App{
	Flags: []cli.Flag{
		&cli.StringFlag{
			Name:    "port",
			Aliases: []string{"p"},
			Value:   "8080",
			Usage:   "hitake-front -p 8080",
			EnvVars: []string{"HITAKE_FRONT_PORT"},
		},
		&cli.StringFlag{
			Name:    "esAddr",
			Aliases: []string{"a"},
			Value:   "http://192.168.10.15:9200",
			Usage:   "hitake-front -esAddr http://192.168.10.15:9200",
			EnvVars: []string{"HITAKE_FRONT_ESADDR"},
		},
		&cli.StringFlag{
			Name:    "backendPoint",
			Aliases: []string{"bp"},
			Value:   "http://192.168.99.254:9090",
			Usage:   "hitake-front -bp http://192.168.10.15:9090",
			EnvVars: []string{"HITAKE_FRONT_ESADDR"},
		},
	},
	Action: func(ctx *cli.Context) error {
		model.ElasticAddr = ctx.String("esAddr")
		router := gin.Default()
		// swagger文档
		router.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerFiles.Handler))

		// 心跳回调
		router.POST("/object/heart", controller.ObjectHeartHook)

		// 分发进度回调
		router.POST("/object/distribute/hook", controller.ObjectDistributeHook)

		iRouter := router.Group("").Use(func(ctx *gin.Context) {
			log.Info(fmt.Sprintf("request header: %+v", ctx.GetHeader("origin")))
			ctx.Header("Access-Control-Allow-Origin", ctx.GetHeader("Origin"))
			if ctx.Request.Method == http.MethodOptions {
				ctx.Header("Access-Control-Allow-Headers", ctx.GetHeader("Access-Control-Request-Headers"))
				ctx.Header("Access-Control-Allow-Methods", ctx.GetHeader("Access-Control-Request-Method"))
				ctx.Header("Access-Control-Max-Age", "1728000")
				ctx.Writer.WriteHeader(204)
				ctx.Abort()
			}
			return
		})
		// 用户登录
		iRouter.POST("/login", controller.AccountLogin)
		iRouter.POST("/register", controller.AccountRegister)
		tRouter := iRouter.Use(func(ctx *gin.Context) {
			token := ctx.GetHeader(controller.TokenHeader)
			if token == "" {
				log.Error("token为空")
				ctx.JSON(http.StatusUnauthorized, &struct {
					Message string `json:"message"`
				}{"invalid token"})
				ctx.Abort()
				return
			}

			_v := new(controller.UserInfo)
			_token, err := jwt.ParseWithClaims(token, _v, func(token *jwt.Token) (v interface{}, err error) {
				return []byte(controller.TokenKey), nil
			})
			if err != nil {
				log.Error(err.Error())
				ctx.JSON(http.StatusUnauthorized, &struct {
					Message string `json:"message"`
				}{"invalid token"})
				ctx.Abort()
				return
			}
			if !_token.Valid {
				log.Error("token验证失败")
				ctx.JSON(http.StatusUnauthorized, &struct {
					Message string `json:"message"`
				}{"invalid token"})
				ctx.Abort()
				return
			}
			ctx.Set("token", _token.Claims)
			return
		})
		// {
		tRouter.GET("/member", controller.AccountInfo)

		// 对象列表
		tRouter.GET("/object", controller.ObjectList)
		tRouter.POST("/object", controller.AddObject)
		tRouter.PUT("/object", controller.RegisterObject)

		// 标签
		tRouter.GET("/tag", controller.TagGet)
		tRouter.GET("/tag/all", controller.AllTag)
		tRouter.POST("/tag", controller.TagAdd)

		// 实例
		tRouter.GET("/object/instance", controller.InstanceList)

		// keystore
		tRouter.GET("/user/keystore", controller.KeystoreList)
		tRouter.POST("/user/keystore", controller.AddKeystore)
		tRouter.PUT("/user/keystore", controller.PutKeystore)
		tRouter.DELETE("/user/keystore", controller.DeleteKeystore)

		// 配置信息
		tRouter.GET("/object/conf", controller.GetConf)
		tRouter.POST("/object/conf", controller.UpdateConf)

		// 存储库文件信息
		tRouter.GET("/parcel", controller.ParcelGet)
		tRouter.GET("/parcel/use", controller.ParcelUse)

		// 分发
		tRouter.GET("/object/distribute", controller.GetDistributeProgress)
		tRouter.POST("/object/distribute", controller.ObjectDistribute)

		// 启停
		tRouter.POST("/object/instance/action", controller.InstanceAction)

		// 数据
		tRouter.GET("/object/log", controller.InstanceLogList)
		// }

		router.GET("/routing/*path", getRouting(router), controller.Get)

		return router.Run(":" + ctx.String("port"))

	},
}


// @title backend_server
// @version 1.0
// @host 8080
func main() {
	err := App.Run(os.Args)
	if err != nil {
		log.Error("启动失败", "err", err)
	}
}

func getRouting(route *gin.Engine) gin.HandlerFunc {
	return func(c *gin.Context) {
		routes := route.Routes()
		c.Set("routes", routes)
		c.Next()
	}

}
