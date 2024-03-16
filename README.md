esbulk
======

Fast parallel command line [bulk loading](https://www.elastic.co/guide/en/elasticsearch/reference/current/docs-bulk.html) utility for elasticsearch. Data is read from a
[newline delimited JSON](http://jsonlines.org/) file or stdin and indexed into elasticsearch in bulk
*and* in parallel. The shortest command would be:

```shell
$ esbulk -index my-index-name < file.ldj
```

Caveat: If indexing *pressure* on the bulk API is too high (dozens or hundreds of
parallel workers, large batch sizes, depending on you setup), esbulk will halt
and report an error:

```shell
$ esbulk -index my-index-name -w 100 file.ldj
2017/01/02 16:25:25 error during bulk operation, try less workers (lower -w value) or
                    increase thread_pool.bulk.queue_size in your nodes
```

Please note that, in such a case, some documents are indexed and some are not.
Your index will be in an inconsistent state, since there is no transactional
bracket around the indexing process.

However, using defaults (parallelism: number of cores) on a single node setup
will just work. For larger clusters, increase the number of workers until you
see full CPU utilization. After that, more workers won't buy any more speed.

Currently, esbulk is [tested against](https://git.io/Jzg2u) elasticsearch
versions 2, 5, 6, 7 and 8 using
[testcontainers](https://github.com/testcontainers/testcontainers-go). Originally written for [Leipzig University
Library](https://en.wikipedia.org/wiki/Leipzig_University_Library), [project
finc](https://finc.info).

[![Project Status: Active – The project has reached a stable, usable state and is being actively developed.](https://www.repostatus.org/badges/latest/active.svg)](https://www.repostatus.org/#active)
![GitHub All Releases](https://img.shields.io/github/downloads/miku/esbulk/total.svg)

Installation
------------

    $ go install github.com/miku/esbulk/cmd/esbulk@latest

For `deb` or `rpm` packages, see: https://github.com/miku/esbulk/releases

Usage
-----

    $ esbulk -h
    Usage of esbulk:
      -0    set the number of replicas to 0 during indexing
      -c string
            create index mappings, settings, aliases, https://is.gd/3zszeu
      -cpuprofile string
            write cpu profile to file
      -id string
            name of field to use as id field, by default ids are autogenerated
      -index string
            index name
      -mapping string
            mapping string or filename to apply before indexing
      -memprofile string
            write heap profile to file
      -optype string
            optype (index - will replace existing data,
                    create - will only create a new doc,
                    update - create new or update existing data)
            (default "index")
      -p string
            pipeline to use to preprocess documents
      -purge
            purge any existing index before indexing
      -purge-pause duration
            pause after purge (default 1s)
      -r string
            Refresh interval after import (default "1s")
      -server value
            elasticsearch server, this works with https as well
      -size int
            bulk batch size (default 1000)
      -skipbroken
            skip broken json
      -type string
            elasticsearch doc type (deprecated since ES7)
      -u string
            http basic auth username:password, like curl -u
      -v    prints current program version
      -verbose
            output basic progress
      -w int
            number of workers to use (default 8)
      -z    unzip gz'd file on the fly

![](https://raw.githubusercontent.com/miku/esbulk/master/docs/asciicast.gif)

To index a JSON file, that contains one document
per line, just run:

    $ esbulk -index example file.ldj

Where `file.ldj` is line delimited JSON, like:

    {"name": "esbulk", "version": "0.2.4"}
    {"name": "estab", "version": "0.1.3"}
    ...

By default `esbulk` will use as many parallel
workers, as there are cores. To tweak the indexing
process, adjust the `-size` and `-w` parameters.

You can index from gzipped files as well, using
the `-z` flag:

    $ esbulk -z -index example file.ldj.gz

Starting with 0.3.7 the preferred method to set a
non-default server hostport is via `-server`, e.g.

    $ esbulk -server https://0.0.0.0:9201

This way, you can use https as well, which was not
possible before. Options `-host` and `-port` are
gone as of [esbulk 0.5.0](https://github.com/miku/esbulk/releases/tag/v0.5.0).

Reusing IDs
-----------

Since version 0.3.8: If you want to reuse IDs from your documents in elasticsearch, you
can specify the ID field via `-id` flag:

    $ cat file.json
    {"x": "doc-1", "db": "mysql"}
    {"x": "doc-2", "db": "mongo"}

Here, we would like to reuse the ID from field *x*.

    $ esbulk -id x -index throwaway -verbose file.json
    ...

    $ curl -s http://localhost:9200/throwaway/_search | jq
    {
      "took": 2,
      "timed_out": false,
      "_shards": {
        "total": 5,
        "successful": 5,
        "failed": 0
      },
      "hits": {
        "total": 2,
        "max_score": 1,
        "hits": [
          {
            "_index": "throwaway",
            "_type": "default",
            "_id": "doc-2",
            "_score": 1,
            "_source": {
              "x": "doc-2",
              "db": "mongo"
            }
          },
          {
            "_index": "throwaway",
            "_type": "default",
            "_id": "doc-1",
            "_score": 1,
            "_source": {
              "x": "doc-1",
              "db": "mysql"
            }
          }
        ]
      }
    }

Nested ID fields
----------------

Version 0.4.3 adds support for nested ID fields:

```
$ cat fixtures/pr-8-1.json
{"a": {"b": 1}}
{"a": {"b": 2}}
{"a": {"b": 3}}

$ esbulk -index throwaway -id a.b < fixtures/pr-8-1.json
...
```

Concatenated ID
---------------

Version 0.4.3 adds support for IDs that are the concatenation of multiple fields:

```
$ cat fixtures/pr-8-2.json
{"a": {"b": 1}, "c": "a"}
{"a": {"b": 2}, "c": "b"}
{"a": {"b": 3}, "c": "c"}

$ esbulk -index throwaway -id a.b,c < fixtures/pr-8-1.json
...

      {
        "_index": "xxx",
        "_type": "default",
        "_id": "1a",
        "_score": 1,
        "_source": {
          "a": {
            "b": 1
          },
          "c": "a"
        }
      },
```

Using X-Pack
------------

Since 0.4.2: support for secured elasticsearch nodes:

```
$ esbulk -u elastic:changeme -index myindex file.ldj
```

----

A similar project has been started for solr, called [solrbulk](https://github.com/miku/solrbulk).

Contributors
------------

* [klaubert](https://github.com/klaubert)
* [sakshambathla](https://github.com/sakshambathla)
* [mumoshu](https://github.com/mumoshu)
* [albertpastrana](https://github.com/albertpastrana)
* [faultlin3](https://github.com/faultlin3)
* [gransy](https://github.com/gransy)
* [Christoph Kepper](https://github.com/ckepper)
* Christian Solomon
* Mikael Byström

and others.

Measurements
------------

```shell
$ csvlook -I measurements.csv
| es    | esbulk | docs      | avg_b | nodes | cores | total_heap_gb | t_s   | docs_per_s | repl |
|-------|--------|-----------|-------|-------|-------|---------------|-------|------------|------|
| 6.1.2 | 0.4.8  | 138000000 | 2000  | 1     | 32    |  64           |  6420 |  22100     | 1    |
| 6.1.2 | 0.4.8  | 138000000 | 2000  | 1     |  8    |  30           | 27360 |   5100     | 1    |
| 6.1.2 | 0.4.8  |   1000000 | 2000  | 1     |  4    |   1           |   300 |   3300     | 1    |
| 6.1.2 | 0.4.8  |  10000000 |   26  | 1     |  4    |   8           |   122 |  81000     | 1    |
| 6.1.2 | 0.4.8  |  10000000 |   26  | 1     | 32    |  64           |    32 | 307000     | 1    |
| 6.2.3 | 0.4.10 | 142944530 | 2000  | 2     | 64    | 128           | 26253 |   5444     | 1    |
| 6.2.3 | 0.4.10 | 142944530 | 2000  | 2     | 64    | 128           | 11113 |  12831     | 0    |
| 6.2.3 | 0.4.13 |  15000000 | 6000  | 2     | 64    | 128           |  2460 |   6400     | 0    |
```

Why not add a [row](https://github.com/miku/esbulk/pulls)?
