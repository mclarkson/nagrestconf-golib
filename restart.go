package nrc

import (
	"bytes"
	"crypto/tls"
	"fmt"
	"io/ioutil"
	"net/http"
	"strings"
)

type restart struct {
	username string
	password string
}

func (r restart) RequiredOptions() []string {
	return []string{}
}

func (r *restart) Options() (arr []string) {
	return []string{"verbose"}
}

func (r restart) OptionsJson() (s string) {
	return s
}

func (r restart) Show(brief bool, filter string) {
}

func (r restart) ShowJson(newline, brief bool, filter string) {
}

func NewNrcRestart(username, password string) *restart {
	r := &restart{}
	r.username = username
	r.password = password
	return r
}

func (r restart) Get(url, endpoint, folder string, data []string) (e error) {
	return nil
}

/*
 * Send HTTP POST request
 */
func (r restart) Post(url, endpoint, folder string, data []string) (e error) {

	for strings.HasSuffix(url, "/") {
		url = strings.TrimSuffix(url, "/")
	}
	for strings.HasSuffix(endpoint, "/") {
		url = strings.TrimSuffix(endpoint, "/")
	}

	fullUrl := url + "/" + endpoint

	// Format data
	dataStr := "json={\"folder\":\"" + folder + "\""
	dataInr, err := FormatData(data, "")
	if err != nil {
		txt := fmt.Sprintf("Could not format data. Check the '-d' option.")
		return HttpError{txt}
	}
	if dataInr != "" {
		dataStr += "," + dataInr
	}
	dataStr += "}"

	//fmt.Printf("URL=%s\n", fullUrl)
	//fmt.Printf("Data=%s\n", dataStr)

	buf := bytes.NewBuffer([]byte(dataStr))

	//fmt.Printf("json=%s\n", buf)

	// accept bad certs
	tr := &http.Transport{
		TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
	}
	client := &http.Client{Transport: tr}

	//resp, err := client.Post(fullUrl, "application/x-www-form-urlencoded", buf)
	req, err := http.NewRequest("POST", fullUrl, buf)
	req.Header.Add("Content-Type", "application/x-www-form-urlencoded")
	if len(r.username) > 0 {
		req.SetBasicAuth(r.username, r.password)
	}
	resp, err := client.Do(req)
	if err != nil {
		txt := fmt.Sprintf("Could not send REST request ('%s').", err.Error())
		return HttpError{txt}
	}

	defer resp.Body.Close()

	var body []byte
	if b, err := ioutil.ReadAll(resp.Body); err != nil {
		txt := fmt.Sprintf("Error reading Body ('%s').", err.Error())
		return HttpError{txt}
	} else {
		body = b
	}

	if resp.StatusCode != 200 {

		response, _ := UrlDecode(string(body))
		txt := fmt.Sprintf("Status (%d): %s", resp.StatusCode, response)
		return HttpError{txt}

	}

	// If we get here then it was a success
	return nil
}
