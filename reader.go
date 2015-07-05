package metascraper

import (
	"golang.org/x/net/html"
)

// TokenReader presents a lightweight version of the usual SAX parser interface,
// with methods for handling the typical events in a token stream.
type TokenReader interface {
	// HandleStart is called when a start or self closing tag token is encountered.
	// It will always be called before HandleEnd for the same tag.
	HandleStart(tagName string, attrs map[string]string, z *html.Tokenizer)

	// HandleEnd is called when an end or self closing tag token is encountered.
	// It will always be called after HandleStart for the same tag.
	HandleEnd(tagName string, z *html.Tokenizer)

	// HandleText is called when a text node is encountered. The contents of the
	// byte slice it is passed could change over time, so you'll want to copy
	// it if you want to hang on to it.
	HandleText(text []byte)

	// Done is called when the token stream ends; it gives your TokenReaders a
	// chance to clean up.
	Done()
}

// ReaderList implements the TokenReader interface over a slice of TokenReaders.
// Calling a method on the ReaderList calls that method on each TokenReader in
// the item list in turn, forwarding the arguments.
type ReaderList struct {
	Readers []TokenReader
}

func (t ReaderList) HandleStart(tn string, attrs map[string]string, z *html.Tokenizer) {
	for _, r := range t.Readers {
		r.HandleStart(tn, attrs, z)
	}
}

func (t ReaderList) HandleEnd(tn string, z *html.Tokenizer) {
	for _, r := range t.Readers {
		r.HandleEnd(tn, z)
	}
}

func (t ReaderList) HandleText(text []byte) {
	for _, r := range t.Readers {
		r.HandleText(text)
	}
}

func (t ReaderList) Done() {
	for _, r := range t.Readers {
		r.Done()
	}
}
