package main

import (
	"flag"
	"log"
	"net/http"
	"path"

	"fmt"

	"github.com/hiwjd/axol"
)

type Conf struct {
	Port    int
	DataDir string
}

func main() {
	conf := Conf{}
	flag.IntVar(&conf.Port, "port", 8888, "server port")
	flag.StringVar(&conf.DataDir, "dataDir", "./data", "data dir, include user info and proj data")
	flag.Parse()

	ss, err := axol.NewStoreService(path.Join(conf.DataDir, "db"))
	if err != nil {
		log.Fatal(err)
	}

	addr := fmt.Sprintf(":%d", conf.Port)
	hh := axol.NewHttpHandler(ss, path.Join(conf.DataDir, "proj"))
	log.Fatal(http.ListenAndServe(addr, hh))
}
