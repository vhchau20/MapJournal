package server

type Server struct {}

func NewServer(root string) *http.Server {
	mux := http.NewServeMux()

	mux.Handle("/assets", http.FileServer(root))
	mux.Handle("/api", ApiHandler{})
	mux.HandleFunc("/~", UserHandler(root))
	mux.Handle("/", http.FileServer(path.Join(root, "web")))

	server := http.Server{
		Handler: mux,
	}

	return &server
}

func UserHandler(root string) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		/*
		/~user_name
		/~user_name/file
		*/

		arr := strings.Split(r.URL.Path, "/")
		if len(arr) < 2 {
			http.Error(w, "invalid username", http.StatusBadRequest)
			return
		}

		user := strings.TrimSuffix(arr[1], "/~")
		userDir := path.Join(root, "userdata", user)

		if _, err := os.Stat(userDir); os.IsExist(err) {
			//not exist
			http.Error(w, "user " + user + " does not exist", http.StatusNotFound)
			return
		}



		htmlFile := path.Join(root, "web", "user", path.Join(arr[3:]...))
	}
}


}
