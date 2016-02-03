package nrc

import (
	"strings"
)

func FormatData(data []string, table string) (string, error) {

	// data string: host:asdf,ipaddress:1.2.3.4
	// to: "host":"asdf","ipaddress":"1.2.3.4"

	m := make(map[string][]string)

	m["hosts"] = []string{"command", "alias"}
	m["services"] = []string{"name", "command", "svcdesc"}
	m["servicesets"] = []string{"name", "command", "svcdesc"}
	m["hosttemplates"] = []string{"checkcommand", "action_url"}
	m["servicetemplates"] = []string{"action_url"}
	m["commands"] = []string{"name", "command"}

	returnString := ""
	comma := ""

	for _, j := range data {
		split := strings.SplitN(j, ":", 2)
		if len(split) < 2 {
			return "", HttpError{}
		}
		var encode = false
		// check whether field should be encoded
		for _, i := range m[table] {
			if i == split[0] {
				encode = true
			}
		}
		if encode == false {
			returnString += comma + `"` + split[0] + `":"`
			returnString += split[1] + `"`
		} else {
			returnString += comma + `"` + split[0] + `":"`
			val := strings.Replace(split[1], `"`, `\"`, -1)
			val = UrlEncodeForce(val)
			returnString += val + `"`
		}
		comma = ","
	}

	return returnString, nil
}
