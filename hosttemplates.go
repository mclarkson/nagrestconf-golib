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

type hosttemplate struct {
	name                  string
	use                   string
	contacts              string
	contactgroups         string
	normchecki            string
	checkinterval         string
	retryinterval         string
	notifperiod           string
	notifopts             string
	disable               string
	checkperiod           string
	maxcheckattempts      string
	checkcommand          string
	notifinterval         string
	passivechecks         string
	obsessoverhost        string
	checkfreshness        string
	freshnessthresh       string
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
	icon_image            string
	icon_image_alt        string
	vrml_image            string
	statusmap_image       string
	coords2d              string
	coords3d              string
	action_url            string
}

type Hosttemplates struct {
	hosttemplates []hosttemplate
}

func HosttemplatesRequiredFields() []string {
	return []string{"name", "checkinterval", "retryinterval", "notifperiod", "checkperiod", "maxcheckattempts", "notifinterval"}
}

func HosttemplatesFields() (arr []string) {

	h := &hosttemplate{}

	n := reflect.TypeOf(h).Elem().NumField()
	for i := 0; i < n; i++ {
		f := reflect.TypeOf(h).Elem().Field(i)
		arr = append(arr, f.Name)
	}

	sort.Strings(arr)

	return arr
}

func HosttemplatesFieldsJson() (s string) {

	f := HosttemplatesFields()

	s = "["
	c := ""
	for _, j := range f {
		s += c + `"` + j + `"`
		c = ","
	}
	s += "]"

	return s
}

func (h *Hosttemplates) FilterHosttemplates(filter string) {

	f := NewFilter(filter)

	newh := []hosttemplate{}

	for _, k := range h.hosttemplates {
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
	h.hosttemplates = newh
}

func (h Hosttemplates) ShowHosttemplatesJson(newline, brief bool, filter string) {

	if filter != "" {
		h.FilterHosttemplates(filter)
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
	for _, r := range h.hosttemplates {
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

func NewNrcHosttemplates() Hosttemplates {
	return Hosttemplates{}
}

/*
 * Send HTTP GET request
 */
func (h *Hosttemplates) GetHosttemplates(url, endpoint, folder, data string) (e error) {

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
			hosttemplate := hosttemplate{}
			for _, j2 := range j.([]interface{}) {
				content := j2.(map[string]interface{})
				for k, v := range content {
					switch k {
					case "name":
						hosttemplate.name = v.(string)
					case "use":
						hosttemplate.use = v.(string)
					case "contacts":
						hosttemplate.contacts = v.(string)
					case "contactgroups":
						hosttemplate.contactgroups = v.(string)
					case "normchecki":
						hosttemplate.normchecki = v.(string)
					case "checkinterval":
						hosttemplate.checkinterval = v.(string)
					case "retryinterval":
						hosttemplate.retryinterval = v.(string)
					case "notifperiod":
						hosttemplate.notifperiod = v.(string)
					case "notifopts":
						hosttemplate.notifopts = v.(string)
					case "disable":
						hosttemplate.disable = v.(string)
					case "checkperiod":
						hosttemplate.checkperiod = v.(string)
					case "maxcheckattempts":
						hosttemplate.maxcheckattempts = v.(string)
					case "checkcommand":
						hosttemplate.checkcommand, _ = UrlDecode(v.(string))
					case "notifinterval":
						hosttemplate.notifinterval = v.(string)
					case "passivechecks":
						hosttemplate.passivechecks = v.(string)
					case "obsessoverhost":
						hosttemplate.obsessoverhost = v.(string)
					case "checkfreshness":
						hosttemplate.checkfreshness = v.(string)
					case "freshnessthresh":
						hosttemplate.freshnessthresh = v.(string)
					case "eventhandler":
						hosttemplate.eventhandler = v.(string)
					case "eventhandlerenabled":
						hosttemplate.eventhandlerenabled = v.(string)
					case "lowflapthresh":
						hosttemplate.lowflapthresh = v.(string)
					case "highflapthresh":
						hosttemplate.highflapthresh = v.(string)
					case "flapdetectionenabled":
						hosttemplate.flapdetectionenabled = v.(string)
					case "flapdetectionoptions":
						hosttemplate.flapdetectionoptions = v.(string)
					case "processperfdata":
						hosttemplate.processperfdata = v.(string)
					case "retainstatusinfo":
						hosttemplate.retainstatusinfo = v.(string)
					case "retainnonstatusinfo":
						hosttemplate.retainnonstatusinfo = v.(string)
					case "firstnotifdelay":
						hosttemplate.firstnotifdelay = v.(string)
					case "notifications_enabled":
						hosttemplate.notifications_enabled = v.(string)
					case "stalkingoptions":
						hosttemplate.stalkingoptions = v.(string)
					case "notes":
						hosttemplate.notes = v.(string)
					case "notes_url":
						hosttemplate.notes_url = v.(string)
					case "icon_image":
						hosttemplate.icon_image = v.(string)
					case "icon_image_alt":
						hosttemplate.icon_image_alt = v.(string)
					case "vrml_image":
						hosttemplate.vrml_image = v.(string)
					case "statusmap_image":
						hosttemplate.statusmap_image = v.(string)
					case "coords2d":
						hosttemplate.coords2d = v.(string)
					case "coords3d":
						hosttemplate.coords3d = v.(string)
					case "action_url":
						hosttemplate.action_url, _ = UrlDecode(v.(string))
					}
				}
			}
			h.hosttemplates = append(h.hosttemplates, hosttemplate)
		}

		return nil
	}
}

/*
 * Send HTTP POST request
 */
func (h Hosttemplates) PostHosttemplates(url, endpoint, folder, data string) (e error) {

	for strings.HasSuffix(url, "/") {
		url = strings.TrimSuffix(url, "/")
	}
	for strings.HasSuffix(endpoint, "/") {
		url = strings.TrimSuffix(endpoint, "/")
	}

	fullUrl := url + "/" + endpoint

	// Format data
	dataStr := "json={\"folder\":\"" + folder + "\""
	dataInr, err := FormatData(data, "hosttemplates")
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
