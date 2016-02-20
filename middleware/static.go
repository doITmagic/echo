package middleware

import (
	"fmt"
	"net/http"
	"os"
	"path/filepath"
	"strings"

	"github.com/webx-top/echo"
)

type (
	StaticOptions struct {
		Root   string `json:"root"`
		Index  string `json:"index"`
		Browse bool   `json:"browse"`
	}
)

func Static(root string, options ...*StaticOptions) echo.MiddlewareFunc {
	return func(next echo.Handler) echo.Handler {
		// Default options
		opts := &StaticOptions{Index: "index.html"}
		if len(options) > 0 {
			opts = options[0]
		}

		opts.Root, _ = filepath.Abs(opts.Root)

		return echo.HandlerFunc(func(c echo.Context) error {
			file := c.Request().URI()
			absFile := filepath.Join(opts.Root, file)
			if !strings.HasPrefix(absFile, opts.Root) {
				return next.Handle(c)
			}
			fi, err := os.Stat(absFile)
			if err != nil {
				return next.Handle(c)
			}
			w := c.Response()
			if fi.IsDir() {
				// Index file
				indexFile := filepath.Join(file, opts.Index)
				fi, err = os.Stat(indexFile)
				if err != nil || fi.IsDir() {
					if opts.Browse {
						fs := http.Dir(opts.Root)
						d, err := fs.Open(file)
						if err != nil {
							return echo.ErrNotFound
						}
						defer d.Close()
						dirs, err := d.Readdir(-1)
						if err != nil {
							return err
						}

						// Create a directory index
						w.Header().Set(echo.ContentType, echo.TextHTMLCharsetUTF8)
						if _, err = fmt.Fprintf(w, "<pre>\n"); err != nil {
							return err
						}
						for _, d := range dirs {
							name := d.Name()
							color := "#212121"
							if d.IsDir() {
								color = "#e91e63"
								name += "/"
							}
							if _, err = fmt.Fprintf(w, "<a href=\"%s\" style=\"color: %s;\">%s</a>\n", name, color, name); err != nil {
								return err
							}
						}
						_, err = fmt.Fprintf(w, "</pre>\n")
						return err
					}
					return next.Handle(c)
				} else {
					absFile = indexFile
				}
			}
			w.ServeFile(absFile)
			return nil
		})
	}
}

// Favicon serves the default favicon - GET /favicon.ico.
func Favicon() echo.HandlerFunc {
	return func(c echo.Context) error {
		return nil
	}
}