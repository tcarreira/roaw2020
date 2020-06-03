package models

import (
	"encoding/json"
	"github.com/gobuffalo/pop/v5"
	"github.com/gobuffalo/validate/v3"
	"github.com/gofrs/uuid"
	"time"
	"github.com/gobuffalo/validate/v3/validators"
)
// Activity is used by pop to map your activities database table to your go code.
type Activity struct {
    ID uuid.UUID `json:"id" db:"id"`
    UserID uuid.UUID `json:"user_id" db:"user_id"`
    Provider string `json:"provider" db:"provider"`
    ProviderID string `json:"provider_id" db:"provider_id"`
    Name string `json:"name" db:"name"`
    Type string `json:"type" db:"type"`
    Datetime time.Time `json:"datetime" db:"datetime"`
    Distance int `json:"distance" db:"distance"`
    MovingTime string `json:"moving_time" db:"moving_time"`
    ElapsedTime string `json:"elapsed_time" db:"elapsed_time"`
    CreatedAt time.Time `json:"created_at" db:"created_at"`
    UpdatedAt time.Time `json:"updated_at" db:"updated_at"`
}

// String is not required by pop and may be deleted
func (a Activity) String() string {
	ja, _ := json.Marshal(a)
	return string(ja)
}

// Activities is not required by pop and may be deleted
type Activities []Activity

// String is not required by pop and may be deleted
func (a Activities) String() string {
	ja, _ := json.Marshal(a)
	return string(ja)
}

// Validate gets run every time you call a "pop.Validate*" (pop.ValidateAndSave, pop.ValidateAndCreate, pop.ValidateAndUpdate) method.
// This method is not required and may be deleted.
func (a *Activity) Validate(tx *pop.Connection) (*validate.Errors, error) {
	return validate.Validate(
		&validators.StringIsPresent{Field: a.Provider, Name: "Provider"},
		&validators.StringIsPresent{Field: a.ProviderID, Name: "ProviderID"},
		&validators.StringIsPresent{Field: a.Name, Name: "Name"},
		&validators.StringIsPresent{Field: a.Type, Name: "Type"},
		&validators.TimeIsPresent{Field: a.Datetime, Name: "Datetime"},
		&validators.IntIsPresent{Field: a.Distance, Name: "Distance"},
		&validators.StringIsPresent{Field: a.MovingTime, Name: "MovingTime"},
		&validators.StringIsPresent{Field: a.ElapsedTime, Name: "ElapsedTime"},
	), nil
}

// ValidateCreate gets run every time you call "pop.ValidateAndCreate" method.
// This method is not required and may be deleted.
func (a *Activity) ValidateCreate(tx *pop.Connection) (*validate.Errors, error) {
	return validate.NewErrors(), nil
}

// ValidateUpdate gets run every time you call "pop.ValidateAndUpdate" method.
// This method is not required and may be deleted.
func (a *Activity) ValidateUpdate(tx *pop.Connection) (*validate.Errors, error) {
	return validate.NewErrors(), nil
}
