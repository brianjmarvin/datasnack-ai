package db

import (
	"database/sql"
	"fmt"
	"log"
	"time"

	"datasnack/models"

	"github.com/google/uuid"
)

// CreateContactUsTable creates the contact_us table if it doesn't exist
func (d *Database) CreateContactUsTable() error {
	query := `CREATE TABLE IF NOT EXISTS contact_us (
		contact_id TEXT PRIMARY KEY,
		full_name TEXT DEFAULT '',
		company_name TEXT DEFAULT '',
		note TEXT DEFAULT '',
		email TEXT DEFAULT '',
		chatbot_url TEXT DEFAULT '',
		created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
		updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
	)`

	_, err := d.Connection.Exec(query)
	if err != nil {
		return fmt.Errorf("failed to create contact_us table: %v", err)
	}

	log.Println("Contact us table created/verified successfully")
	return nil
}

// CreateContactUs creates a new contact message in the database
func (d *Database) CreateContactUs(contact models.ContactUs) (models.ContactUs, error) {
	// Generate contact ID if not provided
	if contact.ContactID == "" {
		contact.ContactID = uuid.New().String()
	}

	contact.CreatedAt = time.Now()
	contact.UpdatedAt = time.Now()

	query := `INSERT INTO contact_us (
		contact_id, full_name, company_name, note, email, chatbot_url, created_at, updated_at
	) VALUES ($1, $2, $3, $4, $5, $6, $7, $8)`

	_, err := d.Connection.Exec(query,
		contact.ContactID,
		contact.FullName,
		contact.CompanyName,
		contact.Note,
		contact.Email,
		contact.ChatbotURL,
		contact.CreatedAt,
		contact.UpdatedAt,
	)
	if err != nil {
		log.Printf("Error creating contact message: %v", err)
		return models.ContactUs{}, fmt.Errorf("failed to create contact message: %v", err)
	}

	log.Printf("Contact message created successfully: %s", contact.ContactID)
	return contact, nil
}

// GetContactUs retrieves a contact message by ID
func (d *Database) GetContactUs(contactID string) (models.ContactUs, error) {
	var contact models.ContactUs
	query := `SELECT 
		contact_id, full_name, company_name, note, email, chatbot_url, created_at, updated_at
		FROM contact_us WHERE contact_id = $1`

	row := d.Connection.QueryRow(query, contactID)
	err := row.Scan(
		&contact.ContactID,
		&contact.FullName,
		&contact.CompanyName,
		&contact.Note,
		&contact.Email,
		&contact.ChatbotURL,
		&contact.CreatedAt,
		&contact.UpdatedAt,
	)
	if err != nil {
		if err == sql.ErrNoRows {
			return models.ContactUs{}, fmt.Errorf("contact message not found")
		}
		log.Printf("Error getting contact message: %v", err)
		return models.ContactUs{}, fmt.Errorf("failed to get contact message: %v", err)
	}

	return contact, nil
}

// GetAllContactUs retrieves all contact messages
func (d *Database) GetAllContactUs() ([]models.ContactUs, error) {
	query := `SELECT 
		contact_id, full_name, company_name, note, email, chatbot_url, created_at, updated_at
		FROM contact_us ORDER BY created_at DESC`

	rows, err := d.Connection.Query(query)
	if err != nil {
		log.Printf("Error getting all contact messages: %v", err)
		return nil, fmt.Errorf("failed to get all contact messages: %v", err)
	}
	defer rows.Close()

	var contacts []models.ContactUs
	for rows.Next() {
		var contact models.ContactUs
		err := rows.Scan(
			&contact.ContactID,
			&contact.FullName,
			&contact.CompanyName,
			&contact.Note,
			&contact.Email,
			&contact.ChatbotURL,
			&contact.CreatedAt,
			&contact.UpdatedAt,
		)
		if err != nil {
			log.Printf("Error scanning contact message: %v", err)
			continue
		}
		contacts = append(contacts, contact)
	}

	if err = rows.Err(); err != nil {
		log.Printf("Error iterating contact messages: %v", err)
		return nil, fmt.Errorf("failed to iterate contact messages: %v", err)
	}

	return contacts, nil
}
