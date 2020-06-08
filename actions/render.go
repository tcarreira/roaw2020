package actions

import (
	"fmt"
	"strings"

	"github.com/gobuffalo/buffalo"
	"github.com/gobuffalo/buffalo/render"
	"github.com/gobuffalo/packr/v2"
	"github.com/gobuffalo/plush"
)

var r *render.Engine
var assetsBox = packr.New("app:assets", "../public")

func weekNumbersList() string {
	list := make([]string, 53)
	for w := 0; w < 53; w++ {
		list[w] = fmt.Sprintf("%d", w)
	}
	return strings.Join(list, ", ")
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
			"appShortName":    "ROAW",
			"appLongName":     "Run Once a Week",
			"appFullName":     "ROAW - Run Once a Week",
			"isLoggedIn":      isLoggedIn,
			"weekNumbersList": weekNumbersList,
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
