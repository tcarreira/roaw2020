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

type userTotalActivityCount struct {
	// RowNumber int    `json:"row_number" db:"row_number"`
	User  string `json:"user" db:"user"`
	Count int    `json:"distance" db:"count"`
}

type userTotalDuration struct {
	// RowNumber int    `json:"row_number" db:"row_number"`
	User     string `json:"user" db:"user"`
	Duration int    `json:"distance" db:"duration"`
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

// SecondsToHuman converts minutes to human duration string (1d 7h 32m )
func SecondsToHuman(duration int) string {
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
		c.Set("secondsToHuman", SecondsToHuman)

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
