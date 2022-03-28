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
	"encoding/json"
	"github.com/bieber/rettention/config"
	"github.com/bieber/rettention/util"
	"github.com/skratchdot/open-golang/open"
	"io"
	"io/ioutil"
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

	request, err := util.APIRequest(c, credential, "GET", "me", nil)
	if err != nil {
		log.Fatal(err)
	}

	response, err := http.DefaultClient.Do(request)
	if err != nil {
		log.Fatal(err)
	}

	defer response.Body.Close()
	if bytes, err := ioutil.ReadAll(response.Body); err == nil {
		log.Println(string(bytes))
	} else {
		log.Fatal(err)
	}
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
	q.Set("scope", "read edit identity")
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
	request, err := util.TokenRequest(
		c,
		"POST",
		"https://www.reddit.com/api/v1/access_token",
		map[string]string{
			"grant_type":   "authorization_code",
			"code":         code,
			"redirect_uri": c.RedirectURI,
		},
	)
	if err != nil {
		log.Fatal(err)
	}

	response, err := http.DefaultClient.Do(request)
	if err != nil {
		log.Fatal(err)
	}

	defer response.Body.Close()
	credential := config.Credential{}
	decoder := json.NewDecoder(response.Body)
	if err := decoder.Decode(&credential); err != nil {
		log.Fatal(err)
	}

	credential.Expiration += time.Now().Unix()
	return credential
}
