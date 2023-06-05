package store

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"time"

	gost_mongo "github.com/bldsoft/gost/mongo"
	"github.com/bldsoft/gost/repository"
	"github.com/gorilla/securecookie"
	"github.com/gorilla/sessions"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

var MongoSessionStoreCollectionName = "session"

// MongoDBStore stores sessions using mongoDB as backend.
type MongoDBStore struct {
	rep     gost_mongo.Repository[sessionDoc, *sessionDoc]
	codecs  []securecookie.Codec
	options sessions.Options
}

// MongoDBStoreConfig is a configuration options for MongoDBStore
type MongoDBStoreConfig struct {

	// whether to create TTL index(https://docs.mongodb.com/manual/core/index-ttl/)
	// for the session document
	IndexTTL bool

	// gorilla-sessions options
	SessionOptions sessions.Options
}

type sessionDoc struct {
	gost_mongo.EntityID `bson:",inline"`
	Data                string    `bson:"data"`
	Modified            time.Time `bson:"modified"`
}

var defaultConfig = MongoDBStoreConfig{
	IndexTTL: true,
	SessionOptions: sessions.Options{
		Path:     "/",
		MaxAge:   3600 * 24 * 30,
		HttpOnly: true,
	},
}

// NewMongoDBStoreWithConfig returns a new NewMongoDBStore with a custom MongoDBStoreConfig
func NewMongoDBStoreWithConfig(db *gost_mongo.Storage, cfg MongoDBStoreConfig, keyPairs ...[]byte) (*MongoDBStore, error) {
	codecs := securecookie.CodecsFromPairs(keyPairs...)
	for _, codec := range codecs {
		if sc, ok := codec.(*securecookie.SecureCookie); ok {
			sc.MaxAge(cfg.SessionOptions.MaxAge)
			sc.MaxLength(0)
		}
	}
	store := &MongoDBStore{gost_mongo.NewRepository[sessionDoc](db, MongoSessionStoreCollectionName), codecs, cfg.SessionOptions}

	if !cfg.IndexTTL {
		return store, nil
	}

	return store, store.ensureIndexTTL()
}

// NewMongoDBStore returns a new NewMongoDBStore with default config
//
//	defaultConfig := MongoDBStoreConfig{
//		IndexTTL: true,
//		SessionOptions: sessions.Options{
//			Path:     "/",
//			MaxAge:   3600 * 24 * 30,
//			HttpOnly: true,
//		},
//	}
func NewMongoDBStore(db *gost_mongo.Storage, keyPairs ...[]byte) (*MongoDBStore, error) {
	return NewMongoDBStoreWithConfig(db, defaultConfig, keyPairs...)
}

// Get returns a session for the given name after adding it to the registry.
//
// It returns a new session if the sessions doesn't exist. Access IsNew on
// the session to check if it is an existing session or a new one.
//
// It returns a new session and an error if the session exists but could
// not be decoded.
func (mstore *MongoDBStore) Get(r *http.Request, name string) (*sessions.Session, error) {
	return sessions.GetRegistry(r).Get(mstore, name)
}

// New returns a session for the given name without adding it to the registry.
//
// The difference between New() and Get() is that calling New() twice will
// decode the session data twice, while Get() registers and reuses the same
// decoded session after the first call.
func (mstore *MongoDBStore) New(r *http.Request, name string) (*sessions.Session, error) {
	session := sessions.NewSession(mstore, name)
	options := mstore.options
	session.Options = &options
	session.IsNew = true

	cookie, err := r.Cookie(name)
	if err != nil {
		return session, nil
	}
	err = securecookie.DecodeMulti(name, cookie.Value, &session.ID, mstore.codecs...)
	if err != nil {
		return session, err
	}

	found, err := mstore.load(r.Context(), session)
	if err != nil {
		return session, err
	}
	session.IsNew = !found

	return session, nil
}

// Save adds a single session to the response and persist session in mongoDB collection
//
// If the Options.MaxAge of the session is <= 0 then the session file will be
// deleted from the store path. With this process it enforces the properly
// session cookie handling so no need to trust in the cookie management in the
// web browser.
func (mstore *MongoDBStore) Save(r *http.Request, w http.ResponseWriter, session *sessions.Session) error {
	ctx := context.Background()

	if session.Options.MaxAge < 0 {
		if err := mstore.rep.Delete(ctx, session.ID); err != nil {
			return err
		}
		http.SetCookie(w, sessions.NewCookie(session.Name(), "", session.Options))

		return nil
	}

	if session.ID == "" {
		session.ID = primitive.NewObjectID().Hex()
	}
	encoded, err := securecookie.EncodeMulti(session.Name(), session.Values, mstore.codecs...)
	if err != nil {
		return err
	}

	sessDoc := &sessionDoc{
		Modified: time.Now(),
		Data:     encoded,
	}
	sessDoc.SetIDFromString(session.ID)
	if val, ok := session.Values["modified"]; ok {
		modified, ok := val.(time.Time)
		if !ok {
			return errors.New("mongodbstore: invalid modified value")
		}
		sessDoc.Modified = modified
	}

	if err = mstore.rep.Upsert(ctx, sessDoc); err != nil {
		return err
	}
	encodedID, err := securecookie.EncodeMulti(session.Name(), session.ID, mstore.codecs...)
	if err != nil {
		return err
	}

	http.SetCookie(w, sessions.NewCookie(session.Name(), encodedID, session.Options))

	return nil
}

func (mstore *MongoDBStore) ensureIndexTTL() error {
	ctx := context.Background()

	indexName := "modified_at_TTL"

	cursor, err := mstore.rep.Collection().Indexes().List(ctx)
	if err != nil {
		return fmt.Errorf("mongodbstore: error ensuring TTL index. Unable to list indexes: %w", err)
	}

	for cursor.Next(ctx) {
		indexInfo := &struct {
			Name string `bson:"name"`
		}{}

		if err = cursor.Decode(indexInfo); err != nil {
			return fmt.Errorf("mongodbstore: error ensuring TTL index. Unable to decode bson index document %w", err)
		}

		if indexInfo.Name == indexName {
			return nil
		}

	}
	indexOpts := options.Index().
		SetExpireAfterSeconds(int32(mstore.options.MaxAge)).
		SetBackground(true).
		SetSparse(true).
		SetName(indexName)

	indexModel := mongo.IndexModel{
		Keys: bson.M{
			"modified_at": 1,
		},
		Options: indexOpts,
	}
	_, err = mstore.rep.Collection().Indexes().CreateOne(ctx, indexModel)
	if err != nil {
		return fmt.Errorf("mongodbstore: error ensuring TTL index. Unable to create index: %w", err)
	}

	return nil
}

func (mstore *MongoDBStore) load(ctx context.Context, sess *sessions.Session) (found bool, err error) {
	sessDoc, err := mstore.rep.FindByID(ctx, sess.ID)
	if err != nil {
		return false, err
	}
	if sessDoc.ID.IsZero() {
		return false, nil
	}
	err = securecookie.DecodeMulti(sess.Name(), sessDoc.Data, &sess.Values, mstore.codecs...)
	if err != nil {
		return false, err
	}

	return true, err
}

func (mstore *MongoDBStore) AllSessions(ctx context.Context, name string, offset, limit int) ([]*sessions.Session, error) {
	findOpt := options.Find().
		SetLimit(int64(limit)).
		SetSkip(int64(offset)).
		SetSort(bson.M{"_id": 1})
	cur, err := mstore.rep.Collection().Find(ctx, bson.M{}, findOpt)
	if err != nil {
		return nil, err
	}
	sessDocs := make([]sessionDoc, 0)
	if err = cur.All(ctx, &sessDocs); err != nil {
		return nil, err
	}

	res := make([]*sessions.Session, 0, len(sessDocs))
	for _, sessDoc := range sessDocs {
		sess := sessions.NewSession(mstore, name)
		sess.ID = sessDoc.StringID()
		err = securecookie.DecodeMulti(name, sessDoc.Data, &sess.Values, mstore.codecs...)
		if err != nil {
			return nil, err
		}
		res = append(res, sess)
	}
	return res, nil
}

func (mstore *MongoDBStore) KillSession(ctx context.Context, id string) error {
	return mstore.rep.Delete(ctx, id, &repository.QueryOptions{Archived: false})
}
