package security

import (
	"bytes"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"io/ioutil"
	"log"

	"github.com/gin-gonic/gin"
)

func AuthMiddleware(secret string) gin.HandlerFunc {
	return func(c *gin.Context) {
		bodyBytes, _ := ioutil.ReadAll(c.Request.Body)

		calculatedHMAC, err := CalculateHMAC(bodyBytes, secret)
		if err != nil {
			log.Printf("Error decoding hmac secret %v", err)
			c.AbortWithStatus(500)
			return
		}
		s := c.Request.Header.Get("Signature")
		receivedHMACBytes, err := hex.DecodeString(s)
		if err != nil {
			log.Printf("Invalid signature, unable to decode hex value. Received %s", s)
			c.AbortWithStatus(401)
			return
		}
		if !hmac.Equal(calculatedHMAC, receivedHMACBytes) {
			log.Printf("Invalid signature. Received %s", s)
			c.AbortWithStatus(401)
			return
		}

		c.Request.Body.Close()
		c.Request.Body = ioutil.NopCloser(bytes.NewBuffer(bodyBytes))
		c.Next()
	}
}

func CalculateHMAC(b []byte, secret string) ([]byte, error) {
	s, err := hex.DecodeString(secret)
	if err != nil {
		return nil, err
	}
	mac := hmac.New(sha256.New, s)
	mac.Write(b)
	return mac.Sum(nil), nil
}
