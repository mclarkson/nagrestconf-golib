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

type servicetemplate struct {
	name                  string
	use                   string
	contacts              string
	contactgroups         string
	notifopts             string
	checkinterval         string
	normchecki            string
	retryinterval         string
	notifinterval         string
	notifperiod           string
	disable               string
	checkperiod           string
	maxcheckattempts      string
	freshnessthresh       string
	activechecks          string
	customvars            string
	isvolatile            string
	initialstate          string
	passivechecks         string
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
	firstnotifdelay       string
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

type Servicetemplates struct {
	servicetemplates []servicetemplate
}

func ServicetemplatesFields() (arr []string) {

	h := &servicetemplate{}

	n := reflect.TypeOf(h).Elem().NumField()
	for i := 0; i < n; i++ {
		f := reflect.TypeOf(h).Elem().Field(i)
		arr = append(arr, f.Name)
	}

	sort.Strings(arr)

	return arr
}

func ServicetemplatesFieldsJson() (s string) {

	f := ServicetemplatesFields()

	s = "["
	c := ""
	for _, j := range f {
		s += c + `"` + j + `"`
		c = ","
	}
	s += "]"

	return s
}

func (h *Servicetemplates) FilterServicetemplates(filter string) {

	f := NewFilter(filter)

	newh := []servicetemplate{}

	for _, k := range h.servicetemplates {
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
	h.servicetemplates = newh
}

func (h Servicetemplates) ShowServicetemplatesJson(newline, brief bool, filter string) {

	if filter != "" {
		h.FilterServicetemplates(filter)
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
	for _, r := range h.servicetemplates {
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

func NewNrcServicetemplates() Servicetemplates {
	return Servicetemplates{}
}

/*
 * Send HTTP GET request
 */
func (h *Servicetemplates) GetServicetemplates(url, endpoint, folder, data string) (e error) {

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
			servicetemplate := servicetemplate{}
			for _, j2 := range j.([]interface{}) {
				content := j2.(map[string]interface{})
				for k, v := range content {
					switch k {
					case "name":
						servicetemplate.name = v.(string)
					case "use":
						servicetemplate.use = v.(string)
					case "contacts":
						servicetemplate.contacts = v.(string)
					case "contactgroups":
						servicetemplate.contactgroups = v.(string)
					case "notifopts":
						servicetemplate.notifopts = v.(string)
					case "checkinterval":
						servicetemplate.checkinterval = v.(string)
					case "normchecki":
						servicetemplate.normchecki = v.(string)
					case "retryinterval":
						servicetemplate.retryinterval = v.(string)
					case "notifinterval":
						servicetemplate.notifinterval = v.(string)
					case "notifperiod":
						servicetemplate.notifperiod = v.(string)
					case "disable":
						servicetemplate.disable = v.(string)
					case "checkperiod":
						servicetemplate.checkperiod = v.(string)
					case "maxcheckattempts":
						servicetemplate.maxcheckattempts = v.(string)
					case "freshnessthresh":
						servicetemplate.freshnessthresh = v.(string)
					case "activechecks":
						servicetemplate.activechecks = v.(string)
					case "customvars":
						servicetemplate.customvars = v.(string)
					case "isvolatile":
						servicetemplate.isvolatile = v.(string)
					case "initialstate":
						servicetemplate.initialstate = v.(string)
					case "passivechecks":
						servicetemplate.passivechecks = v.(string)
					case "obsessoverservice":
						servicetemplate.obsessoverservice = v.(string)
					case "manfreshnessthresh":
						servicetemplate.manfreshnessthresh = v.(string)
					case "checkfreshness":
						servicetemplate.checkfreshness = v.(string)
					case "eventhandler":
						servicetemplate.eventhandler = v.(string)
					case "eventhandlerenabled":
						servicetemplate.eventhandlerenabled = v.(string)
					case "lowflapthresh":
						servicetemplate.lowflapthresh = v.(string)
					case "highflapthresh":
						servicetemplate.highflapthresh = v.(string)
					case "flapdetectionenabled":
						servicetemplate.flapdetectionenabled = v.(string)
					case "flapdetectionoptions":
						servicetemplate.flapdetectionoptions = v.(string)
					case "processperfdata":
						servicetemplate.processperfdata = v.(string)
					case "retainstatusinfo":
						servicetemplate.retainstatusinfo = v.(string)
					case "retainnonstatusinfo":
						servicetemplate.retainnonstatusinfo = v.(string)
					case "firstnotifdelay":
						servicetemplate.firstnotifdelay = v.(string)
					case "notifications_enabled":
						servicetemplate.notifications_enabled = v.(string)
					case "stalkingoptions":
						servicetemplate.stalkingoptions = v.(string)
					case "notes":
						servicetemplate.notes = v.(string)
					case "notes_url":
						servicetemplate.notes_url = v.(string)
					case "action_url":
						servicetemplate.action_url, _ = UrlDecode(v.(string))
					case "icon_image":
						servicetemplate.icon_image = v.(string)
					case "icon_image_alt":
						servicetemplate.icon_image_alt = v.(string)
					case "vrml_image":
						servicetemplate.vrml_image = v.(string)
					case "statusmap_image":
						servicetemplate.statusmap_image = v.(string)
					case "coords2d":
						servicetemplate.coords2d = v.(string)
					case "coords3d":
						servicetemplate.coords3d = v.(string)
					}
				}
			}
			h.servicetemplates = append(h.servicetemplates, servicetemplate)
		}

		return nil
	}
}

/*
 * Send HTTP POST request
 */
func (h Servicetemplates) PostServicetemplates(url, endpoint, folder, data string) (e error) {

	for strings.HasSuffix(url, "/") {
		url = strings.TrimSuffix(url, "/")
	}
	for strings.HasSuffix(endpoint, "/") {
		url = strings.TrimSuffix(endpoint, "/")
	}

	fullUrl := url + "/" + endpoint

	// Format data
	dataStr := "json={\"folder\":\"" + folder + "\""
	dataInr, err := FormatData(data, "servicetemplates")
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
