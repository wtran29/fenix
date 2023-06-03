package session

import (
	"fmt"
	"reflect"
	"testing"

	"github.com/alexedwards/scs/v2"
)

func TestSession_InitSession(t *testing.T) {
	c := &Session{
		CookieLifetime: "100",
		CookiePersist:  "true",
		CookieName:     "session-id",
		CookieDomain:   "localhost",
		SessionType:    "cookie",
		CookieSecure:   "true",
	}

	var sm *scs.SessionManager

	s := c.InitSession()

	var sKind reflect.Kind
	var sType reflect.Type

	rv := reflect.ValueOf(s)

	for rv.Kind() == reflect.Ptr || rv.Kind() == reflect.Interface {
		fmt.Println("For loop:", rv.Kind(), rv.Type(), rv)
		sKind = rv.Kind()
		sType = rv.Type()

		rv = rv.Elem()
	}

	if !rv.IsValid() {
		t.Errorf("invalid type or kind; kind:%v type:%v", rv.Kind(), rv.Type())
	}

	if sKind != reflect.ValueOf(sm).Kind() {
		t.Errorf("wrong KIND returned - testing cookie session. Expected: %v, Got: %v", reflect.ValueOf(sm).Kind(), sKind)
	}

	if sType != reflect.ValueOf(sm).Type() {
		t.Errorf("wrong TYPE returned - testing cookie session. Expected: %v, Got: %v", reflect.ValueOf(sm).Type(), sType)
	}
}
