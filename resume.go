package main

import (
	"crypto/sha256"
	"encoding/hex"
	"io/ioutil"
	"os"
	"path/filepath"
	"strconv"
)

// TODO: save config as json, so we can restore

// Find last stored position
// Returns 0, nil if none is found and no error occurs
func lookupPos(contentHash [sha256.Size]byte) (int, error) {
	cacheDir, err := os.UserCacheDir()
	if err != nil {
		return 0, err
	}
	pos, err := ioutil.ReadFile(filepath.Join(cacheDir, "greedy", hex.EncodeToString(contentHash[:])))
	if err != nil {
		if os.IsNotExist(err) {
			return 0, nil
		}
		return 0, err
	}
	return strconv.Atoi(string(pos))
}

func writePos(contentHash [sha256.Size]byte, pos int) error {
	cacheDir, err := os.UserCacheDir()
	if err != nil {
		return err
	}
	err = os.MkdirAll(filepath.Join(cacheDir, "greedy"), 0777)
	if err != nil {
		return err
	}
	err = ioutil.WriteFile(
		filepath.Join(cacheDir, "greedy", hex.EncodeToString(contentHash[:])),
		[]byte(strconv.Itoa(pos)),
		0777,
	)
	if err != nil {
		return err
	}
	return nil
}
