package trails

import (
	"context"
	"errors"
)

// FindTrailByID finds a trail by its ID.
// This is a public service function used by other packages.
func FindTrailByID(ctx context.Context, id uint) (*Trail, error) {
	return findTrailByID(ctx, id)
}

// FindTrailByName finds a trail by its name.
// This is a public service function used by other packages (e.g., packs for V1 resolution).
func FindTrailByName(ctx context.Context, name string) (*Trail, error) {
	return findTrailByName(ctx, name)
}

// ReturnTrailNames returns a flat list of trail names for V1 pack-options.
func ReturnTrailNames(ctx context.Context) ([]string, error) {
	return returnTrailNames(ctx)
}

// ReturnTrailsGrouped returns trails grouped by continent and country for V2 pack-options.
func ReturnTrailsGrouped(ctx context.Context) (map[string]map[string][]TrailSummary, error) {
	return returnTrailsGrouped(ctx)
}

// IsValidTrailName checks if a trail name exists in the database.
// Returns true if name is nil (optional field) or found in the database.
func IsValidTrailName(ctx context.Context, name *string) (bool, error) {
	if name == nil {
		return true, nil
	}
	_, err := findTrailByName(ctx, *name)
	if err != nil {
		if errors.Is(err, ErrTrailNotFound) {
			return false, nil
		}
		return false, err
	}
	return true, nil
}
