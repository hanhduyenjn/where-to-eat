package mongodb

import (
	"context"
	"fmt"
	"os"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
)

type CategoriesRepo struct {
	configCollection *mongo.Collection
}

func NewCategoriesRepo(client *mongo.Client) *CategoriesRepo {
	mongoDB := os.Getenv("MONGO_DB")
	mongoConfigCollection := os.Getenv("MONGO_COLLECTION_CONFIG")
	return &CategoriesRepo{
		configCollection: client.Database(mongoDB).Collection(mongoConfigCollection),
	}
}

func (r *CategoriesRepo) GetCategoryTypes(ctx context.Context, category string) ([]string, error) {
	var categoryDoc struct {
		Types []string `bson:"types"`
	}
	err := r.configCollection.FindOne(ctx, bson.M{"category": category}).Decode(&categoryDoc)
	if err != nil {
		return nil, fmt.Errorf("error finding category %s in MongoDB config: %w", category, err)
	}
	return categoryDoc.Types, nil
}