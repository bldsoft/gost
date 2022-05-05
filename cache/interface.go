package cache

//ILocalCacheRepository ...
type LocalCacheRepository interface {
	Get(key string) ([]byte, error)
	Set(key string, value []byte) error
	Delete(key string) error
	Reset()
}

//IDistrCacheRepository ...
type DistrCacheRepository interface {
	LocalCacheRepository

	//TODO: add it to ILocalCacheRepository
	GetWithFlags(key string) (data []byte, flags uint32, err error)
	Exist(key string) bool
	Add(key string, value []byte) error
	AddFor(key string, value []byte, expirationSec int32) error
	SetFor(key string, value []byte, expirationSec int32) error
	SetWithFlags(key string, value []byte, flags uint32) error
	CompareAndSwap(key string, handler func(value []byte) ([]byte, error)) error
}
