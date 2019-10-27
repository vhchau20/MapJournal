package server

import (
	"encoding/json"
	"html/template"
	"io"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"path"
	"strconv"
	"strings"

	"github.com/apt4105/journal/tiles"
	"github.com/go-spatial/geom"

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
	mux.HandleFunc("/style/", StyleHandler())

	handler := LogDecorator(mux)

	server := http.Server{
		Handler: handler,
		Addr:    port,
	}

	return &server
}

func StyleHandler() http.HandlerFunc {
	// /style/:user/:view
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != "GET" {
			http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
			return
		}

		// TODO: error check users
		parts := strings.Split(strings.Trim(r.URL.Path, "/"), "/")

		switch len(parts) {
		case 0:
			panic("unreachable")

		case 1:
			http.Error(w, "must provide username", http.StatusBadRequest)
			return

		case 2: // "users" "username"
			// view the user profile
			err := tiles.WriteUserStyle(w, parts[1])
			if err != nil {
				log.Println(err)
			}

		case 3: // "users" "username" "entry"
			// view a user entry
			err := tiles.WriteEntryStyle(w, parts[1], parts[2])
			if err != nil {
				log.Println(err)
			}

		default:
			http.Error(w, "invalid users path", http.StatusBadRequest)
			return
		}
	}
}

func UserdataHandler(root string) http.HandlerFunc {
	fs := http.FileServer(http.Dir(root))

	return func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case "GET":

			p := r.URL.Path

			if p[len(p)-1] != '/' {
				if path.Ext(p) == ".md" {
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
		parts := strings.Split(strings.Trim(r.URL.Path, "/"), "/")

		switch len(parts) {
		case 0:
			panic("unreachable")

		case 1:
			http.Error(w, "must provide username", http.StatusBadRequest)
			return

		case 2: // "users" "username"
			// view the user profile
			switch r.Method {
			case "POST":
				entryName := r.FormValue("entry-name")
				if len(entryName) == 0 {
					http.Error(w, "entry name cannot be empty", http.StatusBadRequest)
					return
				}

				entryPath := path.Join(root, "userdata", parts[1], entryName)
				err := os.Mkdir(entryPath, 0755)
				if err != nil {
					log.Println(err)
					http.Error(w, "could not create directory for entry",
						http.StatusBadRequest)
					return
				}

				// now add coords to the database (maybe)
				lat := r.FormValue("lat")
				lng := r.FormValue("lng")

				if len(lat)*len(lng) == 0 && len(lat)+len(lng) == 0 {
					// both are empty
					http.Redirect(w, r, r.URL.Path, http.StatusSeeOther)
					return
				}

				if (len(lat) == 0) != (len(lng) == 0) {
					http.Error(w,
						"lat and lng must both be empty or non-empty",
						http.StatusBadRequest)
					return
				}

				latf, err := strconv.ParseFloat(lat, 64)
				if err != nil {
					http.Error(w,
						"lat "+lat+" is not a valid float",
						http.StatusBadRequest)
					return
				}

				lngf, err := strconv.ParseFloat(lng, 64)
				if err != nil {
					http.Error(w,
						"lng "+lng+" is not a valid float",
						http.StatusBadRequest)
					return
				}

				pt := geom.Point{lngf, latf}

				err = tiles.AddEntryGeotagPoint(pt,
					parts[1],
					entryName,
					"")
				if err != nil {
					log.Println(err)
					http.Error(w,
						"could not persist points in database",
						http.StatusInternalServerError)
				}

				http.Redirect(w, r, r.URL.Path, http.StatusSeeOther)
				return

			case "GET":
				http.ServeFile(w, r, path.Join(root, "views", "user.html"))
				return

			default:
				http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
				return
			}

		case 3: // "users" "username" "entry"
			// view a user entry
			switch r.Method {
			case "POST":
				file, header, err := r.FormFile("file")
				if err != nil {
					log.Println(err)
					http.Error(w, "coul not read form", http.StatusInternalServerError)
					return
				}

				fpath := path.Join(root,
					"userdata",
					parts[1],
					parts[2],
					header.Filename)

				fd, err := os.OpenFile(fpath,
					os.O_RDWR|os.O_CREATE|os.O_TRUNC,
					0644)
				if err != nil {
					log.Println(err)
					http.Error(w, "could not open file", http.StatusInternalServerError)
					return
				}
				defer fd.Close()

				_, err = io.Copy(fd, file)
				if err != nil {
					log.Println(err)
					http.Error(w, "could not write file", http.StatusInternalServerError)
					return

				}

				// now add coords to the database (maybe)
				lat := r.FormValue("lat")
				lng := r.FormValue("lng")

				if len(lat)*len(lng) == 0 && len(lat)+len(lng) == 0 {
					// both are empty
					http.Redirect(w, r, r.URL.Path, http.StatusSeeOther)
					return
				}

				if (len(lat) == 0) != (len(lng) == 0) {
					http.Error(w,
						"lat and lng must both be empty or non-empty",
						http.StatusBadRequest)
					return
				}

				latf, err := strconv.ParseFloat(lat, 64)
				if err != nil {
					http.Error(w,
						"lat "+lat+" is not a valid float",
						http.StatusBadRequest)
					return
				}

				lngf, err := strconv.ParseFloat(lng, 64)
				if err != nil {
					http.Error(w,
						"lng "+lng+" is not a valid float",
						http.StatusBadRequest)
					return
				}

				pt := geom.Point{lngf, latf}

				err = tiles.AddEntryGeotagPoint(pt,
					parts[1],
					parts[2],
					header.Filename)
				if err != nil {
					http.Error(w,
						"could not persist points in database",
						http.StatusInternalServerError)
				}

				http.Redirect(w, r, r.URL.Path, http.StatusSeeOther)
				return

			case "GET":
				http.ServeFile(w, r, path.Join(root, "views", "entry.html"))
				return

			default:
				http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
				return
			}

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
