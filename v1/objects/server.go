package objects

import (
	"io"
	"log"
	"net/http"
	"os"
	"strings"
)

func Handler(w http.ResponseWriter, r *http.Request) {
	method := r.Method
	if method == http.MethodPut {
		if err := put(w, r); err != nil {
			log.Print(err)
		}
	}

	if method == http.MethodGet {
		if err := get(w, r); err != nil {
					log.Print(err)
		}
	}
}

func put(w http.ResponseWriter, r *http.Request) error {
	// `strings.Split(r.URL.EscapedPath(), "/")[2]` : URL 路径部分的第二级参数
	f, err := os.Create(os.Getenv("STORAGE_ROOT") + "/objects/" + strings.Split(r.URL.EscapedPath(), "/")[2])
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return err
	}
	defer f.Close()
	_, err = io.Copy(f, r.Body)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return err
	}
	return err
}

func get(w http.ResponseWriter, r *http.Request) error {
	f, err := os.Open(os.Getenv("STORAGE_ROOT") + "/objects/" + strings.Split(r.URL.EscapedPath(), "/")[2])
	if err != nil {
		w.WriteHeader(http.StatusNotFound)
		return err
	}
	defer f.Close()
	_, err = io.Copy(w, f)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return err
	}
	return err
}