package stravaclient

import (
	"context"
	"time"

	"github.com/antihax/optional"
	"github.com/tcarreira/roaw2020/strava_client/swagger"
)

// StravaAPI contains a Strava API Client with the necessary context
type StravaAPI struct {
	client *swagger.APIClient
	ctx    context.Context
	opts   *swagger.ActivitiesApiGetLoggedInAthleteActivitiesOpts
}

// NewStravaAPI returns a StravaAPI ready to make API calls (on behalfe of stravaAccessToken)
// with default opts (like after/before dates)
func NewStravaAPI(stravaAccessToken string) *StravaAPI {
	s := &StravaAPI{
		client: swagger.NewAPIClient(swagger.NewConfiguration()),
		ctx:    context.WithValue(context.Background(), swagger.ContextAccessToken, stravaAccessToken),
		opts: &swagger.ActivitiesApiGetLoggedInAthleteActivitiesOpts{
			After:   optional.NewInt32(int32(time.Date(2020, 1, 1, 0, 0, 0, 0, time.UTC).Unix())),
			Before:  optional.NewInt32(int32(time.Date(2021, 1, 1, 0, 0, 0, 0, time.UTC).Unix())),
			Page:    optional.NewInt32(1),
			PerPage: optional.NewInt32(int32(5)),
		},
	}
	return s
}

// FetchActivitiesSinglePage will fetch a single page
func (s *StravaAPI) FetchActivitiesSinglePage(page int) ([]swagger.SummaryActivity, error) {
	s.opts.Page = optional.NewInt32(int32(page))
	activities, _, err := s.client.ActivitiesApi.GetLoggedInAthleteActivities(s.ctx, s.opts)

	return activities, err
}

// FetchAllActivities will fetch and return all activities (within the after/before defined in StravaAPI.opts)
func FetchAllActivities(stravaAccessToken string) ([]swagger.SummaryActivity, error) {
	stravaAPI := NewStravaAPI(stravaAccessToken)
	stravaAPI.opts.PerPage = optional.NewInt32(200) // set to max limit if we are going to fetch all anyway (https://developers.strava.com/docs/#Pagination)

	var allActivities []swagger.SummaryActivity
	for i := 1; ; i++ {
		activities, err := stravaAPI.FetchActivitiesSinglePage(i)

		if err != nil {
			return []swagger.SummaryActivity{}, err
		}

		allActivities = append(allActivities, activities...)

		if int32(len(activities)) != stravaAPI.opts.PerPage.Value() {
			// repeat the cicle until returns less than PerPage
			break
		}

	}

	return allActivities, nil
}
