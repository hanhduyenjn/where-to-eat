package postgres

import (
	"context"
	"fmt"

	"wheretoeat/internal/core/domain"
	"github.com/jmoiron/sqlx"
)

type PlacesRepo struct {
	db *sqlx.DB
}

func NewPlacesRepo(db *sqlx.DB) *PlacesRepo {
	return &PlacesRepo{db: db}
}

func (r *PlacesRepo) SavePlace(ctx context.Context, place domain.Place) error {
	query := `
		INSERT INTO places (
			place_id, name, category, lat, lng, rating, icon_mask_base_uri, primary_type,
			short_address, phone_number, international_phone, takeout, good_for_groups,
			google_maps_uri, utc_offset_minutes, icon_background_color, live_music, restroom,
			dine_in, serves_breakfast, formatted_address, location
		) VALUES (
			:place_id, :name, :category, :lat, :lng, :rating, :icon_mask_base_uri, :primary_type,
			:short_address, :phone_number, :international_phone, :takeout, :good_for_groups,
			:google_maps_uri, :utc_offset_minutes, :icon_background_color, :live_music, :restroom,
			:dine_in, :serves_breakfast, :formatted_address, ST_SetSRID(ST_MakePoint(:lng, :lat), 4326)
		) ON CONFLICT (place_id) DO NOTHING`
	_, err := r.db.NamedExecContext(ctx, query, place)
	if err != nil {
		return fmt.Errorf("failed to insert place: %w", err)
	}
	return nil
}

func (r *PlacesRepo) SaveBatch(ctx context.Context, places []domain.Place, photos []domain.Photo, reviews []domain.Review, openingHours []struct {
	PlaceID string
	Type    string
	Periods string
}, placeTypes []struct {
	PlaceID string
	Type    string
}) error {
	tx, err := r.db.BeginTxx(ctx, nil) // Use BeginTxx to start a sqlx transaction
	if err != nil {
		return err
	}
	defer tx.Rollback()

	// Batch insert into places
	if len(places) > 0 {
		query := `
			INSERT INTO places (
				place_id, name, category, lat, lng, rating, icon_mask_base_uri, primary_type,
				short_address, phone_number, international_phone, takeout, good_for_groups,
				google_maps_uri, utc_offset_minutes, icon_background_color, live_music, restroom,
				dine_in, serves_breakfast, formatted_address, user_rating_count, location 
			) VALUES (
				:place_id, :name, :category, :lat, :lng, :rating, :icon_mask_base_uri, :primary_type,
				:short_address, :phone_number, :international_phone, :takeout, :good_for_groups,
				:google_maps_uri, :utc_offset_minutes, :icon_background_color, :live_music, :restroom,
				:dine_in, :serves_breakfast, :formatted_address, :user_rating_count, ST_SetSRID(ST_MakePoint(:lng, :lat), 4326)
			) ON CONFLICT (place_id) DO NOTHING`
		_, err = tx.NamedExecContext(ctx, query, places)
		if err != nil {
			return fmt.Errorf("failed to batch insert places: %w", err)
		}
	}

	// Batch insert into photos
	if len(photos) > 0 {
		// Validate photos data
		query := `
			INSERT INTO photos (place_id, flag_content_uri)
			VALUES (:place_id, :flag_content_uri)
			ON CONFLICT (flag_content_uri) DO NOTHING`
		_, err = tx.NamedExecContext(ctx, query, photos)
		if err != nil {
			return fmt.Errorf("failed to batch insert photos: %w", err)
		}
	}

	// Batch insert into opening_hours
	if len(openingHours) > 0 {
		query := `
			INSERT INTO opening_hours (place_id, type, periods)
			VALUES (:placeid, :type, :periods)
			ON CONFLICT DO NOTHING`
		_, err = tx.NamedExecContext(ctx, query, openingHours)
		if err != nil {
			return fmt.Errorf("failed to batch insert opening hours: %w", err)
		}
	}

	// Batch insert into place_types
	if len(placeTypes) > 0 {
		query := `
			INSERT INTO place_types (place_id, type)
			VALUES (:placeid, :type)
			ON CONFLICT DO NOTHING`
		_, err = tx.NamedExecContext(ctx, query, placeTypes)
		if err != nil {
			return fmt.Errorf("failed to batch insert place types: %w", err)
		}
	}

	return tx.Commit()
}

func (r *PlacesRepo) GetPhotos(ctx context.Context, limit, offset int) ([]domain.Photo, error) {

	query := `SELECT * FROM photos LIMIT $1 OFFSET $2`
	var photos []domain.Photo

	err := r.db.SelectContext(ctx, &photos, query, limit, offset)
	if err != nil {
		return nil, fmt.Errorf("failed to get photos: %w", err)
	}

	fmt.Printf("Fetched photos: %+v\n", photos)
	return photos, nil
}

func (r *PlacesRepo) UpdatePhotoURL(ctx context.Context, imgUrl, placeID, uuid string) error {
	query := `UPDATE photos SET img_url = $1 WHERE place_id = $2 AND photo_id = $3`
	_, err := r.db.ExecContext(ctx, query, imgUrl, placeID, uuid)
	if err != nil {
		return fmt.Errorf("failed to update photo URL: %w", err)
	}
	return nil
}

func (r *PlacesRepo) GetNearbyPlaces(ctx context.Context, category string, circle domain.Circle, searchString string) ([]domain.Place, error) {
	// Base query
	placesQuery := `
		SELECT place_id, name, lat, lng, rating, user_rating_count, primary_type,
			phone_number, formatted_address, google_maps_uri,
			CASE 
				WHEN COALESCE($4, '') != '' THEN ts_rank(
					to_tsvector('vietnamese', name) || to_tsvector('english', name), 
					plainto_tsquery('vietnamese', $4) || plainto_tsquery('english', $4)
				)
				ELSE NULL
			END AS search_rank
		FROM places
		WHERE ST_DWithin(
				geography(ST_MakePoint(lng, lat)),
				geography(ST_MakePoint($1, $2)),
				$3
			)
			AND (user_rating_count > 100 OR (user_rating_count > 10 AND rating > 4.0))
	`

	args := []interface{}{circle.Lng, circle.Lat, circle.Radius, searchString}

	// Handle category filter
	if category != "" {
		placesQuery += " AND category = $5"
		args = append(args, category)
	}

	// ORDER BY - Prioritize search rank if searchString exists
	placesQuery += `
		ORDER BY 
		CASE 
			WHEN COALESCE($4, '') != '' THEN ts_rank(
				to_tsvector('vietnamese', name) || to_tsvector('english', name), 
				plainto_tsquery('vietnamese', $4) || plainto_tsquery('english', $4)
			) 
		END DESC NULLS LAST,
		user_rating_count DESC,
		ST_Distance(
			geography(ST_MakePoint(lng, lat)),
			geography(ST_MakePoint($1, $2))
		) ASC
		LIMIT 100
	`

	// Fetch places
	var places []domain.Place
	err := r.db.SelectContext(ctx, &places, placesQuery, args...)
	if err != nil {
		return nil, err
	}

	if len(places) == 0 {
		return places, nil // No places found, return early
	}

	// Extract place IDs
	placeIDs := make([]string, len(places))
	placeMap := make(map[string]*domain.Place) // Map to aggregate photos
	for i, p := range places {
		placeIDs[i] = p.ID
		placeMap[p.ID] = &places[i] // Reference for modification
	}

	var photos []domain.Photo
	// Fetch photos for those places
	query, args, err := sqlx.In("SELECT place_id, img_url FROM photos WHERE place_id IN (?)", placeIDs)
	if err != nil {
		return nil, err
	}
	query = r.db.Rebind(query) // Adjust SQL for PostgreSQL compatibility


	err = r.db.SelectContext(ctx, &photos, query, args...)
	if err != nil {
		return nil, err
	}

	// Map photos to their respective places
	for _, photo := range photos {
		if place, exists := placeMap[photo.PlaceID]; exists {
			//cannot use photo.ImageUrl (variable of type sql.NullString) as domain.Photo value in argument to append
			place.PhotoUrls = append(place.PhotoUrls, photo.ImageUrl.String)
		}
	}

	return places, nil
}


func (r *PlacesRepo) GetNumPlaces(ctx context.Context, category string, circle domain.Circle) (int64, error) {
	// not implemented
	return -1, nil
}

func (r *PlacesRepo) AreaHasBeenScanned(ctx context.Context, category string, circle domain.Circle) (bool, error) {
	// not implemented
	return false, nil
}

func (r *PlacesRepo) SavePlaces(ctx context.Context, category string, circle domain.Circle, places []interface{}) error {
	// not implemented
	return nil
}

