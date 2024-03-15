package integration_test

import (
	"log"
	"net/http"
	"os"
	"testing"
	"time"

	. "github.com/Eun/go-hit"
)

const (
	// Attempts connection
	host       = "app:8080"
	healthPath = "http://" + host + "/healthz"
	attempts   = 20

	// HTTP REST
	basePath = "http://" + host + "/v1"
	linkPath = "http://" + host
)

func TestMain(m *testing.M) {
	err := healthCheck(attempts)
	if err != nil {
		log.Fatalf("Integration tests: host %s is not available: %s", host, err)
	}

	log.Printf("Integration tests: host %s is available", host)

	code := m.Run()
	os.Exit(code)
}

func healthCheck(attempts int) error {
	var err error

	for attempts > 0 {
		err = Do(Get(healthPath), Expect().Status().Equal(http.StatusOK))
		if err == nil {
			return nil
		}

		log.Printf("Integration tests: url %s is not available, attempts left: %d", healthPath, attempts)

		time.Sleep(time.Second)

		attempts--
	}

	return err
}

// HTTP POST: /oneTimeLink/create
// HTTP POST: /oneTimeLink/get
func TestHTTPCreateAndGet(t *testing.T) {
	var data, data2 = "some_data", "some_other_data"
	var link, link2 string

	body := map[string]string{
		"data": data,
	}
	Test(t,
		Description("Create Success"),
		Post(basePath+"/oneTimeLink/create"),
		Send().Headers("Content-Type").Add("application/json"),
		Send().Body().JSON(body),
		Expect().Status().Equal(http.StatusOK),
		Expect().Body().JSON().JQ(".link").NotEqual(""),
		Store().Response().Body().JSON().JQ(".link").In(&link),
	)

	body = map[string]string{
		"data": data2,
	}
	Test(t,
		Description("Create Success"),
		Post(basePath+"/oneTimeLink/create"),
		Send().Headers("Content-Type").Add("application/json"),
		Send().Body().JSON(body),
		Expect().Status().Equal(http.StatusOK),
		Expect().Body().JSON().JQ(".link").NotEqual(""),
		Store().Response().Body().JSON().JQ(".link").In(&link2),
	)

	body = map[string]string{
		"link": "invalid_link",
	}
	Test(t,
		Description("Get not found"),
		Post(basePath+"/oneTimeLink/get"),
		Send().Headers("Content-Type").Add("application/json"),
		Send().Body().JSON(body),
		Expect().Status().Equal(http.StatusBadRequest),
		Expect().Body().JSON().JQ(".error").Equal("invalid link"),
	)

	body = map[string]string{
		"link": link,
	}
	Test(t,
		Description("Get Success"),
		Post(basePath+"/oneTimeLink/get"),
		Send().Headers("Content-Type").Add("application/json"),
		Send().Body().JSON(body),
		Expect().Status().Equal(http.StatusOK),
		Expect().Body().JSON().JQ(".data").Equal(data),
	)

	body = map[string]string{
		"link": link,
	}
	Test(t,
		Description("Get Success"),
		Post(basePath+"/oneTimeLink/get"),
		Send().Headers("Content-Type").Add("application/json"),
		Send().Body().JSON(body),
		Expect().Status().Equal(http.StatusGone),
		Expect().Body().JSON().JQ(".error").Equal("link expired"),
	)

	Test(t,
		Description("Get Success"),
		Get(linkPath+"/"+link2),
		Send().Headers("Content-Type").Add("application/json"),
		Expect().Status().Equal(http.StatusOK),
		Expect().Body().JSON().JQ(".data").Equal(data2),
	)
}
