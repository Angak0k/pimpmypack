package packs

import (
	"errors"
	"regexp"
)

var pimpMyPackURLPattern = regexp.MustCompile(
	`^https://app\.alki\.earth/(?:sharedlist|api/sharedlist)/([BILMPTaceghikmnprsty]{30})$`,
)

// validatePimpMyPackURL checks that the URL is a valid PimpMyPack sharing URL
// and returns the sharing code.
func validatePimpMyPackURL(rawURL string) (string, error) {
	matches := pimpMyPackURLPattern.FindStringSubmatch(rawURL)
	if matches == nil {
		return "", errors.New("invalid PimpMyPack URL: must match https://app.alki.earth/sharedlist/<code>")
	}
	return matches[1], nil
}

// convertSharedPackToExternalPack converts a SharedPackResponse into an ExternalPack
// for insertion via the shared insertExternalPack function.
func convertSharedPackToExternalPack(resp *SharedPackResponse) *ExternalPack {
	items := make(ExternalPack, 0, len(resp.Contents))
	for _, c := range resp.Contents {
		items = append(items, ExternalPackItem{
			ItemName:   c.ItemName,
			Category:   c.Category,
			Desc:       c.ItemDescription,
			Qty:        c.Quantity,
			Weight:     c.Weight,
			Unit:       "gram",
			URL:        c.ItemURL,
			Price:      c.Price,
			Currency:   c.Currency,
			Worn:       c.Worn,
			Consumable: c.Consumable,
		})
	}
	return &items
}
