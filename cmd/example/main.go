package main

import (
	"context"
	"flag"
	"fmt"
	"github.com/aaronland/go-http-server"
	"github.com/sfomuseum/go-http-leaflet-geotag"
	"github.com/sfomuseum/go-http-leaflet-geotag/templates/html"
	"html/template"
	"log"
	"net/http"
)

func PageHandler(templates *template.Template, t_name string) (http.Handler, error) {

	t := templates.Lookup(t_name)

	if t == nil {
		return nil, fmt.Errorf("Missing '%s' template", t_name)
	}

	fn := func(rsp http.ResponseWriter, req *http.Request) {

		rsp.Header().Set("Content-type", "text/html")

		err := t.Execute(rsp, nil)

		if err != nil {
			http.Error(rsp, err.Error(), http.StatusInternalServerError)
		}

		return
	}

	return http.HandlerFunc(fn), nil
}

func main() {

	server_uri := flag.String("server-uri", "http://localhost:8080", "A valid aaronland/go-http-server URI")

	flag.Parse()

	ctx := context.Background()

	t := template.New("example")

	t, err := t.ParseFS(html.FS, "*.html")

	geotag_opts := geotag.DefaultLeafletGeotagOptions()

	mux := http.NewServeMux()

	err = geotag.AppendAssetHandlers(mux)

	if err != nil {
		log.Fatalf("Failed to append leaflet-geotag asset handler, %v", err)
	}

	camera_handler, err := PageHandler(t, "camera")

	if err != nil {
		log.Fatalf("Failed to create camera handler, %v", err)
	}

	camera_handler = geotag.AppendResourcesHandler(camera_handler, geotag_opts)

	mux.Handle("/camera/", camera_handler)

	crosshair_handler, err := PageHandler(t, "crosshair")

	if err != nil {
		log.Fatalf("Failed to create crosshair handler, %v", err)
	}

	crosshair_handler = geotag.AppendResourcesHandler(crosshair_handler, geotag_opts)

	mux.Handle("/crosshair/", crosshair_handler)

	index_handler, err := PageHandler(t, "index")

	if err != nil {
		log.Fatalf("Failed to create index handler, %v", err)
	}

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
