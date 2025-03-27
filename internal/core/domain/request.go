package domain

type RequestParams struct {
	Types  []string         // For Nearby Search
	Query  string           // For Text Search
	Circle Circle    // For location-based searches
}