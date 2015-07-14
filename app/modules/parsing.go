package modules

import (
	"encoding/json"
	"regexp"
	"strconv"
	"strings"

	"github.com/PuerkitoBio/goquery"
	"github.com/ThomsonReutersEikon/mailstrip"
)

func SanitizeHTMLForDisplay(html string) string {
	doc, _ := goquery.NewDocumentFromReader(strings.NewReader(html))

	doc.Find("script").Remove()

	doc.Find("img").Each(func(i int, s *goquery.Selection) {
		s.RemoveAttr("onerror")
	})

	html, _ = doc.Html()

	return html
}

func FormatPlainForwardedEmail(subject string, body string) (string, string) {
	subject = formatSubject(subject)

	lines := strings.Split(body, "\n")

	var newBody string
	fwdMatch := regexp.MustCompile(`(?i)-+\s*forwarded\smessage\s*-+`)

	for _, line := range lines {
		if fwdMatch.MatchString(line) {
			break
		}

		newBody += line + "\n"
	}

	newBody = strings.TrimSuffix(newBody, "\n")
	newBody = strings.TrimPrefix(newBody, "\n")
	newBody = strings.Replace(newBody, "\n", "<br/>", -1)
	return subject, newBody
}

func formatSubject(subject string) string {
	return strings.TrimSpace(regexp.MustCompile(`(?i)^Fwd?:?\s?`).ReplaceAllString(subject, ""))
}

func FormatForwardedEmail(subject, body string) (string, string) {
	subject = formatSubject(subject)

	doc, _ := goquery.NewDocumentFromReader(strings.NewReader(body))

	// Let's try the gmail style stuff
	s := doc.Find("div.gmail_quote")

	if s.Length() > 0 {
		if s.Find("div").Length() > 0 {
			if regexp.MustCompile(`(?i)((-*\sforwarded\smessage\s-*|from:|to:|subject:|date:|sent:)[\s\S]*?){4,}`).MatchString(s.Text()) {
				body, _ = s.Find("div:last-child").Html()
				return subject, body
			}
		}
	}

	// Outlook style
	outlookReg := regexp.MustCompile(`From:(.*)[\]>\s]\n*(Sent|Date):(.*)\n*To:\s?(.*?)\n*Subject:`)
	doc.Find("blockquote").EachWithBreak(func(i int, s *goquery.Selection) bool {
		if outlookReg.MatchString(s.Text()) {
			x := s

			for s.Parent().Length() > 0 {
				s.PrevAll().Remove()

				s = s.Parent()
			}

			x.Remove()

			doc.Find("blockquote").Each(func(blockI int, blockS *goquery.Selection) {
				html, _ := blockS.Html()

				blockS.ReplaceWithHtml(html)
			})
			return false
		}

		return true
	})

	doc.Find("div.moz-forward-container").Find("table").First().Remove()

	body, _ = doc.Html()

	return subject, body
}

func SplitPlainEmailBodyAndSignature(originalBody, replyBody string) (signature, body string) {
	// First we should make sure we're on the same page when it comes to line breaks
	originalBody = strings.Replace(originalBody, "\r\n", "\n", -1)
	replyBody = strings.Replace(replyBody, "\r\n", "\n", -1)

	// We should also rip out replies before moving on to paralell matching
	originalBody = StripPlainEmailReplies(originalBody)
	replyBody = StripPlainEmailReplies(replyBody)

	// First we split it up into lines (makes matching easier and faster, although potentially less accurate)
	original := strings.Split(originalBody, "\n")
	reply := strings.Split(replyBody, "\n")

	body = replyBody

	signatures := map[int]string{}

	// Now we need to traverse the lines attempting to match it up
	for originalIdx, originalLine := range original {
		if strings.TrimSpace(originalLine) != "" {
			// Let's try and find the equivelant line in the reply and start
			// a parallel matching process.
			for replyIdx, replyLine := range reply {
				if strings.TrimSpace(originalLine) == strings.TrimSpace(replyLine) {
					// Right, we've got it in the reply, parallel matching time]
					signature := matchPlainText(original, reply, originalIdx, replyIdx)

					signatures[replyIdx] = signature
				}
			}
		}
	}

	// We want to take the longest match as the signature
	if len(signatures) > 0 {
		longest := 0
		signatureLineStart := 0

		for idx, s := range signatures {
			length := len(s)

			if length > longest {
				signature = s
				longest = length
				signatureLineStart = idx
			}
		}

		// Calculate how much percent the matched signature takes up
		percentTakenUp := (float64(len(signature)) / float64(len(replyBody))) * 100

		// If we've taken up less than 80%, let's strip the signature from the text
		if percentTakenUp < 80 {
			body = ""

			for idx, line := range reply {
				if idx >= signatureLineStart {
					break
				}

				body += line + "\n"
			}
		}
	}

	return signature, body
}

func FindEmailMessageIdInBody(bodyContent ...string) (messageId string) {

	for _, body := range bodyContent {
		q, _ := goquery.NewDocumentFromReader(strings.NewReader(body))

		msgIdDIV := q.Find("div#twDeskEmailMessageId")
		found := false

		if msgIdDIV.Length() > 0 {
			msgIdJSON := map[string]string{}

			// We found the div
			err := json.Unmarshal([]byte(msgIdDIV.Text()), &msgIdJSON)

			if err == nil {
				messageId = "<" + msgIdJSON["TWDESKMessageId"] + ">"
				found = true
			}
		}

		if !found {
			// We couldn't find it via DOM parsing, so let's try simple regex matching as a
			// last resort. This is generally the fault of email clients like Outlook which seem
			// to sometimes change the HTML that we sent it.
			result := regexp.MustCompile(`\{\"TWDESKMessageId\"\:\"(.*?)\"\}`).FindStringSubmatch(q.Text())

			// If we have > 0 results then we found it
			if len(result) > 1 && len(strings.TrimSpace(result[1])) > 0 {
				messageId = "<" + result[1] + ">"
			}
		}

		// Now we need to restore the format
		messageId = strings.Replace(messageId, "!", ".", -1)
		messageId = strings.Replace(messageId, "*", "@", 1)

		if strings.TrimSpace(messageId) != "" {
			return messageId
		}
	}

	return messageId
}

func matchPlainText(original, reply []string, originalIdx, replyIdx int) (signature string) {
	for originalIdx < len(original) && replyIdx < len(reply) {
		if strings.TrimSpace(original[originalIdx]) == strings.TrimSpace(reply[replyIdx]) {
			signature += reply[replyIdx]
		} else {
			break
		}

		originalIdx++
		replyIdx++
	}

	return signature
}

func SplitHTMLEmailBodyAndSignature(originalBody string, replyBody string, useParallelMatching bool) (signature, body string) {
	foundSignature := false

	rOriginal := strings.NewReader(originalBody)
	rReply := strings.NewReader(replyBody)

	originalHTML, _ := goquery.NewDocumentFromReader(rOriginal)
	replyHTML, _ := goquery.NewDocumentFromReader(rReply)

	// First we should try the "official" signature patterns
	gmailSignature := replyHTML.Find("div.gmail_signature")

	if gmailSignature.Length() > 0 {
		signature, _ = gmailSignature.Html()
		gmailSignature.Remove()
		foundSignature = true
	}

	oxSignature := replyHTML.Find("div#ox-signature")

	if oxSignature.Length() > 0 {
		signature, _ = oxSignature.Html()
		oxSignature.Remove()
		foundSignature = true
	}

	// Remove "bloop_sign" - Some signature
	bloopSig := replyHTML.Find("div.bloop_sign")

	if bloopSig.Length() > 0 {
		signature, _ = bloopSig.Html()
		bloopSig.Remove()
		foundSignature = true
	}

	// Remove "moz-signature"
	mozSignature := replyHTML.Find("pre.moz-signature")

	if mozSignature.Length() > 0 {
		signature, _ = mozSignature.Html()
		mozSignature.Remove()
		foundSignature = true
	}

	if useParallelMatching {
		// Right, this is a reply, so let's do some harsh reply ripping
		replyHTML = StripEmailReplies(replyHTML)

		// We should rip replies on the originalHTML passed in here also
		// This will result in much better paralell matching performance overall
		// although it may impact signature ripping negatively unfortunately.
		// It is worth it though, from 10 seconds to 283ms in heavy test cases.
		originalHTML = StripEmailReplies(originalHTML)
	}

	if !foundSignature && useParallelMatching {
		return splitViaParallelMatching(originalHTML, replyHTML)
	}

	body, _ = replyHTML.Html()

	return signature, body
}

func StripPlainEmailReplies(body string) string {
	result := mailstrip.Parse(body)
	return result.String()
}

// I'm being somewhat cautious with the expressions used here to strip replies.
// They are fairly specific.
func StripEmailReplies(replyHTML *goquery.Document) *goquery.Document {
	var found bool
	blockReg := regexp.MustCompile(`(?i)^on.*?(<|>|,)+[\s\S]*?wrote:`)
	outlookReg := regexp.MustCompile(`(?i)^(([\s-]*original message[\s-]*|from:|to:|subject:|date:|sent:)[\s\S]*?){4,}`)

	// Generically try to remove agent replies table
	replyHTML.Find("table.twDeskNotificationTable").Remove()

	// Try to find replies from a Teamwork Desk thread
	s := replyHTML.Find("div#twDeskReplyStart table")

	if s.Length() > 0 {
		s.NextAll().Remove()
	}

	// Remove gmail_quotes (thanks Google)
	s = replyHTML.Find("div.gmail_quote,blockquote.gmail_quote")

	if s.Length() > 0 {
		s.Remove()
		found = true
	}

	// Remove zmail_extra
	s = replyHTML.Find("div.zmail_extra").Remove()

	if s.Length() > 0 {
		s.Remove()
		found = true
	}

	// Remove mozilla quotes
	s = replyHTML.Find("div.moz-cite-prefix")

	if s.Length() > 0 {
		text := strings.TrimSpace(s.Text())

		// Do some safety checking
		if blockReg.MatchString(text) {
			deleteAfterSelection(s)
			s.Remove()
			found = true
		}
	}

	// Try to remove Outlook replies based on the line they insert
	s = replyHTML.Find("div.WordSection1 > div div")

	if s.Length() > 0 {
		style, found := s.Attr("style")

		if found {
			if regexp.MustCompile(`border-top:\s*solid.*;`).MatchString(style) {
				deleteAfterSelection(s.Parent())
				s.Parent().Remove()
				found = true
			}
		}
	}

	// First we can try a simple blockquote removal
	replyHTML.Find("blockquote").EachWithBreak(func(i int, s *goquery.Selection) bool {
		searchText := strings.TrimSpace(s.Text())

		if blockReg.MatchString(searchText) {
			deleteAfterSelection(s.Parent())
			s.Remove()
			found = true
			return false
		}

		return true
	})

	if !found {
		replyHTML.Find("p,div").EachWithBreak(func(i int, s *goquery.Selection) bool {
			searchText := strings.TrimSpace(s.Text())

			// Blockquote replies
			if blockReg.MatchString(searchText) {
				s.Remove()
				s.After("blockquote").Remove()
				found = true
				return false
			} else if outlookReg.MatchString(searchText) {
				// Delete everything after this point
				deleteAfterSelection(s.Parent())
				s.Remove()
				found = true
				return false
			}

			return true
		})
	}

	lineByLineReg := regexp.MustCompile(`(?i)^([\s-]*original message[\s-]*|from:|to:|subject:|date:|sent:).*?$`)
	genericWroteReg := regexp.MustCompile(`(?i)on.*?,.*?wrote:$`)

	// If we still haven't found it, last resort
	if !found {
		html, _ := replyHTML.Html()
		htmlLines := strings.Split(html, "\n")

		matchCount := 0

		var newHTML string

		for _, line := range htmlLines {
			doc, _ := goquery.NewDocumentFromReader(strings.NewReader(line))

			if lineByLineReg.MatchString(doc.Text()) {
				matchCount++

				if matchCount >= 4 {
					found = true
					break
				} else {
					continue
				}
			} else if genericWroteReg.MatchString(doc.Text()) {
				found = true
				break
			}

			newHTML += line

			matchCount = 0
		}

		if found {
			replyHTML, _ = goquery.NewDocumentFromReader(strings.NewReader(newHTML))
		}
	}

	// If it's still not found, do a last resort and try ripping out by blockquote
	if !found {
		replyHTML.Find("blockquote").First().Remove()
	}

	return replyHTML
}

func deleteAfterSelection(s *goquery.Selection) {
	if s.Is("body") {
		return
	}

	s.NextAll().Remove()
	deleteAfterSelection(s.Parent())
}

func splitViaParallelMatching(originalHTML, replyHTML *goquery.Document) (signature, body string) {
	signatures := map[string]int{}

	// Rip out the style tag (there tends to be comments in there which mess with parallel matching)
	replyHTML.Find("style").Remove()

	// Rip out title tag (can cause accidental matches)
	replyHTML.Find("title").Remove()

	// Remove everything after this element as it's our reply crap
	replyHTML.Find("div#twDeskReplyStart").NextAll().Remove()
	replyHTML.Find("div#twDeskReplyStart").Remove()

	var indexOfLastElement int

	// Okay, now we need to traverse the HTML to try and find matches
	replyHTML.Find("*").EachWithBreak(func(iReply int, sReply *goquery.Selection) bool {
		if sReply.Find("img").Length() > 0 {
			return true
		}

		// If it's not blank, then we go through the DOM of the original to try and find the same
		if strings.TrimSpace(sReply.Text()) != "" {
			originalHTML.Find("*").EachWithBreak(func(iOriginal int, sOriginal *goquery.Selection) bool {
				if sOriginal.Find("img").Length() > 0 {
					return true
				}

				if strings.TrimSpace(sOriginal.Text()) == strings.TrimSpace(sReply.Text()) {
					// We have a match, let's fork it off into another function which
					// will do the signature matching
					signature = signatureMatch(sReply, sOriginal)

					signatures[signature] = iReply
				}

				return true
			})
		}

		indexOfLastElement = iReply

		return true
	})

	// If we have a slice of signatures, use the longest one?!
	if len(signatures) > 0 {
		longest := 0
		domPosition := 0

		for s, domPos := range signatures {
			length := len(s)

			if length > longest {
				signature = s
				longest = length
				domPosition = domPos
			}
		}

		// Numwords should be based on plaintext
		var numWords int

		signatureHTML, err := goquery.NewDocumentFromReader(strings.NewReader(signature))

		if err == nil {
			numWords = len(strings.Split(signatureHTML.Text(), " "))
		} else {
			numWords = len(strings.Split(signature, " "))
		}

		percentAppearsIn := (float64(domPosition) / float64(indexOfLastElement)) * 100.00

		// Only start stripping if signature length is greater than 0.
		// Let's also make sure we have a good few words, not just "Hey Michael" etc.
		if len(strings.TrimSpace(signature)) > 0 && numWords > 2 && percentAppearsIn > 30 {
			// First we make a clone of the document and see if this will result
			// in empty content before deciding to fully remove
			clonedReply := goquery.CloneDocument(replyHTML)

			clonedReply.Find("*").Each(func(iClone int, sClone *goquery.Selection) {
				if iClone >= domPosition {
					sClone.Remove()
				}
			})

			// If we still have content, then overwrite the replyHTML document
			// with our cleaned one
			if strings.TrimSpace(clonedReply.Text()) != "" {
				replyHTML = clonedReply
			} else {
				signature = ""
			}
		} else {
			signature = ""
		}
	}

	resultHTML, _ := replyHTML.Html()

	return signature, resultHTML
}

func signatureMatch(sReply, sOriginal *goquery.Selection) (signature string) {
	// Traverse both sReply and sOriginal in a loop until one of em
	// runs out of steam
	for sReply.Length() > 0 && sOriginal.Length() > 0 {
		if strings.TrimSpace(sReply.Text()) == strings.TrimSpace(sOriginal.Text()) {
			html, _ := sOriginal.Html()

			// Build current tag (not included in html)
			if sOriginal.Get(0).Type != 1 {
				tag := "<" + sOriginal.Get(0).Data

				for _, attr := range sOriginal.Get(0).Attr {
					tag += " " + attr.Key + "=\"" + attr.Val + "\""
				}

				tag += ">"

				html = tag + html + "</" + sOriginal.Get(0).Data + ">"
			}

			signature += html
		} else if strings.TrimSpace(sReply.Text()) != "" && strings.TrimSpace(sOriginal.Text()) != "" {
			break
		}

		sReply = sReply.Next()
		sOriginal = sOriginal.Next()
	}

	return signature
}

func FormatPlainEmail(body string) string {
	body = strings.Replace(body, "\r\n", "<br />", -1)
	body = strings.Replace(body, "\n", "<br />", -1)

	doc, err := goquery.NewDocumentFromReader(strings.NewReader(body))

	if err != nil {
		return body
	}

	doc.Find("script").Remove()

	body, _ = doc.Html()

	return body
}

func CleanupHTMLEmail(bodyHTML string) string {
	body := bodyHTML

	r := strings.NewReader(body)

	doc, err := goquery.NewDocumentFromReader(r)

	if err != nil {
		panic("Couldn't parse document in GoQuery: " + err.Error())
	}

	// Remove style tags
	doc.Find("style").Remove()

	// Remove base href shite
	doc.Find("base").Remove()

	// Remove head?
	doc.Find("title").Remove()

	// Remove script tags
	doc.Find("script").Remove()

	newHTML := ""
	isOL := false

	// Let's turn "MsoListParagraphs" into a real list (wtf wingdings)
	doc.Find("p.MsoListParagraph").Each(func(i int, s *goquery.Selection) {
		// Remove wingdings
		text := strings.Replace(s.Text(), "§", "", -1)

		// Replace multiple spaces with one
		text = regexp.MustCompile("(&nbsp;|\\s|\u00A0){2,}").ReplaceAllString(text, "")

		if !s.Prev().HasClass("MsoListParagraph") {
			if regexp.MustCompile("^\\d\\.").MatchString(text) {
				isOL = true
			}

			if isOL {
				newHTML += "<ol>"
			} else {
				newHTML += "<ul>"
			}
		}

		if isOL {
			// Remove the starting digits from the text
			text = regexp.MustCompile("^\\d\\.").ReplaceAllString(text, "")
		}

		newHTML += "<li>" + text + "</li>"

		if !s.Next().HasClass("MsoListParagraph") {
			if isOL {
				newHTML += "</ol>"
			} else {
				newHTML += "</ul>"
			}

			s.AfterHtml(newHTML)

			newHTML = ""
		}
	})

	doc.Find("p.MsoListParagraph").Remove()

	// Remove attributes from font tags
	doc.Find("font").Each(func(i int, s *goquery.Selection) {
		for _, node := range s.Nodes {
			for idx := range node.Attr {
				attr := &node.Attr[idx]

				attr.Val = ""
			}
		}
	})

	var fullPreBlock string

	// Want to convert Outlook style codeblocks into a proper block (single pre)
	doc.Find("pre").Each(func(i int, s *goquery.Selection) {
		html, _ := s.Html()
		fullPreBlock += html + "<br />"

		if !s.Next().Is("pre") {
			s.AfterHtml("<pre>" + fullPreBlock + "</pre>")
			fullPreBlock = ""
		}

		s.Remove()
	})

	// Needs to be done last
	doc.Find("*").Each(func(i int, s *goquery.Selection) {
		switch s.Get(0).Data {
		case "br", "img", "span":
			return
		case "div":
			// We're checking this special case, some retard mail client adds in the pattern "<div><br></div>"
			// for what it seems to think are paragraphs, go figure.
			if strings.TrimSpace(s.Text()) == "" && s.Children().Length() > 0 && s.Children().Get(0).Data == "br" {
				s.AfterHtml("<p></p>")
				s.Remove()
				return
			}
		}

		// Don't strip out stuff in pre tags as they are probably already formatted
		if s.Parent().Is("pre") || s.Parent().Parent().Is("pre") {
			return
		}

		// Remove empty elements
		if strings.TrimSpace(s.Text()) == "" {
			if s.Find("img").Length() == 0 {
				s.Remove()
			}
		}
	})

	// Remove the first br
	br := doc.Find("br").First()

	if br.Prev().Length() == 0 {
		br.Remove()
	}

	br = doc.Find("br").Last()

	if br.Next().Length() == 0 {
		br.Remove()
	}

	// Fix Helpscout being cc'd to us
	hsReply, _ := doc.Find("table#hsReplyAbove table").First().Find("tbody tr").Last().Html()

	doc.Find("table#hsReplyAbove").AfterHtml(hsReply)
	doc.Find("table#hsReplyAbove").Remove()

	// Remove header tags
	for i := 0; i < 6; i++ {
		doc.Find("h" + strconv.Itoa(i)).Each(func(i int, s *goquery.Selection) {
			html, _ := s.Html()

			s.ReplaceWithHtml("<p>" + html + "</p>")
		})
	}

	// Set all links to be open in a new window
	doc.Find("a").Each(func(i int, s *goquery.Selection) {
		href, _ := s.Attr("href")
		html, _ := s.Html()

		s.ReplaceWithHtml(`<a href="` + href + `" target="_blank">` + html + `</a>`)
	})

	// Add white-space pre-wrap to code tags so they wrap
	code := doc.Find("code")

	if code.Length() > 0 {
		style, exists := code.Attr("style")

		if exists {
			style += ";"
		}

		style += "white-space:pre-wrap;"

		code.SetAttr("style", style)
	}

	// Remove script tags (we do this at the end again in case we accidentally
	// unstripped a HTML encoded tag for example and actually constructed a script
	// tag incidentally)
	doc.Find("script").Remove()

	html, err := doc.Html()

	// Replace multiple br with a single paragraph
	html = regexp.MustCompile(`(<br\s*\/?>\n?\s?){2,}`).ReplaceAllString(html, "<p></p>")

	return html
}
