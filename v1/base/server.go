package main

import (
	"github.com/gin-gonic/gin"
	"github.com/ooqls/getset/app/app"
)

type RegisterRoutesFunc = func(ctx *app.AppContext, e *gin.Engine) error
