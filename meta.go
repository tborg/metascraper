package metascraper

import (
	"golang.org/x/net/html"
	"log"
	"strings"
)

// Meta represents a `meta` tag in the head of an HTML document.
// Structured properties with extra metadata attached to them are represented
// in HTML documents as a series of `meta` tags where the "extra" metadata properties
// have the same prefix as the preceding tag, followed by an extra ":"; extra
// metadata associated with a property is embedded recursively in the parent
// struct. This nesting is guaranteed to be at most one level deep.
type Meta struct {
	Property string  // The tag's property attribute, if any.
	Content  string  // The tag's content attribute or text content, if any.
	Name     string  // The tag's name attribute, if any.
	Extra    []*Meta // Collects subsequent adjacent metadata with this property prefix.
}

// MetaReader implements the TokenReader interface; it maintains the necessary
// state for extracting structured metadata from a stream of HTML tokens.
type MetaReader struct {
	items  []*Meta
	inside bool
	inHead bool
}

func (r *MetaReader) HandleStart(tn string, attrs map[string]string, z *html.Tokenizer) {
	switch tn {
	// Meta tags outside of the document head are not considered.
	case "head":
		r.inHead = true
	case "meta":
		if !r.inHead {
			return
		}
		// Meta tags are typically self-closing. If the document auther does
		// happen to specify the content as a text node rather than using the
		// standard content attribute, the text node will be recognized as content.
		r.inside = true
		cur, exists := r.current()
		m := r.next()
		// Fun fact: although the property attribute is widely used
		// (for instance by the opengraph spec), it's not actually a standard
		// attribute of the meta tag in XHTML1 or HTML5.
		// http://stackoverflow.com/questions/18448871/meta-tag-there-is-no-attribute-property-and-other-attribute-error
		m.Property = attrs["property"]
		m.Content = attrs["content"]
		m.Name = attrs["name"]
		// Handle structured properties:
		if exists && strings.HasPrefix(m.Property, cur.Property+":") {
			// Pop the current item off the end of the item list and embed it in the previous item.
			r.makeCurrentExtra()
		}
	}
}

func (r *MetaReader) HandleEnd(tn string, z *html.Tokenizer) {
	switch tn {
	case "head":
		r.inHead = false
	case "meta":
		r.inside = false
	}
}

func (r *MetaReader) HandleText(text []byte) {
	// Handles the rare case where a meta tag's content was specified as a text node.
	if r.inside && r.inHead {
		cur, exists := r.current()
		if exists && cur.Content == "" {
			cur.Content += string(text)
		}
	}
}

func (r *MetaReader) Done() {
	// No cleanup.
}

func (r *MetaReader) current() (*Meta, bool) {
	m := &Meta{}
	if len(r.items) == 0 {
		return m, false
	}
	m = r.items[len(r.items)-1]
	return m, true
}

func (r *MetaReader) next() *Meta {
	m := &Meta{}
	r.items = append(r.items, m)
	return m
}

func (r *MetaReader) pop() *Meta {
	c := r.items[len(r.items)-1]
	r.items = r.items[:len(r.items)-1]
	return c
}

func (r *MetaReader) makeCurrentExtra() {
	e := r.pop()
	cur, exists := r.current()
	if !exists {
		log.Fatalln("No prior meta tag to associate the current tag with.")
	}
	cur.Extra = append(cur.Extra, e)
}
