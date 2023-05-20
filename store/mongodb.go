package store

import (
	"context"
	"fmt"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"log"
)

var MgoCli *mongo.Client

func InitEngine() {
	var err error
	clientOptions := options.Client().ApplyURI("mongodb://172.22.114.78:27017")

	// 连接到MongoDB
	MgoCli, err = mongo.Connect(context.TODO(), clientOptions)
	if err != nil {
		log.Fatal(err)
	}
	// 检查连接
	err = MgoCli.Ping(context.TODO(), nil)
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println("------------------------------成功连接到MongoDB------------------")
}

func GetMgoCli() *mongo.Client {
	if MgoCli == nil {
		InitEngine()
	}
	return MgoCli
}
