package aerospike

import (
	"fmt"
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

		err := new(Repository).batchWriteErr([]aero.BatchRecordIfc{
			writeRecord(contKey, aeroTypes.OK, nil),
			writeRecord(mainKey, aeroTypes.OK, nil),
		})
		require.NoError(t, err)
	})

	t.Run("main key exists", func(t *testing.T) {
		t.Parallel()

		err := new(Repository).batchWriteErr([]aero.BatchRecordIfc{
			writeRecord(mainKey, aeroTypes.KEY_EXISTS_ERROR, &aero.AerospikeError{ResultCode: aeroTypes.KEY_EXISTS_ERROR}),
		})

		var asErr *aero.AerospikeError
		require.ErrorAs(t, err, &asErr)
		require.Equal(t, aeroTypes.KEY_EXISTS_ERROR, asErr.ResultCode)
		require.NotContains(t, err.Error(), "continuation")
	})

	t.Run("continuation only", func(t *testing.T) {
		t.Parallel()

		err := new(Repository).batchWriteErr([]aero.BatchRecordIfc{
			writeRecord(contKey, aeroTypes.RECORD_TOO_BIG, &aero.AerospikeError{ResultCode: aeroTypes.RECORD_TOO_BIG}),
			writeRecord(mainKey, aeroTypes.OK, nil),
		})
		require.Error(t, err)
		require.ErrorContains(t, err, "continuation")
		require.ErrorContains(t, err, contKey.String())
	})

	t.Run("result code without err", func(t *testing.T) {
		t.Parallel()

		err := new(Repository).batchWriteErr([]aero.BatchRecordIfc{
			writeRecord(contKey, aeroTypes.KEY_BUSY, nil),
			writeRecord(mainKey, aeroTypes.OK, nil),
		})

		var asErr *aero.AerospikeError
		require.ErrorAs(t, err, &asErr)
		require.Equal(t, aeroTypes.KEY_BUSY, asErr.ResultCode)
		require.ErrorContains(t, err, "continuation")
	})

	t.Run("main error stays first", func(t *testing.T) {
		t.Parallel()

		err := new(Repository).batchWriteErr([]aero.BatchRecordIfc{
			writeRecord(contKey, aeroTypes.RECORD_TOO_BIG, &aero.AerospikeError{ResultCode: aeroTypes.RECORD_TOO_BIG}),
			writeRecord(mainKey, aeroTypes.KEY_EXISTS_ERROR, &aero.AerospikeError{ResultCode: aeroTypes.KEY_EXISTS_ERROR}),
		})

		var asErr *aero.AerospikeError
		require.ErrorAs(t, err, &asErr)
		require.Equal(t, aeroTypes.KEY_EXISTS_ERROR, asErr.ResultCode)
	})
}

func TestPrepBatchWrite(t *testing.T) {
	t.Parallel()

	rep := &Repository{
		cache:         &Storage{namespace: "ns"},
		setName:       "set",
		itemSizeLimit: 4,
		liveTime:      time.Minute,
	}
	generation := uint32(7)
	ttl := 15 * time.Second

	mainVal, continuations := rep.split("main", "abc", []byte("0123456789"))
	records, err := rep.prepBatchWrite(false, "main", mainVal, continuations, 10, &generation, cache.CollectItem(cache.WithTTL(ttl)))
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
		require.Contains(t, cont.Key.String(), "main_abc_")
		require.Equal(t, aero.CREATE_ONLY, cont.Policy.RecordExistsAction)
		require.Equal(t, aero.NONE, cont.Policy.GenerationPolicy)
		require.Equal(t, main.Policy.Expiration, cont.Policy.Expiration)
	}
}

func TestSplitUniqueKeys(t *testing.T) {
	t.Parallel()

	rep := &Repository{itemSizeLimit: 4}
	val := []byte("0123456789abcdef")

	mainA, contA := rep.split("k", "n1", val)
	mainB, contB := rep.split("k", "n2", val)

	require.Equal(t, []byte("0123"), mainA)
	require.Equal(t, mainA, mainB)
	require.Len(t, contA, 3)
	require.Len(t, contB, 3)

	want := [][]byte{[]byte("4567"), []byte("89ab"), []byte("cdef")}
	for i, c := range contA {
		require.Equal(t, want[i], c.Value)
		require.Equal(t, want[i], contB[i].Value)
		require.Equal(t, fmt.Sprintf("k_n1_%d", i+1), c.Key)
		require.Equal(t, fmt.Sprintf("k_n2_%d", i+1), contB[i].Key)
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
