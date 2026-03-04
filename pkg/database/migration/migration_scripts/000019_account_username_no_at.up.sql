ALTER TABLE account ADD CONSTRAINT account_username_no_at CHECK (username NOT LIKE '%@%');
