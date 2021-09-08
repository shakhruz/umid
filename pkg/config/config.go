// Copyright (c) 2021 UMI
//
// Permission is hereby granted, free of charge, to any person obtaining a copy
// of this software and associated documentation files (the "Software"), to deal
// in the Software without restriction, including without limitation the rights
// to use, copy, modify, merge, publish, distribute, sublicense, and/or sell
// copies of the Software, and to permit persons to whom the Software is
// furnished to do so, subject to the following conditions:
//
// The above copyright notice and this permission notice shall be included in all
// copies or substantial portions of the Software.
//
// THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
// IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
// FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
// AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
// LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
// OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN THE
// SOFTWARE.

package config

import (
	"flag"
	"fmt"
	"os"
	"path"
)

type Config struct {
	IndexSize     int
	ChunkSize     int
	Network       string
	StorageType   string
	DataDir       string
	ListenAddress string
	Peer          string
}

func DefaultConfig() *Config {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		homeDir = "/var/cache"
	}

	return &Config{
		IndexSize:     536854528,   // Индекс вместит 38_346_752 блоков, этого хватит минимум на 2 года.
		ChunkSize:     0xFFFF_FFFF, // 4GB
		Network:       "mainnet",
		StorageType:   "file",
		DataDir:       path.Join(homeDir, "umi"),
		ListenAddress: "127.0.0.1:8080",
		Peer:          "https://mainnet.umi.top",
	}
}

func (config *Config) Parse() {
	config.ParseEnvs()
	config.ParseFlags()
}

func (config *Config) ParseEnvs() {
	if network, ok := os.LookupEnv("UMI_NETWORK"); ok {
		config.Network = network
		config.Peer = fmt.Sprintf("https://%s.umi.top", network)
	}

	if dataDir, ok := os.LookupEnv("UMI_DATADIR"); ok {
		config.DataDir = dataDir
	}

	if address, ok := os.LookupEnv("UMI_BIND"); ok {
		config.ListenAddress = address
	}

	if peer, ok := os.LookupEnv("UMI_PEER"); ok {
		config.Peer = peer
	}

	if storage, ok := os.LookupEnv("UMI_STORAGE"); ok {
		config.StorageType = storage
	}
}

func (config *Config) ParseFlags() {
	usage := "The data directory is the location where UMI's data " +
		"files are stored. Overrides environment variable UMI_DATADIR."
	flag.StringVar(&config.DataDir, "datadir", config.DataDir, usage)

	usage = "Bind to given address and always listen on it. " +
		"Use [host]:port notation for IPv6. Overrides environment variable UMI_BIND."
	flag.StringVar(&config.ListenAddress, "bind", config.ListenAddress, usage)

	usage = "Connect only to specific peer. Overrides environment variable UMI_PEER."
	flag.StringVar(&config.Peer, "peer", config.Peer, usage)

	flag.Parse()
}
