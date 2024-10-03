package db

import (
	"context"
	"github.com/joho/godotenv"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"log"
	"os"
	"time"
)

const Dbname = "demo-viewer"

type MongoCollections struct {
	Matches *mongo.Collection
	Users   *mongo.Collection
	Tasks   *mongo.Collection
}

var Collections *MongoCollections

func ConnectMongo() *mongo.Client {
	if err := godotenv.Load(); err != nil {
		log.Fatal("Error loading .env file")
	}

	uri := os.Getenv("MONGO_URI")
	if uri == "" {
		log.Fatal("MONGO_URI is empty")
	}

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	client, err := mongo.Connect(ctx, options.Client().ApplyURI(uri))
	if err != nil {
		return nil
	}

	if err := client.Ping(ctx, nil); err != nil {
		return nil
	}

	Collections = &MongoCollections{
		Matches: client.Database(Dbname).Collection("matches"),
		Users:   client.Database(Dbname).Collection("users"),
		Tasks:   client.Database(Dbname).Collection("tasks"),
	}

	return client
}

func CreateTimeSeriesCollection(db *mongo.Database) error {
	collectionOptions := options.CreateCollection().SetTimeSeriesOptions(
		options.TimeSeries().
			SetTimeField("timestamp").
			SetMetaField("meta").
			SetGranularity("seconds"))

	err := db.CreateCollection(context.TODO(), "match_ticks", collectionOptions)
	if err != nil {
		return err
	}
	return nil
}
