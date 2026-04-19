package gin_tools

import (
	"bytes"
	"net/http"
	"net/http/httptest"
	"net/url"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
)

func TestRequestParams_Query(t *testing.T) {
	gin.SetMode(gin.TestMode)
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)

	req, _ := http.NewRequest("GET", "/test?page=1&limit=10", nil)
	c.Request = req

	params, err := RequestParams(c)
	assert.NoError(t, err)
	assert.Equal(t, "1", params["page"])
	assert.Equal(t, "10", params["limit"])
}

func TestRequestParams_PostJSON(t *testing.T) {
	gin.SetMode(gin.TestMode)
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)

	body := bytes.NewBuffer([]byte(`{"name":"test","age":20}`))
	req, _ := http.NewRequest("POST", "/test?extra=foo", body)
	req.Header.Set("Content-Type", "application/json")
	c.Request = req

	params, err := RequestParams(c)
	assert.NoError(t, err)
	assert.Equal(t, "foo", params["extra"])
	assert.Equal(t, "test", params["name"])
	assert.Equal(t, float64(20), params["age"]) // json decoder handles numbers as float64 mostly
}

func TestRequestParams_PostForm(t *testing.T) {
	gin.SetMode(gin.TestMode)
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)

	form := url.Values{}
	form.Add("keyword", "search")

	req, _ := http.NewRequest("POST", "/test?page=2", bytes.NewBufferString(form.Encode()))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	c.Request = req

	params, err := RequestParams(c)
	assert.NoError(t, err)
	assert.Equal(t, "2", params["page"])
	assert.Equal(t, "search", params["keyword"])
}
