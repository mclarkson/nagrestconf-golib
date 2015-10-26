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

type contact struct {
	name                string
	use                 string
	alias               string
	emailaddr           string
	svcnotifperiod      string
	svcnotifopts        string
	svcnotifcmds        string
	hstnotifperiod      string
	hstnotifopts        string
	hstnotifcmds        string
	cansubmitcmds       string
	disable             string
	svcnotifenabled     string
	hstnotifenabled     string
	pager               string
	address1            string
	address2            string
	address3            string
	address4            string
	address5            string
	address6            string
	retainstatusinfo    string
	retainnonstatusinfo string
	contactgroups       string
}

type Contacts struct {
	contacts []contact
}

func ContactsFields() (arr []string) {

	h := &contact{}

	n := reflect.TypeOf(h).Elem().NumField()
	for i := 0; i < n; i++ {
		f := reflect.TypeOf(h).Elem().Field(i)
		arr = append(arr, f.Name)
	}

	sort.Strings(arr)

	return arr
}

func ContactsFieldsJson() (s string) {

	f := ContactsFields()

	s = "["
	c := ""
	for _, j := range f {
		s += c + `"` + j + `"`
		c = ","
	}
	s += "]"

	return s
}

func (h *Contacts) FilterContacts(filter string) {

	f := NewFilter(filter)

	newh := []contact{}

	for _, k := range h.contacts {
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
	h.contacts = newh
}

func (h Contacts) ShowContactsJson(newline, brief bool, filter string) {

	if filter != "" {
		h.FilterContacts(filter)
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
	for _, r := range h.contacts {
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

func NewNrcContacts() Contacts {
	return Contacts{}
}

/*
 * Send HTTP GET request
 */
func (h *Contacts) GetContacts(url, endpoint, folder, data string) (e error) {

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
	dataStr := FormatData(data, "contacts")
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

		var generic interface{}
		if err := json.Unmarshal(body, &generic); err != nil {
			txt := fmt.Sprintf("Status (%d) Error decoding JSON (%s).",
				resp.StatusCode, err.Error())
			return HttpError{txt}
		}
		genericReply := generic.([]interface{})

		//fmt.Printf("%s\n", body)

		for _, j := range genericReply {
			contact := contact{}
			for _, j2 := range j.([]interface{}) {
				content := j2.(map[string]interface{})
				for k, v := range content {
					switch k {
					case "name":
						contact.name = v.(string)
					case "use":
						contact.use = v.(string)
					case "alias":
						contact.alias = v.(string)
					case "emailaddr":
						contact.emailaddr = v.(string)
					case "svcnotifperiod":
						contact.svcnotifperiod = v.(string)
					case "svcnotifopts":
						contact.svcnotifopts = v.(string)
					case "svcnotifcmds":
						contact.svcnotifcmds = v.(string)
					case "hstnotifperiod":
						contact.hstnotifperiod = v.(string)
					case "hstnotifopts":
						contact.hstnotifopts = v.(string)
					case "hstnotifcmds":
						contact.hstnotifcmds = v.(string)
					case "cansubmitcmds":
						contact.cansubmitcmds = v.(string)
					case "disable":
						contact.disable = v.(string)
					case "svcnotifenabled":
						contact.svcnotifenabled = v.(string)
					case "hstnotifenabled":
						contact.hstnotifenabled = v.(string)
					case "pager":
						contact.pager = v.(string)
					case "address1":
						contact.address1 = v.(string)
					case "address2":
						contact.address2 = v.(string)
					case "address3":
						contact.address3 = v.(string)
					case "address4":
						contact.address4 = v.(string)
					case "address5":
						contact.address5 = v.(string)
					case "address6":
						contact.address6 = v.(string)
					case "retainstatusinfo":
						contact.retainstatusinfo = v.(string)
					case "retainnonstatusinfo":
						contact.retainnonstatusinfo = v.(string)
					case "contactgroups":
						contact.contactgroups = v.(string)
					}
				}
			}
			h.contacts = append(h.contacts, contact)
		}

		return nil
	}
}

/*
 * Send HTTP POST request
 */
func (h Contacts) PostContacts(url, endpoint, folder, data string) (e error) {

	for strings.HasSuffix(url, "/") {
		url = strings.TrimSuffix(url, "/")
	}
	for strings.HasSuffix(endpoint, "/") {
		url = strings.TrimSuffix(endpoint, "/")
	}

	fullUrl := url + "/" + endpoint

	// Format data
	dataStr := "json={\"folder\":\"" + folder + "\""
	dataInr := FormatData(data, "contacts")
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

	resp, err := client.Post(fullUrl, "application/x-www-form-urlencoded", buf)
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
