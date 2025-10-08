package security

import (
	"bytes"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"

	"github.com/stretchr/testify/assert"
)

const s = "a1b2c3d4a1b2c3d4a1b2c3d4a1b2c3d4a1b2c3d4a1b2c3d4a1b2c3d4a1b2c3d4"

type mockBody struct {
	param string
}

func TestCalculateHMAC(t *testing.T) {
	b := []byte("a1b2c3d4a1b2c3d4a1b2c3d4a1b2c3d4a1b2c3d4a1b2c3d4a1b2c3d4a1b2c3d4")
	h, err := CalculateHMAC(b, s)
	if err != nil {
		t.Errorf("failed to calculate HMAC")
	}
	actual := hex.EncodeToString(h)
	sb, _ := hex.DecodeString(s)
	m := hmac.New(sha256.New, sb)
	m.Write(b)
	expected := hex.EncodeToString(m.Sum(nil))
	assert.Equal(t, expected, actual)
}

func TestAuthMiddlewareOk(t *testing.T) {
	r := gin.Default()

	r.GET("/testHMAC", AuthMiddleware(s), func(c *gin.Context) {
		c.JSON(200, "OK")
		return
	})
	requestBody := mockBody{
		param: "asdf",
	}
	requestBodyBytes, _ := json.Marshal(requestBody)

	response := httptest.NewRecorder()
	request, err := http.NewRequest("GET", "/testHMAC", bytes.NewReader(requestBodyBytes))
	if err != nil {
		t.Errorf("failed setting up test request")
	}
	h, err := CalculateHMAC(requestBodyBytes, s)
	if err != nil {
		t.Errorf("failed to calculate HMAC")
	}
	request.Header.Set("Signature", hex.EncodeToString(h))
	r.ServeHTTP(response, request)

	assert.Equal(t, http.StatusOK, response.Code, "Expected code %d, received code %d", http.StatusOK,
		response.Code)
}

func TestAuthMiddlewareUnauthorized(t *testing.T) {
	r := gin.Default()

	r.GET("/testHMAC", AuthMiddleware(s), func(c *gin.Context) {
		c.JSON(200, "OK")
		return
	})
	requestBody := mockBody{
		param: "asdf",
	}
	requestBodyBytes, _ := json.Marshal(requestBody)

	response := httptest.NewRecorder()
	request, err := http.NewRequest("GET", "/testHMAC", bytes.NewReader(requestBodyBytes))
	if err != nil {
		t.Errorf("failed setting up test request")
	}
	request.Header.Set("Signature", "a1b2c3d4a1b2c3d4a1b2c3d4a1b2c3d4a1b2c3d4a1b2c3d4a1b2c3d4a1b2c3d4")
	r.ServeHTTP(response, request)

	assert.Equal(t, http.StatusUnauthorized, response.Code, "Expected code %d, received code %d",
		http.StatusOK, response.Code)
}
