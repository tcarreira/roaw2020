package actions

import (
	"fmt"
	"net/http"

	"github.com/gobuffalo/buffalo"
	"github.com/gobuffalo/pop/v5"
	"github.com/gobuffalo/x/responder"
	"github.com/markbates/goth"

	"github.com/tcarreira/roaw2020/models"
)

// ListUsersHandler gets all Users. This function is mapped to the path
// GET /users
func ListUsersHandler(c buffalo.Context) error {
	// func (v UsersResource) List(c buffalo.Context) error {
	// Get the DB connection from the context
	tx, ok := c.Value("tx").(*pop.Connection)
	if !ok {
		return fmt.Errorf("no transaction found")
	}

	users := &models.Users{}

	// Paginate results. Params "page" and "per_page" control pagination.
	// Default values are "page=1" and "per_page=20".
	q := tx.PaginateFromParams(c.Params())

	// Retrieve all Users from the DB
	if err := q.All(users); err != nil {
		return err
	}

	return responder.Wants("html", func(c buffalo.Context) error {
		// Add the paginator to the context so it can be used in the template.
		c.Set("pagination", q.Paginator)

		c.Set("users", users)
		return c.Render(http.StatusOK, r.HTML("/users/index.plush.html"))
	}).Wants("json", func(c buffalo.Context) error {
		return c.Render(200, r.JSON(users))
	}).Wants("xml", func(c buffalo.Context) error {
		return c.Render(200, r.XML(users))
	}).Respond(c)
}

// Show gets the data for one User. This function is mapped to
// the path GET /users/{user_id}
func ShowUsersHandler(c buffalo.Context) error {
	// Get the DB connection from the context
	tx, ok := c.Value("tx").(*pop.Connection)
	if !ok {
		return fmt.Errorf("no transaction found")
	}

	// Allocate an empty User
	user := &models.User{}

	// To find the User the parameter user_id is used.
	if err := tx.Find(user, c.Param("user_id")); err != nil {
		return c.Error(http.StatusNotFound, err)
	}

	return responder.Wants("html", func(c buffalo.Context) error {
		c.Set("user", user)

		return c.Render(http.StatusOK, r.HTML("/users/show.plush.html"))
	}).Wants("json", func(c buffalo.Context) error {
		return c.Render(200, r.JSON(user))
	}).Wants("xml", func(c buffalo.Context) error {
		return c.Render(200, r.XML(user))
	}).Respond(c)
}

// Show gets the data for one User. This function is mapped to
// the path GET /users/{user_id}
func RefreshUsersHandler(c buffalo.Context) error {
	// Get the DB connection from the context
	tx, ok := c.Value("tx").(*pop.Connection)
	if !ok {
		return fmt.Errorf("no transaction found")
	}

	// Allocate an empty User
	user := &models.User{}

	// To find the User the parameter user_id is used.
	if err := tx.Find(user, c.Param("user_id")); err != nil {
		return c.Error(http.StatusNotFound, err)
	}

	gu := goth.User{}
	gu.AccessToken = user.AccessToken
	gu.RefreshToken = user.RefreshToken

	// stravaProvider := goth.GetProviders()["strava"]
	// stravaProvider.
	// stravaProvider.RefreshToken(user.RefreshToken)

	return c.Redirect(302, "/users")
}
