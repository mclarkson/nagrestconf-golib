package nrc

import (
	"bytes"
	"crypto/tls"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"reflect"
	"regexp"
	"sort"
	"strings"
)

type hostgroup struct {
	name             string
	alias            string
	disable          string
	members          string
	hostgroupmembers string
	notes            string
	notes_url        string
	action_url       string
}

type Hostgroups struct {
	hostgroups []hostgroup
	username   string
	password   string
}

func (h Hostgroups) RequiredOptions() []string {
	return []string{"name", "alias"}
}

func (h *Hostgroups) Options() (arr []string) {

	n := reflect.TypeOf(h.hostgroups).Elem().NumField()
	for i := 0; i < n; i++ {
		f := reflect.TypeOf(h.hostgroups).Elem().Field(i)
		arr = append(arr, f.Name)
	}

	sort.Strings(arr)

	return arr
}

func (h Hostgroups) OptionsJson() (s string) {

	f := h.Options()

	s = "["
	c := ""
	for _, j := range f {
		s += c + `"` + j + `"`
		c = ","
	}
	s += "]"

	return s
}

func (h *Hostgroups) filterHostgroups(filter string) {

	f := NewFilter(filter)

	newh := []hostgroup{}

	for _, k := range h.hostgroups {
		t := &k // Must be a pointer or interface for reflection
		foundCount := 0
		for i := 0; i < len(f.names); i++ {
			_, found := reflect.TypeOf(t).Elem().FieldByName(f.names[i])
			if found == true {
				val := reflect.ValueOf(t).Elem().FieldByName(f.names[i])
				userRegx, _ := UrlDecodeForce(f.regex[i])
				regex := regexp.MustCompile(userRegx)
				if regex.MatchString(val.String()) {
					foundCount += 1
					continue
				}
				/*
					val := reflect.ValueOf(t).Elem().FieldByName(f.names[i])
					if val.String() == f.regex[i] {
						foundCount += 1
						continue
					}
				*/
			}
		}
		if foundCount == len(f.names) {
			newh = append(newh, k)
		}
	}

	// Replace the list we got with this filtered list
	h.hostgroups = newh
}

func (h Hostgroups) Show(brief bool, filter string) {

	if filter != "" {
		h.filterHostgroups(filter)
	}

	var ind4 = "    " // big indent

	fmt.Printf("\n")
	for _, r := range h.hostgroups {
		t := &r // Reflection is only allowed on ptr or interface
		n := reflect.TypeOf(t).Elem().NumField()
		for i := 0; i < n; i++ {
			f := reflect.TypeOf(t).Elem().Field(i)
			g := reflect.ValueOf(t).Elem().Field(i)
			if brief == true || (brief == false && g.String() != "") {
				fmt.Printf("%s%s:%s\n",
					ind4, f.Name, g)
			}
		}
		fmt.Printf("\n")
	}
}

func (h Hostgroups) ShowJson(newline, brief bool, filter string) {

	if filter != "" {
		h.filterHostgroups(filter)
	}

	var nl = ""   // newline
	var ind4 = "" // big indent
	var ind2 = "" // small indent

	if newline == false {
		nl = "\n"
		ind4 = "    "
		ind2 = "  "
	}

	fmt.Printf("[")
	objcomma := ""
	for _, r := range h.hostgroups {
		comma := ""
		nli := ""
		fmt.Printf("%s%s%s{%s", objcomma, nl, ind2, nl)
		t := &r // Reflection is only allowed on ptr or interface
		n := reflect.TypeOf(t).Elem().NumField()
		for i := 0; i < n; i++ {
			f := reflect.TypeOf(t).Elem().Field(i)
			g := reflect.ValueOf(t).Elem().Field(i)
			if brief == true || (brief == false && g.String() != "") {
				fmt.Printf("%s%s%s\"%s\":\"%s\"",
					comma, nli, ind4, f.Name, g)
				comma = ","
				nli = nl
			}
		}
		fmt.Printf("%s%s}", nl, ind2)
		objcomma = ","
	}
	fmt.Printf("%s]\n", nl)
}

func NewNrcHostgroups(username, password string) *Hostgroups {
	h := &Hostgroups{}
	h.username = username
	h.password = password
	return h
}

/*
 * Send HTTP GET request
 */
func (h *Hostgroups) Get(url, endpoint, folder string, data []string) (e error) {

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
	if len(h.username) > 0 {
		req.SetBasicAuth(h.username, h.password)
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

		var generic interface{}
		if err := json.Unmarshal(body, &generic); err != nil {
			txt := fmt.Sprintf("Status (%d) Error decoding JSON (%s).",
				resp.StatusCode, err.Error())
			return HttpError{txt}
		}
		genericReply := generic.([]interface{})

		//fmt.Printf("%s\n", body)

		for _, j := range genericReply {
			hostgroup := hostgroup{}
			for _, j2 := range j.([]interface{}) {
				content := j2.(map[string]interface{})
				for k, v := range content {
					switch k {
					case "name":
						hostgroup.name = v.(string)
					case "alias":
						hostgroup.alias = v.(string)
					case "disable":
						hostgroup.disable = v.(string)
					case "members":
						hostgroup.members = v.(string)
					case "hostgroupmembers":
						hostgroup.hostgroupmembers = v.(string)
					case "notes":
						hostgroup.notes = v.(string)
					case "notes_url":
						hostgroup.notes_url = v.(string)
					case "action_url":
						hostgroup.action_url = v.(string)
					}
				}
			}
			h.hostgroups = append(h.hostgroups, hostgroup)
		}

		return nil
	}
}

/*
 * Send HTTP POST request
 */
func (h Hostgroups) Post(url, endpoint, folder string, data []string) (e error) {

	for strings.HasSuffix(url, "/") {
		url = strings.TrimSuffix(url, "/")
	}
	for strings.HasSuffix(endpoint, "/") {
		url = strings.TrimSuffix(endpoint, "/")
	}

	fullUrl := url + "/" + endpoint

	// Format data
	dataStr := "json={\"folder\":\"" + folder + "\""
	dataInr, err := FormatData(data, "hostgroups")
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
	if len(h.username) > 0 {
		req.SetBasicAuth(h.username, h.password)
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
