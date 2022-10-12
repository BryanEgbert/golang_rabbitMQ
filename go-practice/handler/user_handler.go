package handler

import (
	"database/sql"
	model "myapp/models"
	service "myapp/services"

	"net/http"
	"time"

	"github.com/golang-jwt/jwt"
	"github.com/labstack/echo/v4"

	"github.com/uptrace/bun"
	"github.com/uptrace/bun/dialect/sqlitedialect"
	"github.com/uptrace/bun/driver/sqliteshim"
)

type (
	User struct {
		Name  string `json:"name" validate:"required"`
		Email string `json:"email" validate:"required,email"`
	}

	TokenData struct {
		Token string `json:"token"`
	}

	UserInterface interface {
		GreetUser(c echo.Context) error
		GenerateToken(c echo.Context) (err error)
	}

	TokenInterface interface {
		CreateToken(c echo.Context, user *User) string
	}

	HealthCheckInterface interface {
		GetHealthCheck(c echo.Context) error
	}

	UserHandler struct {
		Db *bun.DB
	}

	HealthCheckHandler struct {
		service.DB
	}
)

func CreateHealthCheckHandler(dbService service.DB) *HealthCheckHandler {
	return &HealthCheckHandler{
		dbService,
	}
}
func CreateUserHandler(sqlDb *sql.DB) *UserHandler {
	var db *bun.DB = bun.NewDB(sqlDb, sqlitedialect.New())

	return &UserHandler{
		Db: db,
	}
}

func (h *UserHandler) CreateToken(c echo.Context, user *User) string {
	token := jwt.New(jwt.SigningMethodHS256)
	claims := token.Claims.(jwt.MapClaims)
	claims["id"] = 100
	claims["nbf"] = time.Now().Unix()
	claims["iat"] = time.Now().Unix()
	claims["exp"] = time.Now().Add(time.Minute * 5).Unix()
	claims["role"] = "guest"
	claims["name"] = user.Name

	tokenString, err := token.SignedString([]byte("secretkey"))

	if err != nil {
		c.Logger().Panic(err)
	}

	return tokenString
}

func (h *UserHandler) GreetUser(c echo.Context) error {
	user := &User{
		Name:  "Bob",
		Email: "bob123@test.com",
	}

	return c.JSON(http.StatusOK, user)
}

func (h *UserHandler) GenerateToken(c echo.Context) (err error) {
	user := new(User)
	if err = c.Bind(user); err != nil {
		return c.String(http.StatusBadRequest, "bad request")
	}

	if err = c.Validate(user); err != nil {
		return err
	}

	token := jwt.New(jwt.SigningMethodHS256)
	claims := token.Claims.(jwt.MapClaims)
	claims["id"] = 100
	claims["nbf"] = time.Now().Unix()
	claims["iat"] = time.Now().Unix()
	claims["exp"] = time.Now().Add(time.Minute * 5).Unix()
	claims["role"] = "guest"
	claims["name"] = user.Name

	tokenString := h.CreateToken(c, user)
	tokenData := &TokenData{
		Token: tokenString,
	}

	return c.JSON(http.StatusCreated, tokenData)
}

type DBData struct {
	DriverName       string
	ConnectionString string
}

func (db *DBData) Connect() (*sql.DB, error) {
	sqldb, err := sql.Open(db.DriverName, db.ConnectionString)

	return sqldb, err
}

func CreateDB(driverName string, connectionString string) service.DB {
	return &DBData{
		DriverName:       driverName,
		ConnectionString: connectionString,
	}
}

var DbData service.DB = CreateDB(sqliteshim.ShimName, "file::memory:")

func (h *HealthCheckHandler) GetHealthCheck(c echo.Context) error {

	var data []model.HealthCheckData = service.HealthCheckService(DbData)

	return c.JSON(http.StatusOK, data)
}
