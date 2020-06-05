package actions

import (
	"context"
	"fmt"
	"net/http"
	"time"

	"github.com/antihax/optional"
	"github.com/gobuffalo/buffalo"
	"github.com/gobuffalo/pop/v5"
	"github.com/gobuffalo/x/responder"
	"github.com/markbates/goth"

	"github.com/tcarreira/roaw2020/models"
	"github.com/tcarreira/roaw2020/strava_client/swagger"
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

	var allActivities []swagger.SummaryActivity

	client := swagger.NewAPIClient(swagger.NewConfiguration())
	ctx := context.WithValue(context.Background(), swagger.ContextAccessToken, user.AccessToken)
	resultsPerPage := 20
	options := &swagger.ActivitiesApiGetLoggedInAthleteActivitiesOpts{
		// Before:  optional.EmptyInt32(),
		Before:  optional.NewInt32(int32(time.Date(2021, 1, 1, 0, 0, 0, 0, time.UTC).Unix())),
		After:   optional.NewInt32(int32(time.Date(2020, 1, 1, 0, 0, 0, 0, time.UTC).Unix())),
		Page:    optional.Int32{},
		PerPage: optional.NewInt32(int32(resultsPerPage)),
	}

	for i := int32(1); ; i++ {
		options.Page = optional.NewInt32(i)
		activities, _, err := client.ActivitiesApi.GetLoggedInAthleteActivities(ctx, options)
		allActivities = append(allActivities, activities...)

		if err != nil {
			c.Flash().Add("error", fmt.Sprintf("Could not fetch activities (%s)", err))
			return c.Redirect(http.StatusTemporaryRedirect, "/users/"+c.Param("user_id"))
		}

		if len(activities) != resultsPerPage {
			break
		}
	}

	return responder.Wants("html", func(c buffalo.Context) error {
		c.Set("user", user)
		c.Set("activities", allActivities)
		return c.Render(http.StatusOK, r.HTML("/users/activities.plush.html"))
	}).Wants("json", func(c buffalo.Context) error {
		return c.Render(http.StatusOK, r.JSON(allActivities))
	}).Wants("xml", func(c buffalo.Context) error {
		return c.Render(http.StatusOK, r.XML(allActivities))
	}).Respond(c)
}
