package controllers

import (
	"crypto/md5"
	"encoding/hex"
	"fmt"
	"io"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/irlTopper/ohlife2/app/interceptors"
	"github.com/irlTopper/ohlife2/app/models"
	"github.com/irlTopper/ohlife2/app/modules"
	"github.com/jmcvetta/randutil"
	"github.com/revel/revel"
	"github.com/revel/revel/cache"
)

type UserController struct {
	*revel.Controller
	interceptors.Authentication
}

type AutoBCCForm struct {
	Id      int64
	Enabled bool
}

func (c *UserController) PUT_v1_userId_resetpassword(
	userId int64,
	token string,
	password string,
) revel.Result {

	c.Validation.Required(token)
	c.Validation.Required(password)

	if c.Validation.HasErrors() {
		return modules.ShowValidationErrorsAsJSON(c.Controller)
	}

	user, _ := modules.DB.DbMap.Get(c.Controller, models.User{}, userId)

	if user == nil {
		return modules.JSONError(c.Controller, http.StatusNotFound, "User with id '"+strconv.FormatInt(userId, 10)+"' not found")
	}

	User, _ := user.(*models.User)

	var cacheToken string

	if err := cache.Get("user."+strconv.FormatInt(userId, 10)+".reset_token", &cacheToken); err != nil {
		return modules.JSONError(c.Controller, http.StatusNotFound, "Token not found and/or expired.")
	}

	if cacheToken != token {
		return modules.JSONError(c.Controller, http.StatusNotFound, "Token not found and/or expired.")
	}

	randInt, err := randutil.IntRange(1000000, 200000000)

	if err == nil {
		User.Salt.Int64 = int64(randInt)
		User.Salt.Valid = true

		hSalt := md5.New()
		io.WriteString(hSalt, strings.ToLower(password)+"-"+strconv.FormatInt(User.Salt.Int64, 10))

		User.Password = strings.ToUpper(hex.EncodeToString(hSalt.Sum(nil)))
	} else {
		hSalt := md5.New()
		io.WriteString(hSalt, strings.ToLower(password))

		User.Password = strings.ToUpper(hex.EncodeToString(hSalt.Sum(nil)))
	}

	modules.DB.DbMap.Update(c.Controller, User)

	deskUser, _ := modules.DB.Get(c.Controller, models.User{}, User.Id)

	dbDeskUser, _ := deskUser.(*models.User)

	// If the user's state is invited then this is actually their first time
	// logging in and setting their password, so we need to set it to active
	if strings.ToLower(dbDeskUser.State) == "invited" {
		dbDeskUser.State = "active"
		cache.Delete("user." + strconv.FormatInt(dbDeskUser.Id, 10) + ".activate")

		modules.DB.Update(c.Controller, dbDeskUser)
	}

	cache.Delete("user." + strconv.FormatInt(userId, 10) + ".reset_token")

	return c.POST_v1_login(User.Email, password, false)
}

func (c *UserController) POST_v1_users_forgotpassword(
	email string,
) revel.Result {
	c.Validation.Required(email)

	if c.Validation.HasErrors() {
		return modules.ShowValidationErrorsAsJSON(c.Controller)
	}

	c.User = &models.Session{
		User: models.User{
			Id: models.RobotUser,
		},
	}

	SQL := "SELECT userId FROM users WHERE userLogin = ? OR userEmail = ?"

	userId, _ := modules.DB.SelectInt(c.Controller, SQL, email, email)

	if userId == 0 {
		return modules.JSONError(c.Controller, http.StatusNotFound, "User with that email not found!")
	}

	return c.POST_v1_userId_resetpassword(userId)
}

func (c *UserController) POST_v1_userId_resetpassword(
	userId int64,
) revel.Result {
	user, _ := modules.DB.DbMap.Get(c.Controller, models.User{}, userId)

	if user == nil {
		return modules.JSONError(c.Controller, http.StatusNotFound, "User with id '"+strconv.FormatInt(userId, 10)+"' not found")
	}

	User, _ := user.(*models.User)

	token, err := randutil.AlphaString(20)

	if err != nil {
		return modules.JSONError(c.Controller, http.StatusInternalServerError, "Couldn't generate random token:"+err.Error())
	}

	cache.Set("user."+strconv.FormatInt(userId, 10)+".reset_token", token, time.Hour*1)

	data := map[string]interface{}{
		"token":     token,
		"firstName": User.FirstName,
		"lastName":  User.LastName,
		"userId":    User.Id,
		"host":      c.Request.Host,
	}

	fmt.Println("TODO", data)

	return c.RenderJson(map[string]interface{}{"status": "ok"})
}

// PUT /v1/users/:userId.json
// Updates a signle user
func (c UserController) PUT_v1_userId(
	userId int64,
	firstName string,
	lastName string,
	email string,
	timezoneId int64,
	password string,
	confirmPassword string,
) revel.Result {

	revel.INFO.Print(c.Params.Values)

	// Only for admins
	if c.User.Id != userId && !true {
		return modules.JSONError(c.Controller, 403, "Access denied")
	}

	// Get the user in TeamworkDesk shard
	obj, _ := modules.DB.Get(c.Controller, models.User{}, userId)
	if obj == nil {
		return modules.JSONError(c.Controller, http.StatusNotFound, "User with id '"+strconv.FormatInt(userId, 10)+"' not found")
	}
	User, _ := obj.(*models.User)

	if firstName != "" && (strings.Contains(firstName, "<") || strings.Contains(firstName, ">")) {
		return modules.JSONError(c.Controller, 400, "User's first name cannot contain < or >!")
	}

	if lastName != "" && (strings.Contains(lastName, "<") || strings.Contains(lastName, ">")) {
		return modules.JSONError(c.Controller, 400, "User's first name cannot contain < or >!")
	}

	// Validate password change if necessary
	if password != "" {
		if password != confirmPassword {
			return modules.JSONError(c.Controller, 400, "Password and confirmPassword must match!")
		}

		randInt, err := randutil.IntRange(1000000, 200000000)

		if err == nil {
			User.Salt.Int64 = int64(randInt)
			User.Salt.Valid = true

			hSalt := md5.New()
			io.WriteString(hSalt, strings.ToLower(password)+"-"+strconv.FormatInt(User.Salt.Int64, 10))

			User.Password = strings.ToUpper(hex.EncodeToString(hSalt.Sum(nil)))
		} else {
			hSalt := md5.New()
			io.WriteString(hSalt, strings.ToLower(password))

			User.Password = strings.ToUpper(hex.EncodeToString(hSalt.Sum(nil)))
		}
	}

	modules.SetStringIfSet(c.Controller, &User.FirstName, "firstName")
	modules.SetStringIfSet(c.Controller, &User.LastName, "lastName")
	modules.SetStringIfSet(c.Controller, &User.Email, "email")

	modules.SetIntIfSet(c.Controller, &User.TimezoneId, "timezoneId")

	// Update the database
	modules.DB.Update(c.Controller, User)

	return c.RenderJson(map[string]interface{}{"status": "ok"})
}

// POST	/v1/login.json
func (c UserController) POST_v1_login(
	username string,
	password string,
	rememberMe bool,
) revel.Result {

	// Validate params
	c.Validation.Required(username)
	c.Validation.Required(password)
	if c.Validation.HasErrors() {
		return modules.ShowValidationErrorsAsJSON(c.Controller)
	}

	// This is just a minor hack to remove the colon from the host.
	// Generally this would be the case when we're on a local development environment,
	// i.e localhost:9000.
	host := c.Request.Host
	if strings.Contains(host, ":") {
		host = host[:strings.LastIndex(host, ":")]
	}

	// Validate that this login works for the domain
	session, err := models.ValidateUser(username, password, true, host, c.Controller)
	if err != nil {
		return modules.JSONError(c.Controller, http.StatusUnauthorized, err.Error())
	}

	// Store the user in the redis cache
	if rememberMe {
		err = cache.Set(c.Session.Id()+"session", session, -1)
		c.Session.SetNoExpiration()
	} else {
		duration := time.Hour * 24 * 30

		if expireAfterDuration, err := time.ParseDuration(revel.Config.StringDefault("session.expires", "7d")); err == nil {
			duration = expireAfterDuration
		}

		err = cache.Set(c.Session.Id()+"session", session, duration)
	}

	if err != nil {
		panic(err)
	}

	return c.RenderJson(map[string]interface{}{"user": session})
}

func (c UserController) GET_v1_logout() revel.Result {
	cache.Delete(c.Session.Id() + "session")

	c.Response.Status = 200
	return c.RenderJson(map[string]interface{}{})
}
