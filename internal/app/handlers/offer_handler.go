package handlers

import (
	"fmt"
	"math"
	"neawn-backend/internal/app/models"
	"neawn-backend/pkg/utils"
	"net/http"
	"slices"
	"sort"

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

type SearchParams struct {
	RegionID              *int32  `form:"regionID" binding:"required"`
	TimeRangeStart        int64   `form:"timeRangeStart" binding:"required"`
	TimeRangeEnd          int64   `form:"timeRangeEnd" binding:"required"`
	NumberDays            uint16  `form:"numberDays" binding:"required"`
	SortOrder             string  `form:"sortOrder" binding:"required,oneof=price-asc price-desc"`
	Page                  *uint32 `form:"page" binding:"required"`
	PageSize              uint32  `form:"pageSize" binding:"required"`
	PriceRangeWidth       uint32  `form:"priceRangeWidth" binding:"required"`
	MinFreeKilometerWidth uint32  `form:"minFreeKilometerWidth" binding:"required"`
	MinNumberSeats        uint8   `form:"minNumberSeats"`
	MinPrice              uint16  `form:"minPrice"`
	MaxPrice              uint16  `form:"maxPrice"`
	CarType               string  `form:"carType" binding:"omitempty,oneof=small sports luxury family"`
	OnlyVollkasko         *bool   `form:"onlyVollkasko"`
	MinFreeKilometer      uint16  `form:"minFreeKilometer"`
}

func (h *OfferHandler) GetOffers(c *gin.Context) {
	var params SearchParams
	if err := c.ShouldBindQuery(&params); err != nil {
		fmt.Println(err)
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// 1. Filter offers
	filteredOffers := filterOffers(h.offers, params)

	// 2. Sort offers
	sortOffers(filteredOffers, params.SortOrder)
	// 3. Paginate results
	start, end := calculatePagination(len(filteredOffers), *params.Page, params.PageSize)
	paginatedOffers := filteredOffers
	if end > 0 {
		paginatedOffers = filteredOffers[start:end]
	}

	// 4. Convert to SearchResultOffers
	searchResults := make([]models.SearchResultOffer, len(paginatedOffers))
	for i, offer := range paginatedOffers {
		searchResults[i] = models.SearchResultOffer{
			ID:   offer.ID,
			Data: offer.Data,
		}
	}

	// 5. Generate aggregations
	response := models.SearchResponse{
		Offers:             searchResults,
		PriceRanges:        generatePriceRanges(filteredOffers, params.PriceRangeWidth),
		CarTypeCounts:      generateCarTypeCounts(filteredOffers),
		SeatsCount:         generateSeatsCount(filteredOffers),
		FreeKilometerRange: generateFreeKilometerRanges(filteredOffers, params.MinFreeKilometerWidth),
		VollkaskoCount:     generateVollkaskoCount(filteredOffers),
	}

	c.JSON(http.StatusOK, response)
}

func (h *OfferHandler) CreateOffers(c *gin.Context) {
	var request models.CreateOffersRequest
	if err := c.ShouldBindJSON(&request); err != nil {
		fmt.Println(err.Error())
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

func generateCarTypeCounts(offers []models.Offer) models.CarTypeCount {
	counts := models.CarTypeCount{
		Small:  0,
		Sports: 0,
		Luxury: 0,
		Family: 0,
	}

	for _, offer := range offers {
		switch offer.CarType {
		case "small":
			counts.Small++
		case "sports":
			counts.Sports++
		case "luxury":
			counts.Luxury++
		case "family":
			counts.Family++
		}
	}

	return counts
}

func generateSeatsCount(offers []models.Offer) []models.SeatsCount {
	// Create a map to count occurrences of each seat number
	seatCounts := make(map[uint8]uint32)
	for _, offer := range offers {
		seatCounts[offer.NumberSeats]++
	}

	// Convert map to slice of SeatsCount
	result := make([]models.SeatsCount, 0, len(seatCounts))
	for seats, count := range seatCounts {
		result = append(result, models.SeatsCount{
			NumberSeats: seats,
			Count:       count,
		})
	}

	// Sort by number of seats
	sort.Slice(result, func(i, j int) bool {
		return result[i].NumberSeats < result[j].NumberSeats
	})

	return result
}

func generateFreeKilometerRanges(offers []models.Offer, width uint32) []models.FreeKilometerRange {
	if len(offers) == 0 {
		return []models.FreeKilometerRange{}
	}

	// Find min and max free kilometers
	minKm := offers[0].FreeKilometers
	maxKm := offers[0].FreeKilometers
	for _, offer := range offers {
		if offer.FreeKilometers < minKm {
			minKm = offer.FreeKilometers
		}
		if offer.FreeKilometers > maxKm {
			maxKm = offer.FreeKilometers
		}
	}

	// Generate ranges
	ranges := make([]models.FreeKilometerRange, 0)
	// Round minKm down and maxKm up to nearest multiple of width
	minKm = uint16(math.Floor(float64(minKm)/float64(width))) * uint16(width)
	maxKm = uint16(math.Ceil(float64(maxKm)/float64(width))) * uint16(width)

	for start := minKm; start <= maxKm; start += uint16(width) {
		end := start + uint16(width) - 1
		count := 0
		for _, offer := range offers {
			if offer.FreeKilometers >= start && offer.FreeKilometers <= end {
				count++
			}
		}
		if count > 0 {
			ranges = append(ranges, models.FreeKilometerRange{
				Start: start,
				End:   end,
				Count: uint32(count),
			})
		}
	}

	return ranges
}

func generateVollkaskoCount(offers []models.Offer) models.VollkaskoCount {
	counts := models.VollkaskoCount{
		TrueCount:  0,
		FalseCount: 0,
	}

	for _, offer := range offers {
		if offer.HasVollkasko {
			counts.TrueCount++
		} else {
			counts.FalseCount++
		}
	}

	return counts
}

// Helper functions
func filterOffers(offers []models.Offer, params SearchParams) []models.Offer {
	filtered := make([]models.Offer, 0)
	for _, offer := range offers {
		if !matchesFilters(offer, params) {
			continue
		}
		filtered = append(filtered, offer)
	}
	return filtered
}

func matchesFilters(offer models.Offer, params SearchParams) bool {
	// Time range check
	// Check if offer is completely outside the time window

	fmt.Println(offer.EndDate, params.TimeRangeStart, offer.StartDate, params.TimeRangeEnd)
	fmt.Println(offer.EndDate < params.TimeRangeStart, offer.StartDate > params.TimeRangeEnd)

	if offer.EndDate < params.TimeRangeStart || offer.StartDate > params.TimeRangeEnd {
		return false
	}

	// Calculate the overlap duration between offer and time window
	overlapStart := max(offer.StartDate, params.TimeRangeStart)
	overlapEnd := min(offer.EndDate, params.TimeRangeEnd)
	overlapDuration := overlapEnd - overlapStart

	fmt.Println(overlapDuration)

	// Check if overlap is long enough for requested number of days
	if overlapDuration < int64(params.NumberDays)*86400000 {
		return false
	}

	// Region check
	_, isParentRegion := utils.RegionToSubregions[*params.RegionID]

	// fmt.Println(isParentRegion, *params.RegionID, offer.MostSpecificRegionID)

	if isParentRegion {
		if !slices.Contains(utils.RegionToSubregions[*params.RegionID], offer.MostSpecificRegionID) {
			return false
		}
	} else {
		if *params.RegionID != offer.MostSpecificRegionID {
			return false
		}
	}

	// Price range check
	if params.MinPrice > 0 && offer.Price < params.MinPrice {
		return false
	}
	if params.MaxPrice > 0 && offer.Price > params.MaxPrice {
		return false
	}

	// Car type check
	if params.CarType != "" && offer.CarType != params.CarType {
		return false
	}

	// Seats check
	if params.MinNumberSeats > 0 && offer.NumberSeats < params.MinNumberSeats {
		return false
	}

	// Free kilometer check
	if params.MinFreeKilometer > 0 && offer.FreeKilometers < params.MinFreeKilometer {
		return false
	}

	// Vollkasko check
	if params.OnlyVollkasko != nil && *params.OnlyVollkasko && !offer.HasVollkasko {
		return false
	}

	return true
}

func sortOffers(offers []models.Offer, sortOrder string) {
	sort.Slice(offers, func(i, j int) bool {
		switch sortOrder {
		case "price-asc":
			return offers[i].Price < offers[j].Price
		case "price-desc":
			return offers[i].Price > offers[j].Price
		default:
			return offers[i].Price < offers[j].Price
		}
	})
}

func calculatePagination(total int, page, pageSize uint32) (int, int) {
	start := int((page - 1) * pageSize)
	if start >= total {
		return 0, 0
	}
	end := int(math.Min(float64(start+int(pageSize)), float64(total)))
	return start, end
}

func generatePriceRanges(offers []models.Offer, width uint32) []models.PriceRange {
	if len(offers) == 0 {
		return []models.PriceRange{}
	}

	// Find min and max prices
	minPrice := offers[0].Price
	maxPrice := offers[0].Price
	for _, offer := range offers {
		if offer.Price < minPrice {
			minPrice = offer.Price
		}
		if offer.Price > maxPrice {
			maxPrice = offer.Price
		}
	}

	// Generate ranges
	ranges := make([]models.PriceRange, 0)
	for start := minPrice; start <= maxPrice; start += uint16(width) {
		end := start + uint16(width) - 1
		count := 0
		for _, offer := range offers {
			if offer.Price >= start && offer.Price <= end {
				count++
			}
		}
		if count > 0 {
			ranges = append(ranges, models.PriceRange{
				Start: start,
				End:   end,
				Count: uint32(count),
			})
		}
	}
	return ranges
}
