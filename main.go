package main

import (
	"bufio"
	"flag"
	"io"
	"log"
	"os"
	"strings"
	"time"

	"github.com/gomodule/redigo/redis"
)

func exec(pool *redis.Pool, path string) error {
	f, err := os.Open(path)
	if err != nil {
		return err
	}
	defer f.Close()

	c := pool.Get()
	defer c.Close()

	r := bufio.NewReader(f)
	for {
		line, _, err := r.ReadLine()
		if err != nil {
			if err == io.EOF {
				return nil
			}
			return err
		}

		str := strings.TrimSpace(string(line))
		if str == "" || strings.HasPrefix(str, "#") {
			continue
		}

		fileds := strings.Fields(str)
		args := make([]interface{}, len(fileds)-1)
		for i, v := range fileds[1:] {
			args[i] = v
		}
		_, err = c.Do(fileds[0], args...)
		if err != nil {
			return err
		}
	}
}

var (
	host     = flag.String("h", "127.0.0.1", "Server hostname")
	port     = flag.String("p", "6379", "Server port")
	password = flag.String("a", "", "Password to use when connecting to the server")
)

func main() {
	flag.Parse()

	pool := redis.NewPool(func() (redis.Conn, error) {
		c, err := redis.Dial("tcp", *host+":"+*port)
		if err != nil {
			return nil, err
		}
		if len(*password) > 0 {
			_, err = c.Do("AUTH", *password)
			if err != nil {
				c.Close()
				return nil, err
			}
		}
		return c, nil
	}, 1)
	defer pool.Close()

	for i := 0; i < flag.NArg(); i++ {
		now := time.Now()
		err := exec(pool, flag.Arg(i))
		if err != nil {
			log.Fatal(err)
		}
		log.Printf("%s %s\n", flag.Arg(i), time.Since(now))
	}
}
