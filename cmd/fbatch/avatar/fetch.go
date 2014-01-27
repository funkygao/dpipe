/*
Fetch users avatars in batch.
*/
package main

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"net/http"
)

func shouldDownloadUser(snsid string) bool {
	rwMutex.RLock()
	_, present = allUsers[snsid]
	rwMutex.RUnlock()
	if present {
		return false
	}

	rwMutex.Lock()
	allUsers[snsid] = true
	rwMutex.Unlock()
	return true
}

func avatarUrl(snsid string) string {
	return fmt.Sprintf("http://graph.facebook.com/%s/picture", snsid)
}

func fetchAvatar(snsid string) {
	if !shouldDownloadUser(snsid) {
		return
	}

	response, _ := http.Get(avatarUrl(snsid))
	if response != nil && response.Body != nil {
		defer response.Body.Close()

		body, _ := ioutil.ReadAll(response.Body)
		if len(body) < 1024 {
			// not a valid avatar image
			return
		}

		ioutil.WriteFile(targetDir+snsid+".jpg", body, 0644)
	}
}

func generateAvatarHtml() {
	buf := new(bytes.Buffer)
	for snsid, _ := range allUsers {
		buf.WriteString("<img src='")
		buf.WriteString(targetDir + snsid)
		buf.WriteString(".jpg' />\n")
	}

	ioutil.WriteFile("index.html", buf.Bytes(), 0644)
}
