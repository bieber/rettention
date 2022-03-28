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
	"net/http"
	"net/url"
	"strings"
)

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

func APIRequest(
	c config.Config,
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
		"https://oauth.reddit.com/api/v1/"+path,
		bodyBuffer,
	)
	if err != nil {
		return r, err
	}

	AddUserAgent(r)
	r.Header.Add("Authorization", "bearer "+credential.AccessToken)

	return r, nil
}

func AddUserAgent(r *http.Request) {
	r.Header.Add("User-Agent", "rettention:v0.0.1 (by /u/robertbieber)")
}
