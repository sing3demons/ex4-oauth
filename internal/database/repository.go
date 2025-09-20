package database

import (
	"errors"
	"time"

	"ex4-oauth2/internal/models"

	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

// Database wraps the GORM database instance
type Database struct {
	db *gorm.DB
}

// NewDatabase creates a new database connection
func NewDatabase(databaseURL string) (*Database, error) {
	db, err := gorm.Open(sqlite.Open(databaseURL), &gorm.Config{})
	if err != nil {
		return nil, err
	}

	// Auto migrate the schemas
	err = db.AutoMigrate(
		&models.User{},
		&models.OAuthState{},
		&models.RefreshToken{},
		&models.OAuth2Client{},
		&models.OAuth2AuthorizationCode{},
		&models.OAuth2AccessToken{},
	)
	if err != nil {
		return nil, err
	}

	return &Database{db: db}, nil
}

// GetDB returns the underlying GORM database instance
func (d *Database) GetDB() *gorm.DB {
	return d.db
}

// Close closes the database connection
func (d *Database) Close() error {
	sqlDB, err := d.db.DB()
	if err != nil {
		return err
	}
	return sqlDB.Close()
}

// UserRepository implements the UserRepository interface
type UserRepository struct {
	db *gorm.DB
}

// NewUserRepository creates a new user repository
func NewUserRepository(db *gorm.DB) models.UserRepository {
	return &UserRepository{db: db}
}

// Create creates a new user
func (r *UserRepository) Create(user *models.User) error {
	return r.db.Create(user).Error
}

// GetByID retrieves a user by ID
func (r *UserRepository) GetByID(id uint) (*models.User, error) {
	var user models.User
	err := r.db.First(&user, id).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.New("user not found")
		}
		return nil, err
	}
	return &user, nil
}

// GetByEmail retrieves a user by email
func (r *UserRepository) GetByEmail(email string) (*models.User, error) {
	var user models.User
	err := r.db.Where("email = ?", email).First(&user).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.New("user not found")
		}
		return nil, err
	}
	return &user, nil
}

// GetByUsername retrieves a user by username
func (r *UserRepository) GetByUsername(username string) (*models.User, error) {
	var user models.User
	err := r.db.Where("username = ?", username).First(&user).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.New("user not found")
		}
		return nil, err
	}
	return &user, nil
}

// GetByProviderID retrieves a user by provider and provider ID
func (r *UserRepository) GetByProviderID(provider, providerID string) (*models.User, error) {
	var user models.User
	err := r.db.Where("provider = ? AND provider_id = ?", provider, providerID).First(&user).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.New("user not found")
		}
		return nil, err
	}
	return &user, nil
}

// Update updates a user
func (r *UserRepository) Update(user *models.User) error {
	return r.db.Save(user).Error
}

// Delete soft deletes a user
func (r *UserRepository) Delete(id uint) error {
	return r.db.Delete(&models.User{}, id).Error
}

// List retrieves users with pagination
func (r *UserRepository) List(limit, offset int) ([]*models.User, error) {
	var users []*models.User
	err := r.db.Limit(limit).Offset(offset).Find(&users).Error
	return users, err
}

// OAuthRepository implements the OAuthRepository interface
type OAuthRepository struct {
	db *gorm.DB
}

// NewOAuthRepository creates a new OAuth repository
func NewOAuthRepository(db *gorm.DB) models.OAuthRepository {
	return &OAuthRepository{db: db}
}

// CreateState creates a new OAuth state
func (r *OAuthRepository) CreateState(state *models.OAuthState) error {
	return r.db.Create(state).Error
}

// GetStateByState retrieves OAuth state by state string
func (r *OAuthRepository) GetStateByState(state string) (*models.OAuthState, error) {
	var oauthState models.OAuthState
	err := r.db.Where("state = ?", state).First(&oauthState).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.New("state not found")
		}
		return nil, err
	}
	return &oauthState, nil
}

// DeleteState deletes OAuth state by state string
func (r *OAuthRepository) DeleteState(state string) error {
	return r.db.Where("state = ?", state).Delete(&models.OAuthState{}).Error
}

// CleanupExpiredStates removes expired OAuth states
func (r *OAuthRepository) CleanupExpiredStates() error {
	return r.db.Where("expires_at < ?", time.Now()).Delete(&models.OAuthState{}).Error
}

// RefreshTokenRepository implements the RefreshTokenRepository interface
type RefreshTokenRepository struct {
	db *gorm.DB
}

// NewRefreshTokenRepository creates a new refresh token repository
func NewRefreshTokenRepository(db *gorm.DB) models.RefreshTokenRepository {
	return &RefreshTokenRepository{db: db}
}

// Create creates a new refresh token
func (r *RefreshTokenRepository) Create(token *models.RefreshToken) error {
	return r.db.Create(token).Error
}

// GetByToken retrieves refresh token by token string
func (r *RefreshTokenRepository) GetByToken(token string) (*models.RefreshToken, error) {
	var refreshToken models.RefreshToken
	err := r.db.Preload("User").Where("token = ?", token).First(&refreshToken).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.New("token not found")
		}
		return nil, err
	}
	return &refreshToken, nil
}

// GetByUserID retrieves all refresh tokens for a user
func (r *RefreshTokenRepository) GetByUserID(userID uint) ([]*models.RefreshToken, error) {
	var tokens []*models.RefreshToken
	err := r.db.Where("user_id = ?", userID).Find(&tokens).Error
	return tokens, err
}

// Delete deletes refresh token by token string
func (r *RefreshTokenRepository) Delete(token string) error {
	return r.db.Where("token = ?", token).Delete(&models.RefreshToken{}).Error
}

// DeleteByUserID deletes all refresh tokens for a user
func (r *RefreshTokenRepository) DeleteByUserID(userID uint) error {
	return r.db.Where("user_id = ?", userID).Delete(&models.RefreshToken{}).Error
}

// CleanupExpiredTokens removes expired refresh tokens
func (r *RefreshTokenRepository) CleanupExpiredTokens() error {
	return r.db.Where("expires_at < ?", time.Now()).Delete(&models.RefreshToken{}).Error
}
