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
	"github.com/spf13/pflag"
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

	c := ReadConfig(*configFile)

	if len(pflag.Args()) != 1 {
		usage()
	}
	commandString := strings.ToLower(pflag.Args()[0])

	switch commandString {
	case "run":
		return c, Run
	case "auth":
		return c, Auth
	default:
		usage()
	}

	// Unreachable, usage() will bail out before we get to this point
	return c, Auth
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
