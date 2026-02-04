package packs

import "context"

// FindPackByID finds a pack by its ID
// This is a public service function used by other packages (e.g., images)
func FindPackByID(ctx context.Context, id uint) (*Pack, error) {
	return findPackByID(ctx, id)
}

// CheckPackOwnership verifies if a user owns a specific pack
// This is a public service function used by other packages (e.g., images)
func CheckPackOwnership(ctx context.Context, packID uint, userID uint) (bool, error) {
	return checkPackOwnership(ctx, packID, userID)
}

// FindPackIDByPackName finds a pack ID by its name
// Returns 0 if not found
func FindPackIDByPackName(packs Packs, packname string) uint {
	for _, pack := range packs {
		if pack.PackName == packname {
			return pack.ID
		}
	}
	return 0
}
