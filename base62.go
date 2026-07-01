package main

import (
	"math/big"
	"strings"
)

var (
	characterSet = "0123456789ABCDEFGHIJKLMNOPQRSTUVWXYZabcdefghijklmnopqrstuvwxyz"
)

func encodeUUIDBytesToBase62(uuidBytes [16]byte) string {
	n := new(big.Int).SetBytes(uuidBytes[:])

	if n.Sign() == 0 {
		return string(characterSet[0])
	}

	var result []byte
	base := big.NewInt(int64(len(characterSet)))
	rem := new(big.Int)

	for n.Sign() > 0 {
		n.QuoRem(n, base, rem)
		result = append(result, characterSet[rem.Int64()])
	}

	for i, j := 0, len(result)-1; i < j; i, j = i+1, j-1 {
		result[i], result[j] = result[j], result[i]
	}

	return string(result)
}

func decodeBase62ToUUID(encoded string) [16]byte {
	var uuid [16]byte
	n := new(big.Int)
	base := big.NewInt(int64(len(characterSet)))

	for i := 0; i < len(encoded); i++ {
		charIndex := int64(strings.IndexByte(characterSet, encoded[i]))
		n.Mul(n, base)
		n.Add(n, big.NewInt(charIndex))
	}

	if n.BitLen() > 128 {
		return uuid
	}

	copy(uuid[:], n.Bytes())
	return uuid
}
