package models

import (
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"github.com/gobuffalo/nulls"
	"github.com/gobuffalo/pop/v5"
	"github.com/gobuffalo/validate/v3"
	"github.com/gobuffalo/validate/v3/validators"
	"github.com/gofrs/uuid"
	"github.com/markbates/goth"
	"github.com/tcarreira/roaw2020/strava_client/swagger"
)

// User is used by pop to map your users database table to your go code.
type User struct {
	ID           uuid.UUID    `json:"id" db:"id"`
	Name         string       `json:"name" db:"name"`
	Email        nulls.String `json:"email" db:"email"`
	Provider     string       `json:"provider" db:"provider"`
	ProviderID   string       `json:"provider_id" db:"provider_id"`
	AccessToken  string       `json:"access_token" db:"access_token"`
	RefreshToken string       `json:"refresh_token" db:"refresh_token"`
	AvatarURL    string       `json:"avatar_url" db:"avatar_url"`
	CreatedAt    time.Time    `json:"created_at" db:"created_at"`
	UpdatedAt    time.Time    `json:"updated_at" db:"updated_at"`
}

// String is not required by pop and may be deleted
func (u User) String() string {
	ju, _ := json.Marshal(u)
	return string(ju)
}

// Users is not required by pop and may be deleted
type Users []User

// String is not required by pop and may be deleted
func (u Users) String() string {
	ju, _ := json.Marshal(u)
	return string(ju)
}

// Validate gets run every time you call a "pop.Validate*" (pop.ValidateAndSave, pop.ValidateAndCreate, pop.ValidateAndUpdate) method.
// This method is not required and may be deleted.
func (u *User) Validate(tx *pop.Connection) (*validate.Errors, error) {
	return validate.Validate(
		&validators.StringIsPresent{Field: u.Name, Name: "Name"},
		&validators.StringIsPresent{Field: u.Provider, Name: "Provider"},
		&validators.StringIsPresent{Field: u.ProviderID, Name: "ProviderID"},
	), nil
}

// ValidateCreate gets run every time you call "pop.ValidateAndCreate" method.
// This method is not required and may be deleted.
func (u *User) ValidateCreate(tx *pop.Connection) (*validate.Errors, error) {
	return validate.NewErrors(), nil
}

// ValidateUpdate gets run every time you call "pop.ValidateAndUpdate" method.
// This method is not required and may be deleted.
func (u *User) ValidateUpdate(tx *pop.Connection) (*validate.Errors, error) {
	return validate.NewErrors(), nil
}

// RefreshAccessToken will refresh user's accessToken and refreshToken auth
func (u *User) RefreshAccessToken(tx *pop.Connection) error {
	// get Strava Auth provider
	provider, ok := goth.GetProviders()[u.Provider]
	if !ok {
		return fmt.Errorf("%s connector is having a problem. Contact the admin", u.Provider)
	}

	// refresh auth tokens
	newTokens, err := provider.RefreshToken(u.RefreshToken)
	if err != nil {
		return fmt.Errorf("The accessToken for user '%s' could not be refreshed. %w", u.Name, err)
	}

	if u.AccessToken != newTokens.AccessToken {
		u.AccessToken = newTokens.AccessToken
		u.RefreshToken = newTokens.RefreshToken
		err = tx.Save(u)
	}

	return err
}

// SyncActivities will fetch provider's activities and store them on database
func (u *User) SyncActivities(tx *pop.Connection, syncFunction func(stravaAccessToken string) ([]swagger.SummaryActivity, error)) error {
	if err := u.RefreshAccessToken(tx); err != nil {
		return err
	}

	stravaActivities, err := syncFunction(u.AccessToken)
	if err != nil {
		return fmt.Errorf("Could not fetch latestActivities for user %s. %w", u.Name, err)
	}

	var errorStrings []string
	for _, stravaActivity := range stravaActivities {
		activity := ParseStravaActivity(stravaActivity, *u)

		if err := activity.CreateOrUpdate(tx); err != nil {
			errorStrings = append(errorStrings, activity.ProviderID)
		}
	}

	if len(errorStrings) > 0 {
		return fmt.Errorf("Error processing activities: %s", strings.Join(errorStrings, ", "))
	}
	return nil
}

// UserStats contains public User data with activities stats
type UserStats struct {
	Distance            int
	Count               int
	ElapsedDuration     int
	MovingDuration      int
	MostDistance        int
	MostElapsedDuration int
	MostMovingDuration  int
}

// GetStats will return a UserStats for this User
func (u *User) GetStats(tx *pop.Connection) (allActivitiesStats UserStats, validActivitiesStats UserStats, err error) {

	activities := Activities{}

	q := tx.Q().Where("users.id = ?", u.ID)
	q = q.Join("users", "activities.user_id = users.id")
	if err := q.All(&activities); err != nil {
		return UserStats{}, UserStats{}, err
	}

	for _, activity := range activities {
		allActivitiesStats.Count++
		allActivitiesStats.Distance += activity.Distance
		allActivitiesStats.ElapsedDuration += activity.ElapsedTime
		allActivitiesStats.MovingDuration += activity.MovingTime

		if activity.Distance > allActivitiesStats.MostDistance {
			allActivitiesStats.MostDistance = activity.Distance
		}
		if activity.ElapsedTime > allActivitiesStats.MostElapsedDuration {
			allActivitiesStats.MostElapsedDuration = activity.ElapsedTime
		}
		if activity.MovingTime > allActivitiesStats.MostMovingDuration {
			allActivitiesStats.MostMovingDuration = activity.MovingTime
		}

		// XXX: hardcoded 15min
		if activity.Type == "Run" && activity.ElapsedTime > (60*15) {
			validActivitiesStats.Count++
			validActivitiesStats.Distance += activity.Distance
			validActivitiesStats.ElapsedDuration += activity.ElapsedTime
			validActivitiesStats.MovingDuration += activity.MovingTime

			if activity.Distance > validActivitiesStats.MostDistance {
				validActivitiesStats.MostDistance = activity.Distance
			}
			if activity.ElapsedTime > validActivitiesStats.MostElapsedDuration {
				validActivitiesStats.MostElapsedDuration = activity.ElapsedTime
			}
			if activity.MovingTime > validActivitiesStats.MostMovingDuration {
				validActivitiesStats.MostMovingDuration = activity.MovingTime
			}
		}
	}

	return
}
