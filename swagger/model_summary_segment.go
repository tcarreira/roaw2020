/*
 * Strava API v3
 *
 * The [Swagger Playground](https://developers.strava.com/playground) is the easiest way to familiarize yourself with the Strava API by submitting HTTP requests and observing the responses before you write any client code. It will show what a response will look like with different endpoints depending on the authorization scope you receive from your athletes. To use the Playground, go to https://www.strava.com/settings/api and change your “Authorization Callback Domain” to developers.strava.com. Please note, we only support Swagger 2.0. There is a known issue where you can only select one scope at a time. For more information, please check the section “client code” at https://developers.strava.com/docs.
 *
 * API version: 3.0.0
 * Generated by: Swagger Codegen (https://github.com/swagger-api/swagger-codegen.git)
 */

package swagger

type SummarySegment struct {
	// The unique identifier of this segment
	Id int64 `json:"id,omitempty"`
	// The name of this segment
	Name string `json:"name,omitempty"`
	ActivityType string `json:"activity_type,omitempty"`
	// The segment's distance, in meters
	Distance float32 `json:"distance,omitempty"`
	// The segment's average grade, in percents
	AverageGrade float32 `json:"average_grade,omitempty"`
	// The segments's maximum grade, in percents
	MaximumGrade float32 `json:"maximum_grade,omitempty"`
	// The segments's highest elevation, in meters
	ElevationHigh float32 `json:"elevation_high,omitempty"`
	// The segments's lowest elevation, in meters
	ElevationLow float32 `json:"elevation_low,omitempty"`
	StartLatlng *LatLng `json:"start_latlng,omitempty"`
	EndLatlng *LatLng `json:"end_latlng,omitempty"`
	// The category of the climb [0, 5]. Higher is harder ie. 5 is Hors catégorie, 0 is uncategorized in climb_category.
	ClimbCategory int32 `json:"climb_category,omitempty"`
	// The segments's city.
	City string `json:"city,omitempty"`
	// The segments's state or geographical region.
	State string `json:"state,omitempty"`
	// The segment's country.
	Country string `json:"country,omitempty"`
	// Whether this segment is private.
	Private bool `json:"private,omitempty"`
	AthletePrEffort *SummarySegmentEffort `json:"athlete_pr_effort,omitempty"`
}
