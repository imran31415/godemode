module github.com/imran31415/godemode/mcp-benchmark

go 1.24.9

toolchain go1.24.10

replace github.com/imran31415/godemode => ../

replace github.com/imran31415/godemode/mcp-benchmark/godemode => ./godemode

replace github.com/imran31415/godemode/mcp-benchmark/filesystem => ./filesystem

require (
	github.com/imran31415/godemode v0.0.0-00010101000000-000000000000
	github.com/imran31415/godemode/mcp-benchmark/filesystem v0.0.0-00010101000000-000000000000
	github.com/imran31415/godemode/mcp-benchmark/godemode v0.0.0-00010101000000-000000000000
)

require (
	github.com/google/uuid v1.6.0 // indirect
	github.com/tetratelabs/wazero v1.9.0 // indirect
	github.com/traefik/yaegi v0.16.1 // indirect
)
