package server

import (
	"encoding/json"
	"html/template"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"path"
	"strings"

	blackfriday "gopkg.in/russross/blackfriday.v2"
)

type Server struct{}

type Decorator func(http.Handler) http.Handler

func LogDecorator(h http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		log.Printf("%s %s", r.Method, r.URL)
		h.ServeHTTP(w, r)
	})
}

func NewServer(root, port string) *http.Server {
	if !path.IsAbs(root) {
		pwd, err := os.Getwd()
		if err != nil {
			panic(err)
		}

		root = path.Join(pwd, root)
	}

	mux := http.NewServeMux()

	// mux.Handle("/", http.FileServer(http.Dir(path.Join(root, "web", "root"))))
	mux.Handle("/assets/", http.FileServer(http.Dir(root)))
	mux.HandleFunc("/userdata/", UserdataHandler(root))
	mux.HandleFunc("/users/", UserHandler(root))

	handler := LogDecorator(mux)

	server := http.Server{
		Handler: handler,
		Addr:    port,
	}

	return &server
}

func UserdataHandler(root string) http.HandlerFunc {
	fs := http.FileServer(http.Dir(root))

	return func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case "GET":

			p := r.URL.Path

			if p[len(p)-1] != '/' {
				if path.Ext(p) == "md" {
					p = path.Join(root, p)
					byt, err := ioutil.ReadFile(p)
					if os.IsExist(err) {
						http.Error(w, "404 not found", http.StatusNotFound)
						return
					}

					w.Header().Add("Content-Type", "text/html")

					out := blackfriday.Run(byt)
					w.Write(out)
					return
				}

				fs.ServeHTTP(w, r)
				return
			}

			fd, err := os.Open(path.Join(root, r.URL.Path))
			if os.IsExist(err) {
				//not exist
				log.Println(err)
				http.Error(w, "path "+r.URL.Path+" does not exist", http.StatusNotFound)
				return
			} else if err != nil {
				log.Println(err)
				http.Error(w, "could not open file", http.StatusInternalServerError)
				return
			}
			defer fd.Close()

			files, err := fd.Readdir(-1)
			if err != nil {
				http.Error(w, "could not open file", http.StatusInternalServerError)
				return
			}

			fileStrs := make([]string, len(files))
			for i, v := range files {
				fileStrs[i] = v.Name()
			}

			err = json.NewEncoder(w).Encode(fileStrs)
			if err != nil {
				log.Println(err)
			}

			return

		case "PUT":

			return

		default:
			http.Error(w, "method "+r.Method+" not allowed", http.StatusMethodNotAllowed)
			return
		}
	}
}

func UserHandler(root string) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		/*
			/users/ -> error

			/users/:username
			/users/:username/:entry
		*/

		if r.Method != "GET" {
			http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
			return
		}

		switch len(strings.Split(strings.Trim(r.URL.Path, "/"), "/")) {
		case 0:
			panic("unreachable")

		case 1:
			http.Error(w, "must provide username", http.StatusBadRequest)
			return

		case 2: // "users" "username"
			// view the user profile
			http.ServeFile(w, r, path.Join(root, "views", "user.html"))

		case 3: // "users" "username" "entry"
			// view a user entry
			http.ServeFile(w, r, path.Join(root, "views", "entry.html"))

		default:
			http.Error(w, "invalid users path", http.StatusBadRequest)
			return
		}
	}
}

const dirHtmlTmplStr = `<!DOCTYPE html>
<html>
<head>
</head>
<body>
	<ul>
		{{ range . }}
		<li><a href="{{ . }}">{{ . }}</a></li>
		{{ end }}
	</ul>
</body
</html>`

var dirHtmlTmpl = template.Must(template.New("").Parse(dirHtmlTmplStr))
