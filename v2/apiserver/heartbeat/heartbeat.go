package heartbeat

import (
	"dot/v2/rabbitmq"
	"math/rand"
	"os"
	"strconv"
	"sync"
	"time"
)
 
var dataServers = make(map[string]time.Time)
var mutex sync.Mutex
 
func ListenHeartbeat() {
	server := os.Getenv("RABBITMQ_SERVER")
	q := rabbitmq.New(server)
	defer q.Close()
	q.Bind("apiServers")
	c := q.Consume()
	go removeExpiredDataServer()
	for msg := range c {
		dataServer, err := strconv.Unquote(string(msg.Body))
		if err != nil {
			panic(err)
		}
		mutex.Lock()
		dataServers[dataServer] = time.Now()
		mutex.Unlock()
	}
}
 
func removeExpiredDataServer() {
	for {
		time.Sleep(5 * time.Second)
		mutex.Lock()
		for s, t := range dataServers {
			if t.Add(10 * time.Second).Before(time.Now()) {
				delete(dataServers, s)
			}
		}
		mutex.Unlock()
	}
}
 
func GetDataServers() []string {
	mutex.Lock()
	defer mutex.Unlock()
	ds := make([]string, 0)
	for s := range dataServers {
		ds = append(ds, s)
	}
	return ds
}
 
func ChooseRandomDataServer() string {
	ds := GetDataServers()
	n := len(ds)
	if n == 0 {
		return ""
	}
	return ds[rand.Intn(n)]
}