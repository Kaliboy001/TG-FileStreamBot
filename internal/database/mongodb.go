package database

import (
	"context"
	"time"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"go.uber.org/zap"
)

// User represents a user document in MongoDB
type User struct {
	UserID    int64     `bson:"user_id"`
	Username  string    `bson:"username"`
	FirstSeen time.Time `bson:"first_seen"`
	LastSeen  time.Time `bson:"last_seen"`
}

// Database holds the MongoDB client and collections
type Database struct {
	client     *mongo.Client
	database   *mongo.Database
	users      *mongo.Collection
	log        *zap.Logger
}

var DB *Database

// Connect initializes the MongoDB connection
func Connect(mongoURI string, log *zap.Logger) error {
	log = log.Named("MongoDB")
	
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	client, err := mongo.Connect(ctx, options.Client().ApplyURI(mongoURI))
	if err != nil {
		return err
	}

	// Ping the database to verify connection
	if err = client.Ping(ctx, nil); err != nil {
		return err
	}

	database := client.Database("shahlink")
	
	DB = &Database{
		client:   client,
		database: database,
		users:    database.Collection("users"),
		log:      log,
	}

	log.Info("Successfully connected to MongoDB")
	return nil
}

// Disconnect closes the MongoDB connection
func Disconnect() error {
	if DB != nil && DB.client != nil {
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		return DB.client.Disconnect(ctx)
	}
	return nil
}

// IsUserSeen checks if a user has been seen before
func (db *Database) IsUserSeen(userID int64) (bool, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	filter := bson.M{"user_id": userID}
	count, err := db.users.CountDocuments(ctx, filter)
	if err != nil {
		return false, err
	}

	return count > 0, nil
}

// AddUser adds a new user to the database
func (db *Database) AddUser(userID int64, username string) error {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	now := time.Now()
	user := User{
		UserID:    userID,
		Username:  username,
		FirstSeen: now,
		LastSeen:  now,
	}

	_, err := db.users.InsertOne(ctx, user)
	if err != nil {
		db.log.Error("Failed to add user", zap.Error(err), zap.Int64("user_id", userID))
		return err
	}

	db.log.Info("New user added", zap.Int64("user_id", userID), zap.String("username", username))
	return nil
}

// UpdateUserLastSeen updates the last seen time for an existing user
func (db *Database) UpdateUserLastSeen(userID int64) error {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	filter := bson.M{"user_id": userID}
	update := bson.M{"$set": bson.M{"last_seen": time.Now()}}

	_, err := db.users.UpdateOne(ctx, filter, update)
	if err != nil {
		db.log.Error("Failed to update user last seen", zap.Error(err), zap.Int64("user_id", userID))
		return err
	}

	return nil
}

// GetTotalUserCount returns the total number of unique users
func (db *Database) GetTotalUserCount() (int64, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	count, err := db.users.CountDocuments(ctx, bson.M{})
	if err != nil {
		db.log.Error("Failed to get total user count", zap.Error(err))
		return 0, err
	}

	return count, nil
}

// GetAllUsers returns all users (for admin purposes)
func (db *Database) GetAllUsers() ([]User, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	cursor, err := db.users.Find(ctx, bson.M{})
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)

	var users []User
	if err = cursor.All(ctx, &users); err != nil {
		return nil, err
	}

	return users, nil
}
