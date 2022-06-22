package util

import "crypto/sha1"

func SHA1Hash(data ...[]byte) []byte {
	hashObj := sha1.New()

	for _, chunk := range data {
		hashObj.Write(chunk)
	}

	return hashObj.Sum(nil)
}
