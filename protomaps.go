package protomaps

import (
	"fmt"
	"github.com/aaronland/go-http-leaflet"
	"github.com/aaronland/go-http-rewrite"
	"github.com/sfomuseum/go-http-protomaps/static"
	"io/fs"
	_ "log"
	"net/http"
	"path/filepath"
	"strings"
)

var INCLUDE_LEAFLET = true

type ProtomapsOptions struct {
	JS  []string
	CSS []string
	TileURL string
}

func DefaultProtomapsOptions() *ProtomapsOptions {

	opts := &ProtomapsOptions{
		CSS: []string{},
		JS: []string{
			"/javascript/protomaps.min.js",
			"/javascript/pmtiles.js",
		},
	}

	return opts
}

func AppendResourcesHandler(next http.Handler, opts *ProtomapsOptions) http.Handler {

	if INCLUDE_LEAFLET {
		leaflet_opts := leaflet.DefaultLeafletOptions()
		next = leaflet.AppendResourcesHandler(next, leaflet_opts)
	}

	return AppendResourcesHandlerWithPrefix(next, opts, "")
}

func AppendResourcesHandlerWithPrefix(next http.Handler, opts *ProtomapsOptions, prefix string) http.Handler {

	if INCLUDE_LEAFLET {
		leaflet_opts := leaflet.DefaultLeafletOptions()
		next = leaflet.AppendResourcesHandlerWithPrefix(next, leaflet_opts, prefix)
	}

	js := opts.JS
	css := opts.CSS

	attrs := map[string]string {
		"protomaps-tile-url": opts.TileURL,
	}
	
	if prefix != "" {

		for i, path := range js {
			js[i] = appendPrefix(prefix, path)
		}

		for i, path := range css {
			css[i] = appendPrefix(prefix, path)
		}

		for k, path := range attrs {

			if strings.HasSuffix(k, "-url") && !strings.HasPrefix(path, "http") {
				attrs[k] = appendPrefix(prefix, path)
			}
		}		
	}

	ext_opts := &rewrite.AppendResourcesOptions{
		JavaScript:  js,
		Stylesheets: css,
		DataAttributes: attrs,		
	}

	return rewrite.AppendResourcesHandler(next, ext_opts)
}

func AssetsHandler() (http.Handler, error) {

	http_fs := http.FS(static.FS)
	return http.FileServer(http_fs), nil
}

func AssetsHandlerWithPrefix(prefix string) (http.Handler, error) {

	fs_handler, err := AssetsHandler()

	if err != nil {
		return nil, err
	}

	prefix = strings.TrimRight(prefix, "/")

	if prefix == "" {
		return fs_handler, nil
	}

	rewrite_func := func(req *http.Request) (*http.Request, error) {
		req.URL.Path = strings.Replace(req.URL.Path, prefix, "", 1)
		return req, nil
	}

	rewrite_handler := rewrite.RewriteRequestHandler(fs_handler, rewrite_func)
	return rewrite_handler, nil
}

func AppendAssetHandlers(mux *http.ServeMux) error {
	return AppendAssetHandlersWithPrefix(mux, "")
}

func AppendAssetHandlersWithPrefix(mux *http.ServeMux, prefix string) error {

	if INCLUDE_LEAFLET {

		err := leaflet.AppendAssetHandlersWithPrefix(mux, prefix)

		if err != nil {
			return err
		}
	}

	asset_handler, err := AssetsHandlerWithPrefix(prefix)

	if err != nil {
		return nil
	}

	walk_func := func(path string, info fs.DirEntry, err error) error {

		if path == "." {
			return nil
		}

		if info.IsDir() {
			return nil
		}

		if prefix != "" {
			path = appendPrefix(prefix, path)
		}

		if !strings.HasPrefix(path, "/") {
			path = fmt.Sprintf("/%s", path)
		}

		// log.Println("APPEND", path)

		mux.Handle(path, asset_handler)
		return nil
	}

	return fs.WalkDir(static.FS, ".", walk_func)
}

func appendPrefix(prefix string, path string) string {

	prefix = strings.TrimRight(prefix, "/")

	if prefix != "" {
		path = strings.TrimLeft(path, "/")
		path = filepath.Join(prefix, path)
	}

	return path
}
