package metascraper

import (
	"golang.org/x/net/html"
	"log"
)

// ItemScope represents a schema.org itemscope.
// see http://schema.org/docs/gs.html for more on the vocabulary.
// An ItemScope may be a complex property of another ItemScope.
type ItemScope struct {
	TagName  string // The tag name of the outer element of the itemscope.
	ItemType string // The itemtype attribute.
	ItemProp string // If this itemscope is nested, this is the name of the complex
	// property belonging to the parent itemscope.
	Props    []*ItemProp  // Collects simple itemprops belonging to this scope.
	Children []*ItemScope // Collects complex itemprops belonging to this scope.
}

// ItemProp represents a simple schema.org itemprop.
type ItemProp struct {
	TagName  string // The tag name of this itemprop.
	ItemProp string // The itemprop attribute.
	Content  string // The text content or content attribute.
	HREF     string // The href attribute if any.
	DateTime string // The datetime attribute if any.
}

// SchemaReader implements the TokenReader interface; it maintains the necessary
// state for extracting schema.org metadata from the body of an HTML document.
type SchemaReader struct {
	items       []*ItemScope // Top-level ItemScopes.
	stack       []*ItemScope // The current hierarchy of ItemScopes.
	breadcrumbs []bool       // Markers indicating whether the current element is an ItemScope.
	insideProp  bool
	inHead      bool
}

func (r *SchemaReader) current() (*ItemScope, bool) {
	s := &ItemScope{}
	n := len(r.stack)
	if n > 0 {
		return r.stack[len(r.stack)-1], true
	} else {
		return s, false
	}
}

func (r *SchemaReader) next() *ItemScope {
	s := &ItemScope{}
	if len(r.stack) > 0 {
		cur := r.stack[len(r.stack)-1]
		cur.Children = append(cur.Children, s)
	} else {
		r.items = append(r.items, s)
	}
	r.stack = append(r.stack, s)
	return s
}

func (r *SchemaReader) pop() *ItemScope {
	s, _ := r.current()
	r.stack = r.stack[:len(r.stack)-1]
	return s
}

func (r *SchemaReader) HandleStart(tn string, attrs map[string]string, z *html.Tokenizer) {
	if tn == "head" {
		r.inHead = true
	}
	if r.inHead || len(attrs) == 0 {
		return
	}
	s, exists := r.current()
	// To maintain the state of the stack, we also maintain a stack of markers
	// for each element encountered, which represents whether that element is an
	// itemscope.
	// A marker is pushed for every start element, and popped for every end element.
	// The itemscope stack is only pushed and popped when the marker is true.
	shouldPopOnEnd := false
	if _, ok := attrs["itemscope"]; ok {
		shouldPopOnEnd = true
		s = r.next()
		s.ItemProp = attrs["itemprop"]
		s.TagName = tn
		s.ItemType = attrs["itemtype"]
	} else if exists {
		itemprop, ok := attrs["itemprop"]
		if ok {
			s.Props = append(s.Props, &ItemProp{
				TagName:  tn,
				ItemProp: itemprop,
				HREF:     attrs["href"],
				Content:  attrs["content"],
				DateTime: attrs["datetime"],
			})
			r.insideProp = true
		}
	}
	r.breadcrumbs = append(r.breadcrumbs, shouldPopOnEnd)
}

func (r *SchemaReader) HandleEnd(tn string, z *html.Tokenizer) {
	if r.inHead {
		if tn == "head" {
			r.inHead = false
		}
		return
	}
	r.insideProp = false
	depth := len(r.breadcrumbs)
	if depth > 0 {
		shouldPop := r.breadcrumbs[depth-1]
		r.breadcrumbs = r.breadcrumbs[:depth-1]
		if shouldPop {
			r.pop()
		}
	}
}

func (r *SchemaReader) HandleText(text []byte) {
	if r.insideProp {
		s, exists := r.current()
		if !exists || len(s.Props) == 0 {
			log.Fatalln("No prop to set content from text node")
		}
		s.Props[len(s.Props)-1].Content = string(text)
	}
}

func (r *SchemaReader) Done() {
	// No cleanup.
}
