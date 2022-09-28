package main

import (
	"io"
	"net/http"
	"os"

	"github.com/gin-gonic/gin"
)

func main() {
	fileLog, _ := os.Create("debug.log")
	gin.DefaultWriter = io.MultiWriter(fileLog)

	router := gin.Default()
	router.GET("/version", VersionHandler)
	router.Run()
}

func VersionHandler(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{"version": os.Getenv("API_VERSION")})
}