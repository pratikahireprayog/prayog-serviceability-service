package shipcube

import (
	"net/http"
	"time"

	"prayog-serviceability-service/internal/shared/config"

	"github.com/sirupsen/logrus"
)

type ShipCubeClient struct {
	httpClient *http.Client
	config     config.ShipCubeConfig
	logger     *logrus.Logger
	token      string
	tokenExpiry time.Time
}


// func NewShipCubeClient(cfg config.ShipCubeConfig) *ShipCubeClient {
// 	return &ShipCubeClient{
// 		httpClient: &http.Client{
// 			Timeout: cfg.Timeout,
// 		},
// 		config: cfg,
// 	}
// }


func NewShipcubeClient(config config.ShipCubeConfig) *ShipCubeClient {
	logger := logrus.New()
	logger.SetLevel(logrus.InfoLevel)

	httpClient := &http.Client{
		Timeout: config.Timeout,
		Transport: &http.Transport{
			MaxIdleConns:        10,
			MaxIdleConnsPerHost: 5,
			IdleConnTimeout:     30 * time.Second,
		},
	}

	return &ShipCubeClient{
		httpClient: httpClient,
		config:     config,
		logger:     logger,
	}
}