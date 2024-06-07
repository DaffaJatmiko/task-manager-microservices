package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"time"

	"github.com/DaffaJatmiko/logger-service/data"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

const (
	webPort  = "80"
	rpcPort  = "5001"
	mongoURL = "mongodb://mongo:27017"
	gRpcPort = "50001"
)

var client *mongo.Client

type Config struct {
	Models data.Models
}

func main() {
	// connect to MongoDB
	mongoClient, err := connectToMongo()
	if err != nil {
		log.Panic(err)
	}
	client = mongoClient

	// create a context in order to connect to MongoDB
	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()

	// close the connection
	defer func(){
		if err = client.Disconnect(ctx); err != nil {
			panic(err)
		}
	}()

	app := Config{
		Models: data.New(client),
	}

	go app.gRPCListen()

	// start web server
	// go app.serve()
	srv := &http.Server{
		Addr:    fmt.Sprintf(":%s", webPort),
		Handler: app.routes(),
	}

	log.Println("Starting web server on port", webPort)
	err = srv.ListenAndServe()
	if err != nil {
			log.Panic("Web server failed to start: ", err)
	}
}

// func (c *Config) serve() {
// 	srv := &http.Server{
// 		Addr:    fmt.Sprintf(":%s", webPort),
// 		Handler: c.routes(),
// 	}

// 	err := srv.ListenAndServe()
// 	if err != nil {
// 		log.Panic()
// 	}
// }

func connectToMongo() (*mongo.Client, error) {
	// TODO: Set client options
	clientOptions := options.Client().ApplyURI(mongoURL)
	clientOptions.SetAuth(options.Credential{
		Username: "admin",
		Password: "password",
	})
	
	// TODO: Connect to MongoDB
	client, err := mongo.Connect(context.TODO(), clientOptions)
	if err != nil {
		log.Println(err)
		return nil, err
	}
	log.Println("Connected to MongoDB!")
	return client, nil
}