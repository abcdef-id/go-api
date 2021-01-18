package controllers

import (
	"net/http"
	"net/http/httptest"
	"testing"
	"strings"

	"github.com/labstack/echo"
	"github.com/stretchr/testify/assert"
)

func TestLoginUser(t *testing.T) {
	req := httptest.NewRequest(http.MethodPost, "/", strings.NewReader(userJSON))
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	c.SetPath("/login")
	h := LoginUser(c)
	
	// Assertions
	if assert.NoError(t, h) {
		assert.Equal(t, http.StatusOK, rec.Code)
	}
	
}