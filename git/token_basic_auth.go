package git

import (
	"encoding/base64"
	"fmt"
	"net/http"
)

type tokenBasicAuth struct {
	Token string
}

func NewTokenBasicAuth(token string) *tokenBasicAuth {
	return &tokenBasicAuth{Token: token}
}

func (a *tokenBasicAuth) SetAuth(r *http.Request) {
	if a == nil {
		return
	}
	token := base64.StdEncoding.EncodeToString([]byte(fmt.Sprintf(`:%s`, a.Token)))
	r.Header.Add("Authorization", fmt.Sprintf("Basic %s", token))
}

// Name is name of the auth
func (a *tokenBasicAuth) Name() string {
	return "http-token-basic-auth"
}

func (a *tokenBasicAuth) String() string {
	masked := "*******"
	if a.Token == "" {
		masked = "<empty>"
	}
	return fmt.Sprintf("%s - %s", a.Name(), masked)
}
