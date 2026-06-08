// Package wire provides compile-time dependency injection definitions.
// Run `go generate ./internal/wire` after adding providers.
// The project currently uses manual wiring in internal/bootstrap for reliability.
package wire

//go:generate go run github.com/google/wire/cmd/wire
