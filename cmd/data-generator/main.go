package main

import (
	"context"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/joho/godotenv"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"go.mongodb.org/mongo-driver/mongo/readpref"
)

func goDotEnv(key string) string {
	//load .ev file
	err := godotenv.Load(".env")

	if err != nil {
		log.Fatalf("Error loading .env file")
	}
	return os.Getenv(key)
}

func main() {

	db_uri := testEnvironmentFile()
	testMongoDB(db_uri)
	testWebServer()

}

func testMongoDB(db_uri string) {

	fmt.Printf("db_uri: %s\n", db_uri)
	client, err := mongo.NewClient(options.Client().ApplyURI(db_uri))
	if err != nil {
		log.Fatal(err)
		panic(err)
	}

	ctx, _ := context.WithTimeout(context.Background(), 10*time.Second)
	err = client.Connect(ctx)
	if err != nil {
		log.Fatal(err)
	}
	defer client.Disconnect(ctx)

	excavatorDatabase := client.Database("excavator")
	goldCollection := excavatorDatabase.Collection("gold")
	dataResult, err := goldCollection.InsertOne(ctx, bson.D{
		{Key: "name", Value: "Gold"},
		{Key: "value", Value: "1"},
		{Key: "tags", Value: bson.A{"development", "gold", "coding"}},
	})
	if err != nil {
		panic(err)
	}

	fmt.Printf("dataResult: %v\n", dataResult)
	testDbClient(client, ctx)

	watchChangesFromCollection(goldCollection)

}

//Allow watch stream of the changes from the database
func watchChangesFromCollection(watchingCollection *mongo.Collection) {
	collStream, err := watchingCollection.Watch(context.TODO(), mongo.Pipeline{})
	if err != nil {
		panic(err)
	}
	defer collStream.Close(context.TODO())
	for collStream.Next(context.TODO()) {
		var data bson.M
		if err := collStream.Decode(&data); err != nil {
			panic(err)
		}
		fmt.Printf("%v\n", data)
	}
}

//Testing the MongoDB Client verify connection to database it's good when no error
func testDbClient(client *mongo.Client, ctx context.Context) {
	//Suspect this line only works with a replica set with ReadPreference: Primary
	if err := client.Ping(context.TODO(), readpref.Primary()); err != nil {
		panic(err)
	}

	databases, err := client.ListDatabaseNames(ctx, bson.M{})
	if err != nil {
		log.Fatal(err)
	}

	fmt.Println(databases)

	client.Database("test").Collection("users")
}

func testEnvironmentFile() string {
	// Hello world, the web server
	dotenv := goDotEnv("MONGO_URL")
	fmt.Printf("Hello World! %s\n\n", dotenv)
	return dotenv
}

func testWebServer() {
	helloHandler := func(w http.ResponseWriter, req *http.Request) {

		io.WriteString(w, "Hello, world! \n")
	}

	http.HandleFunc("/hello", helloHandler)
	log.Println("Listing for requests at http://localhost:8000/hello")
	log.Fatal(http.ListenAndServe(":8000", nil))
}
