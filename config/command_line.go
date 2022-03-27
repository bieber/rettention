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
	"fmt"
	"github.com/go-yaml/yaml"
	"github.com/spf13/pflag"
	"log"
	"os"
	"strings"
)

func CommandLine() (Config, command) {
	var configFile *string = pflag.StringP(
		"config-file",
		"c",
		"rettention.yaml",
		"Path to the configuration file",
	)
	pflag.Parse()

	config := Config{configFile: *configFile}
	fin, err := os.Open(config.configFile)
	if err != nil {
		log.Fatalf(err.Error())
	}

	decoder := yaml.NewDecoder(fin)
	if err := decoder.Decode(&config); err != nil {
		log.Fatalf(err.Error())
	}
	fin.Close()

	if len(pflag.Args()) != 1 {
		usage()
	}
	commandString := strings.ToLower(pflag.Args()[0])

	switch commandString {
	case "run":
		return config, Run
	case "auth":
		return config, Auth
	default:
		usage()
	}

	// Unreachable, usage() will bail out before we get to this point
	return config, Auth
}

func usage() {
	fmt.Fprintf(
		os.Stderr,
		"Usage: %s [options] auth|run\n",
		os.Args[0],
	)
	pflag.PrintDefaults()
	os.Exit(0)
}
