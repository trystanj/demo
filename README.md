# Demo
Small playground to play with querying ES from Go and interfaces.

## Requirements
- Go (if you want to build)
- ElasticSearch
  - `docker run -p 9200:9200 -p 9300:9300 -e "discovery.type=single-node" docker.elastic.co/elasticsearch/elasticsearch:6.2.1`
  - `brew install elasticsearch && elasticsearch` (make sure you're running v6 on port 9200)
- `./demo` if on OS X (alternatively, build and run)
- Go to `http://localhost:8080/search`
  - query params
    - category [restaurant, bar, clerb]
    - token (to get the next page of results, pass the token from the previous request)
