package db

import (
	"context"
	"time"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"go.uber.org/zap"
)

var mongoClient *mongo.Client
var database *mongo.Database

// User represents a user document in the MongoDB collection
type User struct {
	ID        int64     `bson:"_id"`
	JoinedAt  time.Time `bson:"joined_at"`
}

// InitMongoDB connects to the MongoDB database
func InitMongoDB(log *zap.Logger) {
	log = log.Named("MongoDB")
	
	clientOptions := options.Client().ApplyURI("mongodb+srv://mrshokrullah:L7yjtsOjHzGBhaSR@cluster0.aqxyz.mongodb.net/shahlink?retryWrites=true&w=majority&appName=Cluster")
	
	var err error
	mongoClient, err = mongo.Connect(context.TODO(), clientOptions)
	if err != nil {
		log.Fatal("Failed to connect to MongoDB", zap.Error(err))
	}
	
	// Ping the primary to verify connection
	err = mongoClient.Ping(context.TODO(), nil)
	if err != nil {
		log.Fatal("Failed to ping MongoDB", zap.Error(err))
	}
	
	database = mongoClient.Database("shahlink")
	log.Info("Successfully connected to MongoDB")
}

// SaveUser saves a user's ID to the database if it doesn't already exist
func SaveUser(log *zap.Logger, userID int64) {
	log = log.Named("MongoDB")
	usersCollection := database.Collection("users")
	
	// Check if the user already exists to avoid duplicates
	var result User
	err := usersCollection.FindOne(context.TODO(), bson.D{{"_id", userID}}).Decode(&result)
	
	if err == nil {
		log.Sugar().Infof("User %d already exists in the database. Skipping save.", userID)
		return
	}
	
	if err != mongo.ErrNoDocuments {
		log.Error("Error checking for existing user", zap.Error(err))
		return
	}
	
	// If the user doesn't exist, insert the new document
	user := User{
		ID:       userID,
		JoinedAt: time.Now(),
	}
	
	_, err = usersCollection.InsertOne(context.TODO(), user)
	if err != nil {
		log.Error("Failed to save user to MongoDB", zap.Error(err))
	} else {
		log.Sugar().Infof("Successfully saved user %d to MongoDB", userID)
	}
}
