package session

import (
	"database/sql"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/alexedwards/scs/mysqlstore"
	"github.com/alexedwards/scs/postgresstore"
	"github.com/alexedwards/scs/redisstore"
	"github.com/alexedwards/scs/v2"
	"github.com/gomodule/redigo/redis"
)

type Session struct {
	CookieLifetime string
	CookiePersist  string
	CookieName     string
	CookieDomain   string
	SessionType    string
	CookieSecure   string
	DBPool         *sql.DB
	RedisPool      *redis.Pool
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
		session.Store = redisstore.New(f.RedisPool)
	case "mysql", "mariadb":
		session.Store = mysqlstore.New(f.DBPool)
	case "postgres", "postgresql":
		session.Store = postgresstore.New(f.DBPool)
	default:
		// cookie
	}

	return session

}
