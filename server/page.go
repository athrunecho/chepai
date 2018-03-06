package main

import (
	//"fmt"
	//"log"
	"net/http"

	"github.com/gin-gonic/gin"
)

func home(c *gin.Context) {
	c.HTML(http.StatusOK, "home.tmpl", gin.H{"Title": "车牌拍卖"})
}
