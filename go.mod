module github.com/archekb/gipam

go 1.15

// replace github.com/archekb/gipam/pkg/leaser => ./pkg/leaser
// replace github.com/archekb/gipam/pkg/gipam => ./pkg/gipam
// replace github.com/archekb/gipam/pkg/config => ./pkg/config

require (
	github.com/Microsoft/go-winio v0.4.14 // indirect
	github.com/coreos/go-systemd v0.0.0-20191104093116-d3cd4ed1dbcf // indirect
	github.com/docker/go-connections v0.4.0 // indirect
	github.com/docker/go-plugins-helpers v0.0.0-20200102110956-c9a8a2d92ccc
	github.com/dspinhirne/netaddr-go v0.0.0-20200114144454-1f4c8303963f
	github.com/stretchr/testify v1.2.2
	golang.org/x/net v0.0.0-20201024042810-be3efd7ff127 // indirect
	golang.org/x/sys v0.0.0-20201024232916-9f70ab9862d5 // indirect
)
