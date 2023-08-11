/**
 * @author tsukiyo
 * @date 2023-08-12 1:33
 */

package api

import (
	"github.com/stretchr/testify/assert"
	"golang.org/x/crypto/bcrypt"
	"testing"
)

func TestEncrypt(t *testing.T) {
	pwd := "for.nothing"
	encryptedPwd, err := bcrypt.GenerateFromPassword([]byte(pwd), bcrypt.DefaultCost)
	if err != nil {
		t.Fatal(err)
	}
	err = bcrypt.CompareHashAndPassword(encryptedPwd, []byte(pwd))
	assert.NoError(t, err)
}
