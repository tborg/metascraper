package metascraper

import (
	"github.com/kylelemons/godebug/pretty"
	"reflect"
	"testing"
)

// sampled from https://schema.org/docs/gs.html and http://ogp.me/
const testPage = `
    <html>
        <head>
            <title>TestPage</title>
            <meta property="og:title" content="The Rock" />
            <meta property="og:type" content="video.movie" />
            <meta property="og:url" content="http://www.imdb.com/title/tt0117500/" />
            <meta property="og:image" content="http://example.com/rock.jpg" />
            <meta property="og:image:width" content="300" />
            <meta property="og:image:height" content="300" />
            <meta property="og:image" content="http://example.com/rock2.jpg" />
            <meta property="og:image" content="http://example.com/rock3.jpg" />
            <meta property="og:image:height" content="1000" />
            <meta name="keywords" content="a,b,c" />
            <meta name="unusual">special</meta>
        </head>
        <body>
            <div itemscope itemtype="http://schema.org/Offer">
                <span itemprop="name">Blend-O-Matic</span>
                <span itemprop="price">$19.95</span>
                <div itemprop="reviews" itemscope itemtype="http://schema.org/AggregateRating">
                    <img src="four-stars.jpg" />
                    <meta itemprop="ratingValue" content="4" />
                    <meta itemprop="bestRating" content="5" />
                    Based on <span itemprop="ratingCount">25</span> user ratings
                </div>
            </div>
            <div itemscope itemtype="http://schema.org/Event">
                <div itemprop="name">Spinal Tap</div>
                <span itemprop="description">One of the loudest bands ever reunites for an unforgettable two-day show.</span>
                Event date:
                <time itemprop="startDate" datetime="2011-05-08T19:30">May 8, 7:30pm</time>
            </div>
            <div itemscope itemtype="http://schema.org/Person">
              <a href="alice.html" itemprop="url">Alice Jones</a>
            </div>
            <div itemscope itemtype="http://schema.org/Person">
              <a href="bob.html" itemprop="url">Bob Smith</a>
            </div>
        </body>
    </html>
`

var meta = []*Meta{
	&Meta{
		Property: "og:title",
		Content:  "The Rock",
	},
	&Meta{
		Property: "og:type",
		Content:  "video.movie",
	},
	&Meta{
		Property: "og:url",
		Content:  "http://www.imdb.com/title/tt0117500/",
	},
	&Meta{
		Property: "og:image",
		Content:  "http://example.com/rock.jpg",
		Extra: []*Meta{
			&Meta{
				Property: "og:image:width",
				Content:  "300",
			},
			&Meta{
				Property: "og:image:height",
				Content:  "300",
			},
		},
	},
	&Meta{
		Property: "og:image",
		Content:  "http://example.com/rock2.jpg",
	},
	&Meta{
		Property: "og:image",
		Content:  "http://example.com/rock3.jpg",
		Extra: []*Meta{
			&Meta{
				Property: "og:image:height",
				Content:  "1000",
			},
		},
	},
	&Meta{
		Name:    "keywords",
		Content: "a,b,c",
	},
	&Meta{
		Name:    "unusual",
		Content: "special",
	},
}

var schema = []*ItemScope{
	&ItemScope{
		TagName:  "div",
		ItemType: "http://schema.org/Offer",
		Props: []*ItemProp{
			&ItemProp{
				TagName:  "span",
				ItemProp: "name",
				Content:  "Blend-O-Matic",
			},
			&ItemProp{
				TagName:  "span",
				ItemProp: "price",
				Content:  "$19.95",
			},
		},
		Children: []*ItemScope{
			&ItemScope{
				TagName:  "div",
				ItemType: "http://schema.org/AggregateRating",
				ItemProp: "reviews",
				Props: []*ItemProp{
					&ItemProp{
						TagName:  "meta",
						ItemProp: "ratingValue",
						Content:  "4",
					},
					&ItemProp{
						TagName:  "meta",
						ItemProp: "bestRating",
						Content:  "5",
					},
					&ItemProp{
						TagName:  "span",
						ItemProp: "ratingCount",
						Content:  "25",
					},
				},
			},
		},
	},
	&ItemScope{
		TagName:  "div",
		ItemType: "http://schema.org/Event",
		Props: []*ItemProp{
			&ItemProp{
				TagName:  "div",
				ItemProp: "name",
				Content:  "Spinal Tap",
			},
			&ItemProp{
				TagName:  "span",
				ItemProp: "description",
				Content:  `One of the loudest bands ever reunites for an unforgettable two-day show.`,
			},
			&ItemProp{
				TagName:  "time",
				ItemProp: "startDate",
				DateTime: "2011-05-08T19:30",
				Content:  "May 8, 7:30pm",
			},
		},
	},
	&ItemScope{
		TagName:  "div",
		ItemType: "http://schema.org/Person",
		Props: []*ItemProp{
			&ItemProp{
				TagName:  "a",
				ItemProp: "url",
				HREF:     "alice.html",
				Content:  "Alice Jones",
			},
		},
	},
	&ItemScope{
		TagName:  "div",
		ItemType: "http://schema.org/Person",
		Props: []*ItemProp{
			&ItemProp{
				TagName:  "a",
				ItemProp: "url",
				HREF:     "bob.html",
				Content:  "Bob Smith",
			},
		},
	},
}

func TestPage(t *testing.T) {
	p := &Page{
		URL:          "https://www.example.com",
		HTML:         testPage,
		MetaReader:   &MetaReader{},
		SchemaReader: &SchemaReader{},
	}
	p.PageReader = &PageReader{page: p}
	mockbytes := []byte(testPage)
	p.Read(mockbytes)
	if p.Title != "TestPage" {
		t.Errorf("Expected page title %s", p.Title)
	}
	if p.Text != "Blend-O-Matic $19.95 Based on 25 user ratings Spinal Tap One of the loudest bands ever reunites for an unforgettable two-day show. Event date: May 8, 7:30pm Alice Jones Bob Smith" {
		t.Errorf("unexpected page text %s", p.Text)
	}
	for i, m := range p.MetaData() {
		if !reflect.DeepEqual(m, meta[i]) {
			t.Errorf(`Meta result item %d differs:
                expected %+v
                got %+v`, i, meta[i], m)
		}
	}
	for i, s := range p.SchemaData() {
		if !reflect.DeepEqual(s, schema[i]) {
			t.Error(pretty.Compare(s, schema[i]))
		}
	}
}
