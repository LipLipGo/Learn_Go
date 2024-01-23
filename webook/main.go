package main

import (
	"bytes"
	"github.com/fsnotify/fsnotify"
	"github.com/gin-gonic/gin"
	"github.com/spf13/pflag"
	"github.com/spf13/viper"
	_ "github.com/spf13/viper/remote"
	"go.uber.org/zap"
	"log"
	"net/http"
)

func main() {
	initViperRemoteWatch()
	initLogger() // 一般来说，需要先读取一些配置，再初始化日志模块
	server := InitWebServer()
	//db := initDB()
	//rd := initRedis()
	//server := initWebServer()
	//codeSvc := initCodeSvc(rd)
	//initUserHdl(db, server, rd, codeSvc)

	// 首先去除mysql和redis依赖，构造最简单的Web服务部署到k8s上

	server.GET("/hello", func(ctx *gin.Context) {
		ctx.String(http.StatusOK, "hello，启动成功了！")
	})

	server.Run(":8080")
}

func initLogger() {
	Logger, err := zap.NewDevelopment() // 开发环境使用，线上环境使用 NewProduction
	if err != nil {
		panic(err)
	}

	zap.ReplaceGlobals(Logger) // 使用我们创建的 Logger 替换掉里面的全局包变量
}

// 第一种写法
func initViper() {
	viper.SetConfigName("dev")    // 配置文件名
	viper.SetConfigType("yaml")   // 配置文件类型
	viper.AddConfigPath("config") // 当前工作目录的 config 子目录
	err := viper.ReadInConfig()
	if err != nil {
		panic(err)
	}
	val := viper.Get("test.key")
	log.Println(val)
}

// 第二种写法，直接设置配置文件路径
func initViperV1() {
	////viper.SetDefault("db.dsn", "localhost:3306") // 默认值，或者在初始化结构体配置的时候给一个默认值
	//viper.SetConfigFile("config/dev.yaml")
	//err := viper.ReadInConfig()
	//if err != nil {
	//	panic(err)
	//}
	//val := viper.Get("test.key")
	//log.Println(val)

	// 通过读取启动参数来设置不同环境的配置
	cfile := pflag.String("config", "config/config.yaml",
		"配置文件路径") // 这里得到的是指针
	pflag.Parse() // 这一步之后 cfile 里面才有值
	viper.SetConfigFile(*cfile)
	err := viper.ReadInConfig()
	if err != nil {
		panic(err)
	}

}

func initViperWatch() {
	cfile := pflag.String("config", "config/config.yaml",
		"配置文件路径") // 这里得到的是指针
	pflag.Parse() // 这一步之后 cfile 里面才有值
	viper.SetConfigFile(*cfile)

	// 在这里监听配置文件的变更
	viper.WatchConfig()
	viper.OnConfigChange(func(in fsnotify.Event) {
		log.Println(viper.GetString("test.key")) // 输出变更后的配置
	})
	err := viper.ReadInConfig()
	if err != nil {
		panic(err)
	}
}

// 这种方式一般就是在测试或者调试的阶段使用
func initViperV2() {
	// 直接将配置文件内容以字符串的形式定义
	cfg := `		
test:
  key: value1

redis:
  addr: "localhost:6379"

db:
  dsn: "root:root@tcp(localhost:13316)/webook"

`
	err := viper.ReadConfig(bytes.NewReader([]byte(cfg)))
	if err != nil {
		panic(err)
	}
}

// 使用 viper 接入 etcd （远程配置中心）
func initViperRemote() {
	err := viper.AddRemoteProvider("etcd3", "http://127.0.0.1:12379",
		"C:/Program Files/Git/webook") // 整个项目的配置放在 etcd 中的 path 路径下
	if err != nil {
		panic(err)
	}
	viper.SetConfigType("yaml")    // 配置的文件格式
	err = viper.ReadRemoteConfig() // 将远程的配置拉到本地
	if err != nil {
		panic(err)
	}
}

// 监听远程配置中心变更
func initViperRemoteWatch() {
	err := viper.AddRemoteProvider("etcd3", "http://127.0.0.1:12379",
		"C:/Program Files/Git/webook") // 整个项目的配置放在 etcd 中的 path 路径下
	if err != nil {
		panic(err)
	}
	viper.SetConfigType("yaml") // 配置的文件格式

	err = viper.ReadRemoteConfig() // 将远程的配置拉到本地
	if err != nil {
		panic(err)
	}

	// 新开一个 go routine 来监听变更，因为 viper 不是线程安全的，并且需要放在这后面，防止一边读写并发安全问题
	go func() {
		for {
			err = viper.WatchRemoteConfig()
			if err != nil {
				panic(err)
			}
			//log.Println("Watch:", viper.GetString("test.key"))
			//time.Sleep(time.Second * 3)
		}
	}()

}

// 下面是使用wire改造前的代码
//
//	func initUserHdl(db *gorm.DB, server *gin.Engine, rd redis.Cmdable, codeSvc *service.codeService) {
//		ud := dao.NewGORMUserDao(db)
//		uc := cache.NewRedisUserCache(rd)
//		ur := repository.NewCachedUserRepository(ud, uc)
//		us := service.NewuserService(ur)
//
//		//hdl := &web.UserHandler{}
//		hdl := web.NewUserHandler(us, codeSvc) // 设置了正则表达式预编译后需要替换为这个方法
//		hdl.RegisterRouters(server)
//	}
//
//	func initCodeSvc(redisClient redis.Cmdable) *service.codeService {
//		cc := cache.NewRedisCodeCache(redisClient)
//		cRepo := repository.NewCodeRepository(cc)
//		return service.NewcodeService(cRepo, initMemorySms())
//
// }
//
//	func initMemorySms() sms.Service {
//		return localsms.NewService()
//	}
//
//	func initWebServer() *gin.Engine {
//		server := gin.Default()
//
//		// 解决跨域问题
//		// 跨域请求：请求是从 localhost:3000 前端发送到 localhost:8080 后端的
//		// 类似这种就是跨域请求，协议、域名和端口任意一个不同，都是跨域请求
//		// 如果不做额外处理，没办法发送请求的	；浏览器会发送一个预检请求 preflight 给后端，询问是否接收请求
//
//		/*跨域问题要点：
//		1.跨域问题是因为发请求的 协议+域名+端口 和接受请求的 协议+域名+端口 对不上，比如 localhost:3000 发到 localhost:8080  上
//		2.解决跨域问题的关键是在 preflight 请求里告诉浏览器自己愿意接受请求
//		3.Gin 提供了解决跨域问题的 middleware ，可以直接使用
//		4.middleware 是一种机制，可以用来解决一些所有业务都关心的问题，使用 Use 方法来注册 middleware
//		*/
//
//		server.Use(cors.New(cors.Config{ // 通过 Middleware（cors） 处理跨域请求
//			//AllowAllOrigins: true,	允许所有的源头
//			//AllowOrigins: []string{"http://localhost:3000"}, //允许一些
//			//AllowMethods: []string{"POST"},   最好不要设置，允许所有请求方法即可
//			AllowCredentials: true,                                      // cookie的数据是否允许传过来，正常情况下允许
//			AllowHeaders:     []string{"Content-Type", "Authorization"}, //报错，根据报错找到需要添加的headers
//			// 允许前端访问后端响应中带的头部
//			ExposeHeaders: []string{"x-jwt-token"},
//			AllowOriginFunc: func(origin string) bool {
//				if strings.HasPrefix(origin, "http://localhost") {
//					return true
//				}
//
//				return strings.Contains(origin, "your_company.com")
//			},
//			MaxAge: 12 * time.Hour, // preflight检测时长，无影响
//		}), func(ctx *gin.Context) {
//			fmt.Println("这是一个 Middleware")
//		})
//
//		//redisClient := redis.NewClient(&redis.Options{
//		//	Addr: config.Config.Redis.Addr,
//		//})
//		//// 一秒钟100次
//		//server.Use(ratelimit.NewBuilder(redisClient, time.Second, 100).Build())
//
//		//useSession(server)
//		useJWT(server)
//
//		return server
//	}
//
//	func useJWT(server *gin.Engine) {
//		loginJWT := middleware.LoginJWTMiddlewareBuilder{}
//		server.Use(loginJWT.CheckLogin())
//	}
//func useSession(server *gin.Engine) {
//	// 登陆校验
//	login := &middleware.LoginMiddlewareBuilder{}
//	// 存储数据的，也就是 userId 存哪里
//	// 1.直接存 Cookie
//	store := cookie.NewStore([]byte("secret"))
//
//	// 2.使用 memstore 内存的实现， 需要传入两个key ，第一个用于 authentication 用于身份认证，第二个用于 encryption 用于数据加密，可通过工具生成
//	// store := memstore.NewStore([]byte("uF7hZ5sW5fZ7jC1mY1wS9qQ4nQ2gN7lJ"), []byte("yY3aL4pO2iQ5vA5jE9yQ0vN1sC2vW4rN"))
//
//	// 3.使用 redis 实现，需要启动 redis 服务，传入参数 size 表示连接数 + network “tcp” + address “主机+端口” + password 未设置 + 两个key
//	//store, err := redis.NewStore(16, "tcp", "localhost:6379", "",
//	//	[]byte("uF7hZ5sW5fZ7jC1mY1wS9qQ4nQ2gN7lJ"), []byte("yY3aL4pO2iQ5vA5jE9yQ0vN1sC2vW4rN"))
//	//if err != nil {
//	//	panic(err)
//	//}
//	server.Use(sessions.Sessions("ssid", store), login.CheckLogin()) // 先初始化session，再校验
//
//}

//后端处理  1.接受请求并校验	 2.调用业务逻辑处理请求  3.根据业务逻辑处理结果返回响应
