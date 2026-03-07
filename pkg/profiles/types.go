package profiles

import "time"

// PublicProfile represents the public-facing user profile
type PublicProfile struct {
	Username        string              `json:"username"`
	Firstname       string              `json:"firstname"`
	HasProfileImage bool                `json:"has_profile_image"`
	ImagePositionX  int                 `json:"image_position_x"`
	ImagePositionY  int                 `json:"image_position_y"`
	YoutubeURL      *string             `json:"youtube_url,omitempty"`
	InstagramURL    *string             `json:"instagram_url,omitempty"`
	SharedPacks     []SharedPackSummary `json:"shared_packs"`
}

// SharedPackSummary represents a shared pack in a public profile
type SharedPackSummary struct {
	ID              uint      `json:"id"`
	PackName        string    `json:"pack_name"`
	PackDescription string    `json:"pack_description"`
	HasImage        bool      `json:"has_image"`
	PackWeight      int       `json:"pack_weight"`
	PackItemsCount  int       `json:"pack_items_count"`
	SharingCode     string    `json:"sharing_code"`
	Season          *string   `json:"season,omitempty"`
	Trail           *string   `json:"trail,omitempty"`
	Adventure       *string   `json:"adventure,omitempty"`
	CreatedAt       time.Time `json:"created_at"`
}
