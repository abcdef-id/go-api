package controllers

import (
	"github.com/abcdef-id/go-api/presenter"
	"github.com/abcdef-id/go-api/routes/auth"

	"github.com/abcdef-id/go-model/models"

	"github.com/abcdef-id/go-core/config"

	"net/http"
	"time"

	jwt "github.com/dgrijalva/jwt-go"
	"github.com/labstack/echo"
	"github.com/labstack/gommon/log"
	"github.com/spf13/viper"
	"github.com/thedevsaddam/govalidator"
	"golang.org/x/crypto/bcrypt"
)

func LoginUser(c echo.Context) error {
	type LoginData struct {
		Username string `json:"username"`
		Password string `json:"password"`
	}
	var user models.User
	db := config.DB
	login := new(LoginData)
	c.Bind(login)

	if db.Where("username = ?", login.Username).First(&user).RecordNotFound() {
		return c.JSON(http.StatusUnauthorized, map[string]interface{}{"message": "Invalid username or password"})
	} else if !user.IsActive {
		return c.JSON(http.StatusUnauthorized, map[string]interface{}{"message": "User disabled"})
	} else if CheckPasswordHash(login.Password, user.Password) {

		// Set custom claims
		claims := &auth.JwtCustomClaims{
			user.ID,
			user.Username,
			jwt.StandardClaims{
				ExpiresAt: time.Now().Add(time.Hour * 72).Unix(),
			},
		}

		// Create token with claims
		token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)

		// Generate encoded token and send it as response.
		t, err := token.SignedString([]byte(viper.GetString("jwtSign")))
		if err != nil {
			return err
		}
		return c.JSON(http.StatusOK, echo.Map{
			"token": t,
		})
	} else {
		return c.JSON(http.StatusUnauthorized, map[string]interface{}{"message": "Invalid username or password"})
	}

	return echo.ErrUnauthorized
}

func CheckPasswordHash(password, hash string) bool {
	err := bcrypt.CompareHashAndPassword([]byte(hash), []byte(password))
	return err == nil
}

func GetProfile(c echo.Context) error {
	var user presenter.User
	// get database instance
	db := config.DB
	// get user id from token claim
	claim := getUser(c)

	// get user data by user id
	db.First(&user, claim.ID)

	return c.JSON(http.StatusOK, user)
}

func getUser(c echo.Context) *auth.JwtCustomClaims {
	user := c.Get("user").(*jwt.Token)
	claims := user.Claims.(*auth.JwtCustomClaims)
	return claims
}

func GetUser(c echo.Context) error {

	db := config.DB
	name := c.FormValue("name")
	var user []presenter.User
	if name != "" {
		db = db.Where("name LIKE ?", "%"+name+"%")
	}
	return c.JSON(http.StatusOK, user)
}

func AddUser(c echo.Context) error {
	var user models.User
	db := config.DB

	rules := govalidator.MapData{
		"username": []string{"required"},
		"email":    []string{"required", "email"},
		"password": []string{"required"},
	}

	opts := govalidator.Options{
		Request: c.Request(),
		Data:    &user,
		Rules:   rules,
	}

	v := govalidator.New(opts)
	e := v.ValidateJSON()

	if len(e) > 0 {
		err := map[string]interface{}{"validationError": e}
		return c.JSON(http.StatusBadRequest, err)
	}

	var users []models.User
	db.Where("username = ? ", user.Username).Find(&users)
	if len(users) > 0 {
		er := map[string]interface{}{"username": []string{"username already used"}}
		err := map[string]interface{}{"validationError": er}
		return c.JSON(http.StatusBadRequest, err)
	}

	db.Where("email = ? ", user.Email).Find(&users)
	if len(users) > 0 {
		er := map[string]interface{}{"email": []string{"email already used"}}
		err := map[string]interface{}{"validationError": er}
		return c.JSON(http.StatusBadRequest, err)
	}

	user.Password, _ = HashPassword(user.Password)

	if err := db.Create(&user).Error; err != nil {
		log.Error(err)
		return c.JSON(http.StatusBadRequest, err)
	}

	return c.JSON(http.StatusCreated, user)
}

func UpdateUser(c echo.Context) error {
	var newData models.User
	db := config.DB

	rules := govalidator.MapData{
		"username": []string{"required"},
		"email":    []string{"required", "email"},
	}

	opts := govalidator.Options{
		Request: c.Request(),
		Data:    &newData,
		Rules:   rules,
	}

	v := govalidator.New(opts)
	e := v.ValidateJSON()

	if len(e) > 0 {
		err := map[string]interface{}{"validationError": e}
		return c.JSON(http.StatusBadRequest, err)
	}

	var user models.User
	var count int

	id := c.Param("id")
	db.First(&user, id).Count(&count)
	if count != 1 {
		return c.String(404, "user does not exist")
	}
	if len(newData.Password) > 0 {
		newData.Password, _ = HashPassword(newData.Password)
	}
	db.Model(&user).Updates(newData)
	db.Model(&user).Update("is_active", newData.IsActive)

	return c.JSON(http.StatusCreated, nil)
}

func HashPassword(password string) (string, error) {
	bytes, err := bcrypt.GenerateFromPassword([]byte(password), 8)
	return string(bytes), err
}
