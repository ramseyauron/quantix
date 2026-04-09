// MIT License
// Copyright (c) 2024 quantix
// test(PEPPER): Sprint 29 - server accessor methods coverage
package server

import (
	"testing"
)

func TestTCPServer_Accessor_NilServer(t *testing.T) {
	s := &Server{}
	// TCPServer() returns nil for uninitialized server — no panic
	_ = s.TCPServer()
}

func TestWSServer_Accessor_NilServer(t *testing.T) {
	s := &Server{}
	_ = s.WSServer()
}

func TestHTTPServer_Accessor_NilServer(t *testing.T) {
	s := &Server{}
	_ = s.HTTPServer()
}

func TestP2PServer_Accessor_NilServer(t *testing.T) {
	s := &Server{}
	_ = s.P2PServer()
}

func TestClose_NilComponents_NoPanic(t *testing.T) {
	s := &Server{}
	// Close on a server with nil components should not panic
	defer func() {
		if r := recover(); r != nil {
			t.Fatalf("Close panicked with nil components: %v", r)
		}
	}()
	_ = s.Close()
}

func TestStoreAndGetServer_Roundtrip(t *testing.T) {
	s := &Server{}
	StoreServer("test-sprint29", s)
	got := GetServer("test-sprint29")
	if got != s {
		t.Fatal("expected to retrieve the same server that was stored")
	}
}

func TestGetServer_Unknown_Nil(t *testing.T) {
	got := GetServer("definitely-does-not-exist-sprint29")
	if got != nil {
		t.Fatal("expected nil for unknown server name")
	}
}

func TestStoreServer_Overwrite(t *testing.T) {
	s1 := &Server{}
	s2 := &Server{}
	StoreServer("overwrite-test-sprint29", s1)
	StoreServer("overwrite-test-sprint29", s2)
	got := GetServer("overwrite-test-sprint29")
	if got != s2 {
		t.Fatal("expected overwritten server to be s2")
	}
}
