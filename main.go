package main

import (
	"flag"
	"github.com/gin-gonic/gin"
	"github.com/guotie/config"
	"github.com/guotie/deferinit"
	"github.com/smtc/glog"
	"runtime"
	"strings"
	"os"
	"os/signal"
	"syscall"
)

var (
	configFn   = flag.String("config", "./config.json", "config file path") //配置文件地址
	debugFlag  = flag.Bool("d", false, "debug mode")                        //是否为调试模式
	rootPrefix string                                                       //二级目录地址
	tempDir    string                                                       //模版目录
	contentDir string                                                       //脚本目录
	rt         *gin.Engine
temp string //临时文件存储目录
cookieName string //cookie名称
)

/**
主函数入口
创建人:邵炜
创建时间:2016年2月26日11:22:03
*/
func main() {

	if checkPid() { //判断程序是否启动
		return
	}

	flag.Parse()

	serverRun(*configFn, *debugFlag)

	c := make(chan os.Signal, 1)
	writePid()
	// 信号处理
	signal.Notify(c, os.Interrupt, os.Kill, syscall.SIGTERM)
	// 等待信号
	<-c

	serverExit()
	rmPidFile()
	imageFileDelete()
	os.Exit(0)
}

/**
服务启动
创建人:邵炜
创建时间:2016年2月26日11:22:16
输入参数: cfn(配置文件地址) debug(是否为调试模式)
*/
func serverRun(cfn string, debug bool) {

	config.ReadCfg(cfn)

	logInit(debug)

	rootPrefix = strings.TrimSpace(config.GetStringMust("rootPrefix"))
	tempDir = strings.TrimSpace(config.GetStringMust("tempDir"))
	contentDir = strings.TrimSpace(config.GetStringMust("contentDir"))
	port:=strings.TrimSpace(config.GetStringMust("port"))
	temp=strings.TrimSpace(config.GetStringMust("temp"))
	cookieName=strings.TrimSpace(config.GetStringMust("cookieName"))

	if len(rootPrefix) != 0 {
		if !strings.HasPrefix(rootPrefix, "/") {
			rootPrefix = "/" + rootPrefix
		}
		if strings.HasSuffix(rootPrefix, "/") {
			rootPrefix = rootPrefix[0 : len(rootPrefix)-1]
		}
	}

	//初始化所有go文件中的init内方法
	deferinit.InitAll()
	glog.Info("init all module successfully \n")

	//设置多CPU运行
	runtime.GOMAXPROCS(runtime.NumCPU())
	glog.Info("set many cpu successfully \n")

	//启动所有go文件中的init方法
	deferinit.RunRoutines()
	glog.Info("init all run successfully \n")

	if !debug {
		gin.SetMode(gin.ReleaseMode)
	}

	rt = gin.Default()

	router(rt)

	go rt.Run(port)
}

/**
验证授权 重载gin
创建人:邵炜
创建时间:2016年3月8日18:01:55
 */
func verifyAuthorization() gin.HandlerFunc {
	return func(c *gin.Context) {
		_,err:=c.Request.Cookie(cookieName) //未实现
		if err != nil {
			glog.Error("can't find coodie , cookieName: %s ,err: %s \n",cookieName,err.Error())
			return
		}


	}
}


/**
结束进程
创建人:邵炜
创建时间:2016年3月7日14:21:24
 */
func serverExit() {
	// 结束所有go routine
	deferinit.StopRoutines()
	glog.Info("stop routine successfully.\n")

	deferinit.FiniAll()
	glog.Info("fini all modules successfully.\n")

	glog.Close()
}

/**
路由配置
创建人:邵炜
创建时间:2016年3月7日10:15:20
输入参数: gin对象
*/
func router(r *gin.Engine) {
	g := &r.RouterGroup
	if rootPrefix != "" {
		g = r.Group(rootPrefix)
	}

	{
		g.GET("/", func(c *gin.Context) { c.String(200, "ok") })

		g.GET("/assets/*pth", assetsFiles)

		g.GET("/images/:name",imageGet)
	}
}
