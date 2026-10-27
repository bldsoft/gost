package aerospike

import (
	"strings"
	"testing"
	"time"

	aero "github.com/aerospike/aerospike-client-go/v8"
	aeroTypes "github.com/aerospike/aerospike-client-go/v8/types"
	"github.com/stretchr/testify/require"

	"github.com/bldsoft/gost/cache/v2"
)

func TestBatchWriteErr(t *testing.T) {
	t.Parallel()

	mainKey := testKey(t, "ns", "set", "main")
	contKey := testKey(t, "ns", "set", "main_1")

	t.Run("all ok", func(t *testing.T) {
		t.Parallel()

		err := batchWriteErr([]aero.BatchRecordIfc{
			writeRecord(contKey, aeroTypes.OK, nil),
			writeRecord(mainKey, aeroTypes.OK, nil),
		})
		require.NoError(t, err)
	})

	t.Run("main key exists", func(t *testing.T) {
		t.Parallel()

		err := batchWriteErr([]aero.BatchRecordIfc{
			writeRecord(mainKey, aeroTypes.KEY_EXISTS_ERROR, &aero.AerospikeError{ResultCode: aeroTypes.KEY_EXISTS_ERROR}),
		})

		var asErr *aero.AerospikeError
		require.ErrorAs(t, err, &asErr)
		require.Equal(t, aeroTypes.KEY_EXISTS_ERROR, asErr.ResultCode)
	})

	t.Run("continuation only", func(t *testing.T) {
		t.Parallel()

		err := batchWriteErr([]aero.BatchRecordIfc{
			writeRecord(contKey, aeroTypes.RECORD_TOO_BIG, &aero.AerospikeError{ResultCode: aeroTypes.RECORD_TOO_BIG}),
			writeRecord(mainKey, aeroTypes.OK, nil),
		})
		require.Error(t, err)
		require.ErrorContains(t, err, "continuation")
		require.ErrorContains(t, err, contKey.String())
	})

	t.Run("result code without err", func(t *testing.T) {
		t.Parallel()

		err := batchWriteErr([]aero.BatchRecordIfc{
			writeRecord(mainKey, aeroTypes.KEY_BUSY, nil),
		})

		var asErr *aero.AerospikeError
		require.ErrorAs(t, err, &asErr)
		require.Equal(t, aeroTypes.KEY_BUSY, asErr.ResultCode)
	})

	t.Run("main error stays first", func(t *testing.T) {
		t.Parallel()

		err := batchWriteErr([]aero.BatchRecordIfc{
			writeRecord(contKey, aeroTypes.RECORD_TOO_BIG, &aero.AerospikeError{ResultCode: aeroTypes.RECORD_TOO_BIG}),
			writeRecord(mainKey, aeroTypes.KEY_EXISTS_ERROR, &aero.AerospikeError{ResultCode: aeroTypes.KEY_EXISTS_ERROR}),
		})

		var asErr *aero.AerospikeError
		require.ErrorAs(t, err, &asErr)
		require.Equal(t, aeroTypes.KEY_EXISTS_ERROR, asErr.ResultCode)
	})
}

func TestPrepBatchWritePolicies(t *testing.T) {
	t.Parallel()

	rep := &Repository{
		cache:         &Storage{namespace: "ns"},
		setName:       "set",
		itemSizeLimit: 4,
		liveTime:      time.Minute,
	}
	generation := uint32(7)
	ttl := 15 * time.Second

	_, records, err := rep.prepBatchWrite(false, "main", []byte("0123456789"), &generation, cache.WithTTL(ttl))
	require.NoError(t, err)
	require.Len(t, records, 3)

	main := records[len(records)-1].(*aero.BatchWrite)
	require.Contains(t, main.Key.String(), ":main:")
	require.NotContains(t, main.Key.String(), "main_")
	require.Equal(t, aero.CREATE_ONLY, main.Policy.RecordExistsAction)
	require.Equal(t, aero.EXPECT_GEN_EQUAL, main.Policy.GenerationPolicy)
	require.Equal(t, generation, main.Policy.Generation)
	require.Equal(t, truncExpiration(ttl), main.Policy.Expiration)

	for _, rec := range records[:len(records)-1] {
		cont := rec.(*aero.BatchWrite)
		require.True(t, strings.Contains(cont.Key.String(), "main_"))
		require.NotSame(t, main.Policy, cont.Policy)
		require.Equal(t, aero.REPLACE, cont.Policy.RecordExistsAction)
		require.Equal(t, aero.NONE, cont.Policy.GenerationPolicy)
		require.Equal(t, main.Policy.Expiration, cont.Policy.Expiration)
	}
}

func writeRecord(key *aero.Key, code aeroTypes.ResultCode, err aero.Error) *aero.BatchWrite {
	rec := aero.NewBatchWrite(nil, key)
	rec.ResultCode = code
	rec.Err = err

	return rec
}

func testKey(t *testing.T, namespace, set, userKey string) *aero.Key {
	t.Helper()

	key, err := aero.NewKey(namespace, set, userKey)
	require.NoError(t, err)

	return key
}
