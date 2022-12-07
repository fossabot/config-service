package main

import (
	"bytes"
	"fmt"

	"net/http"
	"net/http/httptest"
	"os/exec"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/goccy/go-json"
	"github.com/stretchr/testify/suite"
)

/*
Coverage report cmd
go test -timeout 30s  -coverpkg=./dbhandler,./mongo,./types,./utils,./prob,./login,./cluster,./posture_exception,./vulnerability_exception,./customer_config -coverprofile coverage.out
go tool cover -html=coverage.out -o coverage.html
*/

const (
	mongoDockerCommand = `docker run --name=mongo -d -p 27017:27017 -e  "MONGO_INITDB_ROOT_USERNAME=admin" -e "MONGO_INITDB_ROOT_PASSWORD=admin" mongo`
	mongoStopCommand   = "docker stop mongo && docker rm mongo"
)

func TestConfigServiceWithMongoImage(t *testing.T) {
	suite.Run(t, new(MainTestSuite))
}

type MainTestSuite struct {
	suite.Suite
	router       *gin.Engine
	shutdownFunc func()
	authCookie   string
}

func (suite *MainTestSuite) SetupSuite() {
	out, err := exec.Command("/bin/sh", "-c", mongoDockerCommand).Output()
	if err != nil {
		suite.FailNow("failed to start mongo", err.Error(), string(out))
	}

	//initialize
	suite.shutdownFunc = initialize()
	//Create routes
	suite.router = setupRouter()

	checkReadiness := func() error {
		w := suite.doRequest(http.MethodGet, "/readiness", nil)
		if w.Code != http.StatusOK {
			return fmt.Errorf("failed to get readiness")
		}
		return nil
	}

	err = retry(10, time.Microsecond*10, checkReadiness)
	if err != nil {
		suite.FailNow("service is not ready readiness", err.Error())
	}

	loginDetails := struct {
		CustomerGUID string `json:"customerGUID"`
	}{
		CustomerGUID: "test-customer-guid",
	}

	w := suite.doRequest(http.MethodPost, "/login", loginDetails)
	if w.Code != http.StatusOK {
		suite.FailNow("failed to login")
	}
	suite.authCookie = w.Header().Get("Set-Cookie")
}

func (suite *MainTestSuite) TearDownSuite() {
	suite.shutdownFunc()
	exec.Command("/bin/sh", "-c", mongoStopCommand).Run()
}

func (suite *MainTestSuite) doRequest(method, path string, body interface{}) *httptest.ResponseRecorder {
	w := httptest.NewRecorder()
	var req *http.Request
	var reqErr error
	if body != nil {
		bodyBytes, err := json.Marshal(body)
		if err != nil {
			suite.FailNow("failed to marshal body", err.Error())
		}
		bodyReader := bytes.NewReader(bodyBytes)
		req, reqErr = http.NewRequest(method, path, bodyReader)
	} else {
		req, reqErr = http.NewRequest(method, path, nil)
	}
	if reqErr != nil {
		suite.FailNow("failed to create request", reqErr.Error())
	}

	if suite.authCookie != "" {
		req.Header.Set("Cookie", suite.authCookie)
	}
	suite.router.ServeHTTP(w, req)

	return w
}

func retry(attempt int, delay time.Duration, f func() error) error {
	var err error
	for i := 0; i < attempt; i++ {
		err = f()
		if err == nil {
			return nil
		}
		time.Sleep(delay)
	}
	return err
}
