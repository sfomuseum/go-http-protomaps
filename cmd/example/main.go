package main

import (
	"context"
	"embed"
	"flag"
	"github.com/aaronland/go-http-leaflet"
	"github.com/aaronland/go-http-server"
	"github.com/sfomuseum/go-http-protomaps"
	"log"
	"net/http"
	"net/url"
)

//go:embed index.html sfo.pmtiles
var staticFS embed.FS

func main() {

	server_uri := flag.String("server-uri", "http://localhost:8080", "A valid aaronland/go-http-server URI")

	tile_url := flag.String("protomaps-tile-url", "/sfo.pmtiles", "A custom file://, http:// or https:// URI pointing to a valid Protomaps tiles bundle.")

	append_leaflet := flag.Bool("append-leaflet", true, "Append Leaflet.js assets and resources bundled with the go-http-protomaps package.")

	flag.Parse()

	ctx := context.Background()

	static_fs := http.FS(staticFS)
	static_handler := http.FileServer(static_fs)

	index_handler := static_handler

	mux := http.NewServeMux()

	if !*append_leaflet {

		protomaps.APPEND_LEAFLET_RESOURCES = false
		protomaps.APPEND_LEAFLET_ASSETS = false

		leaflet_opts := leaflet.DefaultLeafletOptions()
		leaflet_opts.EnableHash()

		index_handler = leaflet.AppendResourcesHandler(index_handler, leaflet_opts)

		err := leaflet.AppendAssetHandlers(mux)

		if err != nil {
			log.Fatalf("Failed to append Leaflet asset handlers, %v", err)
		}
	}

	err := protomaps.AppendAssetHandlers(mux)

	if err != nil {
		log.Fatalf("Failed to append leaflet-protomaps asset handler, %v", err)
	}

	u, err := url.Parse(*tile_url)

	if err != nil {
		log.Fatalf("Failed to parse '%s', %v", tile_url, err)
	}

	switch u.Scheme {
	case "":
		// sfo.pmtiles bundled with this example
		mux.Handle(*tile_url, static_handler)
	case "file":

		// file:// URI for custom PM tiles bunle
		mux_url, mux_handler, err := protomaps.FileHandlerFromPath(u.Path, "")

		if err != nil {
			log.Fatalf("Failed to determine absolute path for '%s', %v", *tile_url, err)
		}

		mux.Handle(mux_url, mux_handler)

		*tile_url = mux_url

	case "http", "https":
		// pass - assumed to be remote HTTP/HTTPS PM tiles bundle
	default:
		log.Fatalf("Invalid or unsupported scheme")
	}

	pm_opts := protomaps.DefaultProtomapsOptions()
	pm_opts.TileURL = *tile_url

	pm_opts.LeafletOptions.EnableHash()

	index_handler = protomaps.AppendResourcesHandler(index_handler, pm_opts)
	mux.Handle("/", index_handler)

	s, err := server.NewServer(ctx, *server_uri)

	if err != nil {
		log.Fatalf("Failed to start server '%s', %v", *server_uri, err)
	}

	log.Printf("Listening for requests on %s\n", s.Address())

	err = s.ListenAndServe(ctx, mux)

	if err != nil {
		log.Fatalf("Failed to start server, %v", err)
	}

}
