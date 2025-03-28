package service

import (
	"context"
	"wheretoeat/internal/core/domain"
	"wheretoeat/internal/core/port"
)

type GetPlacesService struct {
	placesRepo port.PlacesRepository
}

func NewGetPlacesService(placesRepo port.PlacesRepository) *GetPlacesService {
	return &GetPlacesService{placesRepo: placesRepo}
}

func (s *GetPlacesService) GetNearbyPlaces(ctx context.Context, lat, lng, radius float64, category string, searchString string) ([]domain.Place, error) {
	// construct the circle from the lat, lng and radius
	circle := domain.Circle{
		Lat:    lat,
		Lng:    lng,
		Radius: radius,
	}
	// call to repository to get the places
	places, err := s.placesRepo.GetNearbyPlaces(ctx, category, circle, searchString)
	if err != nil {
		return nil, err
	}	
	return places, nil
}