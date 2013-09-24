set -e
set -v

CGO_ENABLED=0 go test -purego true
CGO_ENABLED=0 go test -purego true -tags purego 
CGO_ENABLED=1 go test -purego false
CGO_ENABLED=1 go test -purego true -tags purego 
