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

package config

type command int

const (
	Auth command = iota
	Run
)

type User struct {
	Username        string
	CommentDuration string `yaml:"comment_duration"`
	PostDuration    string `yaml:"post_duration"`
}

type Config struct {
	AppID          string `yaml:"app_id"`
	AppSecret      string `yaml:"app_secret"`
	ServeAddress   string `yaml:"serve_address"`
	RedirectURI    string `yaml:"redirect_uri"`
	DeletesPerRun  int    `yaml:"deletes_per_run"`
	CredentialPath string `yaml:"credential_path"`
	Users          []User

	configFile string
}
