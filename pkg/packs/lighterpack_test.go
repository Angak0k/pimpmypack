package packs

import (
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestValidateLighterPackURL(t *testing.T) {
	tests := []struct {
		name    string
		url     string
		wantErr bool
	}{
		{"valid URL", "https://lighterpack.com/r/oo18ii", false},
		{"valid URL alphanumeric", "https://lighterpack.com/r/ABC123xyz", false},
		{"missing scheme", "lighterpack.com/r/oo18ii", true},
		{"HTTP not HTTPS", "http://lighterpack.com/r/oo18ii", true},
		{"wrong domain", "https://evil.com/r/oo18ii", true},
		{"extra path", "https://lighterpack.com/r/oo18ii/extra", true},
		{"query params", "https://lighterpack.com/r/oo18ii?foo=bar", true},
		{"empty ID", "https://lighterpack.com/r/", true},
		{"special chars in ID", "https://lighterpack.com/r/../../etc", true},
		{"empty string", "", true},
		{"user profile URL", "https://lighterpack.com/u/someone", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := validateLighterPackURL(tt.url)
			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestParseLighterPackHTML(t *testing.T) {
	data, err := os.ReadFile("testdata/lighterpack_sample.html")
	require.NoError(t, err)

	packName, packDescription, items, err := parseLighterPackHTML(data)
	require.NoError(t, err)

	assert.Equal(t, "My Thru-Hike Gear", packName)
	assert.Equal(t, "PCT 2026 gear list", packDescription)
	assert.Len(t, items, 6)

	// First item: Zpacks Duplex
	assert.Equal(t, "Zpacks Duplex", items[0].ItemName)
	assert.Equal(t, "Shelter", items[0].Category)
	assert.Equal(t, "2-person tent", items[0].Desc)
	assert.Equal(t, 539, items[0].Weight) // 539000mg -> 539g
	assert.Equal(t, 59900, items[0].Price)
	assert.Equal(t, "USD", items[0].Currency)
	assert.Equal(t, "https://example.com/tent", items[0].URL)
	assert.Equal(t, 1, items[0].Qty)
	assert.False(t, items[0].Worn)
	assert.False(t, items[0].Consumable)

	// Second item: Tent Stakes (no URL, qty 6)
	assert.Equal(t, "Tent Stakes (x6)", items[1].ItemName)
	assert.Empty(t, items[1].URL)
	assert.Equal(t, 6, items[1].Qty)
	assert.Equal(t, 54, items[1].Weight)

	// Third item: quilt with URL
	assert.Equal(t, "Enlightened Equipment Enigma 20", items[2].ItemName)
	assert.Equal(t, "Sleep System", items[2].Category)
	assert.Equal(t, "https://example.com/quilt", items[2].URL)

	// Fourth item: Sleeping Pad
	assert.Equal(t, "Sleeping Pad", items[3].ItemName)
	assert.Equal(t, 19995, items[3].Price)

	// Fifth item: Trail Runners (Worn)
	assert.Equal(t, "Trail Runners", items[4].ItemName)
	assert.Equal(t, "Worn Clothing", items[4].Category)
	assert.True(t, items[4].Worn)
	assert.False(t, items[4].Consumable)

	// Sixth item: Water Filter (Consumable)
	assert.Equal(t, "Water Filter", items[5].ItemName)
	assert.Equal(t, "Food & Water", items[5].Category)
	assert.False(t, items[5].Worn)
	assert.True(t, items[5].Consumable)
}

func TestParseLighterPackHTML_EUR(t *testing.T) {
	data, err := os.ReadFile("testdata/lighterpack_eur.html")
	require.NoError(t, err)

	packName, packDescription, items, err := parseLighterPackHTML(data)
	require.NoError(t, err)

	assert.Equal(t, "Mon sac GR20", packName)
	assert.Equal(t, "Liste pour le GR20", packDescription)
	assert.Len(t, items, 1)

	assert.Equal(t, "Tente", items[0].ItemName)
	assert.Equal(t, 25000, items[0].Price)
	assert.Equal(t, "EUR", items[0].Currency)
	assert.Equal(t, 1200, items[0].Weight) // 1200000mg -> 1200g
}

func TestParseLighterPackHTML_Empty(t *testing.T) {
	data := []byte(`<html><body><div id="lpApp"></div></body></html>`)
	_, _, _, err := parseLighterPackHTML(data)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "no items found")
}

func TestParseLighterPackHTML_MissingName(t *testing.T) {
	data := []byte(`<html><body>
		<ul class="lpCategories">
		<li class="lpCategory">
			<h2 class="lpCategoryName">Gear</h2>
			<ul class="lpItems">
				<li class="lpItem">
					<span class="lpName">Item One</span>
					<span class="lpDescription">Desc</span>
					<input class="lpMG" type="hidden" value="100000"/>
					<span class="lpPriceCell lpNumber">$10.00</span>
					<span class="lpQtyCell">1</span>
					<i class="lpWorn"></i>
					<i class="lpConsumable"></i>
				</li>
			</ul>
		</li>
		</ul>
	</body></html>`)

	packName, packDescription, items, err := parseLighterPackHTML(data)
	require.NoError(t, err)

	assert.Equal(t, "LighterPack Import", packName)
	assert.Equal(t, "LighterPack Import", packDescription) // fallback to packName
	assert.Len(t, items, 1)
}

func TestParsePrice(t *testing.T) {
	tests := []struct {
		input    string
		price    int
		currency string
	}{
		{"$399.50", 39950, "USD"},
		{"€100.00", 10000, "EUR"},
		{"£50.00", 5000, "GBP"},
		{"$0.00", 0, "USD"},
		{"", 0, ""},
		{"199.99", 19999, ""},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			price, currency := parsePrice(tt.input)
			assert.Equal(t, tt.price, price)
			assert.Equal(t, tt.currency, currency)
		})
	}
}
