package mongodb

import (
	"context"
	"os"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
)

// AreasRepo handles interactions with the 'areas' collection in MongoDB
type AreasRepo struct {
	collection *mongo.Collection
}

// NewAreasRepo initializes a new AreasRepo with a MongoDB client and database name
func NewAreasRepo(client *mongo.Client) *AreasRepo {
	mongoDB := os.Getenv("MONGO_DB")
	mongoCollection := os.Getenv("MONGO_COLLECTION_AREAS")
	return &AreasRepo{
		collection: client.Database(mongoDB).Collection(mongoCollection),
	}
}

// SaveArea saves an area document to the 'areas' collection
func (r *AreasRepo) SaveArea(ctx context.Context, area bson.M) error {
	_, err := r.collection.InsertOne(ctx, area)
	if err != nil {
		return err
	}
	return nil
}

// GetAreaByPlaceID retrieves an area by its placeID
func (r *AreasRepo) GetAreaByPlaceID(ctx context.Context, placeID string) (bson.M, error) {
	var result bson.M
	err := r.collection.FindOne(ctx, bson.M{"placeID": placeID}).Decode(&result)
	if err == mongo.ErrNoDocuments {
		return nil, nil // Not found, return nil instead of error
	}
	if err != nil {
		return nil, err
	}
	return result, nil
}