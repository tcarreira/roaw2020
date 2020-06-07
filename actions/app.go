package actions

import (
	"fmt"
	"io"

	"github.com/gobuffalo/buffalo"
	"github.com/gobuffalo/buffalo/render"
	"github.com/gobuffalo/envy"
	forcessl "github.com/gobuffalo/mw-forcessl"
	paramlogger "github.com/gobuffalo/mw-paramlogger"
	"github.com/unrolled/secure"

	"github.com/gobuffalo/buffalo-pop/v2/pop/popmw"
	csrf "github.com/gobuffalo/mw-csrf"
	i18n "github.com/gobuffalo/mw-i18n"
	"github.com/gobuffalo/packr/v2"
	"github.com/tcarreira/roaw2020/models"
)

// ENV is used to help switch settings based on where the
// application is being run. Default is "development".
var ENV = envy.Get("GO_ENV", "development")
var app *buffalo.App

// T - Translator for handling all your i18n needs.
var T *i18n.Translator

// App is where all routes and middleware for buffalo
// should be defined. This is the nerve center of your
// application.
//
// Routing, middleware, groups, etc... are declared TOP -> DOWN.
// This means if you add a middleware to `app` *after* declaring a
// group, that group will NOT have that new middleware. The same
// is true of resource declarations as well.
//
// It also means that routes are checked in the order they are declared.
// `ServeFiles` is a CATCH-ALL route, so it should always be
// placed last in the route declarations, as it will prevent routes
// declared after it to never be called.
func App() *buffalo.App {
	if app == nil {
		buffaloAppOptions := buffalo.NewOptions()
		buffaloAppOptions.SessionName = "_roaw2020_session"

		app = buffalo.New(buffaloAppOptions)

		// Automatically redirect to SSL
		app.Use(forceSSL())

		// Log request parameters (filters apply).
		app.Use(paramlogger.ParameterLogger)

		// Protect against CSRF attacks. https://www.owasp.org/index.php/Cross-Site_Request_Forgery_(CSRF)
		// Remove to disable this.
		app.Use(csrf.New)

		// Wraps each request in a transaction.
		//  c.Value("tx").(*pop.Connection)
		// Remove to disable this.
		app.Use(popmw.Transaction(models.DB))

		// Setup and use translations:
		app.Use(translations())

		// Setup Authorization
		app.Use(SetCurrentUser)
		app.Use(Authorize)

		app.Middleware.Skip(Authorize, DashboardHandler)
		// app.GET("/", HomeHandler)
		app.GET("/", DashboardHandler)

		// /auth/ endpoints
		authRoutes(app)

		activities := app.Group("/activities")
		activities.GET("/sync-all", SyncAllActivitiesHandler)
		activities.GET("/sync", SyncLastActivitiesHandler)
		activities.Resource("", ActivitiesResource{})

		users := app.Group("/users")
		users.GET("", ListUsersHandler)
		users.GET("/{user_id}", ShowUsersHandler)
		users.GET("/{user_id}/refresh", RefreshUsersHandler)
		users.GET("/{user_id}/activities", ListUserActivitiesHandler)
		users.GET("/{user_id}/sync", SyncUserLatestActivitiesHandler)
		users.GET("/{user_id}/sync-all", SyncUserAllActivitiesHandler)

		app.GET("/dashboard", DashboardHandler)

		app.GET("/info", getInfo) // XXX: remove this in production

		app.ServeFiles("/", assetsBox) // serve files from the public directory
	}

	return app
}

// translations will load locale files, set up the translator `actions.T`,
// and will return a middleware to use to load the correct locale for each
// request.
// for more information: https://gobuffalo.io/en/docs/localization
func translations() buffalo.MiddlewareFunc {
	var err error
	if T, err = i18n.New(packr.New("app:locales", "../locales"), "en-US"); err != nil {
		app.Stop(err)
	}
	return T.Middleware()
}

// forceSSL will return a middleware that will redirect an incoming request
// if it is not HTTPS. "http://example.com" => "https://example.com".
// This middleware does **not** enable SSL. for your application. To do that
// we recommend using a proxy: https://gobuffalo.io/en/docs/proxy
// for more information: https://github.com/unrolled/secure/
func forceSSL() buffalo.MiddlewareFunc {
	return forcessl.Middleware(secure.Options{
		SSLRedirect:     ENV == "production",
		SSLProxyHeaders: map[string]string{"X-Forwarded-Proto": "https"},
	})
}

func getInfo(c buffalo.Context) error {
	output := fmt.Sprintf("%+v\n", c)
	html := "<html><body><div style=\"white-space: pre-wrap;\"><pre style=\"white-space: pre-wrap;\">\n" + output + "\n</pre></div>"
	return c.Render(200,
		r.Func("text/html",
			func(w io.Writer, d render.Data) error {
				_, err := w.Write([]byte(html))
				return err
			}))
}
