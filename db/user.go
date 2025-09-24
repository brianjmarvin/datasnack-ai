package db

import (
	"database/sql"
	"fmt"
	"log"
	"strings"
	"time"

	"datasnack/models"
)

// CreateUser creates a new user in the database
func (d *Database) CreateUser(user models.User) (models.User, error) {
	query := `INSERT INTO users (
		user_id, first_name, last_name, full_name, email, 
		subscription_plan, account_validated, created_at, updated_at, 
		last_login, company_id, company_name, credits, is_active
	) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14)
	ON CONFLICT (email) DO NOTHING`

	_, err := d.Connection.Exec(query,
		user.UserID,
		user.FirstName,
		user.LastName,
		user.FullName,
		user.Email,
		user.SubscriptionPlan,
		user.AccountValidated,
		user.CreatedAt,
		user.UpdatedAt,
		user.LastLogin,
		user.CompanyID,
		user.CompanyName,
		user.Credits,
		user.IsActive,
	)
	if err != nil {
		log.Printf("Error creating user: %v", err)
		return models.User{}, fmt.Errorf("failed to create user: %v", err)
	}

	log.Printf("User created successfully: %s", user.UserID)
	return user, nil
}

// GetUser retrieves a user by ID
func (d *Database) GetUser(userID string) (models.User, error) {
	var user models.User
	query := `SELECT 
		user_id, first_name, last_name, full_name, email, 
		subscription_plan, account_validated, created_at, updated_at, 
		last_login, company_id, company_name, credits, is_active
		FROM users WHERE user_id = $1 AND is_active = true`

	row := d.Connection.QueryRow(query, userID)
	err := row.Scan(
		&user.UserID,
		&user.FirstName,
		&user.LastName,
		&user.FullName,
		&user.Email,
		&user.SubscriptionPlan,
		&user.AccountValidated,
		&user.CreatedAt,
		&user.UpdatedAt,
		&user.LastLogin,
		&user.CompanyID,
		&user.CompanyName,
		&user.Credits,
		&user.IsActive,
	)
	if err != nil {
		if err == sql.ErrNoRows {
			return models.User{}, fmt.Errorf("user not found: %s", userID)
		}
		log.Printf("Error getting user: %v", err)
		return models.User{}, fmt.Errorf("failed to get user: %v", err)
	}

	return user, nil
}

// GetUserByEmail retrieves a user by email
func (d *Database) GetUserByEmail(email string) (models.User, error) {
	var user models.User
	query := `SELECT 
		user_id, first_name, last_name, full_name, email, 
		subscription_plan, account_validated, created_at, updated_at, 
		last_login, company_id, company_name, credits, is_active
		FROM users WHERE email = $1 AND is_active = true`

	row := d.Connection.QueryRow(query, email)
	err := row.Scan(
		&user.UserID,
		&user.FirstName,
		&user.LastName,
		&user.FullName,
		&user.Email,
		&user.SubscriptionPlan,
		&user.AccountValidated,
		&user.CreatedAt,
		&user.UpdatedAt,
		&user.LastLogin,
		&user.CompanyID,
		&user.CompanyName,
		&user.Credits,
		&user.IsActive,
	)
	if err != nil {
		if err == sql.ErrNoRows {
			return models.User{}, fmt.Errorf("user not found")
		}
		log.Printf("Error getting user by email: %v", err)
		return models.User{}, fmt.Errorf("failed to get user by email: %v", err)
	}

	return user, nil
}

// GetAllUsers retrieves all active users
func (d *Database) GetAllUsers() ([]models.User, error) {
	query := `SELECT 
		user_id, first_name, last_name, full_name, email, 
		subscription_plan, account_validated, created_at, updated_at, 
		last_login, company_id, company_name, credits, is_active
		FROM users WHERE is_active = true ORDER BY created_at DESC`

	rows, err := d.Connection.Query(query)
	if err != nil {
		log.Printf("Error getting all users: %v", err)
		return nil, fmt.Errorf("failed to get all users: %v", err)
	}
	defer rows.Close()

	var users []models.User
	for rows.Next() {
		var user models.User
		err := rows.Scan(
			&user.UserID,
			&user.FirstName,
			&user.LastName,
			&user.FullName,
			&user.Email,
			&user.SubscriptionPlan,
			&user.AccountValidated,
			&user.CreatedAt,
			&user.UpdatedAt,
			&user.LastLogin,
			&user.CompanyID,
			&user.CompanyName,
			&user.Credits,
			&user.IsActive,
		)
		if err != nil {
			log.Printf("Error scanning user: %v", err)
			continue
		}
		users = append(users, user)
	}

	if err = rows.Err(); err != nil {
		log.Printf("Error iterating users: %v", err)
		return nil, fmt.Errorf("failed to iterate users: %v", err)
	}

	return users, nil
}

// UpdateUser updates an existing user
func (d *Database) UpdateUser(userID string, user models.User) (models.User, error) {
	// Handle name splitting if full name is provided
	if user.FullName != "" {
		nameSplit := strings.Split(user.FullName, " ")
		user.FirstName = nameSplit[0]
		if len(nameSplit) > 1 {
			user.LastName = nameSplit[len(nameSplit)-1]
		}
	}

	user.UpdatedAt = time.Now()

	query := `UPDATE users SET
		first_name = $1, last_name = $2, full_name = $3, email = $4,
		subscription_plan = $5, account_validated = $6, updated_at = $7,
		company_id = $8, company_name = $9, credits = $10, is_active = $11
		WHERE user_id = $12`

	result, err := d.Connection.Exec(query,
		user.FirstName,
		user.LastName,
		user.FullName,
		user.Email,
		user.SubscriptionPlan,
		user.AccountValidated,
		user.UpdatedAt,
		user.CompanyID,
		user.CompanyName,
		user.Credits,
		user.IsActive,
		userID,
	)
	if err != nil {
		log.Printf("Error updating user: %v", err)
		return models.User{}, fmt.Errorf("failed to update user: %v", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		log.Printf("Error getting rows affected: %v", err)
		return models.User{}, fmt.Errorf("failed to get rows affected: %v", err)
	}

	if rowsAffected == 0 {
		return models.User{}, fmt.Errorf("user not found or no changes made")
	}

	log.Printf("User updated successfully: %s", userID)
	return user, nil
}

// UpdateUserCredits updates a user's credits
func (d *Database) UpdateUserCredits(userID string, credits int) error {
	query := `UPDATE users SET
		credits = credits + $1, updated_at = $2
		WHERE user_id = $3`

	result, err := d.Connection.Exec(query, credits, time.Now(), userID)
	if err != nil {
		log.Printf("Error updating user credits: %v", err)
		return fmt.Errorf("failed to update user credits: %v", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		log.Printf("Error getting rows affected: %v", err)
		return fmt.Errorf("failed to get rows affected: %v", err)
	}

	if rowsAffected == 0 {
		return fmt.Errorf("user not found")
	}

	log.Printf("User credits updated successfully: %s (+%d)", userID, credits)
	return nil
}

// UpdateLastLogin updates the user's last login time
func (d *Database) UpdateLastLogin(userID string) error {
	query := `UPDATE users SET last_login = $1 WHERE user_id = $2`

	_, err := d.Connection.Exec(query, time.Now(), userID)
	if err != nil {
		log.Printf("Error updating last login: %v", err)
		return fmt.Errorf("failed to update last login: %v", err)
	}

	return nil
}

// DeleteUser soft deletes a user by setting is_active to false
func (d *Database) DeleteUser(userID string) error {
	query := `UPDATE users SET is_active = false, updated_at = $1 WHERE user_id = $2`

	result, err := d.Connection.Exec(query, time.Now(), userID)
	if err != nil {
		log.Printf("Error deleting user: %v", err)
		return fmt.Errorf("failed to delete user: %v", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		log.Printf("Error getting rows affected: %v", err)
		return fmt.Errorf("failed to get rows affected: %v", err)
	}

	if rowsAffected == 0 {
		return fmt.Errorf("user not found")
	}

	log.Printf("User deleted successfully: %s", userID)
	return nil
}
