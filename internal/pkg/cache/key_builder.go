package cache

import (
	"crypto/sha1"
	"encoding/hex"
	"hash"
	"strings"
	"sync"
)

const (
	DefaultSeparator string = ":"
)

// KeyBuilderInterface ...
type KeyBuilderInterface interface {
	NewKey() KeyInterface
}

var _ KeyBuilderInterface = (*keyBuilder)(nil)

type keyBuilder struct {
	separator string
	hasher    hash.Hash
	lock      sync.Mutex
}

// NewKeyBuilderDefault create new default key builder
func NewKeyBuilderDefault() *keyBuilder {
	return &keyBuilder{
		separator: DefaultSeparator,
		hasher:    sha1.New(),
	}
}

// NewKeyBuilder create new key builder
func NewKeyBuilder(separator string, hasher hash.Hash) *keyBuilder {
	if hasher == nil {
		hasher = sha1.New()
	}
	return &keyBuilder{
		separator: separator,
		hasher:    hasher,
	}
}

// NewKey ...
func (kb *keyBuilder) NewKey() KeyInterface {
	return NewKey(kb)
}

// KeyInterface ...
type KeyInterface interface {
	Add(key string)
	AddHashed(key string) error
	String() string
	Reset()
}

var _ KeyInterface = (*key)(nil)

type key struct {
	builder  *keyBuilder
	keyParts []string
}

// NewKey
func NewKey(builder *keyBuilder) *key {
	return &key{
		builder: builder,
	}
}

// Add add part key
func (k *key) Add(key string) {
	k.keyParts = append(k.keyParts, key)
}

// AddHashed add part key with hashed
func (k *key) AddHashed(key string) error {
	k.builder.lock.Lock()
	defer k.builder.lock.Unlock()
	k.builder.hasher.Reset()
	defer k.builder.hasher.Reset()

	_, err := k.builder.hasher.Write([]byte(key))
	if err != nil {
		return err
	}

	k.keyParts = append(k.keyParts, hex.EncodeToString(k.builder.hasher.Sum(nil)))
	return nil
}

// String ...
func (k *key) String() string {
	return strings.Join(k.keyParts, k.builder.separator)
}

// Reset ...
func (k *key) Reset() {
	k.keyParts = nil
}
