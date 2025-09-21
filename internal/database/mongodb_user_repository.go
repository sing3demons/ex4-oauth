package database

import (
	"context"
	"errors"
	"time"

	"ex4-oauth2/internal/models"

	"github.com/google/uuid"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

// MongoUserRepository implements UserRepository for MongoDB
type MongoUserRepository struct {
	collection *mongo.Collection
}

// NewMongoUserRepository creates a new MongoDB user repository
func NewMongoUserRepository(db *mongo.Database) models.UserRepository {
	return &MongoUserRepository{
		collection: db.Collection("users"),
	}
}

// Create creates a new user
func (r *MongoUserRepository) Create(user *models.User) error {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	user.SetDefaults()
	if err := user.HashPassword(); err != nil {
		return err
	}

	_, err := r.collection.InsertOne(ctx, user)
	return err
}

// GetByID retrieves a user by ID
func (r *MongoUserRepository) GetByID(id string) (*models.User, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	var user models.User
	err := r.collection.FindOne(ctx, bson.M{"_id": id}).Decode(&user)
	if err != nil {
		if errors.Is(err, mongo.ErrNoDocuments) {
			return nil, errors.New("user not found")
		}
		return nil, err
	}
	return &user, nil
}

// GetByEmail retrieves a user by email
func (r *MongoUserRepository) GetByEmail(email string) (*models.User, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	var user models.User
	err := r.collection.FindOne(ctx, bson.M{"email": email}).Decode(&user)
	if err != nil {
		if errors.Is(err, mongo.ErrNoDocuments) {
			return nil, errors.New("user not found")
		}
		return nil, err
	}
	return &user, nil
}

// GetByUsername retrieves a user by username
func (r *MongoUserRepository) GetByUsername(username string) (*models.User, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	var user models.User
	err := r.collection.FindOne(ctx, bson.M{"username": username}).Decode(&user)
	if err != nil {
		if errors.Is(err, mongo.ErrNoDocuments) {
			return nil, errors.New("user not found")
		}
		return nil, err
	}
	return &user, nil
}

// GetByProviderID retrieves a user by provider and provider ID
func (r *MongoUserRepository) GetByProviderID(provider, providerID string) (*models.User, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	var user models.User
	err := r.collection.FindOne(ctx, bson.M{
		"provider":    provider,
		"provider_id": providerID,
	}).Decode(&user)
	if err != nil {
		if errors.Is(err, mongo.ErrNoDocuments) {
			return nil, errors.New("user not found")
		}
		return nil, err
	}
	return &user, nil
}

// Update updates a user
func (r *MongoUserRepository) Update(user *models.User) error {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	user.UpdatedAt = time.Now()

	_, err := r.collection.ReplaceOne(ctx, bson.M{"_id": user.ID}, user)
	return err
}

// Delete deletes a user (soft delete by setting is_active to false)
func (r *MongoUserRepository) Delete(id string) error {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	_, err := r.collection.UpdateOne(
		ctx,
		bson.M{"_id": id},
		bson.M{"$set": bson.M{
			"is_active":  false,
			"updated_at": time.Now(),
		}},
	)
	return err
}

// List retrieves users with pagination
func (r *MongoUserRepository) List(limit, offset int) ([]*models.User, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	opts := options.Find().SetLimit(int64(limit)).SetSkip(int64(offset)).SetSort(bson.M{"created_at": -1})
	cursor, err := r.collection.Find(ctx, bson.M{}, opts)
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)

	var users []*models.User
	for cursor.Next(ctx) {
		var user models.User
		if err := cursor.Decode(&user); err != nil {
			return nil, err
		}
		users = append(users, &user)
	}

	return users, cursor.Err()
}

// MongoOAuthRepository implements OAuthRepository for MongoDB
type MongoOAuthRepository struct {
	collection *mongo.Collection
}

// NewMongoOAuthRepository creates a new MongoDB OAuth repository
func NewMongoOAuthRepository(db *mongo.Database) models.OAuthRepository {
	return &MongoOAuthRepository{
		collection: db.Collection("oauth_states"),
	}
}

// CreateState creates a new OAuth state
func (r *MongoOAuthRepository) CreateState(state *models.OAuthState) error {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if state.ID == "" {
		state.ID = uuid.New().String()
	}
	if state.CreatedAt.IsZero() {
		state.CreatedAt = time.Now()
	}

	_, err := r.collection.InsertOne(ctx, state)
	return err
}

// GetStateByState retrieves OAuth state by state string
func (r *MongoOAuthRepository) GetStateByState(state string) (*models.OAuthState, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	var oauthState models.OAuthState
	err := r.collection.FindOne(ctx, bson.M{"state": state}).Decode(&oauthState)
	if err != nil {
		if errors.Is(err, mongo.ErrNoDocuments) {
			return nil, errors.New("state not found")
		}
		return nil, err
	}
	return &oauthState, nil
}

// DeleteState deletes OAuth state by state string
func (r *MongoOAuthRepository) DeleteState(state string) error {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	_, err := r.collection.DeleteOne(ctx, bson.M{"state": state})
	return err
}

// CleanupExpiredStates removes expired OAuth states
func (r *MongoOAuthRepository) CleanupExpiredStates() error {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	_, err := r.collection.DeleteMany(ctx, bson.M{
		"expires_at": bson.M{"$lt": time.Now()},
	})
	return err
}

// MongoRefreshTokenRepository implements RefreshTokenRepository for MongoDB
type MongoRefreshTokenRepository struct {
	collection *mongo.Collection
}

// NewMongoRefreshTokenRepository creates a new MongoDB refresh token repository
func NewMongoRefreshTokenRepository(db *mongo.Database) models.RefreshTokenRepository {
	return &MongoRefreshTokenRepository{
		collection: db.Collection("refresh_tokens"),
	}
}

// Create creates a new refresh token
func (r *MongoRefreshTokenRepository) Create(token *models.RefreshToken) error {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	token.ID = uuid.New().String()

	if token.CreatedAt.IsZero() {
		token.CreatedAt = time.Now()
	}

	_, err := r.collection.InsertOne(ctx, token)
	return err
}

// GetByToken retrieves refresh token by token string
func (r *MongoRefreshTokenRepository) GetByToken(token string) (*models.RefreshToken, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	var refreshToken models.RefreshToken
	err := r.collection.FindOne(ctx, bson.M{"token": token}).Decode(&refreshToken)
	if err != nil {
		if errors.Is(err, mongo.ErrNoDocuments) {
			return nil, errors.New("token not found")
		}
		return nil, err
	}
	return &refreshToken, nil
}

// GetByUserID retrieves all refresh tokens for a user
func (r *MongoRefreshTokenRepository) GetByUserID(userID string) ([]*models.RefreshToken, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	cursor, err := r.collection.Find(ctx, bson.M{"user_id": userID})
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)

	var tokens []*models.RefreshToken
	for cursor.Next(ctx) {
		var token models.RefreshToken
		if err := cursor.Decode(&token); err != nil {
			return nil, err
		}
		tokens = append(tokens, &token)
	}

	return tokens, cursor.Err()
}

// Delete deletes refresh token by token string
func (r *MongoRefreshTokenRepository) Delete(token string) error {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	_, err := r.collection.DeleteOne(ctx, bson.M{"token": token})
	return err
}

// DeleteByUserID deletes all refresh tokens for a user
func (r *MongoRefreshTokenRepository) DeleteByUserID(userID string) error {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	_, err := r.collection.DeleteMany(ctx, bson.M{"user_id": userID})
	return err
}

// CleanupExpiredTokens removes expired refresh tokens
func (r *MongoRefreshTokenRepository) CleanupExpiredTokens() error {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	_, err := r.collection.DeleteMany(ctx, bson.M{
		"expires_at": bson.M{"$lt": time.Now()},
	})
	return err
}

// MongoEmailVerificationRepository implements EmailVerificationRepository for MongoDB
type MongoEmailVerificationRepository struct {
	collection *mongo.Collection
}

// NewMongoEmailVerificationRepository creates a new MongoDB email verification repository
func NewMongoEmailVerificationRepository(db *mongo.Database) models.EmailVerificationRepository {
	return &MongoEmailVerificationRepository{
		collection: db.Collection("email_verifications"),
	}
}

// Create creates a new email verification record
func (r *MongoEmailVerificationRepository) Create(verification *models.EmailVerification) error {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	verification.ID = uuid.New().String()
	if verification.CreatedAt.IsZero() {
		verification.CreatedAt = time.Now()
	}

	_, err := r.collection.InsertOne(ctx, verification)
	return err
}

// GetByToken retrieves email verification by token
func (r *MongoEmailVerificationRepository) GetByToken(token string) (*models.EmailVerification, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	var verification models.EmailVerification
	err := r.collection.FindOne(ctx, bson.M{"token": token}).Decode(&verification)
	if err != nil {
		if errors.Is(err, mongo.ErrNoDocuments) {
			return nil, errors.New("verification token not found")
		}
		return nil, err
	}
	return &verification, nil
}

// GetByUserID retrieves email verification by user ID and type
func (r *MongoEmailVerificationRepository) GetByUserID(userID string, verificationType string) (*models.EmailVerification, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	var verification models.EmailVerification
	err := r.collection.FindOne(ctx, bson.M{
		"user_id":     userID,
		"type":        verificationType,
		"verified_at": nil,
	}).Decode(&verification)
	if err != nil {
		if errors.Is(err, mongo.ErrNoDocuments) {
			return nil, errors.New("verification not found")
		}
		return nil, err
	}
	return &verification, nil
}

// GetByEmail retrieves email verification by email and type
func (r *MongoEmailVerificationRepository) GetByEmail(email string, verificationType string) (*models.EmailVerification, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	var verification models.EmailVerification
	err := r.collection.FindOne(ctx, bson.M{
		"email":       email,
		"type":        verificationType,
		"verified_at": nil,
	}).Decode(&verification)
	if err != nil {
		if errors.Is(err, mongo.ErrNoDocuments) {
			return nil, errors.New("verification not found")
		}
		return nil, err
	}
	return &verification, nil
}

// MarkAsVerified marks email verification as verified
func (r *MongoEmailVerificationRepository) MarkAsVerified(token string) error {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	now := time.Now()
	_, err := r.collection.UpdateOne(
		ctx,
		bson.M{"token": token},
		bson.M{"$set": bson.M{"verified_at": &now}},
	)
	return err
}

// Delete deletes email verification by token
func (r *MongoEmailVerificationRepository) Delete(token string) error {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	_, err := r.collection.DeleteOne(ctx, bson.M{"token": token})
	return err
}

// DeleteByUserID deletes email verifications by user ID and type
func (r *MongoEmailVerificationRepository) DeleteByUserID(userID string, verificationType string) error {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	_, err := r.collection.DeleteMany(ctx, bson.M{
		"user_id": userID,
		"type":    verificationType,
	})
	return err
}

// CleanupExpiredTokens removes expired email verification tokens
func (r *MongoEmailVerificationRepository) CleanupExpiredTokens() error {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	_, err := r.collection.DeleteMany(ctx, bson.M{
		"expires_at": bson.M{"$lt": time.Now()},
	})
	return err
}
