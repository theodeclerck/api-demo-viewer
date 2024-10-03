package api

import (
	"api-demo-viewer/internal"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"net/http"
)

func Ping(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{
		"message": "pong",
	})
}

func UploadDemo(c *gin.Context) {
	file, _ := c.FormFile("file")
	if internal.CheckFile(file) {
		c.JSON(http.StatusBadRequest, gin.H{
			"message": "something went wrong!",
		})
		return
	}
	newFileName := uuid.New().String() + ".dem"
	err := c.SaveUploadedFile(file, "./files/"+newFileName)
	if err != nil {
		print(err)
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

	go internal.OpenDemo(newFileName, matchID, taskID)

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
		"created_at": "2021-09-23",
	})
}
