# Metascraper

Metascraper is a web scraping utility. It transforms valid HTML markup into
a hierarchy of Go structs. In addition to capturing the raw HTML at the given
endpoint, metascraper will pull out meta tags from the page's head, and
also extracts schema.org metadata embedded in the document body.

### Usage

```go
p, err := metascraper.Scrape(url)
if err != nil {
    log.Fatal(err)
}
log.Println(p.Title)
pretty.Print(p.MetaData())
pretty.Print(p.SchemaData())
```

See [API documentation](http://godoc.org/github.com/tborg/metascraper)

Released under the [MIT License](https://github.com/tborg/metascraper/blob/master/doc.go)
