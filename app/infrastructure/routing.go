package infrastructure

import (
	middleware "go_clean_arch_test/app/infrastructure/auth"
	"go_clean_arch_test/app/interfaces/database"
	"go_clean_arch_test/app/interfaces/database/repository/sql"
	authSql "go_clean_arch_test/app/interfaces/database/repository/sql/auth"
	"go_clean_arch_test/app/interfaces/delivery"
	authDelivery "go_clean_arch_test/app/interfaces/delivery/auth"
	"go_clean_arch_test/app/usecase"
	authUsecase "go_clean_arch_test/app/usecase/auth"
	"time"

	"github.com/gin-contrib/cors"
	"github.com/gin-contrib/sessions"
	"github.com/gin-contrib/sessions/cookie"
	"github.com/gin-gonic/gin"
)

type Routing struct {
	DB  *DB
	Gin *gin.Engine
}

func NewRouting(db *DB) *Routing {
	r := &Routing{
		DB:  db,
		Gin: gin.Default(),
	}
	// Corsの設定
	r.Gin.Use(cors.New(cors.Config{
		// 許可アクセス元
		AllowOrigins: []string{
			"http://localhost:62723",
		},
		// アクセス許可HTTPメソッド(以下PUT,DELETEアクセス不可)
		AllowMethods: []string{
			"POST",
			"GET",
			"OPTIONS",
		},
		// 許可HTTPリクエストヘッダ
		AllowHeaders: []string{
			"Access-Control-Allow-Credentials",
			"Access-Control-Allow-Headers",
			"Content-Type",
			"Content-Length",
			"Accept-Encoding",
			"Authorization",
			"accessToken",
			"Set-Cookie",
			"Cookie",
		},
		// cookie必要許可
		AllowCredentials: true,
		// preflightリクエストの結果をキャッシュする時間
		MaxAge: 24 * time.Hour,
	}))

	// セッションCookieの設定
	store := cookie.NewStore([]byte("secret"))
	store.Options(sessions.Options{
		Secure:   false,
		HttpOnly: false})
	r.Gin.Use(sessions.Sessions("bookMarkAppSessKey", store))

	r.setRouting()
	return r
}

func (r *Routing) setRouting() {

	// repository
	aritcleRepository := sql.NewArticleRepository(SqlConnect())
	authorRepository := sql.NewAuthorRepository(SqlConnect())
	expPoolRepository := sql.NewExpPoolRepository(SqlConnect())
	lvRepository := sql.NewLvRepository(SqlConnect())
	signUpRepository := authSql.NewSignUpRepository(SqlConnect())
	loginRepository := authSql.NewLoginRepository(SqlConnect())

	// usecase
	expPoolUsecase := usecase.NewExpPoolUsecase(expPoolRepository, database.NewTransaction(r.DB.Connection))
	lvUsecase := usecase.NewLvUsecase(lvRepository, database.NewTransaction(r.DB.Connection))
	authoreUsecase := usecase.NewAuthorUsecase(expPoolUsecase, authorRepository)
	articleUsecase := usecase.NewArticleUsecase(authoreUsecase, expPoolUsecase, lvUsecase, aritcleRepository, database.NewTransaction(r.DB.Connection))

	signUpUsecase := authUsecase.NewSignUpUsecase(signUpRepository, loginRepository)
	loginUsecase := authUsecase.NewLoginUsecase(loginRepository)

	// handler
	articleHandler := delivery.NewArticleHandler(articleUsecase, authoreUsecase, loginUsecase)
	authorHandler := delivery.NewAuthorHandler(articleUsecase, authoreUsecase, loginUsecase)
	signUpHandler := authDelivery.NewSignUpHandler(signUpUsecase)
	loginHandler := authDelivery.NewLoginHandler(loginUsecase)

	r.Gin.POST("/signup", func(ctx *gin.Context) { signUpHandler.SignUp(ctx) })
	r.Gin.POST("/login", func(ctx *gin.Context) { loginHandler.Login(ctx) })
	r.Gin.POST("/logout", func(ctx *gin.Context) { authDelivery.Logout(ctx) })
	// 認証済のみアクセス可能なグループ
	authUserGroup := r.Gin.Group("/auth")
	authUserGroup.Use(middleware.LoginCheckMiddleware())
	{
		r.Gin.GET("/", func(ctx *gin.Context) { articleHandler.GetAll(ctx) })
		r.Gin.GET("/article", func(ctx *gin.Context) { articleHandler.GetById(ctx) })
		r.Gin.GET("/article/author", func(ctx *gin.Context) { articleHandler.GetByAuthorId(ctx) })
		r.Gin.GET("/article/search", func(ctx *gin.Context) { articleHandler.GetLikeByTitleAndContent(ctx) })
		r.Gin.POST("/article/input", func(ctx *gin.Context) { articleHandler.Input(ctx) })
		r.Gin.POST("/article/update", func(ctx *gin.Context) { articleHandler.Update(ctx) })
		r.Gin.POST("/article/delete", func(ctx *gin.Context) { articleHandler.Delete(ctx) })

		r.Gin.GET("/author", func(ctx *gin.Context) { authorHandler.GetAllAuthor(ctx) })
		r.Gin.POST("/author/input", func(ctx *gin.Context) { authorHandler.InputAuthor(ctx) })
		r.Gin.POST("/author/update", func(ctx *gin.Context) { authorHandler.UpdateAuthor(ctx) })
		r.Gin.POST("/author/delete", func(ctx *gin.Context) { authorHandler.DeleteAuthor(ctx) })
	}
}

func (r *Routing) Run() {
	r.Gin.Run()
}
