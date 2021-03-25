package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"log"
	"math/rand"
	"net/http"
	"os"
	"runtime"
	"strings"
	"time"

	"github.com/miku/esbulk"
	"github.com/sethgrid/pester"
)

// Version of application.
const Version = "0.6.2"

var (
	version         = flag.Bool("v", false, "prints current program version")
	cpuprofile      = flag.String("cpuprofile", "", "write cpu profile to file")
	memprofile      = flag.String("memprofile", "", "write heap profile to file")
	indexName       = flag.String("index", "", "index name")
	docType         = flag.String("type", "default", "elasticsearch doc type")
	batchSize       = flag.Int("size", 1000, "bulk batch size")
	numWorkers      = flag.Int("w", runtime.NumCPU(), "number of workers to use")
	verbose         = flag.Bool("verbose", false, "output basic progress")
	skipbroken      = flag.Bool("skipbroken", false, "skip broken json")
	gzipped         = flag.Bool("z", false, "unzip gz'd file on the fly")
	mapping         = flag.String("mapping", "", "mapping string or filename to apply before indexing")
	purge           = flag.Bool("purge", false, "purge any existing index before indexing")
	idfield         = flag.String("id", "", "name of field to use as id field, by default ids are autogenerated")
	user            = flag.String("u", "", "http basic auth username:password, like curl -u")
	zeroReplica     = flag.Bool("0", false, "set the number of replicas to 0 during indexing")
	refreshInterval = flag.String("r", "1s", "Refresh interval after import")
	pipeline        = flag.String("p", "", "pipeline to use to preprocess documents")
	serverFlags     esbulk.ArrayFlags
)

// IsJSON checks if a string is valid json.
func IsJSON(str string) bool {
	var js json.RawMessage
	return json.Unmarshal([]byte(str), &js) == nil
}

// indexSettingsRequest runs updates an index setting, given a body and
// options. Body consist of the JSON document, e.g. `{"index":
// {"refresh_interval": "1s"}}`.
func indexSettingsRequest(body string, options esbulk.Options) (*http.Response, error) {
	r := strings.NewReader(body)

	rand.Seed(time.Now().Unix())
	server := options.Servers[rand.Intn(len(options.Servers))]
	link := fmt.Sprintf("%s/%s/_settings", server, options.Index)

	req, err := http.NewRequest("PUT", link, r)
	if err != nil {
		return nil, err
	}
	// Auth handling.
	if options.Username != "" && options.Password != "" {
		req.SetBasicAuth(options.Username, options.Password)
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := pester.Do(req)
	if err != nil {
		return nil, err
	}
	if options.Verbose {
		log.Printf("applied setting: %s with status %s\n", body, resp.Status)
	}
	return resp, nil
}

func main() {
	flag.Var(&serverFlags, "server", "elasticsearch server, this works with https as well")
	flag.Parse()
	var (
		file               *os.File = os.Stdin
		username, password string
	)
	if flag.NArg() > 0 {
		f, err := os.Open(flag.Arg(0))
		if err != nil {
			log.Fatalln(err)
		}
		defer f.Close()
		file = f
	}
	if len(*user) > 0 {
		parts := strings.Split(*user, ":")
		if len(parts) != 2 {
			log.Fatal("http basic auth syntax is: username:password")
		}
		username = parts[0]
		password = parts[1]
	}
	runner := &esbulk.Runner{
		BatchSize:       *batchSize,
		CpuProfile:      *cpuprofile,
		DocType:         *docType,
		File:            file,
		FileGzipped:     *gzipped,
		IdentifierField: *idfield,
		IndexName:       *indexName,
		Mapping:         *mapping,
		MemProfile:      *memprofile,
		NumWorkers:      *numWorkers,
		Password:        password,
		Pipeline:        *pipeline,
		Purge:           *purge,
		RefreshInterval: *refreshInterval,
		Servers:         serverFlags,
		ShowVersion:     *version,
		SkipBroken:      *skipbroken,
		Username:        username,
		Verbose:         *verbose,
		ZeroReplica:     *zeroReplica,
	}
	if err := runner.Run(); err != nil {
		log.Fatal(err)
	}
}
