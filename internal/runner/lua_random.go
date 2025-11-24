package runner

import (
	"crypto/rand"
	"encoding/base64"
	"encoding/hex"
	"math/big"
	mathrand "math/rand"

	"github.com/rs/xid"
	lua "github.com/yuin/gopher-lua"
)

// registerRandom registers the random module with random generation functions
func registerRandom(L *lua.LState) {
	randomModule := L.NewTable()

	// Register random functions
	L.SetField(randomModule, "int", L.NewFunction(randomInt))
	L.SetField(randomModule, "float", L.NewFunction(randomFloat))
	L.SetField(randomModule, "string", L.NewFunction(randomString))
	L.SetField(randomModule, "bytes", L.NewFunction(randomBytes))
	L.SetField(randomModule, "hex", L.NewFunction(randomHex))
	L.SetField(randomModule, "id", L.NewFunction(randomID))

	// Set the random module as a global
	L.SetGlobal("random", randomModule)
}

// randomInt generates a random integer between min and max (inclusive)
// Usage: local num = random.int(min, max)
func randomInt(L *lua.LState) int {
	minValue := L.CheckInt(1)
	maxValue := L.CheckInt(2)

	if minValue > maxValue {
		L.ArgError(1, "min must be less than or equal to max")
		return 0
	}

	// Use crypto/rand for better randomness
	rangeSize := int64(maxValue - minValue + 1)
	n, err := rand.Int(rand.Reader, big.NewInt(rangeSize))
	if err != nil {
		// Fallback to math/rand
		result := mathrand.Intn(int(rangeSize)) + minValue
		L.Push(lua.LNumber(result))
		return 1
	}

	result := int(n.Int64()) + minValue
	L.Push(lua.LNumber(result))
	return 1
}

// randomFloat generates a random float between 0.0 and 1.0
// Usage: local num = random.float()
func randomFloat(L *lua.LState) int {
	result := mathrand.Float64()
	L.Push(lua.LNumber(result))
	return 1
}

// randomString generates a random alphanumeric string of specified length
// Usage: local str = random.string(length)
func randomString(L *lua.LState) int {
	length := L.CheckInt(1)

	if length <= 0 {
		L.ArgError(1, "length must be positive")
		return 0
	}

	// Generate random bytes
	bytes := make([]byte, length)
	const charset = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"

	for i := range bytes {
		n, err := rand.Int(rand.Reader, big.NewInt(int64(len(charset))))
		if err != nil {
			// Fallback to math/rand
			bytes[i] = charset[mathrand.Intn(len(charset))]
		} else {
			bytes[i] = charset[n.Int64()]
		}
	}

	L.Push(lua.LString(string(bytes)))
	return 1
}

// randomBytes generates random bytes and returns them as base64-encoded string
// Usage: local bytes = random.bytes(length)
func randomBytes(L *lua.LState) int {
	length := L.CheckInt(1)

	if length <= 0 {
		L.ArgError(1, "length must be positive")
		return 0
	}

	bytes := make([]byte, length)
	_, err := rand.Read(bytes)
	if err != nil {
		L.Push(lua.LNil)
		L.Push(lua.LString(err.Error()))
		return 2
	}

	// Return as base64 encoded string
	encoded := base64.StdEncoding.EncodeToString(bytes)
	L.Push(lua.LString(encoded))
	return 1
}

// randomHex generates random bytes and returns them as hex-encoded string
// Usage: local hexStr = random.hex(length)
func randomHex(L *lua.LState) int {
	length := L.CheckInt(1)

	if length <= 0 {
		L.ArgError(1, "length must be positive")
		return 0
	}

	bytes := make([]byte, length)
	_, err := rand.Read(bytes)
	if err != nil {
		L.Push(lua.LNil)
		L.Push(lua.LString(err.Error()))
		return 2
	}

	// Return as hex encoded string
	hexStr := hex.EncodeToString(bytes)
	L.Push(lua.LString(hexStr))
	return 1
}

// randomID generates a globally unique sortable ID using xid
// Usage: local id = random.id()
// Returns a 20-character string (smaller than UUID, sortable by creation time)
func randomID(L *lua.LState) int {
	id := xid.New()
	L.Push(lua.LString(id.String()))
	return 1
}
