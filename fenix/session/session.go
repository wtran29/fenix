package session

import (
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/alexedwards/scs/v2"
)

type Session struct {
	CookieLifetime string
	CookiePersist  string
	CookieName     string
	CookieDomain   string
	SessionType    string
	CookieSecure   string
}

func (f *Session) InitSession() *scs.SessionManager {
	var persist, secure bool

	// how long do we keep the session?
	min, err := strconv.Atoi(f.CookieLifetime)
	if err != nil {
		min = 60
	}

	// cookies persist?
	if strings.ToLower(f.CookiePersist) == "true" {
		persist = true
	}

	// cookies secure?
	if strings.ToLower(f.CookieSecure) == "true" {
		secure = true
	}

	// create session
	session := scs.New()
	session.Lifetime = time.Duration(min) * time.Minute
	session.Cookie.Persist = persist
	session.Cookie.Name = f.CookieName
	session.Cookie.Secure = secure
	session.Cookie.Domain = f.CookieDomain
	session.Cookie.SameSite = http.SameSiteLaxMode

	// which session store?
	switch strings.ToLower(f.SessionType) {
	case "redis":
	case "mysql":
	case "mariadb":
	case "postgres", "postgresql":
	default:
		// cookie
	}
	return session
}
