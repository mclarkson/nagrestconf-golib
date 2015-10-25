package nrc

import (
	"crypto/tls"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"reflect"
	"regexp"
	"strings"
)

type host struct {
	name                  string
	alias                 string
	ipaddress             string
	template              string
	hostgroup             string
	contact               string
	contactgroups         string
	activechecks          string
	servicesets           string
	disable               string
	displayname           string
	parents               string
	command               string
	initialstate          string
	maxcheckattempts      string
	checkinterval         string
	retryinterval         string
	passivechecks         string
	checkperiod           string
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
	notifinterval         string
	firstnotifdelay       string
	notifperiod           string
	notifopts             string
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
	customvars            string
}

type Hosts struct {
	hosts []host
}

func (h *Hosts) FilterHosts(filter string) {

	f := NewFilter(filter)

	newh := []host{}

	for _, k := range h.hosts {
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
	h.hosts = newh
}

func (h Hosts) ShowHostsJson(newline, brief bool, filter string) {

	if filter != "" {
		h.FilterHosts(filter)
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
	for _, r := range h.hosts {
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

func NewNrcHosts() Hosts {
	return Hosts{}
}

/*
 * Send HTTP GET request
 */
func (h *Hosts) GetHosts(url, endpoint, folder, data string) (e error) {

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
	dataStr := FormatData(data, "hosts")
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
			host := host{}
			for _, j2 := range j.([]interface{}) {
				content := j2.(map[string]interface{})
				for k, v := range content {
					switch k {
					case "name":
						host.name = v.(string)
					case "alias":
						host.alias = v.(string)
					case "ipaddress":
						host.ipaddress = v.(string)
					case "template":
						host.template = v.(string)
					case "hostgroup":
						host.hostgroup = v.(string)
					case "contact":
						host.contact = v.(string)
					case "contactgroups":
						host.contactgroups = v.(string)
					case "activechecks":
						host.activechecks = v.(string)
					case "servicesets":
						host.servicesets = v.(string)
					case "disable":
						host.disable = v.(string)
					case "displayname":
						host.displayname = v.(string)
					case "parents":
						host.parents = v.(string)
					case "command":
						host.command, _ = UrlDecode(v.(string))
					case "initialstate":
						host.initialstate = v.(string)
					case "maxcheckattempts":
						host.maxcheckattempts = v.(string)
					case "checkinterval":
						host.checkinterval = v.(string)
					case "retryinterval":
						host.retryinterval = v.(string)
					case "passivechecks":
						host.passivechecks = v.(string)
					case "checkperiod":
						host.checkperiod = v.(string)
					case "obsessoverhost":
						host.obsessoverhost = v.(string)
					case "checkfreshness":
						host.checkfreshness = v.(string)
					case "freshnessthresh":
						host.freshnessthresh = v.(string)
					case "eventhandler":
						host.eventhandler = v.(string)
					case "eventhandlerenabled":
						host.eventhandlerenabled = v.(string)
					case "lowflapthresh":
						host.lowflapthresh = v.(string)
					case "highflapthresh":
						host.highflapthresh = v.(string)
					case "flapdetectionenabled":
						host.flapdetectionenabled = v.(string)
					case "flapdetectionoptions":
						host.flapdetectionoptions = v.(string)
					case "processperfdata":
						host.processperfdata = v.(string)
					case "retainstatusinfo":
						host.retainstatusinfo = v.(string)
					case "retainnonstatusinfo":
						host.retainnonstatusinfo = v.(string)
					case "notifinterval":
						host.notifinterval = v.(string)
					case "firstnotifdelay":
						host.firstnotifdelay = v.(string)
					case "notifperiod":
						host.notifperiod = v.(string)
					case "notifopts":
						host.notifopts = v.(string)
					case "notifications_enabled":
						host.notifications_enabled = v.(string)
					case "stalkingoptions":
						host.stalkingoptions = v.(string)
					case "notes":
						host.notes = v.(string)
					case "notes_url":
						host.notes_url = v.(string)
					case "icon_image":
						host.icon_image = v.(string)
					case "icon_image_alt":
						host.icon_image_alt = v.(string)
					case "vrml_image":
						host.vrml_image = v.(string)
					case "statusmap_image":
						host.statusmap_image = v.(string)
					case "coords2d":
						host.coords2d = v.(string)
					case "coords3d":
						host.coords3d = v.(string)
					case "action_url":
						host.action_url = v.(string)
					case "customvars":
						host.customvars = v.(string)
					}
				}
			}
			h.hosts = append(h.hosts, host)
		}

		return nil
	}
}
