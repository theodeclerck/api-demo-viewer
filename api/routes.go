package api

import (
	"api-demo-viewer/internal"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"log"
	"net/http"
	"time"
)

func Ping(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{
		"message": "pong",
	})
}

func UploadDemo(c *gin.Context) {
	file, err := c.FormFile("file")
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"message": "failed to retrieve file",
		})
	}
	if internal.CheckFile(file) {
		c.JSON(http.StatusBadRequest, gin.H{
			"message": "something went wrong!",
		})
		return
	}
	newFileName := uuid.New().String() + ".dem"
	errFile := c.SaveUploadedFile(file, "./files/"+newFileName)
	if errFile != nil {
		print(errFile)
		c.JSON(http.StatusInternalServerError, gin.H{
			"message": "failed to save file",
		})
		return
	}

	matchID, taskID, err := internal.CreateMatch(newFileName)
	if err != nil {
		print(err)
		c.JSON(http.StatusInternalServerError, gin.H{
			"message": "failed to create match",
		})
		return
	}

	errCh := make(chan error)
	go func() {
		errCh <- internal.NewGame(newFileName, matchID, taskID)
	}()

	go func() {
		err := <-errCh
		if err != nil {
			log.Printf("error while parsing demo: %v", err)
		}
	}()

	c.Redirect(http.StatusCreated, "/match/"+matchID.String())
}

func GetMatchName(c *gin.Context) {
	filename := c.Param("filename")

	c.JSON(http.StatusOK, gin.H{
		"message": filename,
	})
}

func GetMatches(c *gin.Context) {
	matches := internal.ListDemosName()

	c.JSON(http.StatusOK, gin.H{
		"message": matches,
	})
}

func GetTask(c *gin.Context) {
	taskId := c.Param("taskId")

	c.JSON(http.StatusOK, gin.H{
		"taskId":     taskId,
		"status":     "done",
		"created_at": time.Now(),
	})
}
