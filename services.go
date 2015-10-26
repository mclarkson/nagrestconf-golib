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

type service struct {
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

type Services struct {
	services []service
}

func ServicesRequiredFields() []string {
	return []string{"name", "template", "command", "svcdesc"}
}

func ServicesFields() (arr []string) {

	h := &service{}

	n := reflect.TypeOf(h).Elem().NumField()
	for i := 0; i < n; i++ {
		f := reflect.TypeOf(h).Elem().Field(i)
		arr = append(arr, f.Name)
	}

	sort.Strings(arr)

	return arr
}

func ServicesFieldsJson() (s string) {

	f := ServicesFields()

	s = "["
	c := ""
	for _, j := range f {
		s += c + `"` + j + `"`
		c = ","
	}
	s += "]"

	return s
}

func (h *Services) FilterServices(filter string) {

	f := NewFilter(filter)

	newh := []service{}

	for _, k := range h.services {
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
	h.services = newh
}

func (h Services) ShowServicesJson(newline, brief bool, filter string) {

	if filter != "" {
		h.FilterServices(filter)
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
	for _, r := range h.services {
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

func NewNrcServices() Services {
	return Services{}
}

/*
 * Send HTTP GET request
 */
func (h *Services) GetServices(url, endpoint, folder, data string) (e error) {

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

		var generic interface{}
		if err := json.Unmarshal(body, &generic); err != nil {
			txt := fmt.Sprintf("Status (%d) Error decoding JSON (%s).",
				resp.StatusCode, err.Error())
			return HttpError{txt}
		}
		genericReply := generic.([]interface{})

		//fmt.Printf("%s\n", body)

		for _, j := range genericReply {
			service := service{}
			for _, j2 := range j.([]interface{}) {
				content := j2.(map[string]interface{})
				for k, v := range content {
					switch k {
					case "name":
						service.name, _ = UrlDecode(v.(string))
					case "template":
						service.template = v.(string)
					case "command":
						service.command, _ = UrlDecode(v.(string))
					case "svcdesc":
						service.svcdesc, _ = UrlDecode(v.(string))
					case "svcgroup":
						service.svcgroup = v.(string)
					case "contacts":
						service.contacts = v.(string)
					case "contactgroups":
						service.contactgroups = v.(string)
					case "freshnessthresh":
						service.freshnessthresh = v.(string)
					case "activechecks":
						service.activechecks = v.(string)
					case "customvars":
						service.customvars = v.(string)
					case "disable":
						service.disable = v.(string)
					case "displayname":
						service.displayname = v.(string)
					case "isvolatile":
						service.isvolatile = v.(string)
					case "initialstate":
						service.initialstate = v.(string)
					case "maxcheckattempts":
						service.maxcheckattempts = v.(string)
					case "checkinterval":
						service.checkinterval = v.(string)
					case "retryinterval":
						service.retryinterval = v.(string)
					case "passivechecks":
						service.passivechecks = v.(string)
					case "checkperiod":
						service.checkperiod = v.(string)
					case "obsessoverservice":
						service.obsessoverservice = v.(string)
					case "manfreshnessthresh":
						service.manfreshnessthresh = v.(string)
					case "checkfreshness":
						service.checkfreshness = v.(string)
					case "eventhandler":
						service.eventhandler = v.(string)
					case "eventhandlerenabled":
						service.eventhandlerenabled = v.(string)
					case "lowflapthresh":
						service.lowflapthresh = v.(string)
					case "highflapthresh":
						service.highflapthresh = v.(string)
					case "flapdetectionenabled":
						service.flapdetectionenabled = v.(string)
					case "flapdetectionoptions":
						service.flapdetectionoptions = v.(string)
					case "processperfdata":
						service.processperfdata = v.(string)
					case "retainstatusinfo":
						service.retainstatusinfo = v.(string)
					case "retainnonstatusinfo":
						service.retainnonstatusinfo = v.(string)
					case "notifinterval":
						service.notifinterval = v.(string)
					case "firstnotifdelay":
						service.firstnotifdelay = v.(string)
					case "notifperiod":
						service.notifperiod = v.(string)
					case "notifopts":
						service.notifopts = v.(string)
					case "notifications_enabled":
						service.notifications_enabled = v.(string)
					case "stalkingoptions":
						service.stalkingoptions = v.(string)
					case "notes":
						service.notes = v.(string)
					case "notes_url":
						service.notes_url = v.(string)
					case "action_url":
						service.action_url = v.(string)
					case "icon_image":
						service.icon_image = v.(string)
					case "icon_image_alt":
						service.icon_image_alt = v.(string)
					case "vrml_image":
						service.vrml_image = v.(string)
					case "statusmap_image":
						service.statusmap_image = v.(string)
					case "coords2d":
						service.coords2d = v.(string)
					case "coords3d":
						service.coords3d = v.(string)
					}
				}
			}
			h.services = append(h.services, service)
		}

		return nil
	}
}

/*
 * Send HTTP POST request
 */
func (h Services) PostServices(url, endpoint, folder, data string) (e error) {

	for strings.HasSuffix(url, "/") {
		url = strings.TrimSuffix(url, "/")
	}
	for strings.HasSuffix(endpoint, "/") {
		url = strings.TrimSuffix(endpoint, "/")
	}

	fullUrl := url + "/" + endpoint

	// Format data
	dataStr := "json={\"folder\":\"" + folder + "\""
	dataInr, err := FormatData(data, "services")
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
