package tests

import (
	"fmt"
	"strings"

	"github.com/PuerkitoBio/goquery"
	"github.com/irlTopper/lifevault/app/modules"
	"github.com/revel/revel/testing"
)

// Must Embed `revel.TestSuite`
type MyAppTest struct {
	testing.TestSuite
}

// Run this before a request
func (t *MyAppTest) Before() {
	println("Set up")
}

// Run this after request
func (t *MyAppTest) After() {
	println("Tear down")
}

// Check main page is kinda there
func (t *MyAppTest) TestIndexPage() {
	t.Get("/")
	t.AssertOk()
	t.AssertContentType("text/html")
}

// Check if robots.txt exists
func (t *MyAppTest) TestRobotsPage() {
	t.Get("/robots.txt")
	t.AssertOk()
	t.AssertContentType("text/html")
}

// Will not appear in panel as it not start with `Test` case sensitive
func (t *MyAppTest) TestFavIcon() {
	t.Get("/favicon.ico")
	t.AssertOk()
	t.AssertContentType("text/html")
}

// Will not appear in panel as it not start with `Test` case sensitive
func (t *MyAppTest) TestEmailParseTest() {
	BodyPlain := "test\n\n>> sig"
	BodyHTML := "<div dir=\"ltr\">Crap sleep.<div>Middle of the night sex with Claire - woke her up.</div><div>Woke up very early in the morning - looking over SaaS spreadsheet.</div><div>Into Costa early.</div><div><br></div><div>Claires sister Orla in hospital - cyst on ovaries. Damo going in later for shoulder op. Claire helping out.</div><div><br></div><div>Ended up chatting about ideas re process to DC for ages. Great to be on the same page.</div><div>We identified that we need somebody in a sales role clawing back ex trial customers.</div><div><br></div><div>Joined lads for coffee outside Costa. Chatted about UFC McGregor fight. </div><div>Into work. Fixed a bug in sub tasks on sub tasks - was locked to our installation for testing so MJ couldn&#39;t test.</div><div><br></div><div>Talked to Adrian and Vic about improving webinar process for sales process.</div><div><br></div><div>After just finishing &quot;Project Pheonix&quot; book it seems that the world is telling me all about &quot;Process&quot; and the need to orgently fix the processes in Teamwork. I see it very clearly now that this is out next biggest challenge.</div><div><br></div><div>So I spent time with Mike and I&#39;ve taken over having meetings with him since Dan only had them informally.</div><div>We went through how the feature selection process should work with my ideas about bringing support and marketing influence in. Then I talk to Adam, then Dan, then Adrian and DC - everybody agrees that the new process is a great thing.</div><div><br></div><blockquote style=\"margin:0 0 0 40px;border:none;padding:0px\"><div><div>Feature Planning</div></div><div><div><br></div></div><div><div>It became apparent that we need a standard process around how we agree on which features are next for development by each team. The goal is to allow everyone to clearly see the product roadmap and allow Marketing and Support Ops to help decide what gets worked on next. The agreed process is as follows:</div></div><div><div><br></div></div><div><div>The heads of product will meet with DC and Adrian on the first working day of every month. Face-to-face preferable, skype as necessary. The purpose is to decide which are the most important features representing the needs of the Marketing Department, Support Ops (representing the customer) and the realities for the Product Manager.</div></div><div><div>Each of the three people at the meeting gets 3 votes for the features they want from the roadmap.</div></div><div><div>The Product Manager takes the votes on board (including his own), decides what the top 3 features are and organises the roadmap accordingly.</div></div><div><div><br></div></div><div><div>Caveats:</div></div><div><div>If a person causes the meeting to be missed for any reason, they are responsible to ensure the meeting is held at the earliest possible date.</div></div><div><div>Only the Product Manager can add features to the Roadmap.</div></div></blockquote><div><br></div><div>Turns out Claire&#39;s sister Orla has a miscarriage. :(</div><div><br></div><div>Home. Have to call Karen back about mam - tried and no answer. Talking to Bev. Dreading it.</div><div><br></div><div>Claire home late - we watched some The Profit and I passed out for a few minutes. </div><div><br></div><div>Bed. Lifevault working - the daily email through. Fixed up the message re X days ago and then the from address. Then I wrote this. So bloody tired.</div><div><br></div><div>zzz</div><div><br></div></div><div class=\"gmail_extra\"><br><div class=\"gmail_quote\">On Mon, Jul 13, 2015 at 9:58 PM,  <span dir=\"ltr\">&lt;<a href=\"mailto:example@lifevaultapp.com\" target=\"_blank\">example@lifevaultapp.com</a>&gt;</span> wrote:<br><blockquote class=\"gmail_quote\" style=\"margin:0 0 0 .8ex;border-left:1px #ccc solid;padding-left:1ex\">Just reply to this email with your entry.<br><br>Remember this? A while back you wrote:<br><br>Lie in with Claire. Fixed bed.\nWent to Hanlys for brunch.\nThen Costa - finished book.\nThen came to WWh to work on LifeVault!\n\nIt&#39;s alive?!\n\n\nOn Sunday, July 12, 2015, <u></u> wrote:\n\n&gt; Just reply to this email with your entry.\n&gt;\n&gt; Remember this? A while back you wrote:\n&gt;\n&gt; This is an example inbound message.\n\n\n\n-- \n\n\n[image: photo]\n*Peter Coppinger*\n\nCEO / FOUNDER / LEAD DEVELOPER\n<a href=\"http://TEAMWORK.COM\" target=\"_blank\">TEAMWORK.COM</a> - The World&#39;s Favourite Way To Get Stuff Done\n\n - - end - -"

	var body string

	fmt.Println("start body:", len(BodyHTML))
	// We don't want to use the new processing on agent emails just yet for safety

	// Check if it's a plaintext email
	if strings.TrimSpace(BodyHTML) == "" {

		body = modules.StripPlainEmailReplies(BodyPlain)
		_, body = modules.SplitPlainEmailBodyAndSignature(BodyPlain, body)

		if strings.TrimSpace(body) == "" {
			body = BodyPlain
		}

		body = modules.FormatPlainEmail(body)

	} else {

		fmt.Println("here001", len(BodyHTML))

		_, body = modules.SplitHTMLEmailBodyAndSignature("", BodyHTML, true)

		doc, _ := goquery.NewDocumentFromReader(strings.NewReader(body))

		if strings.TrimSpace(doc.Text()) == "" {
			fmt.Println("here002")
			body = BodyHTML
		}

		body = modules.CleanupHTMLEmail(body)

	}

	fmt.Println("end body:", len(body))

	t.Assert(true)
}
