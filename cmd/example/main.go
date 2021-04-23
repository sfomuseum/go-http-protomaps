package main

import (
	"context"
	"embed"
	"flag"
	"github.com/aaronland/go-http-server"
	"github.com/sfomuseum/go-http-leaflet-protomaps"
	"log"
	"net/http"
)

//go:embed index.html sfo.pmtiles
var staticFS embed.FS

func main() {

	server_uri := flag.String("server-uri", "http://localhost:8080", "A valid aaronland/go-http-server URI")

	flag.Parse()

	ctx := context.Background()

	static_fs := http.FS(staticFS)
	static_handler := http.FileServer(static_fs)
	
	mux := http.NewServeMux()

	err := protomaps.AppendAssetHandlers(mux)

	if err != nil {
		log.Fatalf("Failed to append leaflet-protomaps asset handler, %v", err)
	}

	protomaps_opts := protomaps.DefaultLeafletProtomapsOptions()
	
	index_handler := protomaps.AppendResourcesHandler(static_handler, protomaps_opts)
	mux.Handle("/", index_handler)
	
	mux.Handle("/sfo.pmtiles", static_handler)
	mux.Handle("/sfo.json", static_handler)		
	
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
