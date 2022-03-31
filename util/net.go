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

package util

import (
	"bytes"
	"encoding/json"
	"github.com/bieber/rettention/config"
	"log"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"time"
)

// Pause when the rate limit drops to this number of requests to
// ensure we don't hit 0
const requestHeadroom = 5

// Add this many seconds to the wait for a rate limit reset to ensure
// we don't send too early
const resetHeadroom = 5 * time.Second

var incomingHeaders chan http.Header = make(chan http.Header)
var consumeRequest chan bool = make(chan bool)

func init() {
	go func() {
		consumeRequest <- true

		for {
			headers := <-incomingHeaders

			remainingRequests, err := strconv.ParseFloat(
				headers.Get("X-Ratelimit-Remaining"),
				64,
			)
			if err != nil {
				log.Fatal(err)
			}

			if remainingRequests < requestHeadroom {
				resetTime, err := strconv.ParseFloat(
					headers.Get("X-Ratelimit-Reset"),
					64,
				)
				if err != nil {
					log.Fatal(err)
				}

				time.Sleep(time.Duration(resetTime))
			}

			consumeRequest <- true
		}
	}()
}

func TokenRequest(
	c config.Config,
	method string,
	uri string,
	body map[string]string,
) (*http.Request, error) {

	bodyValues := url.Values{}
	for k, v := range body {
		bodyValues.Add(k, v)
	}

	r, err := http.NewRequest(
		method,
		uri,
		strings.NewReader(bodyValues.Encode()),
	)
	if err != nil {
		return r, err
	}

	AddUserAgent(r)
	r.SetBasicAuth(c.AppID, c.AppSecret)

	return r, nil
}

func DoTokenRequest(
	c config.Config,
	method string,
	uri string,
	body map[string]string,
	dest any,
) error {
	<-consumeRequest

	request, err := TokenRequest(c, method, uri, body)
	if err != nil {
		return err
	}

	response, err := http.DefaultClient.Do(request)
	if err != nil {
		return err
	}

	incomingHeaders <- response.Header

	defer response.Body.Close()
	decoder := json.NewDecoder(response.Body)
	if err := decoder.Decode(dest); err != nil {
		return err
	}

	return nil
}

func APIRequest(
	credential config.Credential,
	method string,
	path string,
	body any,
) (*http.Request, error) {

	bodyBuffer := bytes.NewBuffer([]byte{})
	if body != nil {
		encoder := json.NewEncoder(bodyBuffer)
		if err := encoder.Encode(body); err != nil {
			return nil, err
		}
	}

	r, err := http.NewRequest(
		method,
		"https://oauth.reddit.com/"+path,
		bodyBuffer,
	)
	if err != nil {
		return r, err
	}

	AddUserAgent(r)
	r.Header.Add("Authorization", "bearer "+credential.AccessToken)

	return r, nil
}

func DoAPIRequest(
	credential config.Credential,
	method string,
	path string,
	body any,
	dest any,
) error {
	<-consumeRequest

	request, err := APIRequest(credential, method, path, body)
	if err != nil {
		return err
	}

	response, err := http.DefaultClient.Do(request)
	if err != nil {
		return err
	}

	incomingHeaders <- response.Header

	defer response.Body.Close()
	decoder := json.NewDecoder(response.Body)
	if err := decoder.Decode(dest); err != nil {
		return err
	}

	return nil
}

func AddUserAgent(r *http.Request) {
	r.Header.Add("User-Agent", "rettention:v0.0.1 (by /u/robertbieber)")
}
