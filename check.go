package nrc

import (
	"crypto/tls"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"strings"
)

type check struct {
	Output   []string
	username string
	password string
}

func NewNrcCheck(username, password string) *check {
	r := &check{}
	r.username = username
	r.password = password
	return r
}

func (c check) RequiredOptions() []string {
	return []string{}
}

func (c *check) Options() (arr []string) {
	return []string{"verbose"}
}

func (c check) OptionsJson() (s string) {
	return s
}

func (c check) ShowJson(newline, brief bool, filter string) {

	var nl = ""   // newline
	var ind4 = "" // big indent

	if newline == false {
		nl = "\n"
		ind4 = "    "
	}

	comma := ""
	fmt.Printf("[")
	for _, j := range c.Output {
		e := new(jsonEncode)
		e.string(j)
		fmt.Printf("%s%s%s%s",
			comma, nl, ind4, e.String())
		comma = ","
	}
	fmt.Printf("%s]\n", nl)
}

func (c check) Show(brief bool, filter string) {

	for _, j := range c.Output {
		fmt.Printf("%s\n", j)
	}
}

/*
 * Send HTTP GET request
 */
func (c *check) Get(url, endpoint, folder string, data []string) (e error) {

	// accept bad certs
	tr := &http.Transport{
		TLSClientConfig:   &tls.Config{InsecureSkipVerify: true},
		DisableKeepAlives: true,
	}
	client := &http.Client{Transport: tr}
	// Not available in Go<1.3
	//client.Timeout = 8 * 1e9

	for strings.HasSuffix(url, "/") {
		url = strings.TrimSuffix(url, "/")
	}
	for strings.HasSuffix(endpoint, "/") {
		url = strings.TrimSuffix(endpoint, "/")
	}

	// Construct url, http://1.2.3.4/rest/show/hosts?json={"folder":"local",...}
	fullUrl := url + "/" + endpoint + "?json={\"folder\":\"" + folder + "\""
	dataStr, err := FormatData(data, "services")
	if err != nil {
		txt := fmt.Sprintf("Could not format data. Check the '-d' option.")
		return HttpError{txt}
	}
	if dataStr != "" {
		fullUrl += "," + dataStr
	}
	fullUrl += "}"

	//fmt.Printf("URL=%s\n", fullUrl)

	//fmt.Printf("%s\n", url+"/"+endpoint)
	//resp, err := client.Get(fullUrl)
	req, err := http.NewRequest("GET", fullUrl, nil)
	if len(c.username) > 0 {
		req.SetBasicAuth(c.username, c.password)
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

	} else {

		if err := json.Unmarshal(body, &c.Output); err != nil {
			txt := fmt.Sprintf("Status (%d) Error decoding JSON (%s).",
				resp.StatusCode, err.Error())
			return HttpError{txt}
		}

		return nil
	}
}

func (c check) Post(url, endpoint, folder string, data []string) (e error) {
	return nil
}
