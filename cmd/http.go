package cmd

import (
	"fmt"
	"github.com/gin-contrib/sessions"
	"github.com/gin-contrib/sessions/cookie"
	"github.com/gin-gonic/gin"
	"github.com/lixiang4u/ShotTv-api/controller"
	"github.com/lixiang4u/ShotTv-api/util"
	go_websocket "github.com/lixiang4u/go-websocket"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"log"
)

var httpServerCmd = &cobra.Command{
	Use:   "serve",
	Short: "start http server",
	Run: func(cmd *cobra.Command, args []string) {
		log.Println(fmt.Sprintf("[AppPath] %s", util.AppPath()))

		//_ = autotls.Run(NewRouter(), "tv.artools.cc")
		//log.Println(viper.GetString("app.addr"))
		_ = NewRouter().Run(viper.GetString("app.addr"))

	},
}

func init() {
	rootCmd.AddCommand(httpServerCmd)
}

// 初始化websocket
func NewRouterW() go_websocket.WSWrapper {
	var ws = go_websocket.WSWrapper{}
	ws.Config.Debug = true

	ws.On("info", new(controller.HomeController).InfoW)           //注册列表数据查询
	ws.On("list", new(controller.HomeController).ListW)           //注册列表数据查询
	ws.On("broadcast", new(controller.HomeController).BroadcastW) //注册广播消息

	return ws
}

// 新建路由表
func NewRouter() *gin.Engine {
	r := gin.Default()

	// 使用session中间件
	r.Use(sessions.Sessions("shot_tv", cookie.NewStore([]byte(viper.GetString("app.secret")))))

	ws := NewRouterW()

	log.Println("[p]", fmt.Sprintf("%s/app/view/**/*", util.AppPath()))
	r.LoadHTMLGlob(fmt.Sprintf("%s/app/view/**/*", util.AppPath()))
	//r.LoadHTMLGlob("D:\\repo\\ShotTv-api\\app\\view\\**\\*")

	r.Static("/html", "./app/public/")
	r.Static("/upload", "./app/upload/")
	r.Static("/static", "./app/static/")
	r.Static("/m3u8", "./app/m3u8/")

	r.GET("/", new(controller.HomeController).Index)      // 默认首页
	r.GET("/hello", new(controller.HomeController).Hello) // 测试页
	r.POST("/api/play", new(controller.HomeController).Play)
	r.POST("/api/play/info", new(controller.HomeController).VideoPlayInfo)

	r.GET("/api/search", new(controller.ResourceController).Search)
	r.GET("/api/tag", new(controller.ResourceController).ListByTag)
	r.GET("/api/tag/:tagName", new(controller.ResourceController).ListByTag)
	r.GET("/api/info/:id", new(controller.ResourceController).Info)
	r.GET("/api/video/:id", new(controller.ResourceController).VideoSource)

	// 统一api
	r.GET("/api/env/predict", new(controller.HomeController).EnvPredict)
	r.GET("/api/video/search", new(controller.VideoController).Search)
	r.GET("/api/video/tag/:tagName", new(controller.VideoController).ListByTag)
	r.GET("/api/video/detail/:id", new(controller.VideoController).Detail) // 视频详细信息
	r.GET("/api/video/source/:id", new(controller.VideoController).Source) // 视频播放信息
	r.GET("/api/video/tag", new(controller.VideoController).ListByTagV2)   // 支持非路径参数
	r.GET("/api/video/detail", new(controller.VideoController).DetailV2)   // 支持非路径参数
	r.GET("/api/video/source", new(controller.VideoController).SourceV2)   // 支持非路径参数
	r.GET("/api/ws", func(context *gin.Context) {
		ws.Run(context.Writer, context.Request, nil)
	})

	r.GET("/tesla/index", new(controller.HomeController).TeslaIndex)
	r.GET("/tesla/fullscreen", new(controller.HomeController).FullScreen)

	r.GET("/home", new(controller.ResourceController).Home2)
	r.GET("/info/:id", new(controller.ResourceController).Info)

	r.GET("/ws", func(context *gin.Context) {
		ws.Run(context.Writer, context.Request, nil)
	})

	return r
}
