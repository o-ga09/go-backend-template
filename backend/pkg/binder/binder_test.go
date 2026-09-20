package binder_test

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/labstack/echo"

	"github.com/o-ga09/go-backend-template/pkg/binder"
)

func TestBinder_Bind(t *testing.T) {
	type request struct {
		ID   string `param:"id" json:"-"`
		Name string `json:"name"`
	}

	tests := []struct {
		name     string
		paramID  string
		body     string
		wantID   string
		wantName string
	}{
		{"パスパラメータとJSONボディを両方バインドする", "user-1", `{"name":"taro"}`, "user-1", "taro"},
		{"パスパラメータが空でもJSONボディはバインドされる", "", `{"name":"taro"}`, "", "taro"},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			e := echo.New()
			e.Binder = binder.New()

			req := httptest.NewRequest(http.MethodPost, "/", strings.NewReader(tc.body))
			req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
			rec := httptest.NewRecorder()
			c := e.NewContext(req, rec)
			c.SetParamNames("id")
			c.SetParamValues(tc.paramID)

			var got request
			if err := c.Bind(&got); err != nil {
				t.Fatalf("Bind() error = %v", err)
			}
			if got.ID != tc.wantID {
				t.Errorf("ID = %q, want %q", got.ID, tc.wantID)
			}
			if got.Name != tc.wantName {
				t.Errorf("Name = %q, want %q", got.Name, tc.wantName)
			}
		})
	}
}
