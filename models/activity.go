package models

import (
	"encoding/json"
	"time"

	"github.com/gobuffalo/pop/v5"
	"github.com/gobuffalo/validate/v3"
	"github.com/gobuffalo/validate/v3/validators"
	"github.com/gofrs/uuid"
)

// Activity is used by pop to map your activities database table to your go code.
type Activity struct {
	ID          uuid.UUID `json:"id" db:"id"`
	UserID      uuid.UUID `json:"user_id" db:"user_id"`
	Provider    string    `json:"provider" db:"provider"`
	ProviderID  string    `json:"provider_id" db:"provider_id"`
	Name        string    `json:"name" db:"name"`
	Type        string    `json:"type" db:"type"`
	Datetime    time.Time `json:"datetime" db:"datetime"`
	Distance    int       `json:"distance" db:"distance"`
	MovingTime  int       `json:"moving_time" db:"moving_time"`
	ElapsedTime int       `json:"elapsed_time" db:"elapsed_time"`
	CreatedAt   time.Time `json:"created_at" db:"created_at"`
	UpdatedAt   time.Time `json:"updated_at" db:"updated_at"`
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
		&validators.IntIsPresent{Field: a.MovingTime, Name: "MovingTime"},
		&validators.IntIsPresent{Field: a.ElapsedTime, Name: "ElapsedTime"},
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

// CreateOrUpdate will create or update the activity
// (based on (provider,provider_id) key)
func (a *Activity) CreateOrUpdate(tx *pop.Connection) error {
	tmpActivity := &Activity{}

	q := tx.Where("provider = ?", a.Provider).Where("provider_id = ?", a.ProviderID)
	q.First(tmpActivity)

	if IsSameActivity(a, tmpActivity) {
		return nil
	}

	a.ID = tmpActivity.ID
	err := tx.Save(a)
	return err
}

// IsSameActivity returns true when activities' relevant fields are equal
func IsSameActivity(a1, a2 *Activity) bool {
	return a1.Provider == a2.Provider &&
		a1.ProviderID == a2.ProviderID &&
		a1.Name == a2.Name &&
		a1.Type == a2.Type &&
		a1.Datetime.Unix() == a2.Datetime.Unix() &&
		a1.Distance == a2.Distance &&
		a1.MovingTime == a2.MovingTime &&
		a1.ElapsedTime == a2.ElapsedTime

}
