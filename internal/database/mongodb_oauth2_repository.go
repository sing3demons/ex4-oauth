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

// MongoOAuth2ClientRepository implements OAuth2ClientRepository for MongoDB
type MongoOAuth2ClientRepository struct {
	collection *mongo.Collection
}

// NewMongoOAuth2ClientRepository creates a new MongoDB OAuth2 client repository
func NewMongoOAuth2ClientRepository(db *mongo.Database) models.OAuth2ClientRepository {
	return &MongoOAuth2ClientRepository{
		collection: db.Collection("oauth2_clients"),
	}
}

// Create creates a new OAuth2 client
func (r *MongoOAuth2ClientRepository) Create(client *models.OAuth2Client) error {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	client.SetDefaults()

	_, err := r.collection.InsertOne(ctx, client)
	return err
}

// GetByID retrieves an OAuth2 client by ID
func (r *MongoOAuth2ClientRepository) GetByID(id string) (*models.OAuth2Client, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	var client models.OAuth2Client
	err := r.collection.FindOne(ctx, bson.M{"_id": id}).Decode(&client)
	if err != nil {
		if errors.Is(err, mongo.ErrNoDocuments) {
			return nil, errors.New("OAuth2 client not found")
		}
		return nil, err
	}
	return &client, nil
}

// GetByClientID retrieves an OAuth2 client by client ID
func (r *MongoOAuth2ClientRepository) GetByClientID(clientID string) (*models.OAuth2Client, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	var client models.OAuth2Client
	err := r.collection.FindOne(ctx, bson.M{"client_id": clientID}).Decode(&client)
	if err != nil {
		if errors.Is(err, mongo.ErrNoDocuments) {
			return nil, errors.New("OAuth2 client not found")
		}
		return nil, err
	}
	return &client, nil
}

// Update updates an OAuth2 client
func (r *MongoOAuth2ClientRepository) Update(client *models.OAuth2Client) error {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	client.UpdatedAt = time.Now()

	_, err := r.collection.ReplaceOne(ctx, bson.M{"_id": client.ID}, client)
	return err
}

// Delete deletes an OAuth2 client
func (r *MongoOAuth2ClientRepository) Delete(id string) error {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	_, err := r.collection.DeleteOne(ctx, bson.M{"_id": id})
	return err
}

// List retrieves OAuth2 clients with pagination
func (r *MongoOAuth2ClientRepository) List(limit, offset int) ([]*models.OAuth2Client, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	opts := options.Find().SetLimit(int64(limit)).SetSkip(int64(offset)).SetSort(bson.M{"created_at": -1})
	cursor, err := r.collection.Find(ctx, bson.M{}, opts)
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)

	var clients []*models.OAuth2Client
	for cursor.Next(ctx) {
		var client models.OAuth2Client
		if err := cursor.Decode(&client); err != nil {
			return nil, err
		}
		clients = append(clients, &client)
	}

	return clients, cursor.Err()
}

// ValidateClientCredentials validates client credentials for confidential clients
func (r *MongoOAuth2ClientRepository) ValidateClientCredentials(clientID, clientSecret string) (*models.OAuth2Client, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	var client models.OAuth2Client
	err := r.collection.FindOne(ctx, bson.M{
		"client_id":     clientID,
		"client_secret": clientSecret,
		"is_active":     true,
		"is_public":     false, // Only confidential clients have secrets
	}).Decode(&client)

	if err != nil {
		if errors.Is(err, mongo.ErrNoDocuments) {
			return nil, errors.New("invalid client credentials")
		}
		return nil, err
	}
	return &client, nil
}

// ValidatePublicClient validates public client (no secret required)
func (r *MongoOAuth2ClientRepository) ValidatePublicClient(clientID string) (*models.OAuth2Client, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	var client models.OAuth2Client
	err := r.collection.FindOne(ctx, bson.M{
		"client_id": clientID,
		"is_active": true,
		"is_public": true, // Only public clients
	}).Decode(&client)

	if err != nil {
		if errors.Is(err, mongo.ErrNoDocuments) {
			return nil, errors.New("invalid public client")
		}
		return nil, err
	}
	return &client, nil
}

// MongoOAuth2AuthorizationCodeRepository implements OAuth2AuthorizationCodeRepository for MongoDB
type MongoOAuth2AuthorizationCodeRepository struct {
	collection *mongo.Collection
}

// NewMongoOAuth2AuthorizationCodeRepository creates a new MongoDB authorization code repository
func NewMongoOAuth2AuthorizationCodeRepository(db *mongo.Database) models.OAuth2AuthorizationCodeRepository {
	return &MongoOAuth2AuthorizationCodeRepository{
		collection: db.Collection("oauth2_authorization_codes"),
	}
}

// Create creates a new authorization code
func (r *MongoOAuth2AuthorizationCodeRepository) Create(code *models.OAuth2AuthorizationCode) error {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	code.ID = uuid.New().String()

	if code.CreatedAt.IsZero() {
		code.CreatedAt = time.Now()
	}

	_, err := r.collection.InsertOne(ctx, code)
	return err
}

// GetByCode retrieves authorization code by code string
func (r *MongoOAuth2AuthorizationCodeRepository) GetByCode(code string) (*models.OAuth2AuthorizationCode, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	var authCode models.OAuth2AuthorizationCode
	err := r.collection.FindOne(ctx, bson.M{"code": code}).Decode(&authCode)
	if err != nil {
		if errors.Is(err, mongo.ErrNoDocuments) {
			return nil, errors.New("authorization code not found")
		}
		return nil, err
	}

	// Check if expired
	if authCode.ExpiresAt.Before(time.Now()) {
		return nil, errors.New("authorization code expired")
	}

	// Check if already used
	if authCode.Used {
		return nil, errors.New("authorization code already used")
	}

	return &authCode, nil
}

// MarkAsUsed marks authorization code as used
func (r *MongoOAuth2AuthorizationCodeRepository) MarkAsUsed(code string) error {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	_, err := r.collection.UpdateOne(
		ctx,
		bson.M{"code": code},
		bson.M{"$set": bson.M{"used": true}},
	)
	return err
}

// CleanupExpiredCodes removes expired authorization codes
func (r *MongoOAuth2AuthorizationCodeRepository) CleanupExpiredCodes() error {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	_, err := r.collection.DeleteMany(ctx, bson.M{
		"expires_at": bson.M{"$lt": time.Now()},
	})
	return err
}

// MongoOAuth2AccessTokenRepository implements OAuth2AccessTokenRepository for MongoDB
type MongoOAuth2AccessTokenRepository struct {
	collection *mongo.Collection
}

// NewMongoOAuth2AccessTokenRepository creates a new MongoDB access token repository
func NewMongoOAuth2AccessTokenRepository(db *mongo.Database) models.OAuth2AccessTokenRepository {
	return &MongoOAuth2AccessTokenRepository{
		collection: db.Collection("oauth2_access_tokens"),
	}
}

// Create creates a new access token
func (r *MongoOAuth2AccessTokenRepository) Create(token *models.OAuth2AccessToken) error {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	token.ID = uuid.New().String()
	if token.CreatedAt.IsZero() {
		token.CreatedAt = time.Now()
	}

	_, err := r.collection.InsertOne(ctx, token)
	return err
}

// GetByToken retrieves access token by token string
func (r *MongoOAuth2AccessTokenRepository) GetByToken(token string) (*models.OAuth2AccessToken, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	var accessToken models.OAuth2AccessToken
	err := r.collection.FindOne(ctx, bson.M{"token": token}).Decode(&accessToken)
	if err != nil {
		if errors.Is(err, mongo.ErrNoDocuments) {
			return nil, errors.New("access token not found")
		}
		return nil, err
	}

	// Check if expired
	if accessToken.ExpiresAt.Before(time.Now()) {
		return nil, errors.New("access token expired")
	}

	return &accessToken, nil
}

// GetByUserAndClient retrieves access tokens by user and client
func (r *MongoOAuth2AccessTokenRepository) GetByUserAndClient(userID string, clientID string) ([]*models.OAuth2AccessToken, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	cursor, err := r.collection.Find(ctx, bson.M{
		"user_id":   userID,
		"client_id": clientID,
	})
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)

	var tokens []*models.OAuth2AccessToken
	for cursor.Next(ctx) {
		var token models.OAuth2AccessToken
		if err := cursor.Decode(&token); err != nil {
			return nil, err
		}
		tokens = append(tokens, &token)
	}

	return tokens, cursor.Err()
}

// Delete deletes access token by token string
func (r *MongoOAuth2AccessTokenRepository) Delete(token string) error {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	_, err := r.collection.DeleteOne(ctx, bson.M{"token": token})
	return err
}

// DeleteByUserAndClient deletes access tokens by user and client
func (r *MongoOAuth2AccessTokenRepository) DeleteByUserAndClient(userID string, clientID string) error {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	_, err := r.collection.DeleteMany(ctx, bson.M{
		"user_id":   userID,
		"client_id": clientID,
	})
	return err
}

// CleanupExpiredTokens removes expired access tokens
func (r *MongoOAuth2AccessTokenRepository) CleanupExpiredTokens() error {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	_, err := r.collection.DeleteMany(ctx, bson.M{
		"expires_at": bson.M{"$lt": time.Now()},
	})
	return err
}
