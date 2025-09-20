package models

import (
	"time"

	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"
)

// User represents the user entity
type User struct {
	ID            uint           `json:"id" gorm:"primaryKey"`
	Email         string         `json:"email" gorm:"unique;not null"`
	Username      string         `json:"username" gorm:"unique;not null"`
	Password      string         `json:"-" gorm:"not null"`
	FirstName     string         `json:"first_name"`
	LastName      string         `json:"last_name"`
	Avatar        string         `json:"avatar"`
	Provider      string         `json:"provider" gorm:"default:'local'"` // 'local', 'google', 'github', etc.
	ProviderID    string         `json:"provider_id"`
	EmailVerified bool           `json:"email_verified" gorm:"default:false"`
	IsActive      bool           `json:"is_active" gorm:"default:true"`
	CreatedAt     time.Time      `json:"created_at"`
	UpdatedAt     time.Time      `json:"updated_at"`
	DeletedAt     gorm.DeletedAt `json:"-" gorm:"index"`
}

// OAuthState represents OAuth2 state for security
type OAuthState struct {
	ID           uint      `json:"id" gorm:"primaryKey"`
	State        string    `json:"state" gorm:"unique;not null"`
	CodeVerifier string    `json:"code_verifier" gorm:"not null"` // For PKCE
	Nonce        string    `json:"nonce"`                         // For OIDC
	UserID       *uint     `json:"user_id"`                       // Optional: for logged-in users
	ExpiresAt    time.Time `json:"expires_at" gorm:"not null"`
	CreatedAt    time.Time `json:"created_at"`
}

// RefreshToken represents refresh token storage
type RefreshToken struct {
	ID        uint           `json:"id" gorm:"primaryKey"`
	UserID    uint           `json:"user_id" gorm:"not null"`
	Token     string         `json:"token" gorm:"unique;not null"`
	ExpiresAt time.Time      `json:"expires_at" gorm:"not null"`
	CreatedAt time.Time      `json:"created_at"`
	DeletedAt gorm.DeletedAt `json:"-" gorm:"index"`
	User      User           `json:"user" gorm:"foreignKey:UserID"`
}

// BeforeCreate is a GORM hook to hash password before creating user
func (u *User) BeforeCreate(tx *gorm.DB) error {
	if u.Password != "" && u.Provider == "local" {
		hashedPassword, err := bcrypt.GenerateFromPassword([]byte(u.Password), bcrypt.DefaultCost)
		if err != nil {
			return err
		}
		u.Password = string(hashedPassword)
	}
	return nil
}

// CheckPassword verifies the provided password against the stored hash
func (u *User) CheckPassword(password string) bool {
	err := bcrypt.CompareHashAndPassword([]byte(u.Password), []byte(password))
	return err == nil
}

// UserRepository defines the interface for user data access
type UserRepository interface {
	Create(user *User) error
	GetByID(id uint) (*User, error)
	GetByEmail(email string) (*User, error)
	GetByUsername(username string) (*User, error)
	GetByProviderID(provider, providerID string) (*User, error)
	Update(user *User) error
	Delete(id uint) error
	List(limit, offset int) ([]*User, error)
}

// OAuthRepository defines the interface for OAuth state management
type OAuthRepository interface {
	CreateState(state *OAuthState) error
	GetStateByState(state string) (*OAuthState, error)
	DeleteState(state string) error
	CleanupExpiredStates() error
}

// RefreshTokenRepository defines the interface for refresh token management
type RefreshTokenRepository interface {
	Create(token *RefreshToken) error
	GetByToken(token string) (*RefreshToken, error)
	GetByUserID(userID uint) ([]*RefreshToken, error)
	Delete(token string) error
	DeleteByUserID(userID uint) error
	CleanupExpiredTokens() error
}
