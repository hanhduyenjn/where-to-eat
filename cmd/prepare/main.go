package prepare

import (
	"context"
	"fmt"
	"log"
	"os"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"

	"github.com/joho/godotenv"
)

var categories = map[string][]string{
	"restaurants": {
		"afghani_restaurant", "african_restaurant", "american_restaurant", "asian_restaurant", "brazilian_restaurant",
		"chinese_restaurant", "french_restaurant", "greek_restaurant", "indian_restaurant", "indonesian_restaurant",
		"italian_restaurant", "japanese_restaurant", "korean_restaurant", "lebanese_restaurant", "mediterranean_restaurant",
		"mexican_restaurant", "middle_eastern_restaurant", "spanish_restaurant", "thai_restaurant", "turkish_restaurant",
		"vietnamese_restaurant", "fine_dining_restaurant", "buffet_restaurant", "barbecue_restaurant", "seafood_restaurant",
		"steakhouse", "sushi_restaurant", "ramen_restaurant",
	},
	"casual_takeaway": {
		"fast_food_restaurant", "diner", "pizza_restaurant", "sandwich_shop", "food_court", "deli", "meal_delivery", "meal_takeaway",
	},
	"alcoholic_places": {
		"bar", "bar_and_grill", "pub", "wine_bar",
	},
	"desserts_sweets_bakery": {
		"acai_shop", "bakery", "bagel_shop", "candy_store", "chocolate_factory", "chocolate_shop",
		"confectionery", "dessert_restaurant", "dessert_shop", "ice_cream_shop", "donut_shop",
	},
	"cafes_beverages": {
		"cafe", "coffee_shop", "cat_cafe", "dog_cafe", "juice_shop", "tea_house",
	},
	"specialty_dietary": {
		"vegetarian_restaurant", "vegan_restaurant", "cafeteria",
	},
}

func insertCategoriesToMongo() {
	if err := godotenv.Load(); err != nil {
		log.Fatal("Error loading .env file")
	}
	
	mongoURI := os.Getenv("MONGO_URI")
	mongoDB := os.Getenv("MONGO_DB")
	mongoConfig := os.Getenv("MONGO_COLLECTION_CONFIG")

	clientOptions := options.Client().ApplyURI(mongoURI)
	client, err := mongo.Connect(context.TODO(), clientOptions)
	if err != nil {
		log.Fatal(err)
	}
	defer client.Disconnect(context.TODO())

	db := client.Database(mongoDB)
	collection := db.Collection(mongoConfig)

	for category, types := range categories {
		doc := bson.M{"category": category, "types": types}
		_, err := collection.InsertOne(context.TODO(), doc)
		if err != nil {
			log.Println("Error inserting category:", category, err)
		} else {
			fmt.Println("Inserted category:", category)
		}
	}
}

func main() {
	insertCategoriesToMongo()
}