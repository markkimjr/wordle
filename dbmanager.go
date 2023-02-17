package main

import (
	"log"
	"os"
	"time"
)

import (
	"context"
	"github.com/joho/godotenv"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

const DatabaseName = "wordle"

type Manager struct {
	Client     *mongo.Client
	Context    context.Context
	CancelFunc context.CancelFunc
}

var instance *Manager

func newManager() *Manager {
	dbManager := Manager{}
	dbManager.Context, dbManager.CancelFunc = context.WithTimeout(context.Background(), 30*time.Second)
	return &dbManager
}

func GetDBManager() (Manager *Manager) {
	if instance == nil {
		instance = mongoConn()
	}

	return instance
}

func GetLocalDBManager() (Manager *Manager) {
	err := godotenv.Load("/Users/markkimjr/dev/wordle/config/local.env")
	if err != nil {
		return nil
	}

	if instance == nil {
		instance = mongoConnByParam() // TODO Add DB connection parameters
	}

	return instance
}

func (Manager *Manager) Collection(dbName string, collectionName string) *mongo.Collection {
	return Manager.Client.Database(dbName).Collection(collectionName)
}

func mongoConnByParam(username string, password string, address string, port string, authMechanism string, authSource string) (Manager *Manager) {
	credential := options.Credential{
		Username: username,
		Password: password,
		//AuthMechanism: authMechanism,
		//AuthSource:    authSource,
	}

	clientOptions := options.Client().ApplyURI("mongodb://" + address + ":" + port).SetAuth(credential)
	client, err := mongo.Connect(context.TODO(), clientOptions)
	if err != nil {
		log.Fatal(err)
	}

	// Check connection
	err = client.Ping(context.TODO(), nil)

	if err != nil {
		log.Fatal(err)
	}
	log.Println("MongoDB Connection Made")

	db := newManager()
	db.Client = client
	return db
}

func mongoConn() (Manager *Manager) {
	return mongoConnByParam(
		os.Getenv("MONGO_USERNAME"), os.Getenv("MONGO_PASSWORD"), os.Getenv("MONGO_ADDRESS"),
		os.Getenv("MONGO_PORT"), os.Getenv("MONGO_AUTH_MECHANISM"), os.Getenv("MONGO_AUTH_SOURCE"),
	)
}
