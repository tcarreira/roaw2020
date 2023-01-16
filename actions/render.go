package actions

import (
	"fmt"
	"math"

	"github.com/gobuffalo/buffalo"
	"github.com/gobuffalo/buffalo/render"
	"github.com/gobuffalo/packr/v2"
	"github.com/gobuffalo/plush"
)

var r *render.Engine
var assetsBox = packr.New("app:assets", "../public")

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

func metersToKm(distance int) string {
	return fmt.Sprintf("%.2f", float64(distance)/1000.0)
}

func speed(distanceMeters, durationSeconds int) string {
	if durationSeconds == 0 {
		return "-"
	}

	speedKmPerHour := float64(distanceMeters) / float64(durationSeconds) * 3.6 // x3600s/1000m
	return fmt.Sprintf("%.2f", speedKmPerHour)
}

func pace(distanceMeters, durationSeconds int) string {
	if durationSeconds == 0 || distanceMeters == 0 {
		return "-"
	}

	paceSecondsKm := float64(durationSeconds) / (float64(distanceMeters) / 1000.0)

	min := int(paceSecondsKm / 60.0)
	sec := int(paceSecondsKm) % 60

	return fmt.Sprintf("%02d:%02d", min, sec)
}

func eq(a, b interface{}) bool {
	return a == b
}

func init() {
	r = render.New(render.Options{
		// HTML layout to be used for all HTML requests:
		HTMLLayout: "application.plush.html",

		// Box containing all of the templates:
		TemplatesBox: packr.New("app:templates", "../templates"),
		AssetsBox:    assetsBox,

		// Add template helpers here:
		Helpers: render.Helpers{
			"appShortName":   "ROAW",
			"appLongName":    "Run Once a Week",
			"appFullName":    "ROAW - Run Once a Week",
			"isLoggedIn":     isLoggedIn,
			"secondsToHuman": SecondsToHuman,
			"metersToKm":     metersToKm,
			"speed":          speed,
			"pace":           pace,
			"eq":             eq,
			"host":           App().Options.Host,
			// "isActive": func(name string, help plush.HelperContext) string {
			// 	if cr, ok := help.Value("current_path").(string); ok {
			// 		if strings.HasPrefix(cr, name) {
			// 			return "active"
			// 		}
			// 	}
			// 	return "inactive"
			// },
			// for non-bootstrap form helpers uncomment the lines
			// below and import "github.com/gobuffalo/helpers/forms"
			// forms.FormKey:     forms.Form,
			// forms.FormForKey:  forms.FormFor,
		},
	})
}

func isLoggedIn(help plush.HelperContext) bool {
	if session, ok := help.Value("session").(*buffalo.Session); ok {
		if u := session.Get("current_user_id"); u != nil {
			return true
		}
	}
	return false
}
