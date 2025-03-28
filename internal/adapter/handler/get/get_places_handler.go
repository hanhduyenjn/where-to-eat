package get

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"wheretoeat/internal/core/port"
)

type GetPlacesHandler struct {
	service port.GetPlacesServicePort
}

func NewGetPlacesHandler(service port.GetPlacesServicePort) *GetPlacesHandler {
	return &GetPlacesHandler{service: service}
}

func (h *GetPlacesHandler) Handle(c *gin.Context) {
	lat, err := strconv.ParseFloat(c.Query("lat"), 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid latitude"})
		return
	}

	lng, err := strconv.ParseFloat(c.Query("lng"), 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid longitude"})
		return
	}

	// if radius missing, get default value 15 km
	radius, err := strconv.ParseFloat(c.Query("radius"), 64)
	if err != nil {
		radius = 15.0 // default radius in km
	}
	if radius <= 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid radius"})
		return
	}

	category := c.Query("category")
	searchString := c.Query("searchString")
	places, err := h.service.GetNearbyPlaces(c.Request.Context(), lat, lng, radius, category, searchString)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, places)
}