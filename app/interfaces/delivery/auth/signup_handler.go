package delivery

import (
	"encoding/json"
	domain "go_clean_arch_test/app/domain/auth"
	"go_clean_arch_test/app/interfaces/delivery"
	signUpUsecase "go_clean_arch_test/app/usecase/auth"
	"net/http"

	"github.com/google/uuid"

	"github.com/gin-contrib/sessions"
	"github.com/gin-gonic/gin"
)

// SignUpHandler interface
type SignUpHandler interface {
	SignUp(ctx *gin.Context)
}

type signUpHandler struct {
	signUpUsecase signUpUsecase.SignUpUsecase
}

// NewSignUpHandler constructor
func NewSignUpHandler(signUpUsecase signUpUsecase.SignUpUsecase) SignUpHandler {
	return &signUpHandler{signUpUsecase: signUpUsecase}
}

func (signUpHandler *signUpHandler) SignUp(ctx *gin.Context) {
	var request domain.SignUp
	err := ctx.BindJSON(&request)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, delivery.NewH("error", http.StatusBadRequest))
	} else {
		// 会員登録処理
		user, err := signUpHandler.signUpUsecase.SignUp(request.Email, request.Password)

		if err != nil {
			ctx.JSON(http.StatusInternalServerError, delivery.NewH(err.Error(), http.StatusInternalServerError))
		} else {
			session := sessions.Default(ctx)
			// セッションに格納する為にユーザー情報をJson化
			loginUser, err := json.Marshal(user)
			if err == nil {
				u, _ := uuid.NewRandom()
				accessToken := u.String()
				session.Set(accessToken, string(loginUser))
				session.Save()

				ctx.JSON(http.StatusOK, delivery.NewH(http.StatusText(http.StatusOK), accessToken))
			} else {
				ctx.JSON(http.StatusInternalServerError, delivery.NewH(err.Error(), http.StatusInternalServerError))
			}
		}
	}
}
