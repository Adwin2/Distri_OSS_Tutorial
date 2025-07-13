package objects

import (
	"io"
	"log"
	"net/http"
	"os"
	"path/filepath"
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
	fn := os.Getenv("STORAGE_ROOT") + "/objects/" + strings.Split(r.URL.EscapedPath(), "/")[2]
	
	// tip: `os.Create()`不可自动创建中间目录
	dir := filepath.Dir(fn)
	err := os.MkdirAll(dir, 0755)
	if err != nil {
		log.Println("目录创建失败", dir)
		w.WriteHeader(http.StatusInternalServerError)
		return err
	}

	// `strings.Split(r.URL.EscapedPath(), "/")[2]` : URL 路径部分的第二级参数
	f, err := os.Create(fn)
	if err != nil {
		log.Println("文件创建失败", strings.Split(r.URL.EscapedPath(), "/")[2])
		w.WriteHeader(http.StatusInternalServerError)
		return err
	}
	defer f.Close()
	// 覆盖写
	_, err = io.Copy(f, r.Body)
	if err != nil {
		log.Println("文件写入失败")
		w.WriteHeader(http.StatusInternalServerError)
		return err
	}

	log.Printf("文件写入成功: %s", fn)
	return err
}

func get(w http.ResponseWriter, r *http.Request) error {
	f, err := os.Open(os.Getenv("STORAGE_ROOT") + "/objects/" + strings.Split(r.URL.EscapedPath(), "/")[2])
	if err != nil {
		log.Println("文件打开失败")
		w.WriteHeader(http.StatusNotFound)
		return err
	}
	defer f.Close()
	_, err = io.Copy(w, f)
	if err != nil {
		log.Println("文件读取失败")
		w.WriteHeader(http.StatusInternalServerError)
		return err
	}

	log.Printf("文件读取成功: %s", f.Name())
	return err
}