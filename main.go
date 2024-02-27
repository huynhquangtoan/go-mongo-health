package main

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"time"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

const databaseName = "_healthcheck"
const collectionName = "tests"

type User struct {
	Name  string `bson:"name"`
	Email string `bson:"email"`
}

func healthCheckHandler(w http.ResponseWriter, r *http.Request) {
	connectionString := os.Getenv("MONGODB_URI")
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	client, err := mongo.Connect(ctx, options.Client().ApplyURI(connectionString))
	if err != nil {
		w.WriteHeader(http.StatusServiceUnavailable)
		fmt.Fprint(w, "MongDB connection error: "+err.Error())
		return
	}
	defer client.Disconnect(ctx)

	err = client.Ping(ctx, nil)
	if err != nil {
		w.WriteHeader(http.StatusServiceUnavailable)
		fmt.Fprint(w, "MongoDB ping failed!")
		return
	}

	// Ghi dữ liệu
	collection := client.Database(databaseName).Collection(collectionName)
	user := User{Name: "John Doe", Email: "john.doe@example.com"}
	_, err = collection.InsertOne(ctx, user)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		fmt.Fprint(w, "Failed to insert data into MongoDB")
		return
	}
	fmt.Fprintln(w, "Write ok")

	// Đọc dữ liệu
	var result User
	filter := bson.D{{"name", "John Doe"}}
	err = collection.FindOne(ctx, filter).Decode(&result)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		fmt.Fprint(w, "Failed to read data from MongoDB")
		return
	}
	fmt.Fprintln(w, "Read ok")

	err = collection.Database().Drop(ctx)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		fmt.Fprint(w, "Failed to delete test db")
		return
	}

	fmt.Fprint(w, "MongoDB is healthy!")
}

func main() {
	http.HandleFunc("/health", healthCheckHandler)

	port := os.Getenv("PORT")
	if port == "" {
		fmt.Println("Port is not set, defaulting to 3000")
		port = "3000"
	}

	fmt.Printf("Server is running on :%s\n", port)
	err := http.ListenAndServe(fmt.Sprintf(":%s", port), nil)
	if err != nil {
		fmt.Printf("Error starting server: %v\n", err)
	}
}
