package gin_inbound_adapter

import (
	"net/http"

	"github.com/gin-gonic/gin"

	"prabogo/internal/domain"
	"prabogo/internal/model"
	inbound_port "prabogo/internal/port/inbound"
	"prabogo/utils/activity"
)

type authAdapter struct {
	domain domain.Domain
}

func NewAuthAdapter(domain domain.Domain) inbound_port.AuthHttpPort {
	return &authAdapter{
		domain: domain,
	}
}

func (a *authAdapter) Login(i any) {
	c := i.(*gin.Context)
	ctx := activity.NewContext("http_login")

	request := struct {
		Email    string `json:"email"`
		Password string `json:"password"`
	}{}

	if err := c.ShouldBindJSON(&request); err != nil {
		c.JSON(http.StatusBadRequest, model.Response{
			Success: false,
			Error:   "Invalid request body",
		})
		return
	}

	token, err := a.domain.Auth().Login(ctx, request.Email, request.Password)
	if err != nil {
		c.JSON(http.StatusUnauthorized, model.Response{
			Success: false,
			Error:   err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, model.Response{
		Success: true,
		Data: gin.H{
			"token": token,
		},
	})
}

func (a *authAdapter) Register(i any) {
	c := i.(*gin.Context)
	ctx := activity.NewContext("http_register")

	var input model.UserInput
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, model.Response{
			Success: false,
			Error:   "Invalid request body",
		})
		return
	}

	user, err := a.domain.Auth().Register(ctx, input)
	if err != nil {
		c.JSON(http.StatusBadRequest, model.Response{
			Success: false,
			Error:   err.Error(),
		})
		return
	}

	c.JSON(http.StatusCreated, model.Response{
		Success: true,
		Data:    user,
	})
}
