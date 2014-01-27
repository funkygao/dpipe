/*
Fetch users avatars in batch.
*/
package main

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"net/http"
	"sync"
)

var (
	allUsers = map[string]bool{}
	rwMutex  = new(sync.RWMutex)
	present  = false
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
	defer response.Body.Close()

	body, _ := ioutil.ReadAll(response.Body)
	if err := ioutil.WriteFile(snsid+".jpg", body, 0644); err != nil {
		panic(err)
	}
}

func generateAvatarHtml() {
	buf := new(bytes.Buffer)
	for snsid, _ := range allUsers {
		buf.WriteString("<img src='")
		buf.WriteString(snsid)
		buf.WriteString(".jpg' />\n")
	}

	ioutil.WriteFile("index.html", buf.Bytes(), 0644)
}
