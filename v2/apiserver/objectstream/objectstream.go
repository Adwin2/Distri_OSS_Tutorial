package objectstream

import (
	"fmt"
	"io"
	"net/http"
)
 
type PutStream struct {
	writer *io.PipeWriter
	c      chan error
}
 
func NewPutStream(server, object string) *PutStream {
	reader, writer := io.Pipe()
	c := make(chan error)
	go func() {
		request, _ := http.NewRequest("PUT", "http://"+server+"/objects/"+object, reader)
		client := http.Client{}
		resp, err := client.Do(request)
		if err != nil && resp.StatusCode != http.StatusOK {
			err = fmt.Errorf("dataserver return http code %d", resp.StatusCode)
		}
		c <- err
	}()
	return &PutStream{writer, c}
}
 
func (w *PutStream) Write(p []byte) (int, error) {
	return w.writer.Write(p)
}
 
func (w *PutStream) Close() error {
	w.writer.Close()
	return <-w.c
}
 
type GetStream struct {
	reader io.Reader
}
 
func newGetStream(url string) (*GetStream, error) {
	resp, err := http.Get(url)
	if err != nil {
		return nil, err
	}
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("dataServer return http code %d", resp.StatusCode)
	}
	return &GetStream{resp.Body}, nil
}
 
func NewGetStream(server, object string) (*GetStream, error) {
	if server == "" || object == "" {
		return nil, fmt.Errorf("invalid server %s object %s", server, object)
	}
	return newGetStream("http://" + server + "/objects" + object)
}
 
func (r *GetStream) Read(p []byte) (int, error) {
	return r.reader.Read(p)
}