package metascraper

import (
	"bytes"
	"golang.org/x/net/html"
	"regexp"
	"strings"
)

var lineFeedReplacer = regexp.MustCompile(`[\n\r]+`)
var whitespaceReplacer = regexp.MustCompile(`\s+`)

// Page represents an HTML document with metadata.
type Page struct {
	URL  string // The web page's URL.
	HTML string // The web page's raw HTML.
	Text string // The text content in the body of the web page, sans markup.
	// Series of more than one line feed are replaced by a single newline.
	// Series of more than one space are replaced by a single space.
	Title        string        // The title of the web page, as given in the head's title element.
	MetaReader   *MetaReader   // A TokenReader for extracting metadata from the document's head.
	SchemaReader *SchemaReader // A TokenReader for extracting schema.org metadata from the document's body.
	PageReader   *PageReader   // A TokenReader for extracting the page's title and text content.
}

// Readers gets a ReaderList aggregating all the TokenReaders associated with
// this page. Client code could add or remove token readers while reusing the
// Read method by embedding the Page struct and overriding this method with one
// that populates the ReaderList differently.
// TODO: Write an example that shows how to extend the Page struct with additional
// token readers.
func (p *Page) Readers() ReaderList {
	return ReaderList{
		Readers: []TokenReader{
			p.PageReader,
			p.MetaReader,
			p.SchemaReader,
		},
	}
}

// MetaData gets the metadata found in this page's head.
func (p *Page) MetaData() []*Meta {
	return p.MetaReader.items
}

// SchemaData gets the schema.org metadata found in this page's body.
func (p *Page) SchemaData() []*ItemScope {
	return p.SchemaReader.items
}

// Read populates the Page struct with content and metadata from the given
// byte array, which the caller is responsible for assuring is well-formed HTML.
func (p *Page) Read(text []byte) error {
	data := bytes.NewReader(text)
	z := html.NewTokenizer(data)
	readers := p.Readers()
	for {
		tt := z.Next()
		text := z.Text()
		switch tt {
		case html.ErrorToken:
			readers.Done()
			// Returning io.EOF indicates success.
			return z.Err()
		case html.TextToken:
			readers.HandleText(text)
		case html.StartTagToken:
			tagName, hasAttr := z.TagName()
			tn := string(tagName)
			attrs := AttrMap(hasAttr, z)
			readers.HandleStart(tn, attrs, z)
		case html.EndTagToken:
			tagName, _ := z.TagName()
			tn := string(tagName)
			readers.HandleEnd(tn, z)
		case html.SelfClosingTagToken:
			tagName, hasAttr := z.TagName()
			tn := string(tagName)
			attrs := AttrMap(hasAttr, z)
			readers.HandleStart(tn, attrs, z)
			readers.HandleEnd(tn, z)
		}
	}
}

// AttrMap parses the attributes of the current element into a friendly map.
// It only makes sense to call this while processing a start or self closing tag token.
func AttrMap(hasAttr bool, z *html.Tokenizer) map[string]string {
	attrs := make(map[string]string)
	if !hasAttr {
		return attrs
	}
	for {
		k, v, more := z.TagAttr()
		attrs[string(k)] = string(v)
		if !more {
			break
		}
	}
	return attrs
}

// PageReader implements the TokenReader interface; it maintains the necessary
// state for extracting the body text and page title from a token stream.
type PageReader struct {
	page     *Page
	inTitle  bool
	inBody   bool
	inScript bool
	text     []byte
}

func (r *PageReader) HandleStart(tn string, attrs map[string]string, z *html.Tokenizer) {
	switch tn {
	case "title":
		r.inTitle = true
	case "body":
		r.inBody = true
	case "script":
		r.inScript = true
	}
}

func (r *PageReader) HandleEnd(tn string, z *html.Tokenizer) {
	switch tn {
	case "title":
		r.inTitle = false
	case "body":
		r.inBody = false
	case "script":
		r.inScript = false
	}
}

func (r *PageReader) HandleText(text []byte) {
	if r.inTitle {
		r.page.Title = string(text)
	} else if r.inBody && !r.inScript {
		r.text = append(r.text, text...)
	}
}

func (r *PageReader) Done() {
	r.text = lineFeedReplacer.ReplaceAll(r.text, []byte("\n"))
	r.text = whitespaceReplacer.ReplaceAll(r.text, []byte{' '})
	r.page.Text = strings.TrimSpace(string(r.text))
}
