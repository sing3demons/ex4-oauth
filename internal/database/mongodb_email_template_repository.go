package database

import (
	"context"
	"time"

	"ex4-oauth2/internal/models"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
)

// MongoEmailTemplateRepository implements EmailTemplateRepository interface
type MongoEmailTemplateRepository struct {
	collection *mongo.Collection
}

// NewMongoEmailTemplateRepository creates a new MongoEmailTemplateRepository
func NewMongoEmailTemplateRepository(db *mongo.Database) models.EmailTemplateRepository {
	return &MongoEmailTemplateRepository{
		collection: db.Collection("email_templates"),
	}
}

// Create creates a new email template
func (r *MongoEmailTemplateRepository) Create(template *models.EmailTemplate) error {
	template.CreatedAt = time.Now()
	template.UpdatedAt = time.Now()

	_, err := r.collection.InsertOne(context.Background(), template)
	return err
}

// GetByName retrieves an email template by name
func (r *MongoEmailTemplateRepository) GetByName(name string) (*models.EmailTemplate, error) {
	var template models.EmailTemplate
	err := r.collection.FindOne(context.Background(), bson.M{"name": name}).Decode(&template)
	if err != nil {
		return nil, err
	}
	return &template, nil
}

// GetAll retrieves all email templates
func (r *MongoEmailTemplateRepository) GetAll() ([]*models.EmailTemplate, error) {
	cursor, err := r.collection.Find(context.Background(), bson.M{})
	if err != nil {
		return nil, err
	}
	defer cursor.Close(context.Background())

	var templates []*models.EmailTemplate
	for cursor.Next(context.Background()) {
		var template models.EmailTemplate
		if err := cursor.Decode(&template); err != nil {
			return nil, err
		}
		templates = append(templates, &template)
	}

	return templates, cursor.Err()
}

// Update updates an existing email template
func (r *MongoEmailTemplateRepository) Update(template *models.EmailTemplate) error {
	template.UpdatedAt = time.Now()

	_, err := r.collection.ReplaceOne(
		context.Background(),
		bson.M{"_id": template.ID},
		template,
	)
	return err
}

// Delete deletes an email template by ID
func (r *MongoEmailTemplateRepository) Delete(id string) error {
	_, err := r.collection.DeleteOne(context.Background(), bson.M{"_id": id})
	return err
}
