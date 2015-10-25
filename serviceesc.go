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

type serviceesc struct {
	hostname      string
	hostgroupname string
	svcdesc       string
	contacts      string
	contactgroups string
	firstnotif    string
	lastnotif     string
	notifinterval string
	period        string
	escopts       string
	disable       string
}

type Serviceesc struct {
	serviceesc []serviceesc
}

func (h *Serviceesc) FilterServiceesc(filter string) {

	f := NewFilter(filter)

	newh := []serviceesc{}

	for _, k := range h.serviceesc {
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
	h.serviceesc = newh
}

func (h Serviceesc) ShowServiceescJson(newline, brief bool, filter string) {

	if filter != "" {
		h.FilterServiceesc(filter)
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
	for _, r := range h.serviceesc {
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

func NewNrcServiceesc() Serviceesc {
	return Serviceesc{}
}

/*
 * Send HTTP GET request
 */
func (h *Serviceesc) GetServiceesc(url, endpoint string) (e error) {

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

	//fmt.Printf("%s\n", url+"/"+endpoint)
	resp, err := client.Get(url + "/" + endpoint)
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
			serviceesc := serviceesc{}
			for _, j2 := range j.([]interface{}) {
				content := j2.(map[string]interface{})
				for k, v := range content {
					switch k {
					case "hostname":
						serviceesc.hostname = v.(string)
					case "hostgroupname":
						serviceesc.hostgroupname = v.(string)
					case "svcdesc":
						serviceesc.svcdesc = v.(string)
					case "contacts":
						serviceesc.contacts = v.(string)
					case "contactgroups":
						serviceesc.contactgroups = v.(string)
					case "firstnotif":
						serviceesc.firstnotif = v.(string)
					case "lastnotif":
						serviceesc.lastnotif = v.(string)
					case "notifinterval":
						serviceesc.notifinterval = v.(string)
					case "period":
						serviceesc.period = v.(string)
					case "escopts":
						serviceesc.escopts = v.(string)
					case "disable":
						serviceesc.disable = v.(string)
					}
				}
			}
			h.serviceesc = append(h.serviceesc, serviceesc)
		}

		return nil
	}
}
