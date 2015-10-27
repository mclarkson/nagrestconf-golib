package nrc

import (
	"crypto/tls"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"strings"
)

type output struct {
	Output []string
}

func NewNrcCheck() *output {
	return &output{}
}

func (o output) RequiredFields() []string {
	return []string{}
}

func (o *output) Fields() (arr []string) {
	return arr
}

func (o output) FieldsJson() (s string) {
	return s
}

func (o output) ShowJson(newline, brief bool, filter string) {

	var nl = ""   // newline
	var ind4 = "" // big indent

	if newline == false {
		nl = "\n"
		ind4 = "    "
	}

	comma := ""
	fmt.Printf("[")
	for _, j := range o.Output {
		e := new(jsonEncode)
		e.string(j)
		fmt.Printf("%s%s%s%s",
			comma, nl, ind4, e.String())
		comma = ","
	}
	fmt.Printf("%s]\n", nl)
}

func (o output) Show(brief bool, filter string) {

	for _, j := range o.Output {
		fmt.Printf("%s\n", j)
	}
}

/*
 * Send HTTP GET request
 */
func (o *output) Get(url, endpoint, folder, data string) (e error) {

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
	resp, err := client.Get(fullUrl)
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

		if err := json.Unmarshal(body, &o.Output); err != nil {
			txt := fmt.Sprintf("Status (%d) Error decoding JSON (%s).",
				resp.StatusCode, err.Error())
			return HttpError{txt}
		}

		return nil
	}
}

func (o output) Post(url, endpoint, folder, data string) (e error) {
	return nil
}
