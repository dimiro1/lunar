package runner

import (
	"github.com/dimiro1/lunar/internal/runtime/crypto"
	lua "github.com/yuin/gopher-lua"
)

// registerCrypto registers the crypto module with hashing and UUID functions.
// This is a thin wrapper around the stdlib/crypto package.
func registerCrypto(L *lua.LState) {
	cryptoModule := L.NewTable()

	// Hashing functions
	L.SetField(cryptoModule, "md5", L.NewFunction(cryptoMD5))
	L.SetField(cryptoModule, "sha1", L.NewFunction(cryptoSHA1))
	L.SetField(cryptoModule, "sha256", L.NewFunction(cryptoSHA256))
	L.SetField(cryptoModule, "sha512", L.NewFunction(cryptoSHA512))

	// HMAC functions
	L.SetField(cryptoModule, "hmac_sha1", L.NewFunction(cryptoHMACSHA1))
	L.SetField(cryptoModule, "hmac_sha256", L.NewFunction(cryptoHMACSHA256))
	L.SetField(cryptoModule, "hmac_sha512", L.NewFunction(cryptoHMACSHA512))

	// UUID function
	L.SetField(cryptoModule, "uuid", L.NewFunction(cryptoUUID))

	L.SetGlobal("crypto", cryptoModule)
}

// cryptoMD5 computes MD5 hash of a string
// Usage: local hash = crypto.md5(str)
func cryptoMD5(L *lua.LState) int {
	L.Push(lua.LString(crypto.MD5(L.CheckString(1))))
	return 1
}

// cryptoSHA1 computes SHA1 hash of a string
// Usage: local hash = crypto.sha1(str)
func cryptoSHA1(L *lua.LState) int {
	L.Push(lua.LString(crypto.SHA1(L.CheckString(1))))
	return 1
}

// cryptoSHA256 computes SHA256 hash of a string
// Usage: local hash = crypto.sha256(str)
func cryptoSHA256(L *lua.LState) int {
	L.Push(lua.LString(crypto.SHA256(L.CheckString(1))))
	return 1
}

// cryptoSHA512 computes SHA512 hash of a string
// Usage: local hash = crypto.sha512(str)
func cryptoSHA512(L *lua.LState) int {
	L.Push(lua.LString(crypto.SHA512(L.CheckString(1))))
	return 1
}

// cryptoHMACSHA1 computes HMAC-SHA1 of a message with a secret key
// Usage: local hash = crypto.hmac_sha1(message, key)
func cryptoHMACSHA1(L *lua.LState) int {
	L.Push(lua.LString(crypto.HMACSHA1(L.CheckString(1), L.CheckString(2))))
	return 1
}

// cryptoHMACSHA256 computes HMAC-SHA256 of a message with a secret key
// Usage: local hash = crypto.hmac_sha256(message, key)
func cryptoHMACSHA256(L *lua.LState) int {
	L.Push(lua.LString(crypto.HMACSHA256(L.CheckString(1), L.CheckString(2))))
	return 1
}

// cryptoHMACSHA512 computes HMAC-SHA512 of a message with a secret key
// Usage: local hash = crypto.hmac_sha512(message, key)
func cryptoHMACSHA512(L *lua.LState) int {
	L.Push(lua.LString(crypto.HMACSHA512(L.CheckString(1), L.CheckString(2))))
	return 1
}

// cryptoUUID generates a new UUID v4
// Usage: local id = crypto.uuid()
func cryptoUUID(L *lua.LState) int {
	L.Push(lua.LString(crypto.UUID()))
	return 1
}
