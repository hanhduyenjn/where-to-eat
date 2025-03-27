package mongodb

import (
	"context"
	"fmt"
	"log"
	"os"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"

	"wheretoeat/internal/core/domain"
)

type PlacesRepo struct {
	client           *mongo.Client
	placesCollection *mongo.Collection
}

func NewMongoAdapter() (*mongo.Client, error) {
	mongoURI := os.Getenv("MONGO_URI")
	clientOptions := options.Client().ApplyURI(mongoURI)
	client, err := mongo.Connect(context.TODO(), clientOptions)
	if err != nil {
		return nil, fmt.Errorf("error connecting to MongoDB: %w", err)
	}
	return client, nil
}

func NewPlacesRepo(client *mongo.Client) *PlacesRepo {
	mongoDB := os.Getenv("MONGO_DB")
	mongoPlacesCollection := os.Getenv("MONGO_COLLECTION_PLACES")
	return &PlacesRepo{
		client:           client,
		placesCollection: client.Database(mongoDB).Collection(mongoPlacesCollection),
	}
}

func (r *PlacesRepo) SavePlaces(ctx context.Context, category string, circle domain.Circle, places []interface{}) error {
	_, err := r.placesCollection.InsertOne(ctx, bson.M{
		"category": category,
		"lat":      circle.Lat,
		"lng":      circle.Lng,
		"radius":   circle.Radius,
		"places":   places,
	})
	if err != nil {
		return fmt.Errorf("error inserting places into MongoDB: %w", err)
	}
	return nil
}

func (r *PlacesRepo) CircleExists(ctx context.Context, category string, circle domain.Circle) (bool, error) {
	
	// This can be done by checking if the given circle's center is within the radius of any existing circle.

	count, err := r.placesCollection.CountDocuments(ctx, bson.M{
		"lat":    circle.Lat,
		"lng":    circle.Lng,
		"radius": circle.Radius,
	})

	if err != nil {
		return false, fmt.Errorf("error counting documents in MongoDB: %w", err)
	}
	return count > 0, nil
}

func (r *PlacesRepo) GetNumPlaces(ctx context.Context, category string, circle domain.Circle) (int64, error) {
	// Find one document matching the category and circle
	var result struct {
		Places []interface{} `bson:"places"`
	}
	err := r.placesCollection.FindOne(ctx, bson.M{
		"category": category,
		"lat":      circle.Lat,
		"lng":      circle.Lng,
		"radius":   circle.Radius,
	}).Decode(&result)
	if err == mongo.ErrNoDocuments {
		// No matching document found, return 0
		return 0, nil
	}
	if err != nil {
		return 0, fmt.Errorf("error finding document in MongoDB: %w", err)
	}

	// Return the length of the 'places' array
	return int64(len(result.Places)), nil
}

func (r *PlacesRepo) GetPlaces(ctx context.Context, limit int, offset int) ([]domain.Place, error) {
	cursor, err := r.placesCollection.Find(ctx, bson.M{}, options.Find().SetLimit(int64(limit)).SetSkip(int64(offset)))
	if err != nil {
		return nil, fmt.Errorf("error finding documents in MongoDB: %w", err)
	}
	defer cursor.Close(ctx)

	var documents []domain.PlacesRawResponse
	if err = cursor.All(ctx, &documents); err != nil {
		return nil, fmt.Errorf("error decoding documents in MongoDB: %w", err)
	}
	// extract the places array inside each document
	var places []domain.Place
	for _, doc := range documents {
		log.Printf("Places: %v", doc.Places)
		places = append(places, doc.Places...)
	}

	return places, nil
}

