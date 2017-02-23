package regex

import "regexp"

const (
	regex_email_pattern        = `(?i)[A-Z0-9._%+-]+@(?:[A-Z0-9-]+\.)+[A-Z]{2,6}`
	regex_strict_email_pattern = `(?i)[A-Z0-9!#$%&'*+/=?^_{|}~-]+` +
		`(?:\.[A-Z0-9!#$%&'*+/=?^_{|}~-]+)*` +
		`@(?:[A-Z0-9](?:[A-Z0-9-]*[A-Z0-9])?\.)+` +
		`[A-Z0-9](?:[A-Z0-9-]*[A-Z0-9])?`
	regex_url_pattern = `(ftp|http|https):\/\/(\w+:{0,1}\w*@)?(\S+)(:[0-9]+)?(\/|\/([\w#!:.?+=&%@!\-\/]))?`
)

var (
	regex_email        *regexp.Regexp
	regex_strict_email *regexp.Regexp
	regex_url          *regexp.Regexp
)

func init() {
	regex_email, _ = regexp.Compile(regex_email_pattern)
	regex_strict_email, _ = regexp.Compile(regex_strict_email_pattern)
	regex_url, _ = regexp.Compile(regex_url_pattern)
}

func IsEmail(email string) bool {
	return regex_email.MatchString(email)
}

func IsEmailRFC(email string) bool {
	return regex_strict_email.MatchString(email)
}

func IsUrl(url string) bool {
	return regex_url.MatchString(url)
}
