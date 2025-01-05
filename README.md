# sitemap helper

Some sitemap helper functions

- Crawler
- Parser

## Install
```bash
go get github.com/liuminhawl/sitemapHelper
```

## Crawler

Crawl the given URL and retrieve links within the same domain to generate
sitemap content.

### Example

See usage example at [example crawler](examples/crawler)

#### Example build

```bash
cd examples/crawler
go build
```

#### Example run

```text
Usage: ./crawler [options] <url>
  -depth int
    the maximum number of links deep to traverse (default 3)
  -outputDir string
    directory where result sitemap will be stored (default "sitemaps")
  -render
    indicate if the site need to be rendered to get actual links, useful for SPA site (default false)
```

## Parser

Parse sitemap content and generate a slice of sitemap entries, ready for
conversion to XML.

### Example

See usage example at [example parser](examples/parser)

- [fileParser](examples/parser/fileParser)
- [urlParser](examples/parser/urlParser)

### Example build

```bash
cd examples/parser/fileParser
# OR
cd examples/parser/urlParser

go build
```

### Example run

File parser
```text
Usage: ./fileParser <filepath>
```

Url parser
```text
Usage: ./urlParser <url>
```
