package main

import (
	"math/big"
	"testing"
)

func TestFileNames(t *testing.T) {
	tests := []struct {
		name    string
		hosts   []string
		options mkcert
		want    [3]string // certificate, private key, PKCS#12
	}{
		{
			name:  "single_dns",
			hosts: []string{"example.test"},
			want:  [3]string{"./example.test.pem", "./example.test-key.pem", "./example.test.p12"},
		},
		{
			name:  "multiple_dns",
			hosts: []string{"example.test", "www.example.test", "api.example.test"},
			want:  [3]string{"./example.test+2.pem", "./example.test+2-key.pem", "./example.test+2.p12"},
		},
		{
			name:  "mixed_sans",
			hosts: []string{"localhost", "127.0.0.1", "::1"},
			want:  [3]string{"./localhost+2.pem", "./localhost+2-key.pem", "./localhost+2.p12"},
		},
		{
			name:  "ipv4",
			hosts: []string{"192.0.2.1"},
			want:  [3]string{"./192.0.2.1.pem", "./192.0.2.1-key.pem", "./192.0.2.1.p12"},
		},
		{
			name:  "ipv6",
			hosts: []string{"2001:db8::1"},
			want:  [3]string{"./2001_db8__1.pem", "./2001_db8__1-key.pem", "./2001_db8__1.p12"},
		},
		{
			name:  "wildcard",
			hosts: []string{"*.example.test"},
			want:  [3]string{"./_wildcard.example.test.pem", "./_wildcard.example.test-key.pem", "./_wildcard.example.test.p12"},
		},
		{
			name:    "client",
			hosts:   []string{"example.test"},
			options: mkcert{client: true},
			want:    [3]string{"./example.test-client.pem", "./example.test-client-key.pem", "./example.test-client.p12"},
		},
		{
			name:    "custom_certificate",
			hosts:   []string{"example.test"},
			options: mkcert{certFile: "output/custom-cert.pem"},
			want:    [3]string{"output/custom-cert.pem", "./example.test-key.pem", "./example.test.p12"},
		},
		{
			name:    "custom_key",
			hosts:   []string{"example.test"},
			options: mkcert{keyFile: "output/custom-key.pem"},
			want:    [3]string{"./example.test.pem", "output/custom-key.pem", "./example.test.p12"},
		},
		{
			name:    "custom_pfx",
			hosts:   []string{"example.test"},
			options: mkcert{pkcs12: true, p12File: "output/custom.pfx"},
			want:    [3]string{"./example.test.pem", "./example.test-key.pem", "output/custom.pfx"},
		},
		{
			name:  "all_custom",
			hosts: []string{"example.test", "www.example.test"},
			options: mkcert{
				client: true, pkcs12: true,
				certFile: "output/custom-cert.pem",
				keyFile:  "output/custom-key.pem",
				p12File:  "output/custom.pfx",
			},
			want: [3]string{"output/custom-cert.pem", "output/custom-key.pem", "output/custom.pfx"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			m := tt.options
			cert, key, pfx := m.fileNames(tt.hosts)
			if got := [3]string{cert, key, pfx}; got != tt.want {
				t.Errorf("fileNames(%q) = %q, want %q", tt.hosts, got, tt.want)
			}
		})
	}
}

func TestRandomSerialNumber(t *testing.T) {
	tests := []struct {
		name    string
		samples int
	}{
		{name: "single_sample", samples: 1},
		{name: "repeated_samples", samples: 256},
	}
	limit := new(big.Int).Lsh(big.NewInt(1), 128)
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			for i := 0; i < tt.samples; i++ {
				serial := randomSerialNumber()
				if serial == nil {
					t.Fatalf("sample %d: randomSerialNumber() returned nil", i)
				}
				// The current implementation allows zero. Do not require uniqueness
				// or a fixed bit length from random samples.
				if serial.Sign() < 0 || serial.Cmp(limit) >= 0 {
					t.Fatalf("sample %d: serial %s is outside [0, 2^128)", i, serial)
				}
			}
		})
	}
}
