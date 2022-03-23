package main

import (
	"context"
	"github.com/kataras/iris/v12"
	"github.com/kataras/iris/v12/mvc"
	"imooc-product/common"

	"github.com/kataras/iris/v12/sessions"
	"imooc-product/fronted/web/controllers"
	"imooc-product/repositories"
	"imooc-product/services"
	"time"
)

func main() {
	//1.创建iris 实例
	app := iris.New()
	//2.设置错误模式，在mvc模式下提示错误
	app.Logger().SetLevel("debug")
	//3.注册模板
	/**
	因为模版中的问题，导入的变量是空长所以不行。
	*/
	tmplate := iris.HTML("./fronted/web/views", ".html").Layout("shared/layout.html").Reload(true)
	app.RegisterView(tmplate)
	//4.设置模板
	app.HandleDir("/public", "./fronted/web/public")
	//app.HandleDir("/html", "./fronted/web/htmlProductShow")
	//出现异常跳转到指定页面
	app.OnAnyErrorCode(func(ctx iris.Context) {
		ctx.ViewData("message", ctx.Values().GetStringDefault("message", "访问的页面出错！"))
		ctx.ViewLayout("")
		ctx.View("shared/error.html")
	})

	//连接数据库
	db, err := common.NewMysqlConn()
	if err != nil {

	}
	sess := sessions.New(sessions.Config{
		Cookie:  "AdminCookie",
		Expires: 600 * time.Minute,
	})
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	user := repositories.NewUserMangerRepository("user", db)
	userService := services.NewUservice(user)
	userPro := mvc.New(app.Party("/user"))
	userPro.Register(ctx, userService, sess.Start)
	userPro.Handle(new(controllers.UserController))

	app.Run(
		iris.Addr("0.0.0.0:8082"),
		iris.WithoutServerError(iris.ErrServerClosed),
		iris.WithOptimizations,
	)

}
