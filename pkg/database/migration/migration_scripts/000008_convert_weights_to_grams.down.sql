-- Add back the weight_unit column
ALTER TABLE inventory ADD COLUMN weight_unit VARCHAR(10) DEFAULT 'METRIC';

-- Convert weights back to original units
UPDATE inventory 
SET weight = CASE 
    WHEN weight_unit = 'IMPERIAL' THEN ROUND(weight / 28.3495) -- Convert grams back to oz
    ELSE weight -- Already in grams
END; 