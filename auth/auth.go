/*

 Copyright 2022, Robert Bieber

 This program is free software: you can redistribute it and/or modify
 it under the terms of the GNU General Public License as published by
 the Free Software Foundation, either version 3 of the License, or (at
 your option) any later version.

 This program is distributed in the hope that it will be useful, but
 WITHOUT ANY WARRANTY; without even the implied warranty of
 MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE. See the GNU
 General Public License for more details.

 You should have received a copy of the GNU General Public License
 along with this program. If not, see <https://www.gnu.org/licenses/>.

*/

package auth

import (
	"context"
	"crypto/rand"
	"github.com/bieber/rettention/config"
	"github.com/bieber/rettention/util"
	"github.com/skratchdot/open-golang/open"
	"io"
	"log"
	"math/big"
	"net/http"
	"net/url"
	"time"
)

type authResponse struct {
	error  string
	code   string
	secret string
}

func Authenticate(c config.Config) {
	code := authUser(c)
	credential := fetchToken(c, code)

	dest := struct {
		Name string `json:"name"`
	}{}
	err := util.DoAPIRequest(credential, "GET", "api/v1/me", nil, &dest)
	if err != nil {
		log.Fatal(err)
	}
	if dest.Name == "" {
		log.Fatal("Failed to fetch username")
	}

	credentials := config.ReadCredentials(c)
	credentials[dest.Name] = credential
	config.WriteCredentials(c, credentials)

	if _, ok := c.Users[dest.Name]; !ok {
		if c.Users == nil {
			c.Users = map[string]config.User{}
		}

		c.Users[dest.Name] = config.User{"forever", "forever", []string{}}
		config.WriteConfig(c)
	}

	log.Printf("Authenticated %s", dest.Name)
}

func Reauthenticate(c config.Config) map[string]config.Credential {
	existingCredentials := config.ReadCredentials(c)
	if len(existingCredentials) == 0 {
		return existingCredentials
	}

	newCredentials := map[string]config.Credential{}
	for username, oldCredential := range existingCredentials {
		newCredential := config.Credential{}
		err := util.DoTokenRequest(
			c,
			"POST",
			"https://www.reddit.com/api/v1/access_token",
			map[string]string{
				"grant_type":    "refresh_token",
				"refresh_token": oldCredential.RefreshToken,
			},
			&newCredential,
		)
		if err != nil {
			log.Printf("Failed to refresh credential for %s", username)
			newCredentials[username] = oldCredential
			continue
		}

		newCredential.Expiration += time.Now().Unix()
		newCredentials[username] = newCredential
	}

	config.WriteCredentials(c, newCredentials)
	return newCredentials
}

func randomSecret(length int) string {
	chars := []rune("abcdefghijklmnopqrstuvwxyz0123456789")
	secret := make([]rune, length)
	for i := 0; i < length; i++ {
		char, err := rand.Int(rand.Reader, big.NewInt(int64(len(chars))))
		if err != nil {
			log.Fatal(err)
		}

		secret[i] = chars[char.Int64()]
	}

	return string(secret)
}

func authUser(c config.Config) string {
	outputSignal := make(chan authResponse)
	closeSignal := make(chan bool)

	s := &http.Server{Addr: c.ServeAddress}

	secret := randomSecret(12)
	handler := func(w http.ResponseWriter, r *http.Request) {
		outputSignal <- authResponse{
			error:  r.FormValue("error"),
			code:   r.FormValue("code"),
			secret: r.FormValue("state"),
		}

		io.WriteString(
			w,
			`
			<html>
				<p>You can now close this window</p>
			</html>
            `,
		)

		closeSignal <- true
	}

	s.Handler = http.HandlerFunc(handler)
	go func() {
		if err := s.ListenAndServe(); err != http.ErrServerClosed {
			log.Fatal(err)
		}
	}()

	go func() {
		<-closeSignal
		ctx, _ := context.WithTimeout(context.Background(), time.Second*5)
		if err := s.Shutdown(ctx); err != nil {
			log.Fatal(err)
		}
	}()

	authURI, err := url.Parse("https://www.reddit.com/api/v1/authorize")
	if err != nil {
		log.Fatal(err)
	}

	q := authURI.Query()
	q.Set("client_id", c.AppID)
	q.Set("response_type", "code")
	q.Set("state", secret)
	q.Set("redirect_uri", c.RedirectURI)
	q.Set("duration", "permanent")
	q.Set("scope", "read edit identity history")
	authURI.RawQuery = q.Encode()

	open.Run(authURI.String())

	response := <-outputSignal

	if response.secret != secret {
		log.Fatalf("Authentication failed: Secrets do not match")
	}
	if response.error != "" {
		log.Fatalf("Authentication failed: %s", response.error)
	}

	return response.code
}

func fetchToken(c config.Config, code string) config.Credential {
	credential := config.Credential{}
	err := util.DoTokenRequest(
		c,
		"POST",
		"https://www.reddit.com/api/v1/access_token",
		map[string]string{
			"grant_type":   "authorization_code",
			"code":         code,
			"redirect_uri": c.RedirectURI,
		},
		&credential,
	)
	if err != nil {
		log.Fatal(err)
	}

	credential.Expiration += time.Now().Unix()
	return credential
}
