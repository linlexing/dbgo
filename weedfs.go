package main

import (
	"io/ioutil"
	"mime"
	"net/http"
	"path"
)

type WeedFS struct {
	hostUrl string
}

func NewWeedFS(url string) *WeedFS {
	return &WeedFS{url}
}
func (f *WeedFS) ProxyFile(filename string, req *http.Request, w http.ResponseWriter) error {
	newReq, err := http.NewRequest(req.Method, f.hostUrl+"/"+filename, nil)
	if err != nil {
		return err
	}
	for k, v := range req.Header {
		for _, str := range v {
			newReq.Header.Add(k, str)
		}
	}
	resp, err := http.DefaultClient.Do(newReq)
	if err != nil {
		return err
	}
	ext := path.Ext(filename)
	if ext != ".gz" && resp.Header.Get("Content-Type") == "application/x-gzip" {
		resp.Header["Content-Type"] = []string{mime.TypeByExtension(ext)}
		resp.Header["content-encoding"] = []string{"gzip"}

	}
	defer resp.Body.Close()
	for k, v := range resp.Header {
		for _, vv := range v {
			w.Header().Add(k, vv)
		}
	}
	w.WriteHeader(resp.StatusCode)
	result, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return err
	}
	w.Write(result)
	return nil
}
