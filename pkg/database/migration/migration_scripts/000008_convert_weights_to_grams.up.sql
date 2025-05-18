-- Convert existing weights to grams
UPDATE inventory 
SET weight = CASE 
    WHEN weight_unit = 'IMPERIAL' THEN ROUND(weight * 28.3495) -- Convert oz to grams and round to nearest integer
    ELSE weight -- Already in grams
END;

-- Remove the weight_unit column
ALTER TABLE inventory DROP COLUMN weight_unit; 