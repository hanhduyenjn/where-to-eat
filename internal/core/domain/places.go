package domain

import (
	"go.mongodb.org/mongo-driver/bson/primitive"
	"database/sql"
)

const (
	OverlapFactor     = 0.9             // Overlap between circles (90%)
	LatMeterToDegree  = 1.0 / 111320    // Conversion factor: meters to latitude degrees
	LngCosineAdjust   = 0.995           // Cosine adjustment for longitude at ~10.8° latitude
	MaxResultsPerReq  = 15              // Google Places API max results per request
	RateLimitDelay    = 100             // Delay in milliseconds between requests
	MinRadius         = 50.0            // Minimum radius in meters
)

type Circle struct {
	Lat    float64
	Lng    float64
	Radius float64
}

// Response of one single near-by-search request
type PlacesRawResponse struct {
	ID       primitive.ObjectID `bson:"_id"`
	Category string             `bson:"category"`
	Lat float64            `bson:"lat"`
	Lng float64           `bson:"lng"`
	Radius   float64            `bson:"radius"`
	Places   []Place            `bson:"places"`
}

// Place represents details of a place.
type Place struct {
	ID                 string         `db:"place_id" bson:"id,omitempty"`
	DisplayName       *DisplayName    `bson:"displayName,omitempty"`
	Name			   string  `db:"name"`
    Category             string  `db:"category"`
    Lat                  float64 `db:"lat"`
    Lng                  float64 `db:"lng"`
	Rating              float64       `db:"rating,omitempty"`
	UserRatingCount	int            `db:"user_rating_count" bson:"userRatingCount,omitempty"`
    IconMaskBaseURI     string        `db:"icon_mask_base_uri,omitempty" bson:"iconMaskBaseUri,omitempty"`
    PrimaryType         string        `db:"primary_type" bson:"primaryType,omitempty"`
    ShortAddress        string        `db:"short_address,omitempty" bson:"shortFormattedAddress,omitempty"`
    PhoneNumber         string        `db:"phone_number,omitempty" bson:"nationalPhoneNumber,omitempty"`
	Reviews            []Review       `bson:"reviews,omitempty"`
	Photos             []Photo        `bson:"photos,omitempty"`
	PhotoUrls		 []string       `bson:"photoUrls,omitempty"`
	TimeZone           TimeZone      `bson:"timeZone,omitempty"`
	OpeningHours       *OpeningHours  `bson:"regularOpeningHours,omitempty"`
	Takeout            bool           `bson:"takeout,omitempty"`
	GoodForGroups      bool           `db:"good_for_groups,omitempty" bson:"goodForGroups,omitempty"`
	InternationalPhone string         `db:"international_phone,omitempty" bson:"internationalPhoneNumber,omitempty"`
	GoogleMapsLinks    *GoogleLinks   `bson:"googleMapsLinks,omitempty"`
	UTCOffsetMinutes   int            `db:"utc_offset_minutes" bson:"utcOffsetMinutes,omitempty"`
	IconBackgroundColor string        `db:"icon_background_color,omitempty" bson:"iconBackgroundColor,omitempty"`
	LiveMusic          bool           `db:"live_music,omitempty" bson:"liveMusic,omitempty"`
	Restroom           bool           `bson:"restroom,omitempty"`
	DineIn 		   bool               `db:"dine_in,omitempty" bson:"dineIn,omitempty"`
	ServesBreakfast    bool            `db:"serves_breakfast" bson:"servesBreakfast,omitempty"`
	Types             []string        `bson:"types,omitempty"`
	FormattedAddress   string         `db:"formatted_address,omitempty" bson:"formattedAddress,omitempty"`
	Location          *Coordinates    `bson:"location,omitempty"`
	GoogleMapsUri      string         `db:"google_maps_uri" bson:"googleMapsUri,omitempty"`
	CurrentOpeningHours *OpeningHours  `bson:"currentOpeningHours,omitempty"`
	SearchRank 	   float64            `db:"search_rank" bson:"searchRank,omitempty"`
	
}

// Review represents a user's review.
type Review struct {
	RelativePublishTimeDescription string   `db:"relative_publish_time_description" bson:"relativePublishTimeDescription,omitempty"`
	Rating                         float64  `bson:"rating,omitempty"`
	AuthorAttribution              Author `bson:"authorAttribution,omitempty"`
	PublishTime                    string   `db:"publish_time" bson:"publishTime,omitempty"`
	FlagContentUri                 string   `bson:"flagContentUri,omitempty"`
	GoogleMapsUri                   string   `bson:"googleMapsUri,omitempty"`
	Name                            string   `bson:"name,omitempty"`
}

// Author represents the author of a review.
type Author struct {
	DisplayName string `bson:"displayName,omitempty"`
	Uri         string `bson:"uri,omitempty"`
	PhotoUri    string `bson:"photoUri,omitempty"`
}

// Photo represents details of a photo.
type Photo struct {
	GoogleMapsUri     string   `bson:"googleMapsUri,omitempty"`
	PhotoId		      string   `db:"photo_id" bson:"photoId,omitempty"`
	PlaceID           string   `db:"place_id" bson:"placeId,omitempty"`
	FlagContentUri    string   `db:"flag_content_uri" bson:"flagContentUri,omitempty"`
	ImageUrl              sql.NullString `db:"img_url" bson:"s3Url,omitempty"`
	Name              string   `bson:"name,omitempty"`
	WidthPx          int      `bson:"widthPx,omitempty"`
	HeightPx         int      `bson:"heightPx,omitempty"`
	AuthorAttributions []Author `bson:"authorAttributions,omitempty"`
}

// TimeZone represents the time zone info.
type TimeZone struct {
	ID string `bson:"id,omitempty"`
}

// OpeningHours represents the opening hours of a place.
type OpeningHours struct {
	NextOpenTime       string    `db:"next_open_time,omitempty" bson:"nextOpenTime,omitempty"`
	OpenNow            bool      `db:"open_now,omitempty" bson:"openNow,omitempty"`
	Periods            []Period  `bson:"periods,omitempty"`
	WeekdayDescriptions []string `db:"weekday_descriptions,omitempty" bson:"weekdayDescriptions,omitempty"`
}

// Period represents the opening and closing time of a place.
type Period struct {
	Open  TimePeriod `bson:"open,omitempty"`
	Close TimePeriod `bson:"close,omitempty"`
}

// TimePeriod represents a specific opening/closing time.
type TimePeriod struct {
	Day    int `bson:"day,omitempty"`
	Hour   int `bson:"hour,omitempty"`
	Minute int `bson:"minute,omitempty"`
}

// Coordinates represents a geographic location.
type Coordinates struct {
	Lat  float64 `bson:"latitude,omitempty"`
	Lng float64 `bson:"longitude,omitempty"`
}

// DisplayName represents the display name of a place.
type DisplayName struct {
	Text         string `bson:"text,omitempty"`
	LanguageCode string `bson:"languageCode,omitempty"`
}

// GoogleLinks represents links related to the place.
type GoogleLinks struct {
	DirectionsUri  string `bson:"directionsUri,omitempty"`
	PlaceUri       string `bson:"placeUri,omitempty"`
	WriteReviewUri string `bson:"writeAReviewUri,omitempty"`
	ReviewsUri     string `bson:"reviewsUri,omitempty"`
	PhotosUri      string `bson:"photosUri,omitempty"`
}