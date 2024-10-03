package main

import (
	"api-demo-viewer/api"
	"api-demo-viewer/db"
	"context"
	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	"log"
)

func main() {
	//internal.OpenDemo("./resources/test.dem")

	client := db.ConnectMongo()
	defer func() {
		if err := client.Disconnect(context.Background()); err != nil {
			log.Fatal("Error disconnecting from MongoDB:", err)
		}
	}()

	r := gin.Default()

	r.Use(cors.New(cors.Config{
		AllowOrigins:     []string{"http://localhost:3000"},
		AllowMethods:     []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
		AllowHeaders:     []string{"Origin", "Content-Length", "Content-Type"},
		ExposeHeaders:    []string{"Content-Length"},
		AllowCredentials: true,
	}))

	r.GET("/ping", api.Ping)
	//matches
	r.POST("/upload", api.UploadDemo)
	r.GET("/match/:filename", api.GetMatchName)
	r.GET("/matches", api.GetMatches)
	//tasks
	r.GET("/tasks/:taskId", api.GetTask)
	//users

	r.Run(":8088")
}
