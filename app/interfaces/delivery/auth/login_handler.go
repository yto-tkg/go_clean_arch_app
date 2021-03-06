package delivery

import (
	"encoding/json"
	domain "go_clean_arch_test/app/domain/auth"
	"go_clean_arch_test/app/interfaces/delivery"
	loginUsecase "go_clean_arch_test/app/usecase/auth"
	"net/http"

	"github.com/google/uuid"

	"github.com/gin-contrib/sessions"
	"github.com/gin-gonic/gin"
	"golang.org/x/crypto/bcrypt"
)

// LoginHandler interface
type LoginHandler interface {
	Login(ctx *gin.Context)
}

type loginHandler struct {
	loginUsecase loginUsecase.LoginUsecase
}

// NewLoginHandler constructor
func NewLoginHandler(loginUsecase loginUsecase.LoginUsecase) LoginHandler {
	return &loginHandler{loginUsecase: loginUsecase}
}

func (loginHandler *loginHandler) Login(ctx *gin.Context) {
	var request domain.Login
	err := ctx.BindJSON(&request)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, delivery.NewH(http.StatusText(http.StatusBadRequest), http.StatusBadRequest))
	} else {
		// メールアドレスでユーザ情報取得
		user, err := loginHandler.loginUsecase.GetByEmail(request.Email)
		if err != nil {
			ctx.JSON(http.StatusInternalServerError, delivery.NewH(err.Error(), nil))
			return
		}

		// ハッシュ値でのパスワード比較
		err = bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(request.Password))
		if err != nil {
			ctx.JSON(http.StatusBadRequest, delivery.NewH(err.Error(), http.StatusBadRequest))
		} else {
			session := sessions.Default(ctx)
			// セッションに格納する為にユーザ情報をJson化
			loginUser, err := json.Marshal(user)
			if err == nil {
				u, _ := uuid.NewRandom()
				accessToken := u.String()
				session.Set(accessToken, string(loginUser))
				session.Save()
				ctx.SetSameSite(http.SameSiteDefaultMode)

				ctx.JSON(http.StatusOK, delivery.NewH(http.StatusText(http.StatusOK), accessToken))
			} else {
				ctx.JSON(http.StatusInternalServerError, delivery.NewH(err.Error(), http.StatusInternalServerError))
			}
		}
	}
}
