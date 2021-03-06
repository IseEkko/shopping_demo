package main

import (
	"github.com/kataras/iris/v12"
)

func main() {
	//1.创建iris 实例
	app := iris.New()

	//2.设置模板
	app.HandleDir("/public", "./fronted/web/public")
	//访问生成好的html静态文件
	app.HandleDir("/html", "./fronted/web/htmlProductShow")
	app.Run(
		iris.Addr("0.0.0.0:3030"),
		iris.WithoutServerError(iris.ErrServerClosed),
		iris.WithOptimizations,
	)

}
