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

func fetchAvatar(area, snsid string) {
	if !shouldDownloadUser(snsid) {
		return
	}

	response, err := http.Get(avatarUrl(snsid))
	if err != nil {
		return
	}

	defer response.Body.Close()

	body, _ := ioutil.ReadAll(response.Body)
	if len(body) < 1024 {
		// not a valid avatar image
		return
	}

	targetFile := targetDir + area + "_" + snsid + jpegExt
	ioutil.WriteFile(targetFile, body, 0644)
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
