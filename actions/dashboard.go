package actions

import (
	"fmt"
	"math"
	"net/http"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/gobuffalo/buffalo"
	"github.com/gobuffalo/pop/v5"
	"github.com/gobuffalo/x/responder"
)

type userTotalDistanceData struct {
	// RowNumber int     `json:"row_number" db:"row_number"`
	User     string `json:"user" db:"user"`
	Distance int    `json:"distance" db:"distance"`
}

func parseThisNextYear(osEnv string) (string, string) {

	thisYear, err := strconv.Atoi(osEnv)
	if err != nil {
		App().Logger.Errorf("%s could not be parsed int int.", osEnv)

		currentYearInt := time.Now().Year()
		return fmt.Sprintf("%d", currentYearInt), fmt.Sprintf("%d", currentYearInt+1)
	}

	return fmt.Sprintf("%d", thisYear), fmt.Sprintf("%d", thisYear+1)
}

func getAllUsersTotalDistance(tx *pop.Connection) ([]userTotalDistanceData, error) {
	thisYear, nextYear := parseThisNextYear(os.Getenv("ROAW_2020"))

	queryString := "SELECT " +
		"  u.name as user, " +
		"  SUM(COALESCE(a.distance,0))/1000 as distance " +
		"FROM users u " +
		"  LEFT JOIN activities a ON a.user_id = u.id " +
		"WHERE a.type IS NULL OR (a.type = 'Run' " +
		"  AND a.datetime >= '" + thisYear + "-01-01' " +
		"  AND a.datetime <  '" + nextYear + "-01-01' ) " +
		"GROUP BY u.id " +
		"ORDER BY distance DESC"

	data := []userTotalDistanceData{}

	err := tx.RawQuery(queryString).All(&data)

	return data, err
}

type userTotalActivityCount struct {
	// RowNumber int    `json:"row_number" db:"row_number"`
	User  string `json:"user" db:"user"`
	Count int    `json:"distance" db:"count"`
}

func getAllUsersActivityCount(tx *pop.Connection) ([]userTotalActivityCount, error) {
	thisYear, nextYear := parseThisNextYear(os.Getenv("ROAW_2020"))

	queryString := "SELECT " +
		"  u.name as user, " +
		"  COUNT(a.distance) as count " +
		"FROM users u " +
		"  LEFT JOIN activities a ON a.user_id = u.id " +
		"WHERE a.type IS NULL OR (a.type = 'Run' " +
		"  AND a.datetime >= '" + thisYear + "-01-01' " +
		"  AND a.datetime <  '" + nextYear + "-01-01' " +
		"  AND a.elapsed_time >= 300 ) " +
		"GROUP BY u.id " +
		"ORDER BY count DESC"

	data := []userTotalActivityCount{}

	err := tx.RawQuery(queryString).All(&data)

	return data, err
}

type userTotalDuration struct {
	// RowNumber int    `json:"row_number" db:"row_number"`
	User     string `json:"user" db:"user"`
	Duration int    `json:"distance" db:"duration"`
}

func getAllUsersTotalDuration(tx *pop.Connection) ([]userTotalDuration, error) {
	thisYear, nextYear := parseThisNextYear(os.Getenv("ROAW_2020"))

	queryString := "SELECT " +
		"  u.name as user, " +
		"  SUM(COALESCE(a.elapsed_time,0)) as duration " +
		"FROM users u " +
		"  LEFT JOIN activities a ON a.user_id = u.id " +
		"WHERE a.type IS NULL OR (a.type = 'Run' " +
		"  AND a.datetime >= '" + thisYear + "-01-01' " +
		"  AND a.datetime <  '" + nextYear + "-01-01' ) " +
		"GROUP BY u.id " +
		"ORDER BY duration DESC"

	data := []userTotalDuration{}

	err := tx.RawQuery(queryString).All(&data)

	return data, err
}

// convertPodiumClass will take the 0-index and convert to podium HTML class name
func convertPodiumClass(i int) string {
	switch i {
	case 0:
		return "table-warning" // 1st place - Gold
	case 1:
		return "table-secondary" // 2nd place - Silver
	case 2:
		return "table-danger" // 3nd place - Bronze
	}
	return ""
}

// secondsToHuman converts minutes to human duration string (1d 7h 32m )
func secondsToHuman(duration int) string {
	if duration == 0 {
		return "0"
	}

	// seconds := int(duration % 60)
	minutes := int(math.Floor(float64(duration%3600.0) / 60))
	hours := int(math.Floor(float64(duration%86400.0) / 3600))
	days := int(math.Floor(float64(duration) / 86400))

	if days > 0 {
		return fmt.Sprintf("%dd %02dh%02dm", days, hours, minutes)
	}
	if hours > 0 {
		return fmt.Sprintf("%02dh%02dm", hours, minutes)
	}
	return fmt.Sprintf("%02dm", minutes)

}

// DashboardHandler shows a dashboard
func DashboardHandler(c buffalo.Context) error {
	// Get the DB connection from the context
	tx, ok := c.Value("tx").(*pop.Connection)
	if !ok {
		return fmt.Errorf("no transaction found")
	}

	allUsersTotalDistance, err := getAllUsersTotalDistance(tx)
	if err != nil {
		c.Flash().Add("error", fmt.Sprintf("Error fetching total distance data: %v", err))
	}

	allUsersActivityCount, err := getAllUsersActivityCount(tx)
	if err != nil {
		c.Flash().Add("error", fmt.Sprintf("Error fetching totalactivity count: %v", err))
	}

	allUsersTotalDuration, err := getAllUsersTotalDuration(tx)
	if err != nil {
		c.Flash().Add("error", fmt.Sprintf("Error fetching total duration data: %v", err))
	}

	weeklyStats, err := getWeeklyDistanceStats(tx)
	if err != nil {
		c.Flash().Add("error", fmt.Sprintf("Error fetching weekly stats: %v", err))
	}

	return responder.Wants("html", func(c buffalo.Context) error {
		c.Set("convertPodiumClass", convertPodiumClass)
		c.Set("secondsToHuman", secondsToHuman)

		c.Set("current_user_id", c.Session().Get("current_user_id"))

		c.Set("totalDistance", allUsersTotalDistance)
		c.Set("totalCount", allUsersActivityCount)
		c.Set("totalDuration", allUsersTotalDuration)

		c.Set("weeklyStats", weeklyStats)

		if strings.Contains(c.Request().RequestURI, "beta") {
			return c.Render(http.StatusOK, r.HTML("/dashboard/index-beta.plush.html"))
		}
		return c.Render(http.StatusOK, r.HTML("/dashboard/index.plush.html"))
	}).Wants("json", func(c buffalo.Context) error {
		return c.Render(200, r.JSON(allUsersTotalDistance))
	}).Wants("xml", func(c buffalo.Context) error {
		return c.Render(200, r.XML(allUsersTotalDistance))
	}).Respond(c)

}

type weekDistance struct {
	Week     int `json:"x" db:"week"`
	Distance int `json:"y" db:"distance"`
}

// map user to struct
type weeklyDistanceStats map[string][]weekDistance

func getWeeklyDistanceStats(tx *pop.Connection) (weeklyDistanceStats, error) {
	thisYear, nextYear := parseThisNextYear(os.Getenv("ROAW_2020"))

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
		"ORDER BY week ASC, u.id ASC"

	data := []struct {
		Week     int    `json:"week" db:"week"`
		User     string `json:"user" db:"user"`
		Distance int    `json:"distance" db:"distance"`
	}{}

	err := tx.RawQuery(queryString).All(&data)

	returnData := weeklyDistanceStats{}
	for _, row := range data {
		_, ok := returnData[row.User]
		if !ok {
			returnData[row.User] = []weekDistance{}
		}
		returnData[row.User] = append(returnData[row.User], weekDistance{
			Week:     row.Week,
			Distance: row.Distance,
		})
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
