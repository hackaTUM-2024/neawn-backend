package models

type CarType string

type Offer struct {
	ID                   string `json:"ID"`
	Data                 string `json:"data"`
	MostSpecificRegionID int32  `json:"mostSpecificRegionID"`
	StartDate            int64  `json:"startDate"`
	EndDate              int64  `json:"endDate"`
	NumberSeats          uint8  `json:"numberSeats"`
	Price                uint16 `json:"price"`
	CarType              string `json:"carType"`
	HasVollkasko         bool   `json:"hasVollkasko"`
	FreeKilometers       uint16 `json:"freeKilometers"`
}

type SearchResultOffer struct {
	ID   string `json:"ID"`
	Data string `json:"data"`
}

type PriceRange struct {
	Start uint16 `json:"start"`
	End   uint16 `json:"end"`
	Count uint32 `json:"count"`
}

type CarTypeCount struct {
	Small  uint32 `json:"small"`
	Sports uint32 `json:"sports"`
	Luxury uint32 `json:"luxury"`
	Family uint32 `json:"family"`
}

type VollkaskoCount struct {
	TrueCount  uint32 `json:"trueCount"`
	FalseCount uint32 `json:"falseCount"`
}

type SeatsCount struct {
	NumberSeats uint8  `json:"numberSeats"`
	Count       uint32 `json:"count"`
}

type FreeKilometerRange struct {
	Start uint16 `json:"start"`
	End   uint16 `json:"end"`
	Count uint32 `json:"count"`
}

type SearchResponse struct {
	Offers             []SearchResultOffer  `json:"offers"`
	PriceRanges        []PriceRange         `json:"priceRanges"`
	CarTypeCounts      CarTypeCount         `json:"carTypeCounts"`
	SeatsCount         []SeatsCount         `json:"seatsCount"`
	FreeKilometerRange []FreeKilometerRange `json:"freeKilometerRange"`
	VollkaskoCount     VollkaskoCount       `json:"vollkaskoCount"`
}

type CreateOffersRequest struct {
	Offers []Offer `json:"offers"`
}
