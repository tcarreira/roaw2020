package actions

import (
	"github.com/gobuffalo/buffalo"
	"github.com/markbates/goth/gothic"
)

// authRoutes declaration of all /auth endpoints
func authRoutes(app *buffalo.App) {
	auth := app.Group("/auth")

	authLogout := auth.Group("/logout")
	authLogout.GET("", AuthDestroy)

	bah := buffalo.WrapHandlerFunc(gothic.BeginAuthHandler)
	auth.GET("/{provider}", bah)
	auth.Middleware.Skip(Authorize, bah, AuthCallback)
	auth.GET("/{provider}/callback", AuthCallback)
}
