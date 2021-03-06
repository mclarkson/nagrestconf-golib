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

type serviceset struct {
	name                  string
	template              string
	command               string
	svcdesc               string
	svcgroup              string
	contacts              string
	contactgroups         string
	freshnessthresh       string
	activechecks          string
	customvars            string
	disable               string
	displayname           string
	isvolatile            string
	initialstate          string
	maxcheckattempts      string
	checkinterval         string
	retryinterval         string
	passivechecks         string
	checkperiod           string
	obsessoverservice     string
	manfreshnessthresh    string
	checkfreshness        string
	eventhandler          string
	eventhandlerenabled   string
	lowflapthresh         string
	highflapthresh        string
	flapdetectionenabled  string
	flapdetectionoptions  string
	processperfdata       string
	retainstatusinfo      string
	retainnonstatusinfo   string
	notifinterval         string
	firstnotifdelay       string
	notifperiod           string
	notifopts             string
	notifications_enabled string
	stalkingoptions       string
	notes                 string
	notes_url             string
	action_url            string
	icon_image            string
	icon_image_alt        string
	vrml_image            string
	statusmap_image       string
	coords2d              string
	coords3d              string
}

type Servicesets struct {
	servicesets []serviceset
	username    string
	password    string
}

func (h Servicesets) RequiredOptions() []string {
	return []string{"name", "template", "command", "svcdesc"}
}

func (h *Servicesets) Options() (arr []string) {

	n := reflect.TypeOf(h.servicesets).Elem().NumField()
	for i := 0; i < n; i++ {
		f := reflect.TypeOf(h.servicesets).Elem().Field(i)
		arr = append(arr, f.Name)
	}

	sort.Strings(arr)

	return arr
}

func (h Servicesets) OptionsJson() (s string) {

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

func (h *Servicesets) filterServicesets(filter string) {

	f := NewFilter(filter)

	newh := []serviceset{}

	for _, k := range h.servicesets {
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
	h.servicesets = newh
}

func (h Servicesets) Show(brief bool, filter string) {

	if filter != "" {
		h.filterServicesets(filter)
	}

	var ind4 = "    " // big indent

	fmt.Printf("\n")
	for _, r := range h.servicesets {
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

func (h Servicesets) ShowJson(newline, brief bool, filter string) {

	if filter != "" {
		h.filterServicesets(filter)
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
	for _, r := range h.servicesets {
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

func NewNrcServicesets(username, password string) *Servicesets {
	h := &Servicesets{}
	h.username = username
	h.password = password
	return h
}

/*
 * Send HTTP GET request
 */
func (h *Servicesets) Get(url, endpoint, folder string, data []string) (e error) {

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
			serviceset := serviceset{}
			for _, j2 := range j.([]interface{}) {
				content := j2.(map[string]interface{})
				for k, v := range content {
					switch k {
					case "name":
						serviceset.name, _ = UrlDecode(v.(string))
					case "template":
						serviceset.template = v.(string)
					case "command":
						serviceset.command, _ = UrlDecode(v.(string))
					case "svcdesc":
						serviceset.svcdesc, _ = UrlDecode(v.(string))
					case "svcgroup":
						serviceset.svcgroup = v.(string)
					case "contacts":
						serviceset.contacts = v.(string)
					case "contactgroups":
						serviceset.contactgroups = v.(string)
					case "freshnessthresh":
						serviceset.freshnessthresh = v.(string)
					case "activechecks":
						serviceset.activechecks = v.(string)
					case "customvars":
						serviceset.customvars = v.(string)
					case "disable":
						serviceset.disable = v.(string)
					case "displayname":
						serviceset.displayname = v.(string)
					case "isvolatile":
						serviceset.isvolatile = v.(string)
					case "initialstate":
						serviceset.initialstate = v.(string)
					case "maxcheckattempts":
						serviceset.maxcheckattempts = v.(string)
					case "checkinterval":
						serviceset.checkinterval = v.(string)
					case "retryinterval":
						serviceset.retryinterval = v.(string)
					case "passivechecks":
						serviceset.passivechecks = v.(string)
					case "checkperiod":
						serviceset.checkperiod = v.(string)
					case "obsessoverservice":
						serviceset.obsessoverservice = v.(string)
					case "manfreshnessthresh":
						serviceset.manfreshnessthresh = v.(string)
					case "checkfreshness":
						serviceset.checkfreshness = v.(string)
					case "eventhandler":
						serviceset.eventhandler = v.(string)
					case "eventhandlerenabled":
						serviceset.eventhandlerenabled = v.(string)
					case "lowflapthresh":
						serviceset.lowflapthresh = v.(string)
					case "highflapthresh":
						serviceset.highflapthresh = v.(string)
					case "flapdetectionenabled":
						serviceset.flapdetectionenabled = v.(string)
					case "flapdetectionoptions":
						serviceset.flapdetectionoptions = v.(string)
					case "processperfdata":
						serviceset.processperfdata = v.(string)
					case "retainstatusinfo":
						serviceset.retainstatusinfo = v.(string)
					case "retainnonstatusinfo":
						serviceset.retainnonstatusinfo = v.(string)
					case "notifinterval":
						serviceset.notifinterval = v.(string)
					case "firstnotifdelay":
						serviceset.firstnotifdelay = v.(string)
					case "notifperiod":
						serviceset.notifperiod = v.(string)
					case "notifopts":
						serviceset.notifopts = v.(string)
					case "notifications_enabled":
						serviceset.notifications_enabled = v.(string)
					case "stalkingoptions":
						serviceset.stalkingoptions = v.(string)
					case "notes":
						serviceset.notes = v.(string)
					case "notes_url":
						serviceset.notes_url = v.(string)
					case "action_url":
						serviceset.action_url = v.(string)
					case "icon_image":
						serviceset.icon_image = v.(string)
					case "icon_image_alt":
						serviceset.icon_image_alt = v.(string)
					case "vrml_image":
						serviceset.vrml_image = v.(string)
					case "statusmap_image":
						serviceset.statusmap_image = v.(string)
					case "coords2d":
						serviceset.coords2d = v.(string)
					case "coords3d":
						serviceset.coords3d = v.(string)
					}
				}
			}
			h.servicesets = append(h.servicesets, serviceset)
		}

		return nil
	}
}

/*
 * Send HTTP POST request
 */
func (h Servicesets) Post(url, endpoint, folder string, data []string) (e error) {

	for strings.HasSuffix(url, "/") {
		url = strings.TrimSuffix(url, "/")
	}
	for strings.HasSuffix(endpoint, "/") {
		url = strings.TrimSuffix(endpoint, "/")
	}

	fullUrl := url + "/" + endpoint

	// Format data
	dataStr := "json={\"folder\":\"" + folder + "\""
	dataInr, err := FormatData(data, "servicesets")
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
