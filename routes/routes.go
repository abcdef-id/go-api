package routes

import (
	c "github.com/abcdef-id/go-api/controllers"
	teakAuth "github.com/abcdef-id/go-api/routes/auth"

	"github.com/labstack/echo"
	"github.com/labstack/echo/middleware"
)

func acl(permission string) echo.MiddlewareFunc {
	return teakAuth.ACL(permission)
}

func Api(e *echo.Echo) {
	e.POST("login", c.LoginUser)
	e.GET("profile", c.GetProfile, middleware.JWTWithConfig(teakAuth.JwtConfig))

	// user
	user := e.Group("user", middleware.JWTWithConfig(teakAuth.JwtConfig))
	user.GET("", c.GetUser)
	user.POST("", c.AddUser, acl("your_permission"))
	user.PUT("/:id", c.UpdateUser, acl("your_permission"))

}
