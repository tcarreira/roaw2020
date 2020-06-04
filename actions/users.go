package actions

import (
	"context"
	"fmt"
	"net/http"

	"github.com/gobuffalo/buffalo"
	"github.com/gobuffalo/pop/v5"
	"github.com/gobuffalo/x/responder"
	"github.com/markbates/goth"

	"github.com/tcarreira/roaw2020/models"
	"github.com/tcarreira/roaw2020/swagger"
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

	// Allocate an empty User
	user := &models.User{}

	// To find the User the parameter user_id is used.
	if err := tx.Find(user, c.Param("user_id")); err != nil {
		return c.Error(http.StatusNotFound, err)
	}

	stravaProvider, ok := goth.GetProviders()[user.Provider]
	if !ok {
		c.Flash().Add("danger", fmt.Sprintf("%s connector is having a problem. Contact the admin", user.Provider))
		// TODO: add mailing here
		return c.Redirect(http.StatusTemporaryRedirect, "/users")
	}

	newToken, err := stravaProvider.RefreshToken(user.RefreshToken)
	if err != nil {
		c.Flash().Add("danger", "The token could not be refreshed")
		return c.Redirect(http.StatusTemporaryRedirect, "/users")
	}

	if user.AccessToken == newToken.AccessToken {
		c.Flash().Add("success", fmt.Sprintf("Token was not renewed. Expires at %+v", newToken.Expiry))
	} else {
		user.AccessToken = newToken.AccessToken
		user.RefreshToken = newToken.RefreshToken
		tx.Save(user)
		c.Flash().Add("success", fmt.Sprintf("New token (%s) expires at %+v", newToken.AccessToken, newToken.Expiry))
	}

	return c.Redirect(http.StatusTemporaryRedirect, "/users")
}

// FetchActivitiesHandler will import all activities from the provider and populate the database
func FetchActivitiesHandler(c buffalo.Context) error {
	tx, ok := c.Value("tx").(*pop.Connection)
	if !ok {
		return fmt.Errorf("no transaction found")
	}

	user := &models.User{}
	// To find the User the parameter user_id is used.
	if err := tx.Find(user, c.Param("user_id")); err != nil {
		return c.Error(http.StatusNotFound, err)
	}

	client := swagger.NewAPIClient(swagger.NewConfiguration())

	ctx := context.WithValue(context.Background(), swagger.ContextAccessToken, user.AccessToken)
	activities, response, err := client.ActivitiesApi.GetLoggedInAthleteActivities(ctx, nil)
	if err != nil {
		c.Flash().Add("error", fmt.Sprintf("Could not fetch activities (%s)", err))
		return c.Redirect(http.StatusTemporaryRedirect, "/users/"+c.Param("user_id"))
	}

	return responder.Wants("html", func(c buffalo.Context) error {
		c.Set("user", user)
		c.Set("a1", fmt.Sprintf("%+v", activities))
		c.Set("a2", fmt.Sprintf("%+v", response))
		c.Set("activities", activities)
		return c.Render(http.StatusOK, r.HTML("/users/activities.plush.html"))
	}).Wants("json", func(c buffalo.Context) error {
		return c.Render(http.StatusOK, r.JSON(activities))
	}).Wants("xml", func(c buffalo.Context) error {
		return c.Render(http.StatusOK, r.XML(activities))
	}).Respond(c)
}
