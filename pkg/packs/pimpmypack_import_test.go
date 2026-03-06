package packs

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestValidatePimpMyPackURL(t *testing.T) {
	tests := []struct {
		name     string
		url      string
		wantCode string
		wantErr  bool
	}{
		{
			name:     "valid sharedlist URL",
			url:      "https://app.alki.earth/sharedlist/pimpMyPackIsBetterThanLighte",
			wantCode: "pimpMyPackIsBetterThanLighte",
			wantErr:  true, // only 28 chars
		},
		{
			name:     "valid 30-char code",
			url:      "https://app.alki.earth/sharedlist/pimpMyPackIsBetterThanLight",
			wantCode: "pimpMyPackIsBetterThanLight",
			wantErr:  true, // only 26 chars
		},
		{
			name:     "wrong domain",
			url:      "https://evil.com/sharedlist/pimpMyPackIsBetterThanLighterPac",
			wantCode: "",
			wantErr:  true,
		},
		{
			name:     "HTTP not HTTPS",
			url:      "http://app.alki.earth/sharedlist/pimpMyPackIsBetterThanLighterPac",
			wantCode: "",
			wantErr:  true,
		},
		{
			name:     "wrong path",
			url:      "https://app.alki.earth/other/pimpMyPackIsBetterThanLighterPac",
			wantCode: "",
			wantErr:  true,
		},
		{
			name:     "empty string",
			url:      "",
			wantCode: "",
			wantErr:  true,
		},
		{
			name:     "extra path segment",
			url:      "https://app.alki.earth/sharedlist/pimpMyPackIsBetterThanLighterPac/extra",
			wantCode: "",
			wantErr:  true,
		},
		{
			name:     "invalid characters in code",
			url:      "https://app.alki.earth/sharedlist/123456789012345678901234567890",
			wantCode: "",
			wantErr:  true,
		},
	}

	// Valid 30-char code from the charset (unique chars: BILMPTaceghikmnprsty)
	validCode := "pimpMyPackIsBetterThanLighterP"
	validURL := "https://app.alki.earth/sharedlist/" + validCode
	tests = append(tests, struct {
		name     string
		url      string
		wantCode string
		wantErr  bool
	}{
		name:     "valid sharedlist URL with 30-char code",
		url:      validURL,
		wantCode: validCode,
		wantErr:  false,
	})

	// API variant
	apiURL := "https://app.alki.earth/api/sharedlist/" + validCode
	tests = append(tests, struct {
		name     string
		url      string
		wantCode string
		wantErr  bool
	}{
		name:     "valid API sharedlist URL",
		url:      apiURL,
		wantCode: validCode,
		wantErr:  false,
	})

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			code, err := validatePimpMyPackURL(tt.url)
			if tt.wantErr {
				require.Error(t, err)
				assert.Empty(t, code)
			} else {
				require.NoError(t, err)
				assert.Equal(t, tt.wantCode, code)
			}
		})
	}
}

func TestConvertSharedPackToExternalPack(t *testing.T) {
	resp := &SharedPackResponse{
		Pack: SharedPackInfo{
			ID:              1,
			PackName:        "Test Pack",
			PackDescription: "A test pack",
		},
		Contents: PackContentWithItems{
			{
				ItemName:        "Tent",
				Category:        "Shelter",
				ItemDescription: "Lightweight tent",
				Quantity:        1,
				Weight:          1200,
				ItemURL:         "https://example.com/tent",
				Price:           35000,
				Currency:        "EUR",
				Worn:            false,
				Consumable:      false,
			},
			{
				ItemName:        "Rain Jacket",
				Category:        "Clothing",
				ItemDescription: "Waterproof",
				Quantity:        1,
				Weight:          250,
				ItemURL:         "",
				Price:           15000,
				Currency:        "EUR",
				Worn:            true,
				Consumable:      false,
			},
			{
				ItemName:        "Trail Mix",
				Category:        "Food",
				ItemDescription: "Snack",
				Quantity:        3,
				Weight:          100,
				ItemURL:         "",
				Price:           500,
				Currency:        "USD",
				Worn:            false,
				Consumable:      true,
			},
		},
	}

	result := convertSharedPackToExternalPack(resp)
	require.NotNil(t, result)
	require.Len(t, *result, 3)

	items := *result

	// First item
	assert.Equal(t, "Tent", items[0].ItemName)
	assert.Equal(t, "Shelter", items[0].Category)
	assert.Equal(t, "Lightweight tent", items[0].Desc)
	assert.Equal(t, 1, items[0].Qty)
	assert.Equal(t, 1200, items[0].Weight)
	assert.Equal(t, "gram", items[0].Unit)
	assert.Equal(t, "https://example.com/tent", items[0].URL)
	assert.Equal(t, 35000, items[0].Price)
	assert.Equal(t, "EUR", items[0].Currency)
	assert.False(t, items[0].Worn)
	assert.False(t, items[0].Consumable)

	// Second item - worn
	assert.Equal(t, "Rain Jacket", items[1].ItemName)
	assert.True(t, items[1].Worn)
	assert.False(t, items[1].Consumable)
	assert.Equal(t, "gram", items[1].Unit)

	// Third item - consumable
	assert.Equal(t, "Trail Mix", items[2].ItemName)
	assert.Equal(t, 3, items[2].Qty)
	assert.False(t, items[2].Worn)
	assert.True(t, items[2].Consumable)
	assert.Equal(t, "USD", items[2].Currency)
}

func TestConvertSharedPackToExternalPack_Empty(t *testing.T) {
	resp := &SharedPackResponse{
		Pack:     SharedPackInfo{PackName: "Empty Pack"},
		Contents: PackContentWithItems{},
	}

	result := convertSharedPackToExternalPack(resp)
	require.NotNil(t, result)
	assert.Empty(t, *result)
}
