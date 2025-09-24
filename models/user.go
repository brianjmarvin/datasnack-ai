package models

import (
	"time"
)

// User represents a user in the system
type User struct {
	UserID           string           `json:"userID,omitempty"`
	FirstName        string           `json:"firstName,omitempty"`
	LastName         string           `json:"lastName,omitempty"`
	FullName         string           `json:"fullName,omitempty"`
	Email            string           `json:"email,omitempty"`
	SubscriptionPlan SubscriptionPlan `json:"subscriptionPlan,omitempty"`
	AccountValidated bool             `json:"accountValidated"`
	CreatedAt        time.Time        `json:"createdAt,omitempty"`
	UpdatedAt        time.Time        `json:"updatedAt,omitempty"`
	LastLogin        time.Time        `json:"lastLogin,omitempty"`
	CompanyID        string           `json:"companyID,omitempty"`
	CompanyName      string           `json:"companyName,omitempty"`
	Credits          int              `json:"credits,omitempty"`
	IsActive         bool             `json:"isActive"`
}

// SubscriptionPlan represents different subscription tiers
type SubscriptionPlan string

const (
	SPARK SubscriptionPlan = "SPARK"
	BOOST SubscriptionPlan = "BOOST"
	SCALE SubscriptionPlan = "SCALE"
)

// UserRequest represents the request payload for user operations
type UserRequest struct {
	FirstName        string           `json:"firstName,omitempty"`
	LastName         string           `json:"lastName,omitempty"`
	FullName         string           `json:"fullName,omitempty"`
	Email            string           `json:"email,omitempty"`
	SubscriptionPlan SubscriptionPlan `json:"subscriptionPlan,omitempty"`
	AccountValidated bool             `json:"accountValidated"`
	CompanyID        string           `json:"companyID,omitempty"`
	CompanyName      string           `json:"companyName,omitempty"`
	Credits          int              `json:"credits,omitempty"`
	IsActive         bool             `json:"isActive"`
}

// UserResponse represents the response from user operations
type UserResponse struct {
	Success bool   `json:"success"`
	Message string `json:"message,omitempty"`
	User    *User  `json:"user,omitempty"`
	Error   string `json:"error,omitempty"`
}

// UsersResponse represents the response for multiple users
type UsersResponse struct {
	Success bool   `json:"success"`
	Message string `json:"message,omitempty"`
	Users   []User `json:"users,omitempty"`
	Error   string `json:"error,omitempty"`
}

// UserClient represents the authenticated user context
type UserClient struct {
	UserID          string
	IsAuthenticated bool
	IsDemo          bool
}
