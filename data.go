package nrc

import (
	//"fmt"
	"strings"
)

func FormatData(data, table string) string {

	// data string: host:asdf,ipaddress:1.2.3.4
	// to: "host":"asdf","ipaddress":"1.2.3.4"

	m := make(map[string][]string)

	m["hosts"] = []string{"command"}
	m["services"] = []string{"name", "command", "svcdesc"}
	m["servicesets"] = []string{"name", "command", "svcdesc"}
	m["hosttemplates"] = []string{"checkcommand", "action_url"}
	m["servicetemplates"] = []string{"action_url"}
	m["commands"] = []string{"name", "command"}

	returnString := ""
	comma := ""

	if len(data) > 0 {
		splitdata := strings.Split(data, ",")
		for _, j := range splitdata {
			split := strings.SplitN(j, ":", 2)
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
				returnString += UrlEncodeForce(split[1]) + `"`
				//fmt.Println("Urlencoding")
			}
			comma = ","
		}
	}

	return returnString
}
