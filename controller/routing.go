package controller

import (
	"fmt"
	"net/http"
	"sort"
	"strings"

	"github.com/gin-gonic/gin"
)

type RoutesInfo []gin.RouteInfo

func Get(c *gin.Context) {
	routes, _ := c.Get("routes")
	routeList := routes.(gin.RoutesInfo)
	routesInfo := RoutesInfo(routeList)
	sort.Sort(routesInfo)
	path := c.Param("path")
	var rString = ""
	if path != "" {
		for _, v := range routesInfo {
			if strings.Index(v.Path, path) == 0 {
				rString += fmt.Sprintf("%s %s\n", v.Path, v.Method)
			}
		}
	} else {
		for _, v := range routesInfo {
			rString += fmt.Sprintf("%s %s\n", v.Path, v.Method)
		}
	}

	c.String(http.StatusOK, rString)
}

func (a RoutesInfo) Len() int { // 重写 Len() 方法
	return len(a)
}
func (a RoutesInfo) Swap(i, j int) { // 重写 Swap() 方法
	a[i], a[j] = a[j], a[i]
}
func (a RoutesInfo) Less(i, j int) bool { // 重写 Less() 方法， 从大到小排序
	return a[j].Path[1:] > a[i].Path[1:]
}
