package main

import (
	"context"
	"flag"
	"fmt"
	"sync"

	"github.com/df3l0p/wmnw/src/wmn"
	"github.com/sirupsen/logrus"
)

var (
	log     = logrus.New()
	verbose = flag.Bool("verbose", false, "enable logs")
	user    = flag.String("user", "", "user to lookup")
)

func setLogLevel() {
	if *verbose {
		log.SetLevel(logrus.InfoLevel)
	} else {
		log.SetLevel(logrus.FatalLevel)
	}
}

func main() {
	flag.Parse()
	setLogLevel()
	ctx := context.Background()

	if *user == "" {
		log.Fatalf("user is empty")
	}

	sites, err := wmn.Sites()
	if err != nil {
		log.Fatalf("unable to get Sites: %v", err)
	}

	var wg sync.WaitGroup
	for _, site := range sites {
		wg.Add(1)
		go func(s wmn.Site) {
			defer wg.Done()
			exist, err := s.CheckUser(ctx, *user)
			if err != nil {
				log.Warn(err)
			}

			if exist {
				fmt.Printf("match for '%s': %s\n", s.Name, s.UrlForUser(*user))
			}
		}(site)
	}
	wg.Wait()

}
