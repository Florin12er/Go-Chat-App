package main

import (
	"log"
	"net/http"

	"GoChatBot/handlers"
	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"
)

func main() {
	gin.SetMode(gin.ReleaseMode)
	// Load .env file
	err := godotenv.Load()
	if err != nil {
		log.Fatal("Error loading .env file")
	}

	r := gin.Default()
	r.LoadHTMLGlob("templates/*")
	r.Static("/static", "./static")
	if err := r.SetTrustedProxies([]string{"127.0.0.1"}); err != nil {
		panic(err)
	}

	r.GET("/", func(c *gin.Context) {
		c.HTML(http.StatusOK, "index.html", gin.H{})
	})

	r.POST("/chat", handlers.HandleChat)

	if err := r.Run(); err != nil {
		log.Fatalf("failed to run server: %v", err)
	}

}
