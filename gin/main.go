package main

import (
	"github.com/gin-gonic/gin"
	"net/http"
)

func main() {
	server := gin.Default() //创建一个web服务器（engine）

	// 静态路由
	server.GET("/hello", func(ctx *gin.Context) { //注册路由	"hello"通过localhost:8080/hello可以访问网页，gin.Context主要负责处理请求，和返回响应
		ctx.String(http.StatusOK, "hello,world") //String方法写回响应到前端
	})

	// 参数路由，路径参数
	server.GET("/users/:name", func(ctx *gin.Context) {
		name := ctx.Param("name")
		ctx.String(http.StatusOK, "hello,"+name) //通过传入的参数来组成路由访问，使用Param方法读取参数
	})

	// 查询参数
	// GET /order?id=123
	server.GET("/order", func(ctx *gin.Context) {
		id := ctx.Query("id")
		ctx.String(http.StatusOK, "订单id是"+id)
	})

	// 通配符路由
	server.GET("/views/*.html", func(ctx *gin.Context) { // 注册路由时 * 不能单独出现
		view := ctx.Param(".html") // 拿到的是 /xx.html 整串
		ctx.String(http.StatusOK, "view 是"+view)
	})

	server.POST("/login", func(ctx *gin.Context) {
		ctx.String(http.StatusOK, "hello, login")
	}) //直接打开网页显示404，发送的是Get请求，可通过postman发送POST请求

	//如果不传参数，实际上监听的是 8080 端口；建议主动传参
	server.Run(":8080")
	//这种写法是错误的
	//server.Run("8080")

}

// 用户是查询数据的，使用 GET 请求，参数放到查询参数里面，即 ?a=123 这种
// 用户是提交数据的，使用 POST 请求，参数全部放到 Body 里面
