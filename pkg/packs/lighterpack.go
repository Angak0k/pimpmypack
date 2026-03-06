package packs

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"html"
	"io"
	"math"
	"net/http"
	"regexp"
	"strconv"
	"strings"
	"time"

	xhtml "golang.org/x/net/html"
)

var lighterPackURLPattern = regexp.MustCompile(`^https://lighterpack\.com/r/[a-zA-Z0-9]+$`)

// validateLighterPackURL checks that the URL is a valid LighterPack sharing URL.
func validateLighterPackURL(rawURL string) error {
	if !lighterPackURLPattern.MatchString(rawURL) {
		return errors.New("invalid LighterPack URL: must match https://lighterpack.com/r/<id>")
	}
	return nil
}

// fetchLighterPackPage fetches the HTML content from a LighterPack sharing URL.
func fetchLighterPackPage(ctx context.Context, rawURL string) ([]byte, error) {
	client := &http.Client{
		Timeout: 15 * time.Second,
		CheckRedirect: func(req *http.Request, via []*http.Request) error {
			// Only allow HTTPS redirects within lighterpack.com
			if req.URL.Scheme != "https" || req.URL.Host != "lighterpack.com" {
				return errors.New("redirect to external or insecure URL blocked")
			}
			if len(via) >= 3 {
				return errors.New("too many redirects")
			}
			return nil
		},
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, rawURL, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch LighterPack page: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("LighterPack returned status %d", resp.StatusCode)
	}

	// Limit response size to 5MB
	body, err := io.ReadAll(io.LimitReader(resp.Body, 5*1024*1024))
	if err != nil {
		return nil, fmt.Errorf("failed to read LighterPack page: %w", err)
	}

	return body, nil
}

// parseLighterPackHTML parses a LighterPack HTML page and extracts pack data.
func parseLighterPackHTML(data []byte) (string, string, LighterPack, error) {
	doc, err := xhtml.Parse(bytes.NewReader(data))
	if err != nil {
		return "", "", nil, fmt.Errorf("failed to parse HTML: %w", err)
	}

	// Extract pack name from h1.lpListName
	packName := "LighterPack Import"
	nameNodes := findNodesByClass(doc, "h1", "lpListName")
	if len(nameNodes) > 0 {
		name := strings.TrimSpace(textContent(nameNodes[0]))
		if name != "" {
			packName = name
		}
	}

	// Extract pack description from span.lpListDescription
	packDescription := packName
	descNodes := findNodesByClass(doc, "span", "lpListDescription")
	if len(descNodes) > 0 {
		desc := strings.TrimSpace(textContent(descNodes[0]))
		if desc != "" {
			packDescription = desc
		}
	}

	// Extract items by category
	var items LighterPack
	categoryNodes := findNodesByClass(doc, "li", "lpCategory")

	for _, catNode := range categoryNodes {
		// Get category name from h2.lpCategoryName
		categoryName := ""
		catNameNodes := findNodesByClass(catNode, "h2", "lpCategoryName")
		if len(catNameNodes) > 0 {
			categoryName = strings.TrimSpace(textContent(catNameNodes[0]))
		}

		// Get items within this category
		itemNodes := findNodesByClass(catNode, "li", "lpItem")
		for _, itemNode := range itemNodes {
			item := parseLighterPackItem(itemNode, categoryName)
			items = append(items, item)
		}
	}

	if len(items) == 0 {
		return "", "", nil, errors.New("no items found in LighterPack page")
	}

	return packName, packDescription, items, nil
}

// parseLighterPackItem extracts a single item from an li.lpItem node.
func parseLighterPackItem(node *xhtml.Node, category string) LighterPackItem {
	item := LighterPackItem{
		Category: category,
		Qty:      1,
		Currency: "EUR", // default
	}

	extractItemName(node, &item)
	extractItemDescription(node, &item)
	extractItemWeight(node, &item)
	extractItemPrice(node, &item)
	extractItemQuantity(node, &item)
	item.Worn = hasActiveFlag(node, "lpWorn")
	item.Consumable = hasActiveFlag(node, "lpConsumable")

	return item
}

func extractItemName(node *xhtml.Node, item *LighterPackItem) {
	nameNodes := findNodesByClass(node, "span", "lpName")
	if len(nameNodes) > 0 {
		item.URL = extractItemURL(nameNodes[0])
		item.ItemName = strings.TrimSpace(textContent(nameNodes[0]))
	}
}

func extractItemDescription(node *xhtml.Node, item *LighterPackItem) {
	descNodes := findNodesByClass(node, "span", "lpDescription")
	if len(descNodes) > 0 {
		item.Desc = strings.TrimSpace(textContent(descNodes[0]))
	}
}

func extractItemWeight(node *xhtml.Node, item *LighterPackItem) {
	mgNodes := findNodesByClass(node, "input", "lpMG")
	if len(mgNodes) > 0 {
		valStr := getAttr(mgNodes[0], "value")
		if mg, err := strconv.ParseFloat(valStr, 64); err == nil {
			item.Weight = int(math.Round(mg / 1000)) // milligrams to grams
		}
	}
}

func extractItemPrice(node *xhtml.Node, item *LighterPackItem) {
	priceNodes := findNodesByClass(node, "span", "lpPriceCell")
	for _, pn := range priceNodes {
		if hasClass(pn, "lpNumber") {
			priceText := strings.TrimSpace(textContent(pn))
			price, currency := parsePrice(priceText)
			item.Price = price
			if currency != "" {
				item.Currency = currency
			}
			break
		}
	}
}

func extractItemQuantity(node *xhtml.Node, item *LighterPackItem) {
	qtyNodes := findNodesByClass(node, "span", "lpQtyCell")
	if len(qtyNodes) > 0 {
		qtyText := strings.TrimSpace(textContent(qtyNodes[0]))
		if q, err := strconv.Atoi(qtyText); err == nil && q > 0 {
			item.Qty = q
		}
	}
}

func hasActiveFlag(node *xhtml.Node, className string) bool {
	nodes := findNodesByClass(node, "i", className)
	return len(nodes) > 0 && hasClass(nodes[0], "lpActive")
}

// extractItemURL looks for an a.lpHref inside a node and returns the decoded href.
func extractItemURL(node *xhtml.Node) string {
	links := findNodesByClass(node, "a", "lpHref")
	if len(links) == 0 {
		return ""
	}
	href := getAttr(links[0], "href")
	return html.UnescapeString(href)
}

// parsePrice parses a price string like "$399.50" or "€100.00" and returns
// the price in cents and the detected currency code.
func parsePrice(s string) (int, string) {
	if s == "" {
		return 0, ""
	}

	currency := ""
	cleaned := s

	// Detect currency from prefix
	switch {
	case strings.HasPrefix(s, "$"):
		currency = "USD"
		cleaned = s[1:]
	case strings.HasPrefix(s, "€"):
		currency = "EUR"
		cleaned = s[len("€"):]
	case strings.HasPrefix(s, "£"):
		currency = "GBP"
		cleaned = s[len("£"):]
	}

	cleaned = strings.TrimSpace(cleaned)
	f, err := strconv.ParseFloat(cleaned, 64)
	if err != nil {
		return 0, currency
	}

	return int(math.Round(f * 100)), currency
}

// HTML helper functions

// hasClass checks if a node has the specified CSS class.
func hasClass(node *xhtml.Node, className string) bool {
	if node == nil {
		return false
	}
	classAttr := getAttr(node, "class")
	for _, c := range strings.Fields(classAttr) {
		if c == className {
			return true
		}
	}
	return false
}

// getAttr returns the value of a named attribute on a node.
func getAttr(node *xhtml.Node, attrName string) string {
	if node == nil {
		return ""
	}
	for _, attr := range node.Attr {
		if attr.Key == attrName {
			return attr.Val
		}
	}
	return ""
}

// textContent recursively collects text content from a node and its children.
func textContent(node *xhtml.Node) string {
	if node == nil {
		return ""
	}
	if node.Type == xhtml.TextNode {
		return node.Data
	}
	var sb strings.Builder
	for child := node.FirstChild; child != nil; child = child.NextSibling {
		sb.WriteString(textContent(child))
	}
	return sb.String()
}

// findNodesByClass finds all descendant elements matching the given tag and class.
func findNodesByClass(node *xhtml.Node, tag, className string) []*xhtml.Node {
	var results []*xhtml.Node
	if node == nil {
		return results
	}

	var walk func(*xhtml.Node)
	walk = func(n *xhtml.Node) {
		if n.Type == xhtml.ElementNode && n.Data == tag && hasClass(n, className) {
			results = append(results, n)
			// Don't recurse into matched nodes to avoid nested matches
			return
		}
		for child := n.FirstChild; child != nil; child = child.NextSibling {
			walk(child)
		}
	}
	walk(node)
	return results
}
