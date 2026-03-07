package profiles

import (
	"context"
	"database/sql"
	"errors"
	"fmt"

	"github.com/Angak0k/pimpmypack/pkg/database"
)

// ErrProfileNotFound is returned when a public profile is not found
// (user doesn't exist, is not active, or has not enabled public profile)
var ErrProfileNotFound = errors.New("public profile not found")

// returnPublicProfileByUsername fetches the public profile for a user.
// Returns ErrProfileNotFound if the user doesn't exist, is not active,
// or has is_profile_public = false (anti-enumeration: same error for all cases).
func returnPublicProfileByUsername(ctx context.Context, username string) (*PublicProfile, error) {
	var profile PublicProfile

	err := database.DB().QueryRowContext(ctx,
		`SELECT a.username, a.firstname, a.youtube_url, a.instagram_url,
		    CASE WHEN ai.account_id IS NOT NULL THEN true ELSE false END AS has_profile_image,
		    a.image_position_x, a.image_position_y
		FROM account a
		LEFT JOIN account_images ai ON a.id = ai.account_id
		WHERE a.username = $1 AND a.status = 'active' AND a.is_profile_public = true;`,
		username,
	).Scan(
		&profile.Username,
		&profile.Firstname,
		&profile.YoutubeURL,
		&profile.InstagramURL,
		&profile.HasProfileImage,
		&profile.ImagePositionX,
		&profile.ImagePositionY,
	)

	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, ErrProfileNotFound
		}
		return nil, fmt.Errorf("failed to fetch public profile: %w", err)
	}

	return &profile, nil
}

// returnSharedPacksByUsername fetches all shared packs for a given username.
// Only returns packs with a non-NULL sharing_code.
func returnSharedPacksByUsername(ctx context.Context, username string) ([]SharedPackSummary, error) {
	rows, err := database.DB().QueryContext(ctx, `
		SELECT p.id, p.pack_name, p.pack_description,
		    CASE WHEN pi.pack_id IS NOT NULL THEN true ELSE false END AS has_image,
		    COALESCE(SUM(i.weight * pc.quantity), 0) AS pack_weight,
		    COALESCE(SUM(pc.quantity), 0) AS pack_items_count,
		    p.sharing_code, p.season, p.trail, p.adventure, p.created_at
		FROM pack p
		JOIN account a ON p.user_id = a.id
		LEFT JOIN pack_content pc ON p.id = pc.pack_id
		LEFT JOIN inventory i ON pc.item_id = i.id
		LEFT JOIN pack_images pi ON p.id = pi.pack_id
		WHERE a.username = $1
		  AND a.status = 'active'
		  AND a.is_profile_public = true
		  AND p.sharing_code IS NOT NULL
		GROUP BY p.id, p.pack_name, p.pack_description, p.sharing_code,
		         pi.pack_id, p.season, p.trail, p.adventure, p.created_at
		ORDER BY p.created_at DESC;`,
		username,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch shared packs: %w", err)
	}
	defer rows.Close()

	var packs []SharedPackSummary
	for rows.Next() {
		var pack SharedPackSummary
		err := rows.Scan(
			&pack.ID,
			&pack.PackName,
			&pack.PackDescription,
			&pack.HasImage,
			&pack.PackWeight,
			&pack.PackItemsCount,
			&pack.SharingCode,
			&pack.Season,
			&pack.Trail,
			&pack.Adventure,
			&pack.CreatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan shared pack: %w", err)
		}
		packs = append(packs, pack)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("rows iteration error: %w", err)
	}

	// Return empty slice instead of nil for consistent JSON output
	if packs == nil {
		packs = []SharedPackSummary{}
	}

	return packs, nil
}
