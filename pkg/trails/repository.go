package trails

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"time"

	"github.com/Angak0k/pimpmypack/pkg/database"
	"github.com/lib/pq"
)

const trailNameConstraint = "idx_trail_name"

func returnTrails(ctx context.Context) (Trails, error) {
	rows, err := database.DB().QueryContext(ctx,
		`SELECT id, name, country, continent, distance_km, url, created_at, updated_at
		FROM trail
		ORDER BY continent, country, name;`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var trails Trails
	for rows.Next() {
		var t Trail
		err := rows.Scan(&t.ID, &t.Name, &t.Country, &t.Continent,
			&t.DistanceKm, &t.URL, &t.CreatedAt, &t.UpdatedAt)
		if err != nil {
			return nil, err
		}
		trails = append(trails, t)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	return trails, nil
}

func findTrailByID(ctx context.Context, id uint) (*Trail, error) {
	var t Trail
	err := database.DB().QueryRowContext(ctx,
		`SELECT id, name, country, continent, distance_km, url, created_at, updated_at
		FROM trail WHERE id = $1;`, id).Scan(
		&t.ID, &t.Name, &t.Country, &t.Continent,
		&t.DistanceKm, &t.URL, &t.CreatedAt, &t.UpdatedAt)

	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, ErrTrailNotFound
		}
		return nil, err
	}

	return &t, nil
}

func findTrailByName(ctx context.Context, name string) (*Trail, error) {
	var t Trail
	err := database.DB().QueryRowContext(ctx,
		`SELECT id, name, country, continent, distance_km, url, created_at, updated_at
		FROM trail WHERE name = $1;`, name).Scan(
		&t.ID, &t.Name, &t.Country, &t.Continent,
		&t.DistanceKm, &t.URL, &t.CreatedAt, &t.UpdatedAt)

	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, ErrTrailNotFound
		}
		return nil, err
	}

	return &t, nil
}

func insertTrail(ctx context.Context, t *Trail) error {
	if t == nil {
		return errors.New("payload is empty")
	}
	t.CreatedAt = time.Now().Truncate(time.Second)
	t.UpdatedAt = time.Now().Truncate(time.Second)

	err := database.DB().QueryRowContext(ctx,
		`INSERT INTO trail (name, country, continent, distance_km, url, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7)
		RETURNING id;`,
		t.Name, t.Country, t.Continent, t.DistanceKm, t.URL,
		t.CreatedAt, t.UpdatedAt).Scan(&t.ID)

	if err != nil {
		var pqErr *pq.Error
		if errors.As(err, &pqErr) && pqErr.Code == "23505" && pqErr.Constraint == trailNameConstraint {
			return ErrTrailNameExists
		}
		return err
	}

	return nil
}

func updateTrailByID(ctx context.Context, id uint, t *Trail) error {
	if t == nil {
		return errors.New("payload is empty")
	}
	t.UpdatedAt = time.Now().Truncate(time.Second)

	result, err := database.DB().ExecContext(ctx,
		`UPDATE trail SET name=$1, country=$2, continent=$3, distance_km=$4, url=$5, updated_at=$6
		WHERE id=$7;`,
		t.Name, t.Country, t.Continent, t.DistanceKm, t.URL, t.UpdatedAt, id)
	if err != nil {
		var pqErr *pq.Error
		if errors.As(err, &pqErr) && pqErr.Code == "23505" && pqErr.Constraint == trailNameConstraint {
			return ErrTrailNameExists
		}
		return err
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return err
	}
	if rowsAffected == 0 {
		return ErrTrailNotFound
	}

	return nil
}

func deleteTrailByID(ctx context.Context, id uint) error {
	// Check if any pack references this trail
	var count int
	err := database.DB().QueryRowContext(ctx,
		`SELECT COUNT(*) FROM pack WHERE trail_id = $1;`, id).Scan(&count)
	if err != nil {
		return err
	}
	if count > 0 {
		return ErrTrailInUse
	}

	result, err := database.DB().ExecContext(ctx,
		`DELETE FROM trail WHERE id = $1;`, id)
	if err != nil {
		return err
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return err
	}
	if rowsAffected == 0 {
		return ErrTrailNotFound
	}

	return nil
}

func insertTrailsBulk(ctx context.Context, trailsToCreate []Trail) ([]Trail, error) {
	tx, err := database.DB().BeginTx(ctx, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer func() {
		if err != nil {
			_ = tx.Rollback()
		}
	}()

	now := time.Now().Truncate(time.Second)
	created := make([]Trail, 0, len(trailsToCreate))

	for i := range trailsToCreate {
		t := &trailsToCreate[i]
		t.CreatedAt = now
		t.UpdatedAt = now

		err = tx.QueryRowContext(ctx,
			`INSERT INTO trail (name, country, continent, distance_km, url, created_at, updated_at)
			VALUES ($1, $2, $3, $4, $5, $6, $7)
			RETURNING id;`,
			t.Name, t.Country, t.Continent, t.DistanceKm, t.URL,
			t.CreatedAt, t.UpdatedAt).Scan(&t.ID)
		if err != nil {
			var pqErr *pq.Error
			if errors.As(err, &pqErr) && pqErr.Code == "23505" && pqErr.Constraint == trailNameConstraint {
				return nil, ErrTrailNameExists
			}
			return nil, fmt.Errorf("failed to insert trail %q: %w", t.Name, err)
		}
		created = append(created, *t)
	}

	if err = tx.Commit(); err != nil {
		return nil, fmt.Errorf("failed to commit transaction: %w", err)
	}

	return created, nil
}

func deleteTrailsBulk(ctx context.Context, ids []uint) error {
	tx, err := database.DB().BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer func() {
		if err != nil {
			_ = tx.Rollback()
		}
	}()

	for _, id := range ids {
		// Check if any pack references this trail
		var count int
		err = tx.QueryRowContext(ctx,
			`SELECT COUNT(*) FROM pack WHERE trail_id = $1;`, id).Scan(&count)
		if err != nil {
			return fmt.Errorf("failed to check trail %d usage: %w", id, err)
		}
		if count > 0 {
			return ErrTrailInUse
		}

		result, execErr := tx.ExecContext(ctx, `DELETE FROM trail WHERE id = $1;`, id)
		if execErr != nil {
			err = execErr
			return fmt.Errorf("failed to delete trail %d: %w", id, err)
		}

		rowsAffected, raErr := result.RowsAffected()
		if raErr != nil {
			err = raErr
			return fmt.Errorf("failed to get rows affected for trail %d: %w", id, err)
		}
		if rowsAffected == 0 {
			err = ErrTrailNotFound
			return err
		}
	}

	if err = tx.Commit(); err != nil {
		return fmt.Errorf("failed to commit transaction: %w", err)
	}

	return nil
}

func returnTrailsGrouped(ctx context.Context) (map[string]map[string][]TrailSummary, error) {
	rows, err := database.DB().QueryContext(ctx,
		`SELECT id, name, continent, country, distance_km, url
		FROM trail
		ORDER BY continent, country, name;`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	grouped := make(map[string]map[string][]TrailSummary)
	for rows.Next() {
		var id uint
		var name, continent, country string
		var distanceKm *int
		var url *string

		err := rows.Scan(&id, &name, &continent, &country, &distanceKm, &url)
		if err != nil {
			return nil, err
		}

		if grouped[continent] == nil {
			grouped[continent] = make(map[string][]TrailSummary)
		}
		grouped[continent][country] = append(grouped[continent][country], TrailSummary{
			ID:         id,
			Name:       name,
			DistanceKm: distanceKm,
			URL:        url,
		})
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	return grouped, nil
}

func returnTrailNames(ctx context.Context) ([]string, error) {
	rows, err := database.DB().QueryContext(ctx,
		`SELECT name FROM trail ORDER BY name;`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var names []string
	for rows.Next() {
		var name string
		if err := rows.Scan(&name); err != nil {
			return nil, err
		}
		names = append(names, name)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	return names, nil
}
