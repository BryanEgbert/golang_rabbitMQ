package service

import (
	"database/sql"
	model "myapp/models"
)

type DB interface {
	Connect() (*sql.DB, error)
}

func HealthCheckService(db DB) []model.HealthCheckData {
	_, err := db.Connect()

	var dbStatus string = "Healthy"
	if err != nil {
		dbStatus = "Unhealthy"
	}

	var data []model.HealthCheckData = []model.HealthCheckData{
		{
			Name:   "API",
			Status: "Healthy",
		},
		{
			Name:   "Database",
			Status: dbStatus,
		},
	}

	return data
}
