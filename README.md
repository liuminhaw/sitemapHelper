# sitemap helper

Some sitemap helper functions

- Generator

## Generator

Crawl from a given URL and get links within the same domain as content for
sitemap generation

### Example

See usage example at [example generator](examples/generator)

#### Example build

```text
cd examples/generator
go build
```

#### Example run

```text
Usage: ./generator [options] <url>
  -depth int
    the maximum number of links deep to traverse (default 3)
  -outputDir string
    directory where result sitemap will be stored (default "sitemaps")
  -render
    indicate if the site need to be rendered to get actual links, useful for SPA site (default false)
```
