package utils

import (
	"bytes"
	"crypto/sha1"
	"fmt"
	"math"
	"net"
	"net/mail"
	"strconv"
	"strings"

	"github.com/PuerkitoBio/goquery"
	"github.com/guregu/null"
	"github.com/guregu/null/zero"
	"github.com/revel/revel"
	"golang.org/x/net/html"
)

type AuthTokenData struct {
	Hash           string `json:"hash"`
	InstallationId int64  `json:"installation_id"`
	UserId         int64  `json:"user_id"`
	Timestamp      string `json:"timestamp"`
}

type ByteSize float64

const (
	REVEL_ADDED_XIP string = "::1" // revel seems to add an XFIP ::1
)

const (
	_           = iota // ignore first value by assigning to blank identifier
	KB ByteSize = 1 << (10 * iota)
	MB
	GB
	TB
	PB
	EB
	ZB
	YB
)

func (b ByteSize) String() string {
	switch {
	case b >= YB:
		return fmt.Sprintf("%.2fYB", b/YB)
	case b >= ZB:
		return fmt.Sprintf("%.2fZB", b/ZB)
	case b >= EB:
		return fmt.Sprintf("%.2fEB", b/EB)
	case b >= PB:
		return fmt.Sprintf("%.2fPB", b/PB)
	case b >= TB:
		return fmt.Sprintf("%.2fTB", b/TB)
	case b >= GB:
		return fmt.Sprintf("%.2fGB", b/GB)
	case b >= MB:
		return fmt.Sprintf("%.2fMB", b/MB)
	case b >= KB:
		return fmt.Sprintf("%.2fKB", b/KB)
	}
	return fmt.Sprintf("%.2fB", b)
}

func IsValidIpAddress(addr string) bool {
	return net.ParseIP(addr) == nil
}

func IsValidEmail(addr string) bool {
	res, err := mail.ParseAddress(addr)
	return err == nil && res.Address != ""
}

/*
Takes a string about to be used in a query in MySQL
and returns a new escaped string
*/
func EscapeSQLQuery(searchTerm string) string {
	s := strings.Replace(searchTerm, `\`, "", -1)
	return strings.Replace(s, "'", `\'`, -1)
}

func ToIntList(ints []int64) string {
	var strIds []string

	for _, i := range ints {
		strIds = append(strIds, strconv.FormatInt(i, 10))
	}

	return strings.Join(strIds, ",")
}

// Go is fucking gay that I had to add this - Topper
func Min(a, b int) int {
	if a > b {
		return b
	}
	return a
}

func Max(a, b int) int {
	if a < b {
		return b
	}
	return a
}

func Round(num float64) int64 {
	return int64(math.Floor(num + .5))
}

func CalculateMaxPages(count int64, pageSize int64) int64 {
	return int64(math.Ceil(float64(count) / float64(pageSize)))
}

func TimeStringToMinutes(timeStr string) (minTime, maxTime int64) {

	switch timeStr {
	case "0-15 min":
		minTime = 0
		maxTime = 15
	case "15-30 min":
		minTime = 15
		maxTime = 30
	case "30-60 min":
		minTime = 30
		maxTime = 60
	case "1-3 hours":
		minTime = 60
		maxTime = 180
	case "3-6 hours":
		minTime = 180
		maxTime = 360
	case "6-12 hours":
		minTime = 360
		maxTime = 720
	case "12-24 hours":
		minTime = 60 * 12
		maxTime = 60 * 24
	case "1-2 days":
		minTime = 60 * 24
		maxTime = 60 * 24 * 2
	case "2+ days":
		minTime = 60 * 24 * 2
		maxTime = 60 * 24 * 365
	}

	return minTime, maxTime
}

func (token *AuthTokenData) GenerateSocketSecret() {
	hash := sha1.Sum([]byte(fmt.Sprintf(
		"i:%v,u:%v,secretKey:%s,timestamp:%s",
		token.InstallationId,
		token.UserId,
		revel.Config.StringDefault("socket.key", "KiaVvR0M7Vp4KiaVvR0M7Vp4"),
		token.Timestamp,
	)))
	token.Hash = fmt.Sprintf("%x", hash)
}

// Sets a nullable valid string bla bla whatever - Sick of writing this shit
func SetNullableString(s *null.String, str string) {
	if str != "" {
		s.String = str
		s.Valid = true
	} else {
		s.Valid = false
	}
}

// Same as string but for integers
func SetNullableInt(i *null.Int, v string) {
	if v != "" {
		var err error

		i.Int64, err = strconv.ParseInt(v, 10, 64)

		if err == nil {
			i.Valid = true
		}
	} else {
		i.Valid = false
	}
}

func SetStringIfSet(rc *revel.Controller, s *string, str string) {
	if len(rc.Request.Form[str]) > 0 {
		*s = rc.Params.Get(str)
	}
}

func SetNullableStringIfSet(rc *revel.Controller, s *null.String, str string) {
	if len(rc.Request.Form[str]) > 0 {
		s.SetValid(rc.Params.Get(str))
	}
}

func SetZeroStringIfSet(rc *revel.Controller, s *zero.String, str string) {
	if len(rc.Request.Form[str]) > 0 {
		s.String = rc.Params.Get(str)
		s.Valid = true
	}
}

func SetBoolIfSet(rc *revel.Controller, b *bool, str string) {
	if len(rc.Request.Form[str]) > 0 {
		val := false

		rc.Params.Bind(&val, str)

		*b = val
	}
}

func SetIntIfSet(rc *revel.Controller, i *int64, str string) {
	if len(rc.Request.Form[str]) > 0 {
		var val int64

		rc.Params.Bind(&val, str)

		*i = val
	}
}

func HTMLtoPlainTextNoNewlines(html string) (output string) {
	q, _ := goquery.NewDocumentFromReader(strings.NewReader(html))
	var buf bytes.Buffer

	for _, n := range q.Nodes {
		buf.WriteString(getNodeText(n))
	}

	return strings.TrimSpace(buf.String())
}

func getNodeText(node *html.Node) string {
	if node.Type == html.TextNode {
		// Quick hack to ignore comments, its not set as a CommentNode when inside a style tag
		if strings.HasPrefix(strings.TrimSpace(node.Data), "<!--") && strings.HasSuffix(strings.TrimSpace(node.Data), "-->") {
			return ""
		}

		// Keep newlines and spaces, like jQuery
		return node.Data
	} else if node.FirstChild != nil {
		var buf bytes.Buffer
		for c := node.FirstChild; c != nil; c = c.NextSibling {
			txt := strings.TrimSpace(getNodeText(c))
			if txt != "" {
				buf.WriteString(txt + " ")
			}
		}
		return buf.String()
	}

	return ""
}

func Int64ArrayToString(ints *[]int64) (output string) {
	l := len(*ints)
	for i := 0; i < l; i++ {
		if i > 0 {
			output = output + ","
		}
		output = output + strconv.FormatInt((*ints)[i], 10)
	}
	return output
}

// Assumes request is behind a load balancer (ELB or HAPROXY) and grabs the original ip
// before falling back to REMOTE_ADDR
func GetClientIP(rc *revel.Controller) (ip string) {
	clientIP := rc.Request.Header.Get("X-Forwarded-For")

	// X-Forwarded-For can actually be an array of addresses separated with ","
	// client IP should be first address
	ip = strings.Split(clientIP, ",")[0]
	if ip != "" && ip != REVEL_ADDED_XIP {
		return ip
	}

	return RemovePortFromIP(rc.Request.RemoteAddr)
}

// Request.RemoteAddress contains port so this strips it off and returns new string
func RemovePortFromIP(s string) string {
	host, _, err := net.SplitHostPort(s)
	if err == nil {
		s = host
	}
	return s
}
