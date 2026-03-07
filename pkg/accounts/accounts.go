package accounts

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"log"
	"strconv"
	"time"

	"github.com/Angak0k/pimpmypack/pkg/config"
	"github.com/Angak0k/pimpmypack/pkg/database"
	"github.com/Angak0k/pimpmypack/pkg/helper"
	"github.com/Angak0k/pimpmypack/pkg/security"
	"github.com/lib/pq"
)

// ErrNoAccountFound is returned when no account is found for a given ID.
var ErrNoAccountFound = errors.New("no account found")

// stageLocal represents the local development stage
const stageLocal = "LOCAL"

// ErrInvalidCredentials is returned when login credentials are invalid.
var ErrInvalidCredentials = errors.New("invalid credentials")

// ErrEmailAlreadyExists is returned when a registration is attempted with an already-used email.
var ErrEmailAlreadyExists = errors.New("email already exists")

// emailSender is the package-level email sender, replaceable for testing.
var emailSender helper.EmailSender

// setEmailSender sets the email sender (used in tests to inject a mock).
func setEmailSender(s helper.EmailSender) {
	emailSender = s
}

// getEmailSender returns the configured email sender, defaulting to SMTP.
func getEmailSender() helper.EmailSender {
	if emailSender != nil {
		return emailSender
	}
	return &helper.SMTPClient{Server: config.MailServerConfig}
}

func registerUser(ctx context.Context, u User) (bool, error) {
	var id int

	confirmationCode, err := helper.GenerateRandomCode(30)
	if err != nil {
		return false, fmt.Errorf("failed to generate confirmation code: %w", err)
	}

	err = database.DB().QueryRowContext(ctx,
		`INSERT INTO account 
		(username, email, firstname, lastname, role, status, confirmation_code, created_at, updated_at) 
		VALUES ($1,$2,$3,$4,$5,$6,$7,$8, $9) 
		RETURNING id;`,
		u.Username,
		u.Email,
		u.Firstname,
		u.Lastname,
		u.Role,
		u.Status,
		confirmationCode,
		u.CreatedAt,
		u.UpdatedAt).Scan(&id)
	if err != nil {
		var pqErr *pq.Error
		if errors.As(err, &pqErr) && pqErr.Code == "23505" && pqErr.Constraint == "account_email_unique" {
			return false, ErrEmailAlreadyExists
		}
		return false, fmt.Errorf("failed to insert user: %w", err)
	}

	//nolint:gosec
	u.ID = uint(id)

	hashedPassword, err := security.HashPassword(u.Password)
	if err != nil {
		return false, fmt.Errorf("failed to hash password: %w", err)
	}

	err = database.DB().QueryRowContext(ctx,
		`INSERT INTO password (user_id, password, updated_at) 
		VALUES ($1,$2,$3) 
		RETURNING id;`,
		id, hashedPassword, u.UpdatedAt).Scan(&id)
	if err != nil {
		return false, fmt.Errorf("failed to insert password: %w", err)
	}

	err = sendConfirmationEmail(u, confirmationCode)
	if err != nil {
		// we haven't succed to send the email but the user is created
		//nolint:nilerr
		return false, nil
	}

	return true, nil
}

// Send confirmation email
func sendConfirmationEmail(u User, code string) error {
	// LOCAL mode: Don't send email, just log the confirmation details
	if config.Stage == stageLocal {
		log.Printf("LOCAL MODE: Email confirmation bypassed for user %s (ID: %d)", u.Username, u.ID)
		log.Printf("LOCAL MODE: Confirm at: /api/confirmemail?id=%d&code=%s", u.ID, code)
		log.Printf("LOCAL MODE: Or use simplified confirmation: /api/confirmemail?username=%s&email=%s", u.Username, u.Email)
		return nil
	}

	// Send confirmation email
	confirmURL := config.Scheme + "://" + config.HostName + "/confirmemail.html?id=" +
		strconv.FormatUint(uint64(u.ID), 10) + "&code=" + code

	mailRcpt := u.Email
	mailSubject := "PimpMyPack - Confirm your email address"
	mailTextBody := helper.BuildConfirmationEmailText(u.Username, confirmURL)
	mailHTMLBody := helper.BuildConfirmationEmailHTML(u.Username, confirmURL)

	err := getEmailSender().SendEmail(mailRcpt, mailSubject, mailTextBody, mailHTMLBody)
	if err != nil {
		return fmt.Errorf("failed to send confirmation email: %w", err)
	}

	return nil
}

func resendConfirmEmail(ctx context.Context, email string) error {
	var user User
	err := database.DB().QueryRowContext(ctx,
		`SELECT id, username, email FROM account WHERE email = $1 AND status = 'pending'`,
		email,
	).Scan(&user.ID, &user.Username, &user.Email)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			// Email not found or not pending — return nil to prevent user enumeration
			return nil
		}
		return fmt.Errorf("failed to query account: %w", err)
	}

	// Generate a new confirmation code (invalidates old one)
	confirmationCode, err := helper.GenerateRandomCode(30)
	if err != nil {
		return fmt.Errorf("failed to generate confirmation code: %w", err)
	}

	// Update the confirmation code in DB
	_, err = database.DB().ExecContext(ctx,
		`UPDATE account SET confirmation_code = $1, updated_at = $2 WHERE id = $3`,
		confirmationCode, time.Now().Truncate(time.Second), user.ID,
	)
	if err != nil {
		return fmt.Errorf("failed to update confirmation code: %w", err)
	}

	err = sendConfirmationEmail(user, confirmationCode)
	if err != nil {
		return fmt.Errorf("failed to send confirmation email: %w", err)
	}

	return nil
}

func confirmEmail(ctx context.Context, id string, code string) error {
	// LOCAL mode: Accept any code
	if config.Stage == stageLocal && code == "LOCAL_BYPASS" {
		log.Printf("LOCAL MODE: Bypassing code verification for user ID %s", id)
		code = "" // Will be ignored in the query
	}

	// Check if the confirmation code is valid
	var query string
	var args []interface{}

	if config.Stage == stageLocal && code == "" {
		// LOCAL mode with bypass: only check ID
		query = "SELECT id FROM account WHERE id = $1;"
		args = []interface{}{id}
	} else {
		// Normal mode: check both ID and code
		query = "SELECT id FROM account WHERE id = $1 AND confirmation_code = $2;"
		args = []interface{}{id, code}
	}

	row := database.DB().QueryRowContext(ctx, query, args...)
	err := row.Scan(&id)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil
		}
		return err
	}
	// Update the DB
	statement, err := database.DB().PrepareContext(ctx, "UPDATE account SET status = 'active' WHERE id = $1;")
	if err != nil {
		return fmt.Errorf("failed to prepare update statement: %w", err)
	}
	defer statement.Close()

	_, err = statement.ExecContext(ctx, id)
	if err != nil {
		return fmt.Errorf("failed to execute update query: %w", err)
	}

	return nil
}

// confirmUserByUsernameAndEmail confirms a user by username and email (LOCAL mode only)
func confirmUserByUsernameAndEmail(ctx context.Context, username, email string) error {
	result, err := database.DB().ExecContext(ctx,
		`UPDATE account SET status = 'active'
		 WHERE username = $1 AND email = $2 AND status = 'pending'`,
		username, email)
	if err != nil {
		return fmt.Errorf("failed to confirm user: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}

	if rowsAffected == 0 {
		return errors.New("user not found or already confirmed")
	}

	log.Printf("LOCAL MODE: Confirmed user %s (%s) via username/email", username, email)
	return nil
}

func forgotPassword(ctx context.Context, email string) error {
	newPassword, err := helper.GenerateRandomCode(10)
	if err != nil {
		return fmt.Errorf("failed to generate new password: %w", err)
	}

	userID, err := getUserIDByEmail(ctx, email)
	if err != nil {
		// email not found but we don't want to leak this information
		return fmt.Errorf("failed to process the request %w", err)
	}

	err = updatePassword(context.Background(), userID, newPassword)
	if err != nil {
		return fmt.Errorf("failed to generate new password: %w", err)
	}

	mailRcpt := email
	mailSubject := "PimpMyPack - Your password has been reset"
	mailTextBody := helper.BuildPasswordResetEmailText(newPassword)
	mailHTMLBody := helper.BuildPasswordResetEmailHTML(newPassword)

	err = getEmailSender().SendEmail(mailRcpt, mailSubject, mailTextBody, mailHTMLBody)
	if err != nil {
		return fmt.Errorf("failed to send password reset email: %w", err)
	}
	return nil
}

func loginCheck(
	ctx context.Context, username string, password string, rememberMe bool,
) (*security.TokenPairResponse, uint, bool, error) {
	var err error
	var status string
	var storedPassword string
	var id uint

	row := database.DB().QueryRowContext(ctx,
		`SELECT p.password, p.user_id, a.status
		FROM password AS p JOIN account AS a ON p.user_id = a.id
		WHERE a.username = $1 OR a.email = $1;`,
		username)
	err = row.Scan(&storedPassword, &id, &status)

	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, 0, false, ErrInvalidCredentials
		}
		return nil, 0, false, fmt.Errorf("failed to query user: %w", err)
	}

	err = security.VerifyPassword(password, storedPassword)

	if err != nil {
		return nil, id, false, ErrInvalidCredentials
	}

	// Check account status before generating tokens
	if status != "active" {
		return nil, id, true, nil
	}

	// Only generate tokens for active accounts
	tokenPair, err := security.GenerateTokenPair(ctx, id, rememberMe)

	if err != nil {
		return nil, id, false, err
	}

	return tokenPair, id, false, nil
}

func updatePassword(ctx context.Context, userID uint, updatedPassword string) error {
	var lastPassword string
	// Get old password
	row := database.DB().QueryRowContext(ctx, "SELECT password FROM password WHERE user_id = $1;", userID)
	err := row.Scan(&lastPassword)
	if err != nil {
		return fmt.Errorf("failed to get old password: %w", err)
	}

	// Update DB
	statement, err := database.DB().PrepareContext(ctx,
		`UPDATE password 
		SET password = $1, last_password = $2, updated_at = $3 
		WHERE user_id = $4;`)
	if err != nil {
		return fmt.Errorf("failed to prepare update statement: %w", err)
	}
	defer statement.Close()

	hashedPassword, err := security.HashPassword(updatedPassword)
	if err != nil {
		return fmt.Errorf("failed to hash password: %w", err)
	}

	_, err = statement.ExecContext(ctx, hashedPassword, lastPassword, time.Now().Truncate(time.Second), userID)
	if err != nil {
		return fmt.Errorf("failed to execute update query: %w", err)
	}

	return nil
}

func returnAccounts(ctx context.Context) (*Accounts, error) {
	var accounts Accounts

	rows, err := database.DB().QueryContext(ctx,
		`SELECT a.id, a.username, a.email, a.firstname, a.lastname, a.role, a.status,
		    a.preferred_currency, a.preferred_unit_system, a.youtube_url, a.instagram_url,
		    CASE WHEN ai.account_id IS NOT NULL THEN true ELSE false END AS has_profile_image,
		    a.image_position_x, a.image_position_y, a.is_profile_public,
		    CASE WHEN abi.account_id IS NOT NULL THEN true ELSE false END AS has_banner_image,
		    a.banner_position_y,
		    a.created_at, a.updated_at
		FROM account a
		LEFT JOIN account_images ai ON a.id = ai.account_id
		LEFT JOIN account_banner_images abi ON a.id = abi.account_id;`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	for rows.Next() {
		var account Account
		err := rows.Scan(
			&account.ID,
			&account.Username,
			&account.Email,
			&account.Firstname,
			&account.Lastname,
			&account.Role,
			&account.Status,
			&account.PreferredCurrency,
			&account.PreferredUnitSystem,
			&account.YoutubeURL,
			&account.InstagramURL,
			&account.HasProfileImage,
			&account.ImagePositionX,
			&account.ImagePositionY,
			&account.IsProfilePublic,
			&account.HasBannerImage,
			&account.BannerPositionY,
			&account.CreatedAt,
			&account.UpdatedAt)
		if err != nil {
			return nil, err
		}
		accounts = append(accounts, account)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	return &accounts, nil
}

func findAccountByID(ctx context.Context, id uint) (*Account, error) {
	var account Account

	row := database.DB().QueryRowContext(ctx,
		`SELECT a.id, a.username, a.email, a.firstname, a.lastname, a.role, a.status,
		    a.preferred_currency, a.preferred_unit_system, a.youtube_url, a.instagram_url,
		    CASE WHEN ai.account_id IS NOT NULL THEN true ELSE false END AS has_profile_image,
		    a.image_position_x, a.image_position_y, a.is_profile_public,
		    CASE WHEN abi.account_id IS NOT NULL THEN true ELSE false END AS has_banner_image,
		    a.banner_position_y,
		    a.created_at, a.updated_at
		FROM account a
		LEFT JOIN account_images ai ON a.id = ai.account_id
		LEFT JOIN account_banner_images abi ON a.id = abi.account_id
		WHERE a.id = $1;`,
		id)
	err := row.Scan(
		&account.ID,
		&account.Username,
		&account.Email,
		&account.Firstname,
		&account.Lastname,
		&account.Role,
		&account.Status,
		&account.PreferredCurrency,
		&account.PreferredUnitSystem,
		&account.YoutubeURL,
		&account.InstagramURL,
		&account.HasProfileImage,
		&account.ImagePositionX,
		&account.ImagePositionY,
		&account.IsProfilePublic,
		&account.HasBannerImage,
		&account.BannerPositionY,
		&account.CreatedAt,
		&account.UpdatedAt)

	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			// Handle case when no rows are returned
			return nil, ErrNoAccountFound
		}
		return nil, err
	}

	return &account, nil
}

func insertAccount(ctx context.Context, a *Account) error {
	if a == nil {
		return errors.New("payload is empty")
	}
	a.CreatedAt = time.Now().Truncate(time.Second)
	a.UpdatedAt = time.Now().Truncate(time.Second)

	err := database.DB().QueryRowContext(ctx,
		`INSERT INTO account (username, email, firstname, lastname, role, status, preferred_currency,
		    preferred_unit_system, youtube_url, instagram_url, created_at, updated_at)
		VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11,$12)
		RETURNING id;`,
		a.Username, a.Email, a.Firstname, a.Lastname, a.Role, a.Status, "EUR", "METRIC",
		a.YoutubeURL, a.InstagramURL, a.CreatedAt, a.UpdatedAt).Scan(&a.ID)

	if err != nil {
		return err
	}
	return nil
}

// updateAccountByID updates ALL fields of an account including role and status
// ADMIN ONLY: Used by PutAccountByID (admin endpoint) to update any user's account
// SECURITY: NEVER call from user-facing endpoints - use updateMyAccount() instead
func updateAccountByID(ctx context.Context, id uint, a *Account) error {
	if a == nil {
		return errors.New("payload is empty")
	}

	a.ID = id
	a.UpdatedAt = time.Now().Truncate(time.Second)
	statement, err := database.DB().PrepareContext(ctx,
		`UPDATE account SET email=$1, firstname=$2, lastname=$3, status=$4, role=$5, preferred_currency=$6,
		    preferred_unit_system=$7, youtube_url=$8, instagram_url=$9, image_position_x=$10,
		    image_position_y=$11, is_profile_public=$12, banner_position_y=$13, updated_at=$14
		WHERE id=$15 RETURNING username;`)
	if err != nil {
		return err
	}

	defer statement.Close()

	err = statement.QueryRowContext(ctx, a.Email, a.Firstname, a.Lastname, a.Status, a.Role, a.PreferredCurrency,
		a.PreferredUnitSystem, a.YoutubeURL, a.InstagramURL, a.ImagePositionX, a.ImagePositionY,
		a.IsProfilePublic, a.BannerPositionY, a.UpdatedAt, a.ID).Scan(&a.Username)
	if err != nil {
		return err
	}
	return nil
}

// isValidationError checks if the error is a user input validation error
func isValidationError(err error) bool {
	msg := err.Error()
	return msg == "invalid email format" ||
		msg == "image_position_x must be between 0 and 100" ||
		msg == "image_position_y must be between 0 and 100" ||
		msg == "banner_position_y must be between 0 and 100"
}

// normalizeEmptyStringPtr converts empty string pointers to nil for clean DB storage
func normalizeEmptyStringPtr(s *string) *string {
	if s != nil && *s == "" {
		return nil
	}
	return s
}

// updateMyAccount updates only user-controllable fields of an account
// SECURITY: This function explicitly excludes role, status, username to prevent privilege escalation
// Only safe fields can be updated: email, firstname, lastname, preferences, social URLs
func updateMyAccount(ctx context.Context, userID uint, input *AccountUpdateInput) (*Account, error) {
	if input == nil {
		return nil, errors.New("input is empty")
	}

	// Validate email format
	if !helper.IsValidEmail(input.Email) {
		return nil, errors.New("invalid email format")
	}

	// Normalize empty strings to nil for optional URL fields
	input.YoutubeURL = normalizeEmptyStringPtr(input.YoutubeURL)
	input.InstagramURL = normalizeEmptyStringPtr(input.InstagramURL)

	// Validate image position range when provided
	if input.ImagePositionX != nil && (*input.ImagePositionX < 0 || *input.ImagePositionX > 100) {
		return nil, errors.New("image_position_x must be between 0 and 100")
	}
	if input.ImagePositionY != nil && (*input.ImagePositionY < 0 || *input.ImagePositionY > 100) {
		return nil, errors.New("image_position_y must be between 0 and 100")
	}
	if input.BannerPositionY != nil && (*input.BannerPositionY < 0 || *input.BannerPositionY > 100) {
		return nil, errors.New("banner_position_y must be between 0 and 100")
	}

	now := time.Now().Truncate(time.Second)

	// Update ONLY safe fields - explicitly excludes role, status, username
	// COALESCE keeps existing values when client omits them (nil)
	statement, err := database.DB().PrepareContext(ctx,
		`UPDATE account
		 SET email=$1, firstname=$2, lastname=$3, preferred_currency=$4,
		     preferred_unit_system=$5, youtube_url=$6, instagram_url=$7,
		     image_position_x=COALESCE($8, image_position_x),
		     image_position_y=COALESCE($9, image_position_y),
		     is_profile_public=COALESCE($10, is_profile_public),
		     banner_position_y=COALESCE($11, banner_position_y),
		     updated_at=$12
		 WHERE id=$13
		 RETURNING id, username, email, firstname, lastname, role, status,
		           preferred_currency, preferred_unit_system, youtube_url, instagram_url,
		           image_position_x, image_position_y, is_profile_public,
		           banner_position_y,
		           created_at, updated_at;`)
	if err != nil {
		return nil, fmt.Errorf("failed to prepare statement: %w", err)
	}
	defer statement.Close()

	var account Account
	err = statement.QueryRowContext(ctx,
		input.Email,
		input.Firstname,
		input.Lastname,
		input.PreferredCurrency,
		input.PreferredUnitSystem,
		input.YoutubeURL,
		input.InstagramURL,
		input.ImagePositionX,
		input.ImagePositionY,
		input.IsProfilePublic,
		input.BannerPositionY,
		now,
		userID,
	).Scan(
		&account.ID,
		&account.Username,
		&account.Email,
		&account.Firstname,
		&account.Lastname,
		&account.Role,
		&account.Status,
		&account.PreferredCurrency,
		&account.PreferredUnitSystem,
		&account.YoutubeURL,
		&account.InstagramURL,
		&account.ImagePositionX,
		&account.ImagePositionY,
		&account.IsProfilePublic,
		&account.BannerPositionY,
		&account.CreatedAt,
		&account.UpdatedAt,
	)

	if errors.Is(err, sql.ErrNoRows) {
		return nil, ErrNoAccountFound
	}
	if err != nil {
		return nil, fmt.Errorf("failed to update account: %w", err)
	}

	// Compute has_profile_image and has_banner_image separately
	var hasProfileImage, hasBannerImage bool
	err = database.DB().QueryRowContext(ctx,
		`SELECT
		    EXISTS(SELECT 1 FROM account_images WHERE account_id = $1),
		    EXISTS(SELECT 1 FROM account_banner_images WHERE account_id = $1)`,
		userID,
	).Scan(&hasProfileImage, &hasBannerImage)
	if err != nil {
		return nil, fmt.Errorf("failed to check images: %w", err)
	}
	account.HasProfileImage = hasProfileImage
	account.HasBannerImage = hasBannerImage

	return &account, nil
}

func deleteAccountByID(ctx context.Context, id string) error {
	statement, err := database.DB().PrepareContext(ctx, "DELETE FROM account WHERE id=$1;")
	if err != nil {
		return err
	}

	defer statement.Close()

	_, err = statement.ExecContext(ctx, id)
	if err != nil {
		return err
	}

	return nil
}

func getUserIDByEmail(ctx context.Context, email string) (uint, error) {
	var id uint
	row := database.DB().QueryRowContext(ctx, "SELECT id FROM account WHERE email = $1;", email)
	err := row.Scan(&id)
	if err != nil {
		return 0, err
	}
	return id, nil
}

// FindUserIDByUsername finds a user ID by username
// Returns 0 if not found
func FindUserIDByUsername(users []User, username string) uint {
	for _, user := range users {
		if user.Username == username {
			return user.ID
		}
	}
	return 0
}
