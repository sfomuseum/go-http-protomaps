package protomaps

import (
	"fmt"
	"io"
	"log"
	"net/http"
	"path/filepath"

	"github.com/aaronland/go-http-leaflet"
	aa_static "github.com/aaronland/go-http-static"
	"github.com/sfomuseum/go-http-protomaps/static"
)

// ProtomapsOptions provides a list of JavaScript and CSS link to include with HTML output as well as a URL referencing a specific Protomaps PMTiles database to include a data attribute.
type ProtomapsOptions struct {
	// A list of relative JavaScript files to reference in one or more <script> tags
	JS []string
	// A list of relative CSS files to reference in one or more <link rel="stylesheet"> tags
	CSS []string
	// A URL for a specific PMTiles database to include as a 'data-protomaps-tile-url' attribute on the <body> tag.
	TileURL string
	// A leaflet.LeafletOptions struct
	LeafletOptions *leaflet.LeafletOptions
	// AppendJavaScriptAtEOF is a boolean flag to append JavaScript markup at the end of an HTML document
	// rather than in the <head> HTML element. Default is false
	AppendJavaScriptAtEOF bool
	RollupAssets          bool
	Prefix                string
	Logger                *log.Logger
	// By default the go-http-protomaps package will also include and reference Leaflet.js resources using the aaronland/go-http-leaflet package. If you want or need to disable this behaviour set this variable to false.
	AppendLeafletResources bool
	// By default the go-http-protomaps package will also include and reference Leaflet.js assets using the aaronland/go-http-leaflet package. If you want or need to disable this behaviour set this variable to false.
	AppendLeafletAssets bool
}

// Return a *ProtomapsOptions struct with default paths and URIs.
func DefaultProtomapsOptions() *ProtomapsOptions {

	logger := log.New(io.Discard, "", 0)

	leaflet_opts := leaflet.DefaultLeafletOptions()

	opts := &ProtomapsOptions{
		CSS: []string{},
		JS: []string{
			"/javascript/protomaps.min.js",
		},
		LeafletOptions:         leaflet_opts,
		Logger:                 logger,
		AppendLeafletResources: true,
		AppendLeafletAssets:    true,
	}

	return opts
}

// AppendResourcesHandlerWithPrefix will rewrite any HTML produced by previous handler to include the necessary markup to load Protomaps JavaScript files and related assets ensuring that all URIs are prepended with a prefix.
func AppendResourcesHandler(next http.Handler, opts *ProtomapsOptions) http.Handler {

	if opts.AppendLeafletResources {

		opts.LeafletOptions.AppendJavaScriptAtEOF = opts.AppendJavaScriptAtEOF
		opts.LeafletOptions.RollupAssets = opts.RollupAssets
		opts.LeafletOptions.Prefix = opts.Prefix
		opts.LeafletOptions.Logger = opts.Logger

		next = leaflet.AppendResourcesHandler(next, opts.LeafletOptions)
	}

	static_opts := aa_static.DefaultResourcesOptions()
	static_opts.DataAttributes["protomaps-tile-url"] = opts.TileURL
	static_opts.AppendJavaScriptAtEOF = opts.AppendJavaScriptAtEOF

	static_opts.CSS = opts.CSS
	static_opts.JS = opts.JS

	return aa_static.AppendResourcesHandlerWithPrefix(next, static_opts, opts.Prefix)
}

// Append all the files in the net/http FS instance containing the embedded Protomaps assets to an *http.ServeMux instance ensuring that all URLs are prepended with prefix.
func AppendAssetHandlers(mux *http.ServeMux, opts *ProtomapsOptions) error {

	if opts.AppendLeafletAssets {

		opts.LeafletOptions.AppendJavaScriptAtEOF = opts.AppendJavaScriptAtEOF
		opts.LeafletOptions.RollupAssets = opts.RollupAssets
		opts.LeafletOptions.Prefix = opts.Prefix
		opts.LeafletOptions.Logger = opts.Logger

		err := leaflet.AppendAssetHandlers(mux, opts.LeafletOptions)

		if err != nil {
			return fmt.Errorf("Failed to append Leaflet assets, %w", err)
		}
	}

	return aa_static.AppendStaticAssetHandlersWithPrefix(mux, static.FS, opts.Prefix)
}

// FileHandlerFromPath will take a path and create a http.FileServer handler
// instance for the files in its root directory. The handler is returned with
// a relative URI for the filename in 'path' to be assigned to a net/http
// ServeMux instance.
func FileHandlerFromPath(path string, prefix string) (string, http.Handler, error) {

	abs_path, err := filepath.Abs(path)

	if err != nil {
		return "", nil, fmt.Errorf("Failed to determine absolute path for '%s', %v", path, err)
	}

	fname := filepath.Base(abs_path)
	root := filepath.Dir(abs_path)

	tile_dir := http.Dir(root)
	tile_handler := http.FileServer(tile_dir)

	tile_url := fmt.Sprintf("/%s", fname)

	if prefix != "" {
		tile_handler = http.StripPrefix(prefix, tile_handler)
		tile_url = filepath.Join(prefix, fname)
	}

	return tile_url, tile_handler, nil
}
