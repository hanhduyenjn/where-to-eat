-- Enable PostGIS for geospatial queries
CREATE EXTENSION IF NOT EXISTS postgis;

-- Places table (flattened with category, lat, lng, radius from raw response)
CREATE TABLE places (
    place_id VARCHAR(255) PRIMARY KEY,
    name TEXT NOT NULL,
    category VARCHAR(255),
    lat DOUBLE PRECISION, 
    lng DOUBLE PRECISION, 
    radius DOUBLE PRECISION, 
    user_rating_count INT,
    rating DOUBLE PRECISION,
    icon_mask_base_uri TEXT,
    primary_type VARCHAR(255),
    short_address TEXT,
    phone_number TEXT,
    international_phone TEXT,
    takeout BOOLEAN,
    good_for_groups BOOLEAN,
    google_maps_uri TEXT,
    utc_offset_minutes INT,
    icon_background_color TEXT,
    live_music BOOLEAN,
    restroom BOOLEAN,
    dine_in BOOLEAN,
    serves_breakfast BOOLEAN,
    formatted_address TEXT,
    location GEOMETRY(POINT, 4326) -- PostGIS point for lat/lng
);

-- Photos table
CREATE TABLE photos (
    photo_id SERIAL PRIMARY KEY,
    place_id VARCHAR(255) REFERENCES places(place_id),
    flag_content_uri TEXT,
    img_url TEXT
);

-- Reviews table
CREATE TABLE reviews (
    review_id SERIAL PRIMARY KEY,
    place_id VARCHAR(255) REFERENCES places(place_id),
    text TEXT,
    rating INT,
    author TEXT -- Add more fields as needed
);

-- Opening Hours table
CREATE TABLE opening_hours (
    opening_hours_id SERIAL PRIMARY KEY,
    place_id VARCHAR(255) REFERENCES places(place_id),
    type VARCHAR(50), -- 'regular' or 'current'
    periods TEXT -- JSON or text representation of periods
);

-- Place Types table (many-to-many)
CREATE TABLE place_types (
    place_id VARCHAR(255) REFERENCES places(place_id),
    type VARCHAR(255),
    PRIMARY KEY (place_id, type)
);

-- Indexes for performance
CREATE INDEX places_location_idx ON places USING GIST (location);
CREATE INDEX places_category_idx ON places (category);