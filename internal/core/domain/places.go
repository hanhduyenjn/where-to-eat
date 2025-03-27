package domain

import (
	"go.mongodb.org/mongo-driver/bson/primitive"
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
	Latitude float64            `bson:"lat"`
	Longitude float64           `bson:"lng"`
	Radius   float64            `bson:"radius"`
	Places   []Place            `bson:"places"`
}

// Place represents details of a place.
type Place struct {
	Rating              float64       `bson:"rating,omitempty"`
	IconMaskBaseUri     string        `bson:"iconMaskBaseUri,omitempty"`
	Reviews            []Review       `bson:"reviews,omitempty"`
	Photos             []Photo        `bson:"photos,omitempty"`
	PrimaryType        string         `bson:"primaryType,omitempty"`
	ShortAddress       string         `bson:"shortFormattedAddress,omitempty"`
	TimeZone           TimeZone      `bson:"timeZone,omitempty"`
	PhoneNumber        string         `bson:"nationalPhoneNumber,omitempty"`
	InternationalPhone string         `bson:"internationalPhoneNumber,omitempty"`
	OpeningHours       *OpeningHours  `bson:"regularOpeningHours,omitempty"`
	Takeout            bool           `bson:"takeout,omitempty"`
	GoodForGroups      bool           `bson:"goodForGroups,omitempty"`
	GoogleMapsLinks    *GoogleLinks   `bson:"googleMapsLinks,omitempty"`
	ID                 string         `bson:"id,omitempty"`
	UTCOffsetMinutes   int            `bson:"utcOffsetMinutes,omitempty"`
	IconBackgroundColor string        `bson:"iconBackgroundColor,omitempty"`
	LiveMusic          bool           `bson:"liveMusic,omitempty"`
	Restroom           bool           `bson:"restroom,omitempty"`
	DineIn             bool           `bson:"dineIn,omitempty"`
	ServesBreakfast    bool           `bson:"servesBreakfast,omitempty"`
	Types             []string        `bson:"types,omitempty"`
	FormattedAddress   string         `bson:"formattedAddress,omitempty"`
	Location          *Coordinates    `bson:"location,omitempty"`
	GoogleMapsUri      string         `bson:"googleMapsUri,omitempty"`
	DisplayName       *DisplayName    `bson:"displayName,omitempty"`
	CurrentOpeningHours *OpeningHours  `bson:"currentOpeningHours,omitempty"`
}

// Review represents a user's review.
type Review struct {
	RelativePublishTimeDescription string   `bson:"relativePublishTimeDescription,omitempty"`
	Rating                         float64  `bson:"rating,omitempty"`
	AuthorAttribution              Author `bson:"authorAttribution,omitempty"`
	PublishTime                    string   `bson:"publishTime,omitempty"`
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
	Name              string   `bson:"name,omitempty"`
	WidthPx          int      `bson:"widthPx,omitempty"`
	HeightPx         int      `bson:"heightPx,omitempty"`
	AuthorAttributions []Author `bson:"authorAttributions,omitempty"`
	FlagContentUri    string   `bson:"flagContentUri,omitempty"`
}

// TimeZone represents the time zone info.
type TimeZone struct {
	ID string `bson:"id,omitempty"`
}

// OpeningHours represents the opening hours of a place.
type OpeningHours struct {
	NextOpenTime       string    `bson:"nextOpenTime,omitempty"`
	OpenNow            bool      `bson:"openNow,omitempty"`
	Periods            []Period  `bson:"periods,omitempty"`
	WeekdayDescriptions []string `bson:"weekdayDescriptions,omitempty"`
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
	Latitude  float64 `bson:"latitude,omitempty"`
	Longitude float64 `bson:"longitude,omitempty"`
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