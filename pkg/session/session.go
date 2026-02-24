package session

import (
	"github.com/gorilla/sessions"
	"github.com/labstack/echo/v4"
)

const (
	AdminSessionName = "admin_session"
	WebSessionName   = "web_session"
	FlashSuccess     = "flash_success"
	FlashError       = "flash_error"
)

var (
	AdminStore *sessions.CookieStore
	WebStore   *sessions.CookieStore
)

func Init(secret string) {
	AdminStore = sessions.NewCookieStore([]byte("admin-" + secret))
	AdminStore.Options = &sessions.Options{
		Path:     "/admin",
		MaxAge:   86400 * 7,
		HttpOnly: true,
	}

	WebStore = sessions.NewCookieStore([]byte("web-" + secret))
	WebStore.Options = &sessions.Options{
		Path:     "/",
		MaxAge:   86400 * 30,
		HttpOnly: true,
	}
}

func GetAdminSession(c echo.Context) *sessions.Session {
	sess, _ := AdminStore.Get(c.Request(), AdminSessionName)
	return sess
}

func GetWebSession(c echo.Context) *sessions.Session {
	sess, _ := WebStore.Get(c.Request(), WebSessionName)
	return sess
}

func SetFlash(c echo.Context, sess *sessions.Session, key, value string) {
	sess.AddFlash(value, key)
	sess.Save(c.Request(), c.Response())
}

func GetFlash(c echo.Context, sess *sessions.Session, key string) []string {
	flashes := sess.Flashes(key)
	if len(flashes) > 0 {
		sess.Save(c.Request(), c.Response())
		result := make([]string, len(flashes))
		for i, f := range flashes {
			result[i] = f.(string)
		}
		return result
	}
	return nil
}
