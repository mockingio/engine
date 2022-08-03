# Mockingio Engine

[![CI](https://github.com/mockingio/engine/actions/workflows/main.yml/badge.svg)](https://github.com/mockingio/engine/actions/workflows/main.yml)
[![codecov](https://codecov.io/gh/mockingio/engine/branch/main/graph/badge.svg?token=0AXGI7UR85)](https://codecov.io/gh/mockingio/engine)
[![Go Report Card](https://goreportcard.com/badge/github.com/mockingio/engine)](https://goreportcard.com/report/github.com/mockingio/engine)
[![License](https://img.shields.io/badge/License-Apache_2.0-blue.svg)](https://opensource.org/licenses/Apache-2.0)

### Go package usage

```go
import (
	"net/http"
	"testing"

	"github.com/mockingio/engine/pkg/engine"
)

func Test_Example(t *testing.T) {
	srv := engine.
		New().
		Get("/hello").
		Response(http.StatusOK, "hello world").
		Start(t)
	defer srv.Close()

	req, _ := http.NewRequest("GET", srv.URL, nil)
	client := &http.Client{}

	resp, err := client.Do(req)
}

func Test_Example_WithRules(t *testing.T) {
	srv := engine.
		New().
		Get("/hello").
		When("cookie", "name", "equal", "Chocolate").
		Response(http.StatusOK, "hello world").
		Start(t)
	defer srv.Close()

	req, _ := http.NewRequest("GET", srv.URL, nil)
	client := &http.Client{}

	resp, err := client.Do(req)
}
```