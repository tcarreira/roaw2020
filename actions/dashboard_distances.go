package actions

import (
	"fmt"
	"net/http"

	"github.com/gobuffalo/buffalo"
	"github.com/gobuffalo/envy"
	"github.com/gobuffalo/pop/v5"
	"github.com/gobuffalo/x/responder"
)

type weekDistance struct {
	Week     int `json:"x" db:"week"`
	Distance int `json:"y" db:"distance"`
}

// map user to struct
type weeklyDistanceStats map[string][]weekDistance

func getWeeklyDistanceStats(tx *pop.Connection) (weeklyDistanceStats, error) {
	thisYear, nextYear := parseThisNextYear(envy.Get("ROAW_YEAR", ""))

	queryString := "SELECT " +
		"  COALESCE(EXTRACT(WEEK FROM a.datetime),0) AS week, " +
		"  u.name as user, " +
		"  SUM(COALESCE(a.distance,0))/1000 as distance " +
		"FROM users u " +
		"  LEFT JOIN activities a ON a.user_id = u.id " +
		"WHERE a.type IS NULL OR (a.type = 'Run' " +
		"  AND a.datetime >= '" + thisYear + "-01-01' " +
		"  AND a.datetime <  '" + nextYear + "-01-01' ) " +
		"GROUP BY u.id, week " +
		"ORDER BY u.name ASC, week ASC"

	data := []struct {
		Week     int    `json:"week" db:"week"`
		User     string `json:"user" db:"user"`
		Distance int    `json:"distance" db:"distance"`
	}{}
	err := tx.RawQuery(queryString).All(&data)

	weekIdx := 0
	returnData := weeklyDistanceStats{}
	for _, row := range data {
		_, ok := returnData[row.User]
		if !ok {
			returnData[row.User] = []weekDistance{}
			weekIdx = 0 // assuming ordered by user first
		}

		// fill empty weeks (until the last one)
		for ; weekIdx < row.Week; weekIdx++ {
			returnData[row.User] = append(returnData[row.User], weekDistance{Week: weekIdx, Distance: 0})
		}

		returnData[row.User] = append(returnData[row.User], weekDistance{
			Week:     row.Week,
			Distance: row.Distance,
		})
		weekIdx++
	}

	return returnData, err

}

func getWeeklyCumulativeDistanceStats(tx *pop.Connection) (weeklyDistanceStats, error) {
	distanceStats, err := getWeeklyDistanceStats(tx)
	if err != nil {
		return weeklyDistanceStats{}, nil
	}

	// everyone gets last week point
	latestWeek := 0
	for _, weeksDistances := range distanceStats {
		if latestWeek < weeksDistances[len(weeksDistances)-1].Week {
			latestWeek = weeksDistances[len(weeksDistances)-1].Week
		}
	}

	// fill cumulative values
	for user, weeksDistances := range distanceStats {
		cumulative := 0
		for idx, weekDistance := range weeksDistances {
			cumulative += weekDistance.Distance
			distanceStats[user][idx].Distance = cumulative
		}
		if weeksDistances[len(weeksDistances)-1].Week < latestWeek {
			// everyone gets last week point
			distanceStats[user] = append(weeksDistances, weekDistance{latestWeek, cumulative})
		}
	}

	return distanceStats, nil
}

// WeeklyDistanceStatsHandler shows a weekly stats by user
func WeeklyDistanceStatsHandler(c buffalo.Context) error {
	// Get the DB connection from the context
	tx, ok := c.Value("tx").(*pop.Connection)
	if !ok {
		return fmt.Errorf("no transaction found")
	}

	weeklyStats, err := getWeeklyDistanceStats(tx)
	if err != nil {
		c.Flash().Add("error", fmt.Sprintf("Error fetching weekly stats: %v", err))
	}

	return responder.Wants("html", func(c buffalo.Context) error {
		c.Set("weeklyStats", weeklyStats)

		return c.Render(http.StatusOK, r.HTML("/dashboard/weekly-stats.plush.html"))
	}).Wants("json", func(c buffalo.Context) error {
		return c.Render(http.StatusOK, r.JSON(weeklyStats))
	}).Wants("xml", func(c buffalo.Context) error {
		return c.Render(http.StatusOK, r.XML(weeklyStats))
	}).Respond(c)

}

// WeeklyCumulativeDistanceStatsHandler shows a weekly stats by user
func WeeklyCumulativeDistanceStatsHandler(c buffalo.Context) error {
	// Get the DB connection from the context
	tx, ok := c.Value("tx").(*pop.Connection)
	if !ok {
		return fmt.Errorf("no transaction found")
	}

	weeklyStats, err := getWeeklyCumulativeDistanceStats(tx)
	// weeklyStats, err := getWeeklyCumulativeDistanceStats(tx)
	if err != nil {
		c.Flash().Add("error", fmt.Sprintf("Error fetching weekly stats: %v", err))
	}

	return responder.Wants("html", func(c buffalo.Context) error {
		c.Set("weeklyStats", weeklyStats)

		return c.Render(http.StatusOK, r.HTML("/dashboard/weekly-stats.plush.html"))
	}).Wants("json", func(c buffalo.Context) error {
		return c.Render(http.StatusOK, r.JSON(weeklyStats))
	}).Wants("xml", func(c buffalo.Context) error {
		return c.Render(http.StatusOK, r.XML(weeklyStats))
	}).Respond(c)

}
