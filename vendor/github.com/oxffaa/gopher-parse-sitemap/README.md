# gopher-parse-sitemap

[![Build Status](https://travis-ci.org/oxffaa/gopher-parse-sitemap.svg?branch=master)](https://travis-ci.org/oxffaa/gopher-parse-sitemap)

A high effective golang library for parsing big-sized sitemaps and avoiding high memory usage. The sitemap parser was written on golang without external dependencies. See https://www.sitemaps.org/ for more information about the sitemap format.

## Why yet another sitemaps parsing library?

Time by time needs to parse really huge sitemaps. If you just unmarshal the whole file to an array of structures it produces high memory usage and the application can crash due to OOM (out of memory error). 


The solution is to handle sitemap entries on the fly. That is read one entity, consume it, repeat while there are unhandled items in the sitemap.

```golang
err := sitemap.ParseFromFile("./testdata/sitemap.xml", func(e Entry) error {
    return fmt.Println(e.GetLocation())
})
```

### I need parse only small and medium-sized sitemaps. Should I use this library?

Yes. Of course, you can just load a sitemap to memory.

```golang
result := make([]string, 0, 0)
err := sitemap.ParseIndexFromFile("./testdata/sitemap-index.xml", func(e IndexEntry) error {
    result = append(result, e.GetLocation())
    return nil
})
```

But if you are pretty sure that you don't need to handle big-sized sitemaps, may be better to choose a library with simpler and more suitable API. In that case, you can try projects like https://github.com/yterajima/go-sitemap, https://github.com/snabb/sitemap, and https://github.com/decaseal/go-sitemap-parser.

## Install

Installation is pretty easy, just do:

```bash
go get -u github.com/oxffaa/gopher-parse-sitemap
```

After that import it:
```golang
import "github.com/oxffaa/gopher-parse-sitemap"
```

Well done, you can start to create something awesome.

## Documentation

Please, see [here](https://godoc.org/github.com/oxffaa/gopher-parse-sitemap) for documentation.
