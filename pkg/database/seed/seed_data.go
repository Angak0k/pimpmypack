package seed

// seedUser holds the seed account data.
type seedUser struct {
	username  string
	password  string
	email     string
	firstname string
	lastname  string
	role      string
	status    string
}

// seedItem represents an inventory item for seeding.
type seedItem struct {
	itemName    string
	category    string
	description string
	weight      int
	url         string
	price       int
	currency    string
}

// seedPackDef represents a pack definition for seeding.
type seedPackDef struct {
	packName        string
	packDescription string
	season          *string
	trail           *string
	adventure       *string
	isFavorite      bool
}

// seedPackContentDef links items to packs by composite key.
type seedPackContentDef struct {
	packIndex   int // index into seedPackDefs
	itemName    string
	category    string
	description string
	quantity    int
	worn        bool
	consumable  bool
}

func strPtr(s string) *string { return &s }

// Shorthand constructors to keep data declarations under 120 chars.
func item(name, cat, desc string, w int) seedItem {
	return seedItem{
		itemName: name, category: cat,
		description: desc, weight: w, currency: "EUR",
	}
}

func pc(
	idx int, name, cat, desc string, qty int,
) seedPackContentDef {
	return seedPackContentDef{
		packIndex: idx, itemName: name,
		category: cat, description: desc, quantity: qty,
	}
}

func pcWorn(idx int, name, desc string) seedPackContentDef {
	return seedPackContentDef{
		packIndex: idx, itemName: name,
		category: "Worn", description: desc,
		quantity: 1, worn: true,
	}
}

func pcCons(
	idx int, name, cat, desc string, qty int,
) seedPackContentDef {
	return seedPackContentDef{
		packIndex: idx, itemName: name,
		category: cat, description: desc,
		quantity: qty, consumable: true,
	}
}

var defaultUser = seedUser{
	username:  "Romain",
	password:  "romain1234",
	email:     "john.doe@angakok.net",
	firstname: "John",
	lastname:  "Doe",
	role:      "standard",
	status:    "active",
}

// packGR10 and packGR54 are indexes into seedPackDefs.
const (
	packGR10 = 0
	packGR54 = 1
)

var seedPackDefs = []seedPackDef{
	{
		packName:        "GR10 - Saison 5",
		packDescription: "GR10 Thru-hike gear list",
		season:          strPtr("3-Season"),
		trail:           strPtr("GR10"),
		adventure:       strPtr("Thru-hike"),
		isFavorite:      true,
	},
	{
		packName:        "GR54",
		packDescription: "GR54 Backpacking gear list",
		season:          strPtr("3-Season"),
		trail:           strPtr("GR54"),
		adventure:       strPtr("Backpacking"),
		isFavorite:      false,
	},
}

// seedInventoryItems is the deduplicated list from both CSVs.
// Categories normalized: Wear->Clothes, Healthpack->Healthcare,
// Food->Food & Drinks, MISC->Various.
var seedInventoryItems = []seedItem{
	// Backpack
	item("ALD Hybrid 30", "Backpack", "Backpack", 525),
	item("Sakasec", "Backpack", "Liner", 33),

	// Shelter
	item("X-MID-PRO 1", "Shelter", "Tent", 480),
	item("Stake", "Shelter", "Stake", 9),
	item("Polycree Footprint 1p", "Shelter", "Tent", 45),
	item("Stakes bag", "Shelter", "Stake", 3),

	// Sleeping
	item("Thermarest Neoair X-Light R", "Sleeping", "Sleeping pad", 384),
	item("Seatosummit Aero Ultralight", "Sleeping", "Pillow", 82),
	item("Seatosummit Reactor", "Sleeping", "Sleep Liner", 275),
	item("Seatosummit Spark 0", "Sleeping", "Sleeping Bag", 240),
	item("Seatosummit Ember Eb II", "Sleeping", "Quilt", 610),

	// Water
	item("Katadyn BeFree", "Water", "Hydratation", 38),
	item("Hydrapak Seeker 3L", "Water", "Hydratation", 92),
	item("Hydrapak Ultraflask", "Water", "Hydratation", 55),
	item("Hydrapak Skyflask IT", "Water", "Hydratation", 69),
	item("Water Bottle", "Water", "Hydratation", 28),

	// Cooking
	item("Sakabouf", "Cooking", "Bag", 50),
	item("Wildo Fold a cup", "Cooking", "Popote", 24),
	item("Forclaz Trek 500 Folding", "Cooking", "Popote", 10),
	item("TOAKS Titanium", "Cooking", "Popote", 120),
	item("MSR Pocket Rocket 2", "Cooking", "Réchaud", 73),
	item("Briquet BIC", "Cooking", "Réchaud", 10),
	item("Wind Screen", "Cooking", "Réchaud", 36),
	item("Optimus Support Cartouche", "Cooking", "Réchaud", 25),
	item("Cartouche", "Cooking", "Réchaud", 186),
	item("Allume feu", "Cooking", "Fire", 30),
	item("Couteau Leatherman", "Cooking", "Various", 56),

	// Clothes (normalized from Wear)
	item("Buff Merinos", "Clothes", "Couche 2-3", 60),
	item("Veste Millet (hardshell)", "Clothes", "Couche 2-3", 380),
	item("Short Millet", "Clothes", "Bas", 120),
	item("Boxer Forclaz", "Clothes", "Couche 1 - rechange", 37),
	item("Chaussettes Quechua Hike 900", "Clothes", "Couche 1 - rechange", 30),
	item("Filets", "Clothes", "Bag", 44),
	item("Tongs Olaian", "Clothes", "Shoes", 220),
	item("Polo manches longues merinos / Forclaz Travel 500",
		"Clothes", "Nuit", 190),
	item("Collant Forclaz MT500 / Merinos / Taille L",
		"Clothes", "Nuit", 174),
	item("Doudoune Montbell Superior", "Clothes", "Couche 2-3", 243),
	item("filets", "Clothes", "sac à dos", 20),
	item("dry sack 20L", "Clothes", "Sac à dos", 86),

	// Healthcare (normalized from Healthpack)
	item("Scrubba Wash Bag", "Healthcare", "Hygiene", 145),
	item("Serviette Microfibre D4", "Healthcare", "Hygiene", 100),
	item("Eponge naturelle", "Healthcare", "Hygiene", 5),
	item("Repulsif Insectes", "Healthcare", "Hygiene", 40),
	item("Brosse à dents + dentifrice", "Healthcare", "Hygiene", 35),
	item("Gel Douche", "Healthcare", "Hygiene", 80),
	item("Savon de Marseille", "Healthcare", "Hygiene", 80),
	item("Mouchoirs en papier", "Healthcare", "Hygiene", 23),
	item("Solar Lipstick", "Healthcare", "Healthcare", 15),
	item("Compeed", "Healthcare", "Healthcare", 36),
	item("Compresse", "Healthcare", "Healthcare", 1),
	item("Tire-tiques", "Healthcare", "Healthcare", 2),
	item("Steristrip", "Healthcare", "Healthcare", 1),
	item("Chlorexidine", "Healthcare", "Healthcare", 50),
	item("Bande", "Healthcare", "Healthcare", 8),
	item("Niflugel", "Healthcare", "Healthcare", 15),
	item("Nurofen", "Healthcare", "Healthcare", 1),
	item("Serum physiologique", "Healthcare", "Healthcare", 6),
	item("Sparadrap", "Healthcare", "Healthcare", 24),
	item("Creme solaire", "Healthcare", "Santé", 100),

	// Worn
	item("Chaussure La sportiva TX4", "Worn", "Bas", 820),
	item("Picture Shooner Strecht Pants", "Worn", "Bas", 280),
	item("Baton BD FLZ", "Worn", "Divers", 385),
	item("Casquette", "Worn", "Protection Solaire", 80),
	item("Lunettes de soleil", "Worn", "Protection Solaire", 200),
	item("Boxer Forclaz", "Worn", "Couche 1", 37),
	item("Hoodie Blackcrows", "Worn", "Couche 1", 200),
	item("Chaussettes Quechua Hike 900", "Worn", "Couche 1", 30),
	item("Montre Suunto 9", "Worn", "Montre", 80),
	item("La Sportiva Akasha II", "Worn", "Chaussure", 620),

	// Electronics
	item("iPhone13", "Electronics", "Telephone", 173),
	item("Seatosummit DrySack UlytaSil Nano 2L",
		"Electronics", "Bag", 23),
	item("Anker Nano3 Charger", "Electronics", "Energy", 38),
	item("Kit Cables USB-C", "Electronics", "Energy", 52),
	item("Nitecore Carbon 10000", "Electronics", "Energy", 157),
	item("TOMTOP SOLAR PANEL 7.8 W CUSTOM",
		"Electronics", "Energy", 125),
	item("Flextail gear Tiny pump", "Electronics", "Various", 88),
	item("Headset", "Electronics", "Telephone", 30),
	item("Petzl Bindi", "Electronics", "Lamp", 35),

	// Various (normalized from MISC)
	item("Patchs", "Various", "DIY", 30),
	item("String", "Various", "DIY", 1),
	item("Couteau Leatherman", "Various", "Various", 56),
	item("Ziplocs", "Various", "Bag", 30),
	item("Pelle TheTentLab", "Various", "Outils", 27),
	item("Bouchons oreille", "Various", "Healthcare", 7),
	item("Cordelette", "Various", "Brico", 30),
	item("Argent/Cheques", "Various", "", 20),
	item("Lunettes de soleil", "Various", "Protection Solaire", 200),

	// Food & Drinks (normalized from Food)
	item("Water", "Food & Drinks", "Water", 500),
	item("cliffbar", "Food & Drinks", "food", 68),
	item("Lyophilisés soir", "Food & Drinks", "food", 130),
	item("Lyophilisés matin", "Food & Drinks", "food", 80),
	item("Tea & Coffee", "Food & Drinks", "food", 100),
}

// seedPackContentDefs links items to packs.
// Items identified by (itemName, category, description).
//
//nolint:dupl
var seedPackContentDefs = []seedPackContentDef{
	// ==========================================================
	// GR10 Pack Contents
	// ==========================================================

	// Backpack
	pc(packGR10, "Sakasec", "Backpack", "Liner", 1),
	pc(packGR10, "ALD Hybrid 30", "Backpack", "Backpack", 1),

	// Shelter
	pc(packGR10, "X-MID-PRO 1", "Shelter", "Tent", 1),
	pc(packGR10, "Stake", "Shelter", "Stake", 8),
	pc(packGR10, "Polycree Footprint 1p", "Shelter", "Tent", 1),
	pc(packGR10, "Stakes bag", "Shelter", "Stake", 1),

	// Sleeping
	pc(packGR10, "Thermarest Neoair X-Light R",
		"Sleeping", "Sleeping pad", 1),
	pc(packGR10, "Seatosummit Aero Ultralight",
		"Sleeping", "Pillow", 1),
	pc(packGR10, "Seatosummit Reactor",
		"Sleeping", "Sleep Liner", 1),
	pc(packGR10, "Seatosummit Spark 0",
		"Sleeping", "Sleeping Bag", 1),

	// Water
	pc(packGR10, "Katadyn BeFree", "Water", "Hydratation", 1),
	pc(packGR10, "Hydrapak Seeker 3L", "Water", "Hydratation", 1),
	pc(packGR10, "Hydrapak Ultraflask", "Water", "Hydratation", 1),
	pc(packGR10, "Hydrapak Skyflask IT", "Water", "Hydratation", 1),
	pc(packGR10, "Water Bottle", "Water", "Hydratation", 1),

	// Cooking
	pc(packGR10, "Sakabouf", "Cooking", "Bag", 1),
	pc(packGR10, "Wildo Fold a cup", "Cooking", "Popote", 1),
	pc(packGR10, "Forclaz Trek 500 Folding", "Cooking", "Popote", 1),
	pc(packGR10, "TOAKS Titanium", "Cooking", "Popote", 1),
	pc(packGR10, "MSR Pocket Rocket 2", "Cooking", "Réchaud", 1),
	pc(packGR10, "Briquet BIC", "Cooking", "Réchaud", 1),
	pc(packGR10, "Wind Screen", "Cooking", "Réchaud", 1),
	pc(packGR10, "Optimus Support Cartouche",
		"Cooking", "Réchaud", 1),
	pc(packGR10, "Cartouche", "Cooking", "Réchaud", 1),
	pcCons(packGR10, "Allume feu", "Cooking", "Fire", 5),

	// Clothes
	pc(packGR10, "Buff Merinos", "Clothes", "Couche 2-3", 1),
	pc(packGR10, "Veste Millet (hardshell)",
		"Clothes", "Couche 2-3", 1),
	pc(packGR10, "Short Millet", "Clothes", "Bas", 1),
	pc(packGR10, "Boxer Forclaz",
		"Clothes", "Couche 1 - rechange", 1),
	pc(packGR10, "Chaussettes Quechua Hike 900",
		"Clothes", "Couche 1 - rechange", 1),
	pc(packGR10, "Filets", "Clothes", "Bag", 1),
	pc(packGR10, "Tongs Olaian", "Clothes", "Shoes", 1),
	pc(packGR10, "Polo manches longues merinos / Forclaz Travel 500",
		"Clothes", "Nuit", 1),
	pc(packGR10, "Collant Forclaz MT500 / Merinos / Taille L",
		"Clothes", "Nuit", 1),
	pc(packGR10, "Doudoune Montbell Superior",
		"Clothes", "Couche 2-3", 1),

	// Healthcare
	pc(packGR10, "Scrubba Wash Bag", "Healthcare", "Hygiene", 1),
	pc(packGR10, "Serviette Microfibre D4",
		"Healthcare", "Hygiene", 1),
	pc(packGR10, "Eponge naturelle", "Healthcare", "Hygiene", 1),
	pc(packGR10, "Repulsif Insectes", "Healthcare", "Hygiene", 1),
	pc(packGR10, "Brosse à dents + dentifrice",
		"Healthcare", "Hygiene", 1),
	pc(packGR10, "Gel Douche", "Healthcare", "Hygiene", 1),
	pc(packGR10, "Savon de Marseille", "Healthcare", "Hygiene", 1),
	pc(packGR10, "Mouchoirs en papier", "Healthcare", "Hygiene", 2),
	pc(packGR10, "Solar Lipstick", "Healthcare", "Healthcare", 1),
	pc(packGR10, "Compeed", "Healthcare", "Healthcare", 1),
	pc(packGR10, "Compresse", "Healthcare", "Healthcare", 2),
	pc(packGR10, "Tire-tiques", "Healthcare", "Healthcare", 1),
	pc(packGR10, "Steristrip", "Healthcare", "Healthcare", 1),
	pc(packGR10, "Chlorexidine", "Healthcare", "Healthcare", 1),
	pc(packGR10, "Bande", "Healthcare", "Healthcare", 1),
	pc(packGR10, "Niflugel", "Healthcare", "Healthcare", 1),
	pc(packGR10, "Nurofen", "Healthcare", "Healthcare", 8),
	pc(packGR10, "Serum physiologique",
		"Healthcare", "Healthcare", 2),
	pc(packGR10, "Sparadrap", "Healthcare", "Healthcare", 1),

	// Worn
	pcWorn(packGR10, "Chaussure La sportiva TX4", "Bas"),
	pcWorn(packGR10, "Picture Shooner Strecht Pants", "Bas"),
	pcWorn(packGR10, "Baton BD FLZ", "Divers"),
	pcWorn(packGR10, "Casquette", "Protection Solaire"),
	pcWorn(packGR10, "Lunettes de soleil", "Protection Solaire"),
	pcWorn(packGR10, "Boxer Forclaz", "Couche 1"),
	pcWorn(packGR10, "Hoodie Blackcrows", "Couche 1"),
	pcWorn(packGR10, "Chaussettes Quechua Hike 900", "Couche 1"),
	pcWorn(packGR10, "Montre Suunto 9", "Montre"),

	// Electronics
	pc(packGR10, "iPhone13", "Electronics", "Telephone", 1),
	pc(packGR10, "Seatosummit DrySack UlytaSil Nano 2L",
		"Electronics", "Bag", 2),
	pc(packGR10, "Anker Nano3 Charger",
		"Electronics", "Energy", 1),
	pc(packGR10, "Kit Cables USB-C", "Electronics", "Energy", 1),
	pc(packGR10, "Nitecore Carbon 10000",
		"Electronics", "Energy", 1),
	pc(packGR10, "TOMTOP SOLAR PANEL 7.8 W CUSTOM",
		"Electronics", "Energy", 1),
	pc(packGR10, "Flextail gear Tiny pump",
		"Electronics", "Various", 1),
	pc(packGR10, "Headset", "Electronics", "Telephone", 1),
	pc(packGR10, "Petzl Bindi", "Electronics", "Lamp", 1),

	// Various
	pc(packGR10, "Patchs", "Various", "DIY", 1),
	pc(packGR10, "String", "Various", "DIY", 14),
	pc(packGR10, "Couteau Leatherman", "Various", "Various", 1),
	pc(packGR10, "Ziplocs", "Various", "Bag", 1),
	pc(packGR10, "Pelle TheTentLab", "Various", "Outils", 1),
	pc(packGR10, "Bouchons oreille", "Various", "Healthcare", 1),

	// Food & Drinks
	pcCons(packGR10, "Water", "Food & Drinks", "Water", 2),
	pcCons(packGR10, "cliffbar", "Food & Drinks", "food", 6),
	pcCons(packGR10, "Lyophilisés soir",
		"Food & Drinks", "food", 3),
	pcCons(packGR10, "Lyophilisés matin",
		"Food & Drinks", "food", 3),
	pcCons(packGR10, "Tea & Coffee", "Food & Drinks", "food", 1),

	// ==========================================================
	// GR54 Pack Contents
	// ==========================================================

	// Backpack
	pc(packGR54, "ALD Hybrid 30", "Backpack", "Backpack", 1),
	pc(packGR54, "Sakasec", "Backpack", "Liner", 1),

	// Shelter
	pc(packGR54, "X-MID-PRO 1", "Shelter", "Tent", 1),
	pc(packGR54, "Stake", "Shelter", "Stake", 8),
	pc(packGR54, "Stakes bag", "Shelter", "Stake", 1),
	pc(packGR54, "Polycree Footprint 1p", "Shelter", "Tent", 1),

	// Sleeping
	pc(packGR54, "Seatosummit Aero Ultralight",
		"Sleeping", "Pillow", 1),
	pc(packGR54, "Thermarest Neoair X-Light R",
		"Sleeping", "Sleeping pad", 1),
	pc(packGR54, "Seatosummit Ember Eb II",
		"Sleeping", "Quilt", 1),

	// Water
	pc(packGR54, "Hydrapak Ultraflask", "Water", "Hydratation", 1),
	pc(packGR54, "Katadyn BeFree", "Water", "Hydratation", 1),
	pc(packGR54, "Hydrapak Seeker 3L", "Water", "Hydratation", 1),
	pc(packGR54, "Hydrapak Skyflask IT", "Water", "Hydratation", 1),

	// Clothes (normalized from Wear)
	pc(packGR54, "Doudoune Montbell Superior",
		"Clothes", "Couche 2-3", 1),
	pc(packGR54, "Veste Millet (hardshell)",
		"Clothes", "Couche 2-3", 1),
	pc(packGR54, "Short Millet", "Clothes", "Bas", 1),
	pc(packGR54, "Chaussettes Quechua Hike 900",
		"Clothes", "Couche 1 - rechange", 1),
	pc(packGR54, "Boxer Forclaz",
		"Clothes", "Couche 1 - rechange", 1),
	pc(packGR54, "Collant Forclaz MT500 / Merinos / Taille L",
		"Clothes", "Nuit", 1),
	pc(packGR54, "Polo manches longues merinos / Forclaz Travel 500",
		"Clothes", "Nuit", 1),
	pc(packGR54, "Tongs Olaian", "Clothes", "Shoes", 1),
	pc(packGR54, "Buff Merinos", "Clothes", "Couche 2-3", 1),
	pc(packGR54, "filets", "Clothes", "sac à dos", 1),
	pc(packGR54, "dry sack 20L", "Clothes", "Sac à dos", 1),

	// Healthcare
	pc(packGR54, "Creme solaire", "Healthcare", "Santé", 1),
	pc(packGR54, "Solar Lipstick", "Healthcare", "Healthcare", 1),
	pc(packGR54, "Gel Douche", "Healthcare", "Hygiene", 1),
	pc(packGR54, "Brosse à dents + dentifrice",
		"Healthcare", "Hygiene", 1),
	pc(packGR54, "Repulsif Insectes", "Healthcare", "Hygiene", 1),
	pc(packGR54, "Savon de Marseille", "Healthcare", "Hygiene", 1),
	pc(packGR54, "Eponge naturelle", "Healthcare", "Hygiene", 1),
	pc(packGR54, "Scrubba Wash Bag", "Healthcare", "Hygiene", 1),
	pc(packGR54, "Mouchoirs en papier", "Healthcare", "Hygiene", 2),
	pc(packGR54, "Serviette Microfibre D4",
		"Healthcare", "Hygiene", 1),
	pc(packGR54, "Compeed", "Healthcare", "Healthcare", 1),
	pc(packGR54, "Serum physiologique",
		"Healthcare", "Healthcare", 2),
	pc(packGR54, "Chlorexidine", "Healthcare", "Healthcare", 1),
	pc(packGR54, "Tire-tiques", "Healthcare", "Healthcare", 1),
	pc(packGR54, "Niflugel", "Healthcare", "Healthcare", 1),
	pc(packGR54, "Bande", "Healthcare", "Healthcare", 1),
	pc(packGR54, "Sparadrap", "Healthcare", "Healthcare", 1),
	pc(packGR54, "Nurofen", "Healthcare", "Healthcare", 8),

	// Cooking
	pc(packGR54, "Cartouche", "Cooking", "Réchaud", 1),
	pc(packGR54, "MSR Pocket Rocket 2", "Cooking", "Réchaud", 1),
	pc(packGR54, "Sakabouf", "Cooking", "Bag", 1),
	pc(packGR54, "Wind Screen", "Cooking", "Réchaud", 1),
	pc(packGR54, "Briquet BIC", "Cooking", "Réchaud", 1),
	pc(packGR54, "Couteau Leatherman", "Cooking", "Various", 1),
	pc(packGR54, "Wildo Fold a cup", "Cooking", "Popote", 1),
	pc(packGR54, "Forclaz Trek 500 Folding", "Cooking", "Popote", 1),

	// Electronics
	pc(packGR54, "Petzl Bindi", "Electronics", "Lamp", 1),
	pc(packGR54, "Headset", "Electronics", "Telephone", 1),
	pc(packGR54, "TOMTOP SOLAR PANEL 7.8 W CUSTOM",
		"Electronics", "Energy", 1),
	pc(packGR54, "Flextail gear Tiny pump",
		"Electronics", "Various", 1),
	pc(packGR54, "Nitecore Carbon 10000",
		"Electronics", "Energy", 1),
	pc(packGR54, "Kit Cables USB-C", "Electronics", "Energy", 1),
	pc(packGR54, "Anker Nano3 Charger",
		"Electronics", "Energy", 1),
	pc(packGR54, "iPhone13", "Electronics", "Telephone", 1),
	pc(packGR54, "Seatosummit DrySack UlytaSil Nano 2L",
		"Electronics", "Bag", 2),

	// Various (normalized from MISC)
	pc(packGR54, "Lunettes de soleil",
		"Various", "Protection Solaire", 1),
	pc(packGR54, "Bouchons oreille", "Various", "Healthcare", 1),
	pc(packGR54, "Cordelette", "Various", "Brico", 1),
	pc(packGR54, "Patchs", "Various", "DIY", 1),
	pc(packGR54, "Argent/Cheques", "Various", "", 1),

	// Worn
	pcWorn(packGR54, "Chaussettes Quechua Hike 900", "Couche 1"),
	pcWorn(packGR54, "Hoodie Blackcrows", "Couche 1"),
	pcWorn(packGR54, "Boxer Forclaz", "Couche 1"),
	pcWorn(packGR54, "Picture Shooner Strecht Pants", "Bas"),
	pcWorn(packGR54, "Casquette", "Protection Solaire"),
	pcWorn(packGR54, "La Sportiva Akasha II", "Chaussure"),
	pcWorn(packGR54, "Baton BD FLZ", "Divers"),
	pcWorn(packGR54, "Montre Suunto 9", "Montre"),

	// Food & Drinks (normalized from Food)
	pcCons(packGR54, "Water", "Food & Drinks", "Water", 2),
	pcCons(packGR54, "Tea & Coffee", "Food & Drinks", "food", 1),
	pcCons(packGR54, "Lyophilisés soir",
		"Food & Drinks", "food", 1),
	pcCons(packGR54, "Lyophilisés matin",
		"Food & Drinks", "food", 1),
}
