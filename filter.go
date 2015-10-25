package nrc

import (
	"strings"
)

type filterStruct struct {
	names []string
	regex []string
}

func NewFilter(filter string) filterStruct {

	f := filterStruct{}

	if len(filter) > 0 {
		filters := strings.Split(filter, ",")
		for _, j := range filters {
			split := strings.SplitN(j, ":", 2)
			if encode == false {
				f.names = append(f.names, split[0])
				f.regex = append(f.regex, split[1])
			} else {
				f.names = append(f.names, UrlEncodeForce(split[0]))
				f.regex = append(f.regex, UrlEncodeForce(split[1]))
			}
		}
	}

	return f
}
