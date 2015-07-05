package metascraper

import (
	"io"
	"net/http"
)

// Scrape creates a new page and populates its fields from the content found at
// the given URL.
func Scrape(url string) (*Page, error) {
	p := &Page{
		URL:          url,
		MetaReader:   &MetaReader{},
		SchemaReader: &SchemaReader{},
	}
	// Unlike the other TokenReaders, the PageReader must manipulate its parent.
	p.PageReader = &PageReader{page: p}
	resp, err := http.Get(url)
	if err != nil {
		return p, err
	}
	// TODO: Can this be done with fewer conversions?
	htmlBytes := []byte{}
	if _, err = resp.Body.Read(htmlBytes); err != nil {
		return p, err
	}
	p.HTML = string(htmlBytes)
	if err = p.Read(htmlBytes); err != io.EOF {
		return p, err
	}
	return p, nil
}
