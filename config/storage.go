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

import (
	"github.com/go-yaml/yaml"
	"log"
	"os"
)

func ReadConfig(path string) Config {
	c := Config{configFile: path}
	fin, err := os.Open(path)
	if err != nil {
		log.Fatal(err)
	}
	defer fin.Close()

	decoder := yaml.NewDecoder(fin)
	if err := decoder.Decode(&c); err != nil {
		log.Println("Error reading credentials", err)
	}

	return c
}

func WriteConfig(c Config) {
	fout, err := os.Create(c.configFile)
	if err != nil {
		log.Fatal(err)
	}
	defer fout.Close()

	encoder := yaml.NewEncoder(fout)
	if err := encoder.Encode(c); err != nil {
		log.Fatal(err)
	}
}

func ReadCredentials(c Config) map[string]Credential {
	fin, err := os.Open(c.CredentialPath)
	if err != nil {
		return map[string]Credential{}
	}
	defer fin.Close()

	credentials := map[string]Credential{}
	decoder := yaml.NewDecoder(fin)
	if err := decoder.Decode(credentials); err != nil {
		log.Fatal(err)
	}

	return credentials
}

func WriteCredentials(c Config, credentials map[string]Credential) {
	fout, err := os.Create(c.CredentialPath)
	if err != nil {
		log.Fatal(err)
	}
	defer fout.Close()

	encoder := yaml.NewEncoder(fout)
	if err := encoder.Encode(credentials); err != nil {
		log.Fatal(err)
	}
}
