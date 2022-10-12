package handler

import (
	"database/sql"
	"errors"
	"myapp/handler"
	model "myapp/models"
	service "myapp/services"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/go-playground/validator"
	"github.com/labstack/echo/v4"
	"github.com/stretchr/testify/assert"
	"github.com/uptrace/bun/driver/sqliteshim"
)

// Mock functions (don't use if not needed)
type (
	MockHandler     struct{}
	MockDB          struct{}
	CustomValidator struct {
		validator *validator.Validate
	}
)

func (mockDb *MockDB) Connect() (*sql.DB, error) {
	return nil, errors.New("Error connecting to database")
}

func (MockHandler *MockHandler) GreetUser(c echo.Context) error {
	return c.JSON(http.StatusOK, "test")
}

func (mockHandler *MockHandler) GenerateToken(c echo.Context) error {
	tokenData := &handler.TokenData{
		Token: "fake token",
	}

	return c.JSON(http.StatusCreated, tokenData)
}

func (MockHandler *MockHandler) CreateToken(c echo.Context, user *handler.User) string {
	return "mock token"
}

func (cv *CustomValidator) Validate(i interface{}) error {
	if err := cv.validator.Struct(i); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, strings.Split(err.Error(), ":")[2])
	}

	return nil
}

// =================================================================

var (
	expected    = "{\"name\":\"Bob\",\"email\":\"bob123@test.com\"}\n"
	fakeRequest = `{"name":"Bob","email":"bonb@test.com"}`
)

func TestGreetUserHandler(t *testing.T) {
	rec, c := SetupTest("/api/v1/", http.MethodGet)
	userHandler := &handler.UserHandler{}
	if assert.NoError(t, userHandler.GreetUser(c)) {
		assert.Equal(t, http.StatusOK, rec.Code)
		assert.Equal(t, expected, rec.Body.String())
	}
}

func TestGenerateTokenHandler(t *testing.T) {
	e := echo.New()
	e.Validator = &CustomValidator{validator: validator.New()}

	req := httptest.NewRequest(http.MethodPost, "/api/v1/token", strings.NewReader(fakeRequest))
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)

	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	var userHandler handler.UserInterface = &MockHandler{}

	if assert.NoError(t, userHandler.GenerateToken(c)) {
		assert.Equal(t, http.StatusCreated, rec.Code)
		assert.Equal(t, "{\"token\":\"fake token\"}\n", rec.Body.String())
	}
}

// func TestHealthCheckHandler(t *testing.T) {
// 	e := echo.New()
// 	e.Validator = &CustomValidator{validator: validator.New()}

// 	req := httptest.NewRequest(http.MethodGet, "/api/v1/health", nil)
// 	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)

// 	rec := httptest.NewRecorder()
// 	c := e.NewContext(req, rec)

// 	var healthHandler handler.HealthCheckInterface = handler.CreateHealthCheckHandler("check error")

// 	if assert.NoError(t, healthHandler.GetHealthCheck(c)) {
// 		assert.Equal(t, http.StatusOK, rec.Code)
// 	}
// }

func TestHealthCheckService(t *testing.T) {
	t.Run("Database should be unhealthy", func(t *testing.T) {
		var mockDB *MockDB = &MockDB{}
		var expectedData []model.HealthCheckData = []model.HealthCheckData{
			{
				Name:   "API",
				Status: "Healthy",
			},
			{
				Name:   "Database",
				Status: "Unhealthy",
			},
		}

		data := service.HealthCheckService(mockDB)

		assert.Equal(t, expectedData, data)
	})

	t.Run("Database should be healthy", func(t *testing.T) {
		db := handler.CreateDB(sqliteshim.ShimName, "file::memory:")

		var expectedData []model.HealthCheckData = []model.HealthCheckData{
			{
				Name:   "API",
				Status: "Healthy",
			},
			{
				Name:   "Database",
				Status: "Healthy",
			},
		}

		data := service.HealthCheckService(db)

		assert.Equal(t, expectedData, data)
	})
}

func SetupTest(path string, method string) (*httptest.ResponseRecorder, echo.Context) {
	e := echo.New()

	req := httptest.NewRequest(method, path, nil)
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)

	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	return rec, c
}
