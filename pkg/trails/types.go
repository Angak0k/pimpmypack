package trails

import (
	"errors"
	"time"
)

// Domain errors
var (
	ErrTrailNotFound   = errors.New("trail not found")
	ErrTrailNameExists = errors.New("trail name already exists")
	ErrTrailInUse      = errors.New("trail is in use by one or more packs")
)

// Trail represents a trail with its metadata
type Trail struct {
	ID         uint      `json:"id"`
	Name       string    `json:"name"`
	Country    string    `json:"country"`
	Continent  string    `json:"continent"`
	DistanceKm *int      `json:"distance_km,omitempty"`
	URL        *string   `json:"url,omitempty"`
	CreatedAt  time.Time `json:"created_at"`
	UpdatedAt  time.Time `json:"updated_at"`
}

// Trails represents a collection of trails
type Trails []Trail

// TrailCreateRequest represents the input for creating a trail
type TrailCreateRequest struct {
	Name       string  `json:"name" binding:"required"`
	Country    string  `json:"country" binding:"required"`
	Continent  string  `json:"continent" binding:"required"`
	DistanceKm *int    `json:"distance_km"`
	URL        *string `json:"url"`
}

// TrailUpdateRequest represents the input for updating a trail
type TrailUpdateRequest struct {
	Name       string  `json:"name" binding:"required"`
	Country    string  `json:"country" binding:"required"`
	Continent  string  `json:"continent" binding:"required"`
	DistanceKm *int    `json:"distance_km"`
	URL        *string `json:"url"`
}

// TrailBulkCreateRequest represents the input for bulk creating trails
type TrailBulkCreateRequest struct {
	Trails []TrailCreateRequest `json:"trails" binding:"required,min=1"`
}

// TrailBulkDeleteRequest represents the input for bulk deleting trails
type TrailBulkDeleteRequest struct {
	IDs []uint `json:"ids" binding:"required,min=1"`
}

// TrailSummary is a lightweight trail representation for options endpoints
type TrailSummary struct {
	ID         uint    `json:"id"`
	Name       string  `json:"name"`
	DistanceKm *int    `json:"distance_km,omitempty"`
	URL        *string `json:"url,omitempty"`
}

// GroupedResponse represents trails grouped by continent and country
type GroupedResponse struct {
	Continents map[string]map[string][]TrailSummary `json:"continents"`
}
