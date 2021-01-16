package actions

import (
	"fmt"
	"net/http"

	"github.com/gobuffalo/buffalo"
	"github.com/gobuffalo/envy"
	"github.com/gobuffalo/pop/v5"
	"github.com/gobuffalo/x/responder"
)

type weekCount struct {
	Week  int `json:"x" db:"week"`
	Count int `json:"y" db:"count"`
}

// map user to struct
type weeklyCountStats map[string][]weekCount

func getWeeklyCountStats(tx *pop.Connection) (weeklyCountStats, error) {
	thisYear, nextYear := parseThisNextYear(envy.Get("ROAW_YEAR", ""))

	queryString := "SELECT " +
		"  COALESCE(" +
		"    CASE " +
		"      WHEN DATE_PART('isoyear', a.datetime) < " + thisYear + " then 0 " +
		"      ELSE DATE_PART('week', a.datetime) " +
		"    END " +
		"  , 0) AS week, " +
		"  u.name as user, " +
		"  COUNT(a.id) as count " +
		"FROM users u " +
		"  LEFT JOIN activities a ON a.user_id = u.id " +
		"WHERE a.type IS NULL OR (a.type = 'Run' " +
		"  AND a.datetime >= '" + thisYear + "-01-01' " +
		"  AND a.datetime <  '" + nextYear + "-01-01' ) " +
		"GROUP BY u.id, week " +
		"ORDER BY u.name ASC, week ASC"

	data := []struct {
		Week  int    `json:"week" db:"week"`
		User  string `json:"user" db:"user"`
		Count int    `json:"count" db:"count"`
	}{}
	err := tx.RawQuery(queryString).All(&data)

	weekIdx := 0
	returnData := weeklyCountStats{}
	for _, row := range data {
		_, ok := returnData[row.User]
		if !ok {
			returnData[row.User] = []weekCount{}
			weekIdx = 0 // assuming ordered by user first
		}

		// fill empty weeks (until the last one)
		for ; weekIdx < row.Week; weekIdx++ {
			returnData[row.User] = append(returnData[row.User], weekCount{Week: weekIdx, Count: 0})
		}

		returnData[row.User] = append(returnData[row.User], weekCount{
			Week:  row.Week,
			Count: row.Count,
		})
		weekIdx++
	}

	return returnData, err

}

func getWeeklyCumulativeCountStats(tx *pop.Connection) (weeklyCountStats, error) {
	countStats, err := getWeeklyCountStats(tx)
	if err != nil {
		return weeklyCountStats{}, nil
	}

	// everyone gets last week point
	latestWeek := 0
	for _, weeksCounts := range countStats {
		if latestWeek < weeksCounts[len(weeksCounts)-1].Week {
			latestWeek = weeksCounts[len(weeksCounts)-1].Week
		}
	}

	// fill cumulative values
	for user, weeksCounts := range countStats {
		cumulative := 0
		for idx, weekCount := range weeksCounts {
			cumulative += weekCount.Count
			countStats[user][idx].Count = cumulative
		}
		if weeksCounts[len(weeksCounts)-1].Week < latestWeek {
			// everyone gets last week point
			countStats[user] = append(weeksCounts, weekCount{latestWeek, cumulative})
		}
	}

	return countStats, nil
}

// WeeklyCountStatsHandler shows a weekly stats by user
func WeeklyCountStatsHandler(c buffalo.Context) error {
	// Get the DB connection from the context
	tx, ok := c.Value("tx").(*pop.Connection)
	if !ok {
		return fmt.Errorf("no transaction found")
	}

	weeklyStats, err := getWeeklyCountStats(tx)
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

// WeeklyCumulativeCountStatsHandler shows a weekly stats by user
func WeeklyCumulativeCountStatsHandler(c buffalo.Context) error {
	// Get the DB connection from the context
	tx, ok := c.Value("tx").(*pop.Connection)
	if !ok {
		return fmt.Errorf("no transaction found")
	}

	weeklyStats, err := getWeeklyCumulativeCountStats(tx)
	// weeklyStats, err := getWeeklyCumulativeCountStats(tx)
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
