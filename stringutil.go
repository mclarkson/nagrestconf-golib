package nrc

import (
	"net/url"
)

var encode bool

func SetEncode(tf bool) {
	encode = tf
}

func UrlDecodeForce(str string) (string, error) {
	u, err := url.QueryUnescape(str)
	if err != nil {
		return "", err
	}
	return u, nil

	return str, nil
}

func UrlDecode(str string) (string, error) {
	if encode == false {
		u, err := url.QueryUnescape(str)
		if err != nil {
			return "", err
		}
		return u, nil
	}

	return str, nil
}

func UrlEncodeForce(str string) string {
	u := url.QueryEscape(str)
	return u
}

func UrlEncode(str string) string {
	u := url.QueryEscape(str)
	return u
}
