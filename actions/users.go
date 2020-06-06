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
	if err := q.Order("name asc").All(users); err != nil {
		return err
	}

	return responder.Wants("html", func(c buffalo.Context) error {
		// Add the paginator to the context so it can be used in the template.
		c.Set("pagination", q.Paginator)

		c.Set("users", users)
		return c.Render(http.StatusOK, r.HTML("/users/index.plush.html"))
	}).Wants("json", func(c buffalo.Context) error {
		return c.Render(http.StatusOK, r.JSON(users))
	}).Wants("xml", func(c buffalo.Context) error {
		return c.Render(http.StatusOK, r.XML(users))
	}).Respond(c)
}

// ShowUsersHandler gets the data for one User. This function is mapped to
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
		return c.Render(http.StatusOK, r.JSON(user))
	}).Wants("xml", func(c buffalo.Context) error {
		return c.Render(http.StatusOK, r.XML(user))
	}).Respond(c)
}

// RefreshUsersHandler refresh user access tokens
func RefreshUsersHandler(c buffalo.Context) error {
	// Get the DB connection from the context
	tx, ok := c.Value("tx").(*pop.Connection)
	if !ok {
		return fmt.Errorf("no transaction found")
	}

	user := &models.User{}
	// To find the User the parameter user_id is used.
	if err := tx.Find(user, c.Param("user_id")); err != nil {
		return c.Error(http.StatusNotFound, err)
	}

	// get Strava Auth provider
	stravaProvider, ok := goth.GetProviders()[user.Provider]
	if !ok {
		c.Flash().Add("danger", fmt.Sprintf("%s connector is having a problem. Contact the admin", user.Provider))
		// TODO: add mailing here
		return c.Redirect(http.StatusTemporaryRedirect, "/users")
	}

	// refresh auth tokens
	newTokens, err := stravaProvider.RefreshToken(user.RefreshToken)
	if err != nil {
		c.Flash().Add("danger", "The token could not be refreshed")
		return c.Redirect(http.StatusTemporaryRedirect, "/users")
	}

	if user.AccessToken == newTokens.AccessToken {
		c.Flash().Add("success", fmt.Sprintf("Token was not renewed. Expires at %+v", newTokens.Expiry))
	} else {
		user.AccessToken = newTokens.AccessToken
		user.RefreshToken = newTokens.RefreshToken
		tx.Save(user)
		c.Flash().Add("success", fmt.Sprintf("New token for user %s expires at %+v", user.Name, newTokens.Expiry))
	}

	return c.Redirect(http.StatusTemporaryRedirect, "/users")
}

// ListUserActivitiesHandler will list all activities (does NOT call the provider)
func ListUserActivitiesHandler(c buffalo.Context) error {
	// Get the DB connection from the context
	tx, ok := c.Value("tx").(*pop.Connection)
	if !ok {
		return fmt.Errorf("no transaction found")
	}

	user := &models.User{}
	// To find the User the parameter user_id is used.
	if err := tx.Find(user, c.Param("user_id")); err != nil {
		return c.Error(http.StatusNotFound, err)
	}

	activities := &models.Activities{}
	// To find the User the parameter user_id is used.
	if err := tx.Where("user_id = ?", c.Param("user_id")).All(activities); err != nil {
		c.Flash().Add("error", fmt.Sprintf("Could not fetch activities (%s)", err))
		c.Logger().Error(err)
		return c.Redirect(http.StatusTemporaryRedirect, "/users/"+c.Param("user_id"))
	}

	return responder.Wants("html", func(c buffalo.Context) error {
		c.Set("meters2km", func(distance int) string {
			return fmt.Sprintf("%.2f", float64(distance)/1000.0)
		})
		c.Set("user", user)
		c.Set("activities", activities)
		return c.Render(http.StatusOK, r.HTML("/users/activities.plush.html"))
	}).Wants("json", func(c buffalo.Context) error {
		return c.Render(http.StatusOK, r.JSON(activities))
	}).Wants("xml", func(c buffalo.Context) error {
		return c.Render(http.StatusOK, r.XML(activities))
	}).Respond(c)
}

// RefreshAccessToken will refresh user's accessToken and refreshToken auth
func RefreshAccessToken(tx *pop.Connection, user *models.User) error {
	// get Strava Auth provider
	provider, ok := goth.GetProviders()[user.Provider]
	if !ok {
		return fmt.Errorf("%s connector is having a problem. Contact the admin", user.Provider)
	}

	// refresh auth tokens
	newTokens, err := provider.RefreshToken(user.RefreshToken)
	if err != nil {
		return fmt.Errorf("The accessToken for user '%s' could not be refreshed. %w", user.Name, err)
	}

	if user.AccessToken != newTokens.AccessToken {
		user.AccessToken = newTokens.AccessToken
		user.RefreshToken = newTokens.RefreshToken
		err = tx.Save(user)
	}

	return err
}
