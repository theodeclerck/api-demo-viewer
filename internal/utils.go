package internal

import (
	"api-demo-viewer/db"
	"context"
	"fmt"
	_ "github.com/google/uuid"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"mime/multipart"
	"os"
	"time"
)

type Task struct {
	ID        primitive.ObjectID `bson:"_id"`
	MatchID   primitive.ObjectID `bson:"match_id"`
	Status    string             `bson:"status"`
	CreatedAt string             `bson:"created_at"`
}

type Match struct {
	ID        primitive.ObjectID `bson:"_id"`
	FileName  string             `bson:"file_name"`
	CreatedAt string             `bson:"created_at"`
}

type User struct {
	ID        primitive.ObjectID `bson:"_id"`
	Username  string             `bson:"username"`
	Password  string             `bson:"password"`
	CreatedAt string             `bson:"created_at"`
}

func CheckFile(file *multipart.FileHeader) bool {
	if !CheckFileSize(file) || !CheckFileName(file) || AlreadyExist(file) {
		return true
	}
	return false
}

func CheckFileSize(file *multipart.FileHeader) bool {
	if file.Size > 1024*1024*1000 { // 1GB
		print("file size is too big")
		return false
	}
	return true
}

func CheckFileName(file *multipart.FileHeader) bool {
	if file.Filename[len(file.Filename)-4:] != ".dem" {
		print("file extension is not .dem")
		return false
	}
	return true
}

func AlreadyExist(file *multipart.FileHeader) bool {
	if _, err := os.Stat("/files/" + file.Filename); !os.IsNotExist(err) {
		print(err)
		return true
	}
	return false
}

func ListDemosName() []string {
	files, err := os.ReadDir("./files")
	if err != nil {
		print(err)
	}
	var demos []string
	for _, file := range files {
		demos = append(demos, file.Name())
	}
	return demos
}

func CreateMatch(filename string) (primitive.ObjectID, primitive.ObjectID, error) {
	if db.Collections.Matches == nil {
		return primitive.NilObjectID, primitive.NilObjectID, fmt.Errorf("MongoDB collection 'Matches' is not initialized")
	}

	match := Match{
		ID:        primitive.NewObjectID(),
		FileName:  filename,
		CreatedAt: time.Now().String(),
	}

	result, err := db.Collections.Matches.InsertOne(context.Background(), match)
	if err != nil {
		return primitive.NilObjectID, primitive.NilObjectID, err
	}

	taskID, err := CreateTask(match.ID)
	if err != nil {
		return primitive.NilObjectID, primitive.NilObjectID, err
	}

	return result.InsertedID.(primitive.ObjectID), taskID, nil
}

func CreateTask(matchID primitive.ObjectID) (primitive.ObjectID, error) {
	task := Task{
		ID:        primitive.NewObjectID(),
		MatchID:   matchID,
		Status:    "in_progress",
		CreatedAt: time.Now().String(),
	}

	_, err := db.Collections.Tasks.InsertOne(context.Background(), task)
	if err != nil {
		print(err)
		return primitive.NilObjectID, err
	}

	return task.ID, nil
}

func UpdateTask(taskID primitive.ObjectID, status string) error {
	_, err := db.Collections.Tasks.UpdateOne(context.Background(), primitive.M{"_id": taskID}, primitive.M{"$set": primitive.M{"status": status}})
	if err != nil {
		print(err)
		return err
	}
	return nil
}

func CreateGame(game *Game) error {
	doc := prepareGameDocument(game)

	_, err := db.Collections.MatchTicks.InsertMany(
		context.Background(),
		doc,
	)
	if err != nil {
		print(err)
		return err
	}
	return nil
}

func prepareGameDocument(game *Game) []interface{} {
	var doc []interface{}
	for _, state := range game.States {
		doc = append(doc, state)
	}
	return doc
}
