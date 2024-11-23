package handlers

import (
	"fmt"
	"math"
	"neawn-backend/internal/app/models"
	"neawn-backend/pkg/utils"
	"net/http"
	"slices"
	"sort"
	"sync"

	"github.com/gin-gonic/gin"
)

type OfferHandler struct {
	offers []models.Offer
	mutex  sync.RWMutex
}

func NewOfferHandler() *OfferHandler {
	return &OfferHandler{
		offers: make([]models.Offer, 0),
		mutex:  sync.RWMutex{},
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

	// 1. Filter offers (required filters)
	filteredOffers := filterOffers(h.offers, params)

	// Optional filters
	var filteredOptionalOffers []models.Offer
	for _, offer := range filteredOffers {
		if params.OnlyVollkasko != nil && *params.OnlyVollkasko && !offer.HasVollkasko {
			continue
		}

		if params.CarType != "" && offer.CarType != params.CarType {
			continue
		}

		if params.MinPrice > 0 && offer.Price < params.MinPrice {
			continue
		}
		if params.MaxPrice > 0 && offer.Price >= params.MaxPrice {
			continue
		}

		if params.MinNumberSeats > 0 && offer.NumberSeats < params.MinNumberSeats {
			continue
		}

		if params.MinFreeKilometer > 0 && offer.FreeKilometers < params.MinFreeKilometer {
			continue
		}

		filteredOptionalOffers = append(filteredOptionalOffers, offer)
	}

	// 2. Sort offers
	sortOffers(filteredOptionalOffers, params.SortOrder)

	// 3. Paginate results
	start, end := calculatePagination(len(filteredOptionalOffers), *params.Page, params.PageSize)
	paginatedOffers := filteredOptionalOffers

	if end > 0 {
		paginatedOffers = filteredOptionalOffers[start:end]
	} else {
		paginatedOffers = []models.Offer{}
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
		PriceRanges:        generatePriceRanges(filteredOffers, params),
		SeatsCount:         generateSeatsCount(filteredOffers, params),
		FreeKilometerRange: generateFreeKilometerRanges(filteredOffers, params),
		CarTypeCounts:      generateCarTypeCounts(filteredOffers, params),
		VollkaskoCount:     generateVollkaskoCount(filteredOffers, params),
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

	h.mutex.Lock()
	defer h.mutex.Unlock()

	h.offers = append(h.offers, request.Offers...)

	fmt.Println("# of offers:", len(h.offers))

	c.Status(http.StatusOK)
}

func (h *OfferHandler) CleanupData(c *gin.Context) {
	h.mutex.Lock()
	defer h.mutex.Unlock()

	h.offers = make([]models.Offer, 0)
	c.Status(http.StatusOK)
}

func generateCarTypeCounts(offers []models.Offer, params SearchParams) models.CarTypeCount {
	counts := models.CarTypeCount{
		Small:  0,
		Sports: 0,
		Luxury: 0,
		Family: 0,
	}

	for _, offer := range offers {
		if params.OnlyVollkasko != nil && *params.OnlyVollkasko && !offer.HasVollkasko {
			continue
		}

		if params.MinPrice > 0 && offer.Price < params.MinPrice {
			continue
		}
		if params.MaxPrice > 0 && offer.Price >= params.MaxPrice {
			continue
		}

		if params.MinNumberSeats > 0 && offer.NumberSeats < params.MinNumberSeats {
			continue
		}

		if params.MinFreeKilometer > 0 && offer.FreeKilometers < params.MinFreeKilometer {
			continue
		}

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

func generateSeatsCount(offers []models.Offer, params SearchParams) []models.SeatsCount {
	// Create a map to count occurrences of each seat number
	seatCounts := make(map[uint8]uint32)
	for _, offer := range offers {
		if params.MinFreeKilometer > 0 && offer.FreeKilometers < params.MinFreeKilometer {
			continue
		}

		if params.CarType != "" && offer.CarType != params.CarType {
			continue
		}

		if params.OnlyVollkasko != nil && *params.OnlyVollkasko && !offer.HasVollkasko {
			continue
		}

		if params.MinPrice > 0 && offer.Price < params.MinPrice {
			continue
		}
		if params.MaxPrice > 0 && offer.Price >= params.MaxPrice {
			continue
		}

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
	// sort.Slice(result, func(i, j int) bool {
	// 	return result[i].NumberSeats < result[j].NumberSeats
	// })

	return result
}

func generateFreeKilometerRanges(offers []models.Offer, params SearchParams) []models.FreeKilometerRange {
	if len(offers) == 0 {
		return []models.FreeKilometerRange{}
	}

	var filteredOffers []models.Offer
	for _, offer := range offers {
		if params.CarType != "" && offer.CarType != params.CarType {
			continue
		}

		if params.OnlyVollkasko != nil && *params.OnlyVollkasko && !offer.HasVollkasko {
			continue
		}

		if params.MinPrice > 0 && offer.Price < params.MinPrice {
			continue
		}
		if params.MaxPrice > 0 && offer.Price >= params.MaxPrice {
			continue
		}

		if params.MinNumberSeats > 0 && offer.NumberSeats < params.MinNumberSeats {
			continue
		}

		filteredOffers = append(filteredOffers, offer)
	}

	if len(filteredOffers) == 0 {
		return []models.FreeKilometerRange{}
	}

	// Find min and max free kilometers
	minKm := filteredOffers[0].FreeKilometers
	maxKm := filteredOffers[0].FreeKilometers
	for _, offer := range filteredOffers {
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
	minKm = uint16(params.MinFreeKilometerWidth) * (minKm / uint16(params.MinFreeKilometerWidth))
	maxKm = uint16(params.MinFreeKilometerWidth) * ((maxKm + uint16(params.MinFreeKilometerWidth) - 1) / uint16(params.MinFreeKilometerWidth))

	// minKm = uint16(math.Floor(float64(minKm)/float64(params.MinFreeKilometerWidth))) * uint16(params.MinFreeKilometerWidth)
	// maxKm = uint16(math.Ceil(float64(maxKm)/float64(params.MinFreeKilometerWidth))) * uint16(params.MinFreeKilometerWidth)

	for start := minKm; start <= maxKm; start += uint16(params.MinFreeKilometerWidth) {
		end := start + uint16(params.MinFreeKilometerWidth) - 1
		count := 0
		for _, offer := range filteredOffers {
			if offer.FreeKilometers >= start && offer.FreeKilometers <= end {
				count++
			}
		}
		if count > 0 {
			ranges = append(ranges, models.FreeKilometerRange{
				Start: start,
				End:   end + 1,
				Count: uint32(count),
			})
		}
	}

	return ranges
}

func generateVollkaskoCount(offers []models.Offer, params SearchParams) models.VollkaskoCount {
	counts := models.VollkaskoCount{
		TrueCount:  0,
		FalseCount: 0,
	}

	for _, offer := range offers {
		if params.CarType != "" && offer.CarType != params.CarType {
			continue
		}

		if params.MinPrice > 0 && offer.Price < params.MinPrice {
			continue
		}
		if params.MaxPrice > 0 && offer.Price >= params.MaxPrice {
			continue
		}

		if params.MinNumberSeats > 0 && offer.NumberSeats < params.MinNumberSeats {
			continue
		}

		if params.MinFreeKilometer > 0 && offer.FreeKilometers < params.MinFreeKilometer {
			continue
		}

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

	if offer.StartDate < params.TimeRangeStart || offer.EndDate > params.TimeRangeEnd || (offer.EndDate-offer.StartDate) != int64(params.NumberDays)*86400000 {
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

	/*

		// Price range check
		if params.MinPrice > 0 && offer.Price < params.MinPrice {
			return false
		}
		if params.MaxPrice > 0 && offer.Price > params.MaxPrice {
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
			// Car type check
			if params.CarType != "" && offer.CarType != params.CarType {
				return false
			}

			// Vollkasko check
			if params.OnlyVollkasko != nil && *params.OnlyVollkasko && !offer.HasVollkasko {
				return false
			}
	*/

	return true
}

func sortOffers(offers []models.Offer, sortOrder string) {
	sort.Slice(offers, func(i, j int) bool {
		switch sortOrder {
		case "price-asc":
			if offers[i].Price != offers[j].Price {
				return offers[i].Price < offers[j].Price
			}
			return offers[i].ID < offers[j].ID
		case "price-desc":
			if offers[i].Price != offers[j].Price {
				return offers[i].Price > offers[j].Price
			}
			return offers[i].ID < offers[j].ID
		default:
			if offers[i].Price != offers[j].Price {
				return offers[i].Price < offers[j].Price
			}
			return offers[i].ID < offers[j].ID
		}
	})
}

func calculatePagination(total int, page, pageSize uint32) (int, int) {
	start := int((page) * pageSize)
	if start >= total {
		return 0, 0
	}
	end := int(math.Min(float64(start+int(pageSize)), float64(total)))
	return start, end
}

func generatePriceRanges(offers []models.Offer, params SearchParams) []models.PriceRange {
	if len(offers) == 0 {
		return []models.PriceRange{}
	}

	var filteredOffers []models.Offer
	for _, offer := range offers {
		if params.CarType != "" && offer.CarType != params.CarType {
			continue
		}

		if params.MinNumberSeats > 0 && offer.NumberSeats < params.MinNumberSeats {
			continue
		}

		if params.MinFreeKilometer > 0 && offer.FreeKilometers < params.MinFreeKilometer {
			continue
		}

		if params.OnlyVollkasko != nil && *params.OnlyVollkasko && !offer.HasVollkasko {
			continue
		}

		filteredOffers = append(filteredOffers, offer)
	}

	if len(filteredOffers) == 0 {
		return []models.PriceRange{}
	}

	// Find min and max prices
	minPrice := filteredOffers[0].Price
	maxPrice := filteredOffers[0].Price
	for _, offer := range filteredOffers {
		if offer.Price < minPrice {
			minPrice = offer.Price
		}
		if offer.Price > maxPrice {
			maxPrice = offer.Price
		}
	}

	// Generate ranges
	ranges := make([]models.PriceRange, 0)
	// Round minPrice down and maxPrice up to nearest multiple of width
	startPrice := uint16(params.PriceRangeWidth) * (minPrice / uint16(params.PriceRangeWidth))
	endPrice := uint16(params.PriceRangeWidth) * ((maxPrice + uint16(params.PriceRangeWidth) - 1) / uint16(params.PriceRangeWidth))

	for start := startPrice; start <= endPrice; start += uint16(params.PriceRangeWidth) {
		end := start + uint16(params.PriceRangeWidth) - 1
		count := 0
		for _, offer := range filteredOffers {
			if offer.Price >= start && offer.Price <= end {
				count++
			}
		}
		if count > 0 {
			ranges = append(ranges, models.PriceRange{
				Start: start,
				End:   end + 1,
				Count: uint32(count),
			})
		}
	}
	return ranges
}
