package mongodb

import (
	"context"
	"fmt"
	"os"
	"math"
	"log"

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
	mongoPlacesCollection := os.Getenv("MONGO_COLLECTION_SEARCH_RESULTS")
	return &PlacesRepo{
		client:           client,
		placesCollection: client.Database(mongoDB).Collection(mongoPlacesCollection),
	}
}

func (r *PlacesRepo) SaveSearchResults(ctx context.Context, category string, circle domain.Circle, places []interface{}) error {
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

func (r *PlacesRepo) GetRawPlaces(ctx context.Context) ([]domain.PlacesRawResponse, error) {
	cursor, err := r.placesCollection.Find(ctx, bson.M{})
	if err != nil {
		return nil, fmt.Errorf("error finding documents in MongoDB: %w", err)
	}
	defer cursor.Close(ctx)

	var documents []domain.PlacesRawResponse
	if err = cursor.All(ctx, &documents); err != nil {
		return nil, fmt.Errorf("error decoding documents in MongoDB: %w", err)
	}

	return documents, nil
}

func (r *PlacesRepo) AreaHasBeenScanned(ctx context.Context, category string, circle domain.Circle) (bool, error) {
    // Find all circles that might overlap with the given circle
	log.Println("Searching for overlapping circles...")
    var existingCircles []domain.Circle
    cursor, err := r.placesCollection.Find(ctx, bson.M{
        "$and": []bson.M{
            {"lat": bson.M{"$gte": circle.Lat - circle.Radius, "$lte": circle.Lat + circle.Radius}},
            {"lng": bson.M{"$gte": circle.Lng - circle.Radius, "$lte": circle.Lng + circle.Radius}},
			{"category": category},
        },
    })
	log.Println("Number of overlapping circles found: ", cursor.RemainingBatchLength())

    if err != nil {
        return false, fmt.Errorf("error finding overlapping circles in MongoDB: %w", err)
    }
    defer cursor.Close(ctx)

    if err = cursor.All(ctx, &existingCircles); err != nil {
        return false, fmt.Errorf("error decoding circles from MongoDB: %w", err)
    }

    // Calculate total overlapped area
    givenArea := math.Pi * circle.Radius * circle.Radius
    coveredArea := 0.0

    for _, existing := range existingCircles {
        overlap := calculateCircleIntersectionArea(circle, existing)
        coveredArea += overlap
    }
	log.Printf("Percentage of area covered: %.2f%%\n", (coveredArea/givenArea)*100)
    return coveredArea >= givenArea, nil
}

// calculateCircleIntersectionArea calculates the overlapping area between two circles
func calculateCircleIntersectionArea(c1, c2 domain.Circle) float64 {
    d := haversineDistance(c1.Lat, c1.Lng, c2.Lat, c2.Lng)
    
    if d >= c1.Radius+c2.Radius {
        return 0 // No overlap
    }
    
    if d <= math.Abs(c1.Radius-c2.Radius) {
        // One circle is completely inside the other
        return math.Pi * math.Min(c1.Radius, c2.Radius) * math.Min(c1.Radius, c2.Radius)
    }
    
    r1, r2 := c1.Radius, c2.Radius
    
    a1 := r1 * r1 * math.Acos((d*d + r1*r1 - r2*r2) / (2 * d * r1))
    a2 := r2 * r2 * math.Acos((d*d + r2*r2 - r1*r1) / (2 * d * r2))
    a3 := 0.5 * math.Sqrt((-d+r1+r2) * (d+r1-r2) * (d-r1+r2) * (d+r1+r2))
    
    return a1 + a2 - a3
}

// haversineDistance calculates the great-circle distance between two points
func haversineDistance(lat1, lon1, lat2, lon2 float64) float64 {
    const R = 6371e3 // Earth radius in meters
    
    dLat := (lat2 - lat1) * (math.Pi / 180)
    dLon := (lon2 - lon1) * (math.Pi / 180)
    
    a := math.Sin(dLat/2)*math.Sin(dLat/2) +
        math.Cos(lat1*(math.Pi/180))*math.Cos(lat2*(math.Pi/180))*
            math.Sin(dLon/2)*math.Sin(dLon/2)
    
    c := 2 * math.Atan2(math.Sqrt(a), math.Sqrt(1-a))
    return R * c
}
