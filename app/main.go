package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"strconv"
	"time"

	"go.etcd.io/etcd/clientv3"
)

var (
	dialTimeout    = 2 * time.Second
	requestTimeout = 10 * time.Second
)

type CounterStore struct {
	etcdCli clientv3.Client
}

func (cs CounterStore) get(w http.ResponseWriter, req *http.Request) {
	log.Printf("get %v", req)
	name := req.URL.Query().Get("name")
	ctx, _ := context.WithTimeout(context.Background(), requestTimeout)
	gr, err := cs.etcdCli.Get(ctx, name)
	if err != nil {
		fmt.Fprintf(w, "%s not found\n", err)
	}
	fmt.Fprintf(w, "Value: %s", gr.Kvs[0].Value)
}

func (cs CounterStore) set(w http.ResponseWriter, req *http.Request) {
	log.Printf("set %v", req)
	name := req.URL.Query().Get("name")
	val := req.URL.Query().Get("val")
	intval, err := strconv.Atoi(val)
	if err != nil {
		fmt.Fprintf(w, "%s %s \n", err, intval)
	} else {
		ctx, _ := context.WithTimeout(context.Background(), requestTimeout)
		pr, _ := cs.etcdCli.Put(ctx, name, val)
		rev := pr.Header.Revision
		fmt.Fprintf(w, "Revision: %d \n", rev)
		fmt.Fprintf(w, "ok\n")
	}
}

func (cs CounterStore) inc(w http.ResponseWriter, req *http.Request) {
	log.Printf("inc %v", req)
	name := req.URL.Query().Get("name")
	if name == "" {
		fmt.Fprintf(w, "Enter a name e.g. http://localhost:8000/inc?name=x")
		return
	}
	ctx, _ := context.WithTimeout(context.Background(), requestTimeout)
	gr, _ := cs.etcdCli.Get(ctx, name)
	if len(gr.Kvs) > 0 {
		val := string(gr.Kvs[0].Value)
		intval, err := strconv.Atoi(val)
		intval = intval + 1
		incrementedVal := strconv.Itoa(intval)
		if err != nil {
			fmt.Fprintf(w, "%s\n", err)
		} else {
			ctx, _ := context.WithTimeout(context.Background(), requestTimeout)
			pr, _ := cs.etcdCli.Put(ctx, name, incrementedVal)
			rev := pr.Header.Revision
			fmt.Fprintf(w, "Revision: %d \n", rev)
			fmt.Fprintf(w, "ok\n")
		}
	} else {
		fmt.Fprintf(w, "'%s'  not found", name)
	}
}

func main() {
	etcdCli, _ := clientv3.New(clientv3.Config{
		DialTimeout: dialTimeout,
		Endpoints:   []string{"localhost:2379"},
	})
	defer etcdCli.Close()
	store := CounterStore{etcdCli: *etcdCli}
	http.HandleFunc("/get", store.get)
	http.HandleFunc("/set", store.set)
	http.HandleFunc("/inc", store.inc)

	portnum := 8000
	if len(os.Args) > 1 {
		portnum, _ = strconv.Atoi(os.Args[1])
	}
	log.Printf("Going to listen on port %d\n", portnum)
	log.Fatal(http.ListenAndServe("localhost:"+strconv.Itoa(portnum), nil))
}
