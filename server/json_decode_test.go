package server

import (
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/gin-gonic/gin"
)

func TestDecodeJSONRejectsTrailingDocumentsAndUnknownFields(t *testing.T) {
	gin.SetMode(gin.TestMode)
	for _, test := range []struct {
		name string
		body string
	}{
		{name: "trailing object", body: `{"value":"ok"} {}`},
		{name: "unknown field", body: `{"value":"ok","unexpected":true}`},
	} {
		t.Run(test.name, func(t *testing.T) {
			context, _ := gin.CreateTestContext(httptest.NewRecorder())
			context.Request = httptest.NewRequest("POST", "/api/agent/v1/test", strings.NewReader(test.body))
			var target struct {
				Value string `json:"value"`
			}
			if err := decodeJSON(context, &target); err == nil {
				t.Fatal("expected strict JSON rejection")
			}
		})
	}
}

func TestDecodeJSONAcceptsOneDocumentWithWhitespace(t *testing.T) {
	gin.SetMode(gin.TestMode)
	context, _ := gin.CreateTestContext(httptest.NewRecorder())
	context.Request = httptest.NewRequest("POST", "/api/agent/v1/test", strings.NewReader(" \n{\"value\":\"ok\"}\n \t"))
	var target struct {
		Value string `json:"value"`
	}
	if err := decodeJSON(context, &target); err != nil {
		t.Fatal(err)
	}
	if target.Value != "ok" {
		t.Fatalf("decoded value=%q", target.Value)
	}
}

func TestDecodeJSONRejectsOversizedBody(t *testing.T) {
	gin.SetMode(gin.TestMode)
	body := `{"value":"` + strings.Repeat("x", 4<<20) + `"}`
	context, _ := gin.CreateTestContext(httptest.NewRecorder())
	context.Request = httptest.NewRequest("POST", "/api/agent/v1/test", strings.NewReader(body))
	var target struct {
		Value string `json:"value"`
	}
	if err := decodeJSON(context, &target); err == nil {
		t.Fatal("expected oversized JSON body rejection")
	}
}
