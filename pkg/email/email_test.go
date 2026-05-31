package email

import (
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestNewEmail(t *testing.T) {
	info := &SMTPInfo{
		Host:     "smtp.test.com",
		Port:     587,
		IsSSL:    true,
		UserName: "user",
		Password: "password",
		From:     "sender@test.com",
	}

	email := NewEmail(info)
	assert.NotNil(t, email)
	assert.Equal(t, "smtp.test.com", email.Host)
	assert.Equal(t, 587, email.Port)
	assert.True(t, email.IsSSL)
}
