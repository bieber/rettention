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

package run

import (
	"github.com/bieber/rettention/auth"
	"github.com/bieber/rettention/config"
	"github.com/bieber/rettention/util"
	"log"
	"time"
)

func Run(c config.Config) {
	credentials := auth.Reauthenticate(c)

	users := map[string]config.User{}
	for username, user := range c.Users {
		processComments := user.CommentDuration != "forever"
		processPosts := user.PostDuration != "forever"

		if _, ok := credentials[username]; !ok {
			log.Printf("Skipping %s due to missing credential", username)
		} else if !processComments && !processPosts {
			log.Printf("Skipping %s, no retention configured", username)
		} else {
			users[username] = user
		}
	}

	for username, user := range users {
		err := RunUser(username, user, credentials[username])
		if err != nil {
			log.Printf("Error processing %s: %s", username, err.Error())
		}
	}
}

func RunUser(
	username string,
	user config.User,
	credential config.Credential,
) error {
	log.Printf("Processing %s", username)
	toDelete, err := fetchToDelete(username, user, credential)
	if err != nil {
		return err
	}

	return deleteAll(credential, toDelete)
}

func fetchToDelete(
	username string,
	user config.User,
	credential config.Credential,
) (toDelete []string, err error) {
	commentDuration, err := time.ParseDuration(user.CommentDuration)
	if err != nil {
		return
	}
	postDuration, err := time.ParseDuration(user.PostDuration)
	if err != nil {
		return
	}

	commentMin := time.Now().Add(-commentDuration).Unix()
	postMin := time.Now().Add(-postDuration).Unix()

	after := "null"
	page := 1
	for after != "" {
		dest := struct {
			Data struct {
				After    string
				Children []struct {
					Data struct {
						Name    string
						Created float64
					}
					Kind string
				}
			}
		}{}

		uri := "user/" + username + "/overview?after=" + after
		err = util.DoAPIRequest(credential, "GET", uri, nil, &dest)
		if err != nil {
			return
		}

		for _, child := range dest.Data.Children {
			switch child.Kind {
			case "t1":
				if int64(child.Data.Created) < commentMin {
					toDelete = append(toDelete, child.Data.Name)
				}
			case "t3":
				if int64(child.Data.Created) < postMin {
					toDelete = append(toDelete, child.Data.Name)
				}
			}
		}

		log.Printf(
			"Processed page %d, %d entries queued to delete",
			page,
			len(toDelete),
		)
		page++

		after = dest.Data.After
	}

	return
}

func deleteAll(credential config.Credential, toDelete []string) error {
	log.Printf("Deleting %d entries", len(toDelete))

	for i, id := range toDelete {
		err := util.DoAPIRequest(
			credential,
			"POST",
			"api/del",
			map[string]string{"id": id},
			nil,
		)
		if err != nil {
			return err
		}

		if (i+1)%10 == 0 || i == len(toDelete)-1 {
			log.Printf("Deleted %d/%d entries", i+1, len(toDelete))
		}
	}

	return nil
}
