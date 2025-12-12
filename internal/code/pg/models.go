package code

import (
	"time"

	"github.com/weni-ai/flows-code-actions/internal/code"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

// CodePG represents the PostgreSQL version of the Code struct
type CodePG struct {
	ID          string            `json:"id"`
	Name        string            `json:"name"`
	Type        code.CodeType     `json:"type"`
	Source      string            `json:"source"`
	Language    code.LanguageType `json:"language"`
	URL         string            `json:"url,omitempty"`
	ProjectUUID string            `json:"project_uuid"`
	CreatedAt   time.Time         `json:"created_at"`
	UpdatedAt   time.Time         `json:"updated_at"`
	Timeout     int               `json:"timeout"`
}

// ToMongoCode converts PostgreSQL Code to MongoDB Code format
func (c *CodePG) ToMongoCode() *code.Code {
	mongoCode := &code.Code{
		Name:        c.Name,
		Type:        c.Type,
		Source:      c.Source,
		Language:    c.Language,
		URL:         c.URL,
		ProjectUUID: c.ProjectUUID,
		CreatedAt:   c.CreatedAt,
		UpdatedAt:   c.UpdatedAt,
		Timeout:     c.Timeout,
	}

	// Convert string ID to ObjectID if valid
	if c.ID != "" {
		if oid, err := primitive.ObjectIDFromHex(c.ID); err == nil {
			mongoCode.ID = oid
		}
	}

	return mongoCode
}

// FromMongoCode converts MongoDB Code to PostgreSQL Code format
func FromMongoCode(mongoCode *code.Code) *CodePG {
	return &CodePG{
		ID:          mongoCode.ID.Hex(),
		Name:        mongoCode.Name,
		Type:        mongoCode.Type,
		Source:      mongoCode.Source,
		Language:    mongoCode.Language,
		URL:         mongoCode.URL,
		ProjectUUID: mongoCode.ProjectUUID,
		CreatedAt:   mongoCode.CreatedAt,
		UpdatedAt:   mongoCode.UpdatedAt,
		Timeout:     mongoCode.Timeout,
	}
}

// FromMongoCodes converts slice of MongoDB Code to PostgreSQL Code format
func FromMongoCodes(mongoCodes []code.Code) []code.Code {
	codes := make([]code.Code, len(mongoCodes))
	for i, mongoCode := range mongoCodes {
		pgCode := FromMongoCode(&mongoCode)
		codes[i] = *pgCode.ToMongoCode()
	}
	return codes
}
