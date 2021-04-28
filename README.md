# go-http-protomaps

![](docs/images/go-http-leaflet-protomaps-sfo.png)

`go-http-protomaps` is an HTTP middleware package for including Protomaps.js assets in web applications.

## Documentation

[![Go Reference](https://pkg.go.dev/badge/github.com/sfomuseum/go-http-protomaps.svg)](https://pkg.go.dev/github.com/sfomuseum/go-http-protomaps)

`go-http-protomaps` is an HTTP middleware package for including Protomaps.js assets in web applications. It exports two principal methods: 

* `protomaps.AppendAssetHandlers(*http.ServeMux)` which is used to append HTTP handlers to a `http.ServeMux` instance for serving Protomaps JavaScript files, and related assets.
* `protomaps.AppendResourcesHandler(http.Handler, *ProtomapsOptions)` which is used to rewrite any HTML produced by previous handler to include the necessary markup to load Protomaps JavaScript files and related assets.

## Example

```
package main

import (
	"embed"
	"github.com/sfomuseum/go-http-protomaps"
	"log"
	"net/http"
	"net/url"
)

//go:embed index.html sfo.pmtiles
var staticFS embed.FS

func main() {

	tile_url := "/sfo.pmtiles"

	static_fs := http.FS(staticFS)
	static_handler := http.FileServer(static_fs)

	mux := http.NewServeMux()

	mux.Handle(*tile_url, static_handler)

	protomaps.AppendAssetHandlers(mux)
	
	pm_opts := protomaps.DefaultProtomapsOptions()
	pm_opts.TileURL = *tile_url

	index_handler := protomaps.AppendResourcesHandler(static_handler, pm_opts)
	mux.Handle("/", index_handler)

	err = http.ListenAndServe(":8080", mux)
}
```

_Error handling omitted for brevity._

## Tools

```
$> make cli
go build -mod vendor -o bin/example cmd/example/main.go
```

### example

A simple example web application that uses the `go-http-protomaps` middleware and displays a map.

```
$> ./bin/example -h
Usage of ./bin/example:
  -protomaps-tile-url string
    	A custom file://, http:// or https:// URI pointing to a valid Protomaps tiles bundle. (default "/sfo.pmtiles")
  -server-uri string
    	A valid aaronland/go-http-server URI (default "http://localhost:8080")
```

The `example` application contains an embedded Protomaps tiles database file so the easiest way to test things is like this:

```
$> ./bin/example 
2021/04/28 07:45:57 Listening for requests on http://localhost:8080
```

And then when you open the URL `http://localhost:8080` in your web browser you should see something like this:

![](docs/images/go-http-leaflet-protomaps-sfo.png)

The `example` application also supports loading `.pmtiles` databases from remote locations. For example:

```
$> ./bin/example -protomaps-tile-url https://static.sfomuseum.org/pmtiles/sfo.pmtiles
2021/04/28 07:48:11 Listening for requests on http://localhost:8080
```

If you want to load a local file you'd do this:

```
$> ./bin/example -protomaps-tile-url file:///usr/local/data/sfo.pmtiles
2021/04/28 07:48:53 Listening for requests on http://localhost:8080
```

In the case of `file://` URLs the `example` application will create an `http.Dir` handler for the root folder of the URL (`/usr/local/data/)` and then route the filename (`/sfo.pmtiles`) to that handler.

## See also

* https://protomaps.com/blog/new-way-to-make-maps/
* https://github.com/protomaps/protomaps.js
* https://github.com/aaronland/go-http-leaflet