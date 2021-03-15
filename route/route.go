package route
//
// import (
// 	"ylsz/hit/route"
// 	"ylsz/hitake/front/controller"
// )
//
// // Init 注册路由
// func Init(r *route.Route) {
// 	r.Use(route.CORS)
//
// 	// 注册账户
// 	r.Post("/register", controller.AccountRegister)
// 	// 登录
// 	r.Post("/login", controller.AccountLogin)
//
// 	// 后端回调结果
// 	r.Post("/progress/_hook", controller.ProgressPost)
// 	r.Post("/server/discover/_hook", controller.DiscoverHook)
// 	r.Post("/server/distribute/_hook", controller.InstanceDistributeHook)
//
// 	g := r.Group("", route.Token(controller.TokenHeader, controller.TokenKey, controller.UserInfo{}))
// 	{
// 		// 获取用户信息
// 		g.Get("/member", controller.AccountInfo)
// 		// 更新存储库信息
// 		g.Post("/member/parcel", controller.AccountUpdateParcel)
//
// 		// 提交分发记录
// 		g.Post("/parcel", controller.AccountUpdateParcel)
// 		// 获取存储库目录文件
// 		g.Get("/parcel", controller.ParcelGet)
//
// 		// 获取标签
// 		g.Get("/tag", controller.TagGet)
// 		// 添加标签
// 		g.Post("/tag", controller.TagAdd)
//
// 		// 提交服务发现条件
// 		g.Post("/server/discover", controller.DiscoverPost)
// 		// 获取发现结果
// 		g.Get("/server/discover", controller.DiscoverGet)
// 		// 获取服务发现进度
// 		g.Get("/progress", controller.ProgressGet)
//
// 		// 获取对象列表
// 		g.Get("/instance", controller.InstanceGet)
// 		// 对象中安装的实例列表
// 		// g.Get("/object", func(ctx *route.Context) {
// 		// 	data := []struct {
// 		// 		ID         string `json:"id"`
// 		// 		Name       string `json:"name"`
// 		// 		CreateTime string `json:"createTime"`
// 		// 		Status     string `json:"status"`
// 		// 		Message    string `json:"message"`
// 		// 	}{
// 		// 		{"910dto1j3a", "agent", "2021-02-25 12:05:12", "install", ""},
// 		// 		{"910dto1j3b", "filebeat", "2021-02-25 12:05:12", "runing", ""},
// 		// 		{"910dto1j3c", "metribeat", "2021-02-25 12:05:12", "stop", ""},
// 		// 	}
// 		// 	ctx.JSON(&data)
// 		// })
// 		g.Get("/object", controller.InstanceObjectList)
// 		// 注册对象
// 		g.Post("/instance", controller.InstancePost)
// 		// 执行分发任务
// 		g.Post("/instance/distribute", controller.InstanceDistribute)
// 		// 获取分发结果
// 		g.Get("/instance/distribute", controller.InstanceDistributeGet)
// 		// 启停服务
// 		// g.Post("/instance/ss", func(ctx *route.Context) {
// 		// 	// data := struct {
// 		// 	// 	ID      string `json:"id"`
// 		// 	// 	Status  string `json:"status"`
// 		// 	// 	Message string `json:"message"`
// 		// 	// }{
// 		// 	// 	ID:     ctx.GetString("id"),
// 		// 	// 	Status: "runing",
// 		// 	// }
//
// 		// 	data := []struct {
// 		// 		ID         string `json:"id"`
// 		// 		Name       string `json:"name"`
// 		// 		CreateTime string `json:"createTime"`
// 		// 		Status     string `json:"status"`
// 		// 		Message    string `json:"message"`
// 		// 	}{
// 		// 		{"910dto1j3a", "agent", "2021-02-25 12:05:12", "install", ""},
// 		// 		{"910dto1j3b", "filebeat", "2021-02-25 12:05:12", "stop", ""},
// 		// 		{"910dto1j3c", "metribeat", "2021-02-25 12:05:12", "runing", ""},
// 		// 	}
// 		// 	ctx.JSON(&data)
// 		// })
// 		g.Post("/instance/ss", controller.InstanceStart)
//
// 		// 获取采集日志
// 		// g.Get("/instance/log", func(ctx *route.Context) {
// 		// 	data := []struct {
// 		// 		ID         string `json:"id"`
// 		// 		CreateTime string `json:"createTime"`
// 		// 		Message    string `json:"message"`
// 		// 	}{
// 		// 		{"910dto1j3a", "2021-02-25 12:05:12", "{ETag: Key:test/ LastModified:0001-01-01 00:00:00 +0000 UTC Size:0 ContentType: Expires:0001-01-01 00:00:00 +0000 UTC Metadata:map[] UserMetadata:map[] UserTags:map[] UserTagCount:0 Owner:{DisplayName: ID:} Grant:[] StorageClass: IsLatest:false IsDeleteMarker:false VersionID: ReplicationStatus: Expiration:0001-01-01 00:00:00 +0000 UTC ExpirationRuleID: Err:<nil>}"},
// 		// 		{"910dto1j3b", "2021-02-25 12:05:12", "{ETag: Key:test/ LastModified:0001-01-01 00:00:00 +0000 UTC Size:0 ContentType: Expires:0001-01-01 00:00:00 +0000 UTC Metadata:map[] UserMetadata:map[] UserTags:map[] UserTagCount:0 Owner:{DisplayName: ID:} Grant:[] StorageClass: IsLatest:false IsDeleteMarker:false VersionID: ReplicationStatus: Expiration:0001-01-01 00:00:00 +0000 UTC ExpirationRuleID: Err:<nil>}"},
// 		// 		{"910dto1j3c", "2021-02-25 12:05:12", "{ETag: Key:test/ LastModified:0001-01-01 00:00:00 +0000 UTC Size:0 ContentType: Expires:0001-01-01 00:00:00 +0000 UTC Metadata:map[] UserMetadata:map[] UserTags:map[] UserTagCount:0 Owner:{DisplayName: ID:} Grant:[] StorageClass: IsLatest:false IsDeleteMarker:false VersionID: ReplicationStatus: Expiration:0001-01-01 00:00:00 +0000 UTC ExpirationRuleID: Err:<nil>}"},
// 		// 		{"910dto1j3d", "2021-02-25 12:05:12", "{ETag: Key:test/ LastModified:0001-01-01 00:00:00 +0000 UTC Size:0 ContentType: Expires:0001-01-01 00:00:00 +0000 UTC Metadata:map[] UserMetadata:map[] UserTags:map[] UserTagCount:0 Owner:{DisplayName: ID:} Grant:[] StorageClass: IsLatest:false IsDeleteMarker:false VersionID: ReplicationStatus: Expiration:0001-01-01 00:00:00 +0000 UTC ExpirationRuleID: Err:<nil>}"},
// 		// 		{"910dto1j3e", "2021-02-25 12:05:12", "{ETag: Key:test/ LastModified:0001-01-01 00:00:00 +0000 UTC Size:0 ContentType: Expires:0001-01-01 00:00:00 +0000 UTC Metadata:map[] UserMetadata:map[] UserTags:map[] UserTagCount:0 Owner:{DisplayName: ID:} Grant:[] StorageClass: IsLatest:false IsDeleteMarker:false VersionID: ReplicationStatus: Expiration:0001-01-01 00:00:00 +0000 UTC ExpirationRuleID: Err:<nil>}"},
// 		// 		{"910dto1j3f", "2021-02-25 12:05:12", "{ETag: Key:test/ LastModified:0001-01-01 00:00:00 +0000 UTC Size:0 ContentType: Expires:0001-01-01 00:00:00 +0000 UTC Metadata:map[] UserMetadata:map[] UserTags:map[] UserTagCount:0 Owner:{DisplayName: ID:} Grant:[] StorageClass: IsLatest:false IsDeleteMarker:false VersionID: ReplicationStatus: Expiration:0001-01-01 00:00:00 +0000 UTC ExpirationRuleID: Err:<nil>}"},
// 		// 		{"910dto1j3g", "2021-02-25 12:05:12", "{ETag: Key:test/ LastModified:0001-01-01 00:00:00 +0000 UTC Size:0 ContentType: Expires:0001-01-01 00:00:00 +0000 UTC Metadata:map[] UserMetadata:map[] UserTags:map[] UserTagCount:0 Owner:{DisplayName: ID:} Grant:[] StorageClass: IsLatest:false IsDeleteMarker:false VersionID: ReplicationStatus: Expiration:0001-01-01 00:00:00 +0000 UTC ExpirationRuleID: Err:<nil>}"},
// 		// 		{"910dto1j3h", "2021-02-25 12:05:12", "{ETag: Key:test/ LastModified:0001-01-01 00:00:00 +0000 UTC Size:0 ContentType: Expires:0001-01-01 00:00:00 +0000 UTC Metadata:map[] UserMetadata:map[] UserTags:map[] UserTagCount:0 Owner:{DisplayName: ID:} Grant:[] StorageClass: IsLatest:false IsDeleteMarker:false VersionID: ReplicationStatus: Expiration:0001-01-01 00:00:00 +0000 UTC ExpirationRuleID: Err:<nil>}"},
// 		// 		{"910dto1j3i", "2021-02-25 12:05:12", "{ETag: Key:test/ LastModified:0001-01-01 00:00:00 +0000 UTC Size:0 ContentType: Expires:0001-01-01 00:00:00 +0000 UTC Metadata:map[] UserMetadata:map[] UserTags:map[] UserTagCount:0 Owner:{DisplayName: ID:} Grant:[] StorageClass: IsLatest:false IsDeleteMarker:false VersionID: ReplicationStatus: Expiration:0001-01-01 00:00:00 +0000 UTC ExpirationRuleID: Err:<nil>}"},
// 		// 	}
// 		// 	ctx.JSON(&data)
// 		// })
//
// 		g.Get("/instance/log", controller.InstanceLogList)
// 		g.Get("/parcel/use", func(ctx *route.Context) {
// 			type instance struct {
// 				ID      string `json:"id"`
// 				Name    string `json:"name"`
// 				Message string `json:"message"`
// 				Status  string `json:"status"`
// 			}
// 			data := []struct {
// 				ID          string     `json:"id"`
// 				Name        string     `json:"name"`
// 				URL         string     `json:"url"`
// 				AllNum      int        `json:"allNum"`
// 				InstallNum  int        `json:"installNum"`
// 				DownloadNum int        `json:"downloadNum"`
// 				RunNum      int        `json:"runNum"`
// 				Instances   []instance `json:"instances"`
// 				CreateTime  string     `json:"createTime"`
// 				Message     string     `json:"message"`
// 			}{
// 				{
// 					ID:          "910rt72m1",
// 					Name:        "agent.tar.gz",
// 					URL:         "hitake-0.1.10/agent/agent.tar.gz",
// 					AllNum:      3,
// 					InstallNum:  1,
// 					DownloadNum: 1,
// 					RunNum:      1,
// 					Instances: []instance{
// 						{"3o29d8d", "192.168.10.120", "", "distributing"},
// 					},
// 				},
// 				{
// 					ID:          "910rt73m1",
// 					Name:        "filebeat.tar.gz",
// 					URL:         "hitake-0.1.10/agent/agent.tar.gz",
// 					AllNum:      3,
// 					InstallNum:  1,
// 					DownloadNum: 1,
// 					RunNum:      1,
// 					Instances: []instance{
// 						{"3o29d8d", "192.168.10.120", "", "distributing"},
// 					},
// 				},
// 			}
// 			ctx.JSON(&data)
// 		})
//
// 		g.Post("/parcel/distribute", controller.InstanceDistribute)
// 		// 获取实例列表
// 	}
//
// }
