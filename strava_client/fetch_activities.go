package stravaclient

import (
	"context"
	"strconv"
	"time"

	"github.com/antihax/optional"
	"github.com/gobuffalo/envy"
	"github.com/tcarreira/roaw2020/strava_client/swagger"
)

// StravaAPI contains a Strava API Client with the necessary context
type StravaAPI struct {
	client *swagger.APIClient
	ctx    context.Context
	opts   *swagger.ActivitiesApiGetLoggedInAthleteActivitiesOpts
}

func getThisYear() int {

	osEnv := envy.Get("ROAW_YEAR", "")
	thisYear, err := strconv.Atoi(osEnv)
	if err != nil {
		// Fail silently
		return time.Now().Year()
	}

	return thisYear
}

// NewStravaAPI returns a StravaAPI ready to make API calls (on behalf of stravaAccessToken)
// with default opts (like after/before dates)
func NewStravaAPI(stravaAccessToken string) *StravaAPI {
	thisYear := getThisYear()

	s := &StravaAPI{
		client: swagger.NewAPIClient(swagger.NewConfiguration()),
		ctx:    context.WithValue(context.Background(), swagger.ContextAccessToken, stravaAccessToken),
		opts: &swagger.ActivitiesApiGetLoggedInAthleteActivitiesOpts{
			After:   optional.NewInt32(int32(time.Date(thisYear, 1, 1, 0, 0, 0, 0, time.UTC).Unix())),
			Before:  optional.NewInt32(int32(time.Date(thisYear+1, 1, 1, 0, 0, 0, 0, time.UTC).Unix())),
			Page:    optional.NewInt32(1),
			PerPage: optional.NewInt32(int32(5)),
		},
	}
	return s
}

// fetchActivitiesSinglePage will fetch a single page
func (s *StravaAPI) fetchActivitiesSinglePage(page int) ([]swagger.SummaryActivity, error) {
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
		activities, err := stravaAPI.fetchActivitiesSinglePage(i)

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

// FetchLatestActivities will fetch and return the latest N activities (N defined in StravaAPI.opts)
func FetchLatestActivities(stravaAccessToken string) ([]swagger.SummaryActivity, error) {
	stravaAPI := NewStravaAPI(stravaAccessToken)

	activities, err := stravaAPI.fetchActivitiesSinglePage(1)

	return activities, err
}
