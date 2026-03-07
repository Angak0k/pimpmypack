CREATE TABLE "trail" (
    "id"           SERIAL PRIMARY KEY,
    "name"         TEXT NOT NULL,
    "country"      TEXT NOT NULL,
    "continent"    TEXT NOT NULL,
    "distance_km"  INT,
    "url"          TEXT,
    "created_at"   TIMESTAMP NOT NULL DEFAULT NOW(),
    "updated_at"   TIMESTAMP NOT NULL DEFAULT NOW()
);

CREATE UNIQUE INDEX idx_trail_name ON trail(name);

-- Seed with curated trail data
INSERT INTO trail (name, country, continent, distance_km) VALUES
-- North America
('Appalachian Trail', 'United States', 'North America', 3500),
('Pacific Crest Trail', 'United States', 'North America', 4265),
('Continental Divide Trail', 'United States', 'North America', 4989),
('John Muir Trail', 'United States', 'North America', 340),
('Colorado Trail', 'United States', 'North America', 782),
('Long Trail', 'United States', 'North America', 439),
('Wonderland Trail', 'United States', 'North America', 150),
('Superior Hiking Trail', 'United States', 'North America', 499),
('Tahoe Rim Trail', 'United States', 'North America', 266),
('Arizona Trail', 'United States', 'North America', 1287),
('Ice Age Trail', 'United States', 'North America', 1900),
('Pacific Northwest Trail', 'United States', 'North America', 1931),
('Hayduke Trail', 'United States', 'North America', 1300),
('Great Divide Trail', 'Canada', 'North America', 1130),
('West Coast Trail', 'Canada', 'North America', 75),
('Bruce Trail', 'Canada', 'North America', 900),

-- Europe - France
('GR20', 'France', 'Europe', 180),
('GR10', 'France', 'Europe', 866),
('GR34', 'France', 'Europe', 2000),
('GR5', 'France', 'Europe', 620),
('GR54', 'France', 'Europe', 176),
('HRP', 'France', 'Europe', 800),
('Hexatrek', 'France', 'Europe', 3034),
('Tour du Mont Blanc', 'France', 'Europe', 170),

-- Europe - Spain
('Camino de Santiago', 'Spain', 'Europe', 800),
('Camino del Norte', 'Spain', 'Europe', 825),
('GR11', 'Spain', 'Europe', 840),

-- Europe - Italy
('Alta Via 1', 'Italy', 'Europe', 120),
('Alta Via 2', 'Italy', 'Europe', 160),
('Sentiero Italia', 'Italy', 'Europe', 7000),
('Via Francigena', 'Italy', 'Europe', 1000),

-- Europe - UK
('West Highland Way', 'United Kingdom', 'Europe', 154),
('Pennine Way', 'United Kingdom', 'Europe', 431),
('South West Coast Path', 'United Kingdom', 'Europe', 1014),
('Cape Wrath Trail', 'United Kingdom', 'Europe', 370),

-- Europe - Scandinavia
('Kungsleden', 'Sweden', 'Europe', 440),
('Nordkalottleden', 'Sweden', 'Europe', 800),
('Laugavegur Trail', 'Iceland', 'Europe', 55),

-- Europe - Other
('E5 European Long Distance Path', 'Germany', 'Europe', 600),
('Walker''s Haute Route', 'Switzerland', 'Europe', 180),
('Via Alpina Red Trail', 'Switzerland', 'Europe', 2500),
('Lycian Way', 'Turkey', 'Europe', 540),
('Rota Vicentina', 'Portugal', 'Europe', 450),

-- Asia
('Kumano Kodo', 'Japan', 'Asia', 70),
('Nakasendo Trail', 'Japan', 'Asia', 335),
('Shikoku Pilgrimage', 'Japan', 'Asia', 1200),
('Annapurna Circuit', 'Nepal', 'Asia', 230),
('Everest Base Camp Trek', 'Nepal', 'Asia', 130),
('Langtang Valley Trek', 'Nepal', 'Asia', 75),
('Markha Valley Trek', 'India', 'Asia', 75),

-- Oceania
('Te Araroa', 'New Zealand', 'Oceania', 3000),
('Milford Track', 'New Zealand', 'Oceania', 53),
('Routeburn Track', 'New Zealand', 'Oceania', 32),
('Tongariro Northern Circuit', 'New Zealand', 'Oceania', 43),
('Overland Track', 'Australia', 'Oceania', 65),
('Larapinta Trail', 'Australia', 'Oceania', 223),
('Bibbulmun Track', 'Australia', 'Oceania', 1003),

-- South America
('Torres del Paine W Trek', 'Chile', 'South America', 80),
('Torres del Paine O Circuit', 'Chile', 'South America', 130),
('Huemul Circuit', 'Argentina', 'South America', 65),
('Inca Trail', 'Peru', 'South America', 43),
('Santa Cruz Trek', 'Peru', 'South America', 50),
('Lares Trek', 'Peru', 'South America', 33),

-- Africa
('Mount Kilimanjaro', 'Tanzania', 'Africa', 62),
('Mount Kenya Circuit', 'Kenya', 'Africa', 50),
('Otter Trail', 'South Africa', 'Africa', 42),
('Fish River Canyon Trail', 'Namibia', 'Africa', 85);
