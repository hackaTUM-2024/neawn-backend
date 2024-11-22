package handlers

import (
	"neawn-backend/internal/app/models"
	"net/http"

	"github.com/gin-gonic/gin"
)

type OfferHandler struct {
	offers []models.Offer // This should be replaced with a proper database
}

func NewOfferHandler() *OfferHandler {
	return &OfferHandler{
		offers: make([]models.Offer, 0),
	}
}

func (h *OfferHandler) GetOffers(c *gin.Context) {
	var params struct {
		RegionID              int32  `form:"regionID" binding:"required"`
		TimeRangeStart        int64  `form:"timeRangeStart" binding:"required"`
		TimeRangeEnd          int64  `form:"timeRangeEnd" binding:"required"`
		NumberDays            uint16 `form:"numberDays" binding:"required"`
		SortOrder             string `form:"sortOrder" binding:"required,oneof=price-asc price-desc"`
		Page                  uint32 `form:"page" binding:"required"`
		PageSize              uint32 `form:"pageSize" binding:"required"`
		PriceRangeWidth       uint32 `form:"priceRangeWidth" binding:"required"`
		MinFreeKilometerWidth uint32 `form:"minFreeKilometerWidth" binding:"required"`
		MinNumberSeats        uint8  `form:"minNumberSeats"`
		MinPrice              uint16 `form:"minPrice"`
		MaxPrice              uint16 `form:"maxPrice"`
		CarType               string `form:"carType" binding:"omitempty,oneof=small sports luxury family"`
		OnlyVollkasko         *bool  `form:"onlyVollkasko"`
		MinFreeKilometer      uint16 `form:"minFreeKilometer"`
	}

	if err := c.ShouldBindQuery(&params); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// TODO: Implement the actual filtering, sorting, and aggregation logic
	response := models.SearchResponse{
		Offers:             []models.SearchResultOffer{},
		PriceRanges:        []models.PriceRange{},
		CarTypeCounts:      models.CarTypeCount{},
		SeatsCount:         []models.SeatsCount{},
		FreeKilometerRange: []models.FreeKilometerRange{},
		VollkaskoCount:     models.VollkaskoCount{},
	}

	c.JSON(http.StatusOK, response)
}

func (h *OfferHandler) CreateOffers(c *gin.Context) {
	var request models.CreateOffersRequest
	if err := c.ShouldBindJSON(&request); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if len(request.Offers) == 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "at least one offer is required"})
		return
	}

	// TODO: Implement actual offer creation logic
	h.offers = append(h.offers, request.Offers...)

	c.Status(http.StatusOK)
}

func (h *OfferHandler) CleanupData(c *gin.Context) {
	// TODO: Implement cleanup logic
	h.offers = make([]models.Offer, 0)
	c.Status(http.StatusOK)
}
