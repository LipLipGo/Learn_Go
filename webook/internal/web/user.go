package web

import (
	"Learn_Go/webook/internal/domain"
	"Learn_Go/webook/internal/service"
	"fmt"
	regexp "github.com/dlclark/regexp2"
	"github.com/gin-contrib/sessions"
	"github.com/gin-gonic/gin"
	jwt "github.com/golang-jwt/jwt/v5"
	"net/http"
	"time"
)

// 校验数据的正则表达式

const (
	emailRegexPattern = "^\\w+([-+.]\\w+)*@\\w+([-.]\\w+)*\\.\\w+([-.]\\w+)*$"
	// 和上面比起来，用 ` 看起来就比较清爽
	passwordRegexPattern = `^(?=.*[A-Za-z])(?=.*\d)(?=.*[$@$!%*#?&])[A-Za-z\d$@$!%*#?&]{8,}$` // 官方的正则表达式不支持 ?= 这种写法，因此会报错，可以使用开源的正则表达式匹配库
	//bizLogin             = "login"
)

type UserHandler struct {
	emailRexExp    *regexp.Regexp
	passwordRexExp *regexp.Regexp
	svc            *service.UserService
}

func NewUserHandler(svc *service.UserService) *UserHandler { // 预编译正则表达式，保证正则表达式正确，性能优化
	return &UserHandler{
		emailRexExp:    regexp.MustCompile(emailRegexPattern, regexp.None),
		passwordRexExp: regexp.MustCompile(passwordRegexPattern, regexp.None),
		svc:            svc,
	}
}

// 分散注册路由 由各自的 Handler 注册自己的路由；优点：有条理  缺点：不好找

func (h *UserHandler) RegisterRouters(server *gin.Engine) {
	//server.POST("/users/signup", h.SignUp)
	//server.POST("/users/login", h.LogIn)
	//server.POST("/users/edit", h.Edit)
	//server.GET("users/profile", h.Profile)

	// 为了处理 /users 前缀写错，可以使用分组路由
	ug := server.Group("/users")
	// POST /users/signup
	ug.POST("/signup", h.SignUp)
	// POST /users/login
	//ug.POST("/login", h.LogIn)
	ug.POST("/login", h.LogInJWT)
	// POST /users/edit
	ug.POST("/edit", h.Edit)
	// POST /users/profile
	ug.GET("/profile", h.Profile)

}

func (h *UserHandler) SignUp(ctx *gin.Context) {

	//内部类，除了SignUp，其它方法用不了； 用于接收前端数据
	type SignUpReq struct {
		Email           string `json:"email"` //标签
		Password        string `json:"password"`
		ConfirmPassword string `json:"confirmPassword"`
	}

	// Bind方法是gin中最常用于接收请求的方法

	var req SignUpReq
	if err := ctx.Bind(&req); err != nil { // 将前端传来的数据绑定到 req 上，填充 req 中的字段值；bind 方法会根据Content-Type规定的数据格式判断格式是否一致，若不一致，返回400码
		return
	}
	// 拿到前端传来的数据，开始校验请求，
	// 1.邮箱需要复合一定能的格式（合法的邮箱）	2.密码和确认密码需要相等		3.密码需要符合一定的规律			现在多用二次验证
	// 使用正则表达式来校验

	// 邮箱校验
	isEmail, err := h.emailRexExp.MatchString(req.Email) // 如果设置了正则表达式预编译，官方的regexp包就不会返回 err 信息；这里替换为开源的包，这里返回的err是超时处理

	//if err != nil {
	//	ctx.String(http.StatusOK, "系统错误！") // 当正则表达式不正确，会返回错误
	//	return
	//}
	if err != nil {
		ctx.String(http.StatusOK, "超时错误！") // 使用开源regexp包，当超时时，会返回err
		return
	}
	if !isEmail {
		ctx.String(http.StatusOK, "非法邮箱格式！") // 当正则表达式不正确，会返回错误
		return
	}
	// 密码校验

	if req.Password != req.ConfirmPassword {
		ctx.String(http.StatusOK, "两次密码输入不一致！")
		return
	}

	isPassword, err := h.passwordRexExp.MatchString(req.Password)
	if err != nil {
		ctx.String(http.StatusOK, "超时错误")
		return
	}
	if !isPassword {
		ctx.String(http.StatusOK, "密码必须包含字母、数字、特殊字符，并且不少于八位")
		return
	}

	err = h.svc.Signup(ctx, domain.User{Email: req.Email, Password: req.Password})

	// 要判定邮箱冲突
	switch err {
	case nil:
		ctx.String(http.StatusOK, "注册成功！")
	case service.ErrDuplicateEmail: // 最好不要跨层调用，通过逐层传导实现调用
		ctx.String(http.StatusOK, "该邮箱已被注册，请更换一个邮箱！")
	default:
		ctx.String(http.StatusOK, "系统错误")

	}

}

func (h *UserHandler) LogIn(ctx *gin.Context) {
	type LoginReq struct {
		Email    string `json:"email"` //标签
		Password string `json:"password"`
	}

	var req LoginReq

	if err := ctx.Bind(&req); err != nil {
		return
	}

	u, err := h.svc.Login(ctx, req.Email, req.Password) // 这里的 u 用于登录校验

	// 判定登陆时，账户和密码是否输入正确
	switch err {
	case service.ErrInvalidUserOrPassword:
		ctx.String(http.StatusOK, "账号或密码错误，请重新输入！")
	case nil:
		// 记录登陆状态
		sess := sessions.Default(ctx) // 通过 gin 的 middleware 来设置 Session
		sess.Set("userId", u.Id)
		sess.Options(sessions.Options{ // 这里Option控制cookie，其中MaxAge同时控制cookie和session中数据的过期时间
			// 设置最长有效期
			MaxAge: 30,
			//HttpOnly: true,
		})
		err = sess.Save() // 需要主动Save才能生效
		if err != nil {
			ctx.String(http.StatusOK, "系统错误！")
			return
		}
		ctx.String(http.StatusOK, "登陆成功！")
	default:
		ctx.String(http.StatusOK, "系统错误！")

	}

}

func (h *UserHandler) LogInJWT(ctx *gin.Context) {
	type LoginReq struct {
		Email    string `json:"email"` //标签
		Password string `json:"password"`
	}

	var req LoginReq

	if err := ctx.Bind(&req); err != nil {
		return
	}

	u, err := h.svc.Login(ctx, req.Email, req.Password) // 这里的 u 用于登录校验

	// 判定登陆时，账户和密码是否输入正确
	switch err {
	case service.ErrInvalidUserOrPassword:
		ctx.String(http.StatusOK, "账号或密码错误，请重新输入！")
	case nil:
		uc := UserClaims{ // Claims就表示数据
			Uid:       u.Id,
			UserAgent: ctx.GetHeader("User-Agent"),
			// 设置过期时间
			RegisteredClaims: jwt.RegisteredClaims{ExpiresAt: jwt.NewNumericDate(time.Now().Add(time.Minute * 30))},
		}
		token := jwt.NewWithClaims(jwt.SigningMethodHS512, uc) // SigningMethod的安全性和性能有一些差异，没有要求可随意选
		tokenStr, err := token.SignedString(JWTKey)            // 这里token是一个结构体，但是传到前端需要一串字符，通过这个方法转换，其中不同的SigningMethod有不同的参数类型

		if err != nil {
			ctx.String(http.StatusOK, "系统错误！")
		}
		ctx.Header("x-jwt-token", tokenStr) // 希望后端将token放在x-jwt-token里面，前端在请求的Authorization头部带上Bearer token

		ctx.String(http.StatusOK, "登陆成功！")
	default:
		ctx.String(http.StatusOK, "系统错误！")

	}
}

type UserClaims struct {
	jwt.RegisteredClaims // 正常就这么使用
	Uid                  int64
	UserAgent            string
}

var JWTKey = []byte("uF7hZ5sW5fZ7jC1mY1wS9qQ4nQ2gN7lJ")

type Result struct {
	code int
	msg  string
}

func (h *UserHandler) Edit(ctx *gin.Context) {

	type EditReq struct {
		NikeName string `json:"nickname"`
		BirthDay string `json:"birthday"`
		AboutMe  string `json:"aboutMe"`
	}

	var req EditReq
	if err := ctx.Bind(&req); err != nil {
		//ctx.String(http.StatusOK, "系统错误!")
		return
	}

	uc, ok := ctx.MustGet("user").(UserClaims)

	if !ok {
		ctx.AbortWithStatus(http.StatusUnauthorized)
		return
	}
	fmt.Println(req.BirthDay)

	//// 获取用户的UserId
	//sess := sessions.Default(ctx)
	//UserId := sess.Get("UserId")
	//
	//if UserId == nil {
	//	ctx.String(http.StatusOK, "未登录，请先登陆！")
	//	ctx.AbortWithStatus(http.StatusUnauthorized) // http.StatusUnauthorized 通常用于代表没登陆
	//	return
	//}
	//userId := UserId.(int64)

	// 校验昵称
	if req.NikeName == "" {
		ctx.JSON(http.StatusOK, Result{
			code: 4,
			msg:  "昵称不可为空！",
		})
		return
	}

	// 校验生日字段格式，通过time.parse校验，不需要使用正则表达式

	birthday, err := time.Parse(time.DateOnly, req.BirthDay)
	if err != nil {
		ctx.String(http.StatusOK, "生日格式输入错误！")
		return
	}

	//校验 aboutMe 长度
	if len(req.AboutMe) > 1024 {
		ctx.JSON(http.StatusOK, Result{
			code: 4,
			msg:  "描述过长！",
		})
		return
	}

	err = h.svc.UpdateNonSensitiveInfo(ctx, domain.User{
		Id:       uc.Uid,
		NickName: req.NikeName,
		BirthDay: birthday,
		AboutMe:  req.AboutMe,
	})

	if err != nil {
		ctx.JSON(http.StatusOK, Result{
			code: 5,
			msg:  "系统异常",
		})
		return
	}
	ctx.String(http.StatusOK, "更新成功")

}

// http协议时无状态的，那每次登陆都需要登录校验，那怎么来保存这个状态？
// cookie和session
// cookie：浏览器存储数据到本地，这些数据就是cookie，键值对，缺点：不安全
/*
## Cookie关键配置（要注意“安全使用”）
- Domain：也就是Cookie可以使用在什么域名下，按照**最小化原则**来设定（比如二级域名）
- Path：Cookie可以用在什么路径下，同样按照**最小化原则**来设定
- Max-Age和Expires：过期时间，只保留必要时间
- Http-Only：设置为true的话，那么浏览器上的JS代码将无法使用这个Cookie。**永远设置为true**
- Secure：只能用于HTTPS协议，**生产环境永远设置为true**
- SameSite：是否允许跨站发送Cookie，尽量避免
*/

// Session 用于登陆

func (h *UserHandler) Profile(ctx *gin.Context) {

	uc, ok := ctx.MustGet("user").(UserClaims)
	if !ok {
		ctx.AbortWithStatus(http.StatusUnauthorized)
		return
	}

	u, err := h.svc.FindById(ctx, uc.Uid)
	if err != nil {
		ctx.String(http.StatusOK, "系统异常！")
		return
	}

	type User struct {
		Nickname string `json:"nickname"`
		Email    string `json:"email"`
		AboutMe  string `json:"aboutMe"`
		Birthday string `json:"birthday"`
	}
	ctx.JSON(http.StatusOK, User{
		Nickname: u.NickName,
		Email:    u.Email,
		AboutMe:  u.AboutMe,
		Birthday: u.BirthDay.Format(time.DateOnly),
	})

}
