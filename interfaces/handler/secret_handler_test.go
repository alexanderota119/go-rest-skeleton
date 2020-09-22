package handler_test

import (
	"encoding/json"
	"go-rest-skeleton/interfaces/handler"
	"go-rest-skeleton/pkg/encoder"
	"go-rest-skeleton/pkg/security"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
)

func TestGenerateSecret_Success(t *testing.T) {
	var secretData security.SecretKey
	secretHandler := handler.NewSecretHandler()

	gin.SetMode(gin.TestMode)
	w := httptest.NewRecorder()
	c, r := gin.CreateTestContext(w)
	v1 := r.Group("/api/v1/")
	v1.GET("/secret", secretHandler.GenerateSecret)

	var err error
	c.Request, err = http.NewRequest(http.MethodGet, "/api/v1/secret", nil)
	if err != nil {
		t.Errorf("this is the error: %v\n", err)
	}
	r.ServeHTTP(w, c.Request)

	response := encoder.ResponseDecoder(w.Body)
	data, _ := json.Marshal(response["data"])

	_ = json.Unmarshal(data, &secretData)

	assert.Equal(t, w.Code, http.StatusOK)
	assert.NotNil(t, secretData.PublicKey)
	assert.NotNil(t, secretData.PrivateKey)
}
