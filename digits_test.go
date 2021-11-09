package main

import (
	"net"
	"testing"

	// boilerplate for asserting tests
	"github.com/stretchr/testify/require"
)

// TODO: non-functional
func TestAccept(t *testing.T) {
	t.Run("only accepts 5 clients", func(t *testing.T) {
		_, err := net.Dial("tcp4", "127.0.0.1:8000")
		require.NoError(t, err)
	})
}
