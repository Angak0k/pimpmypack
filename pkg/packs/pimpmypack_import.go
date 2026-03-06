package packs

import (
	"errors"
	"net/url"
	"regexp"
)

var pimpMyPackPathPattern = regexp.MustCompile(
	`^/(?:sharedlist|api/sharedlist)/([BILMPTaceghikmnprsty]{30})$`,
)

var pimpMyPackCodePattern = regexp.MustCompile(
	`^[BILMPTaceghikmnprsty]{30}$`,
)

// validatePimpMyPackURL checks that the URL is a valid PimpMyPack sharing URL
// and returns the sharing code.
// Supported formats:
//   - https://app.alki.earth/sharedlist/<code>
//   - https://app.alki.earth/api/sharedlist/<code>
//   - https://app.alki.earth/public-pack.html?code=<code>
func validatePimpMyPackURL(rawURL string) (string, error) {
	parsed, err := url.Parse(rawURL)
	if err != nil || parsed.Scheme != "https" || parsed.Host != "app.alki.earth" {
		return "", errors.New("invalid PimpMyPack URL: must be from https://app.alki.earth")
	}

	// Try path-based format: /sharedlist/<code> or /api/sharedlist/<code>
	if matches := pimpMyPackPathPattern.FindStringSubmatch(parsed.Path); matches != nil {
		return matches[1], nil
	}

	// Try query-based format: /public-pack.html?code=<code>
	if parsed.Path == "/public-pack.html" {
		code := parsed.Query().Get("code")
		if pimpMyPackCodePattern.MatchString(code) {
			return code, nil
		}
	}

	return "", errors.New("invalid PimpMyPack URL: must match https://app.alki.earth/sharedlist/<code> or https://app.alki.earth/public-pack.html?code=<code>")
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
