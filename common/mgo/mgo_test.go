package mgo

import (
	"errors"
	"sync"
	"testing"

	mgo "gopkg.in/mgo.v2"

	"github.com/andeya/gust/result"
	"github.com/andeya/pholcus/common/pool"
)

type mockQuery struct {
	count    int
	docs     []interface{}
	countErr error
	allErr   error
}

func (m *mockQuery) Count() (int, error) {
	if m.countErr != nil {
		return 0, m.countErr
	}
	return m.count, nil
}

func (m *mockQuery) Sort(_ ...string) queryProvider { return m }

func (m *mockQuery) Skip(_ int) queryProvider { return m }

func (m *mockQuery) Limit(_ int) queryProvider { return m }

func (m *mockQuery) Select(_ interface{}) queryProvider { return m }

func (m *mockQuery) All(result interface{}) error {
	if m.allErr != nil {
		return m.allErr
	}
	*(result.(*[]interface{})) = m.docs
	return nil
}

type mockCollection struct {
	insertErr    error
	removeErr    error
	updateErr    error
	updateAll    *mgo.ChangeInfo
	updateAllErr error
	upsert       *mgo.ChangeInfo
	upsertErr    error
	findQuery    *mockQuery
}

func (m *mockCollection) Find(query interface{}) queryProvider {
	if m.findQuery != nil {
		return m.findQuery
	}
	return &mockQuery{count: 0, docs: []interface{}{}}
}

func (m *mockCollection) Insert(_ ...interface{}) error { return m.insertErr }

func (m *mockCollection) Remove(_ interface{}) error { return m.removeErr }

func (m *mockCollection) Update(_, _ interface{}) error { return m.updateErr }

func (m *mockCollection) UpdateAll(_, _ interface{}) (*mgo.ChangeInfo, error) {
	return m.updateAll, m.updateAllErr
}

func (m *mockCollection) Upsert(_, _ interface{}) (*mgo.ChangeInfo, error) {
	return m.upsert, m.upsertErr
}

type mockDatabase struct {
	collections map[string]collectionProvider
	names       []string
	namesErr    error
}

func (m *mockDatabase) C(name string) collectionProvider {
	if m.collections != nil {
		if c, ok := m.collections[name]; ok {
			return c
		}
	}
	return &mockCollection{findQuery: &mockQuery{}}
}

func (m *mockDatabase) CollectionNames() ([]string, error) {
	return m.names, m.namesErr
}

type mockSession struct {
	mu         sync.Mutex
	usable     bool
	dbs        map[string]dbProvider
	dbNames    []string
	dbNamesErr error
}

func (m *mockSession) Usable() bool {
	m.mu.Lock()
	defer m.mu.Unlock()
	return m.usable
}

func (m *mockSession) Reset() {}

func (m *mockSession) Close() {}

func (m *mockSession) DB(name string) dbProvider {
	m.mu.Lock()
	defer m.mu.Unlock()
	if m.dbs != nil {
		if db, ok := m.dbs[name]; ok {
			return db
		}
	}
	return &mockDatabase{collections: map[string]collectionProvider{}}
}

func (m *mockSession) DatabaseNames() ([]string, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	return m.dbNames, m.dbNamesErr
}

type mockPool struct {
	src pool.Src
}

func (m *mockPool) Call(fn func(pool.Src) error) result.VoidResult {
	return result.RetVoid(fn(m.src))
}

func (m *mockPool) Close() {}

func (m *mockPool) Len() int { return 0 }

func setupMockPool(s sessionProvider) {
	testPool = &mockPool{src: s.(pool.Src)}
	getSessionFunc = func(src pool.Src) sessionProvider { return src.(sessionProvider) }
}

func teardownMockPool() {
	testPool = nil
	getSessionFunc = func(src pool.Src) sessionProvider {
		return &mgoSessionAdapter{src.(*MgoSrc)}
	}
}

func TestGetOperator(t *testing.T) {
	tests := []struct {
		op     string
		expect bool
	}{
		{"list", true}, {"LIST", true}, {"List", true},
		{"count", true}, {"COUNT", true},
		{"find", true}, {"Find", true},
		{"insert", true}, {"INSERT", true},
		{"update", true}, {"Update", true},
		{"update_all", true}, {"UPDATE_ALL", true},
		{"upsert", true}, {"Upsert", true},
		{"remove", true}, {"Remove", true},
		{"unknown", false}, {"", false},
	}
	for _, tt := range tests {
		o := getOperator(tt.op)
		if (o != nil) != tt.expect {
			t.Errorf("getOperator(%q) = %v, expect non-nil=%v", tt.op, o, tt.expect)
		}
	}
}

func TestMgo_InvalidOp(t *testing.T) {
	r := Mgo(nil, "invalid_op", nil)
	if !r.IsErr() {
		t.Error("expected error for invalid operation")
	}
}

func TestMgo_UnknownOptionField(t *testing.T) {
	ms := &mockSession{
		usable: true,
		dbs: map[string]dbProvider{
			"db": &mockDatabase{
				collections: map[string]collectionProvider{
					"coll": &mockCollection{findQuery: &mockQuery{count: 0}},
				},
			},
		},
	}
	setupMockPool(ms)
	defer teardownMockPool()

	var count int
	r := Mgo(&count, "count", map[string]interface{}{
		"Database":     "db",
		"Collection":   "coll",
		"Query":        map[string]interface{}{},
		"UnknownField": "ignored",
	})
	if r.IsErr() {
		t.Fatalf("Mgo with unknown option: %v", r.UnwrapErr())
	}
}

func TestMgo_OptionReflection(t *testing.T) {
	ms := &mockSession{
		usable: true,
		dbs: map[string]dbProvider{
			"testdb": &mockDatabase{
				collections: map[string]collectionProvider{
					"testcoll": &mockCollection{findQuery: &mockQuery{count: 5}},
				},
			},
		},
	}
	setupMockPool(ms)
	defer teardownMockPool()

	var count int
	r := Mgo(&count, "count", map[string]interface{}{
		"Database":   "testdb",
		"Collection": "testcoll",
		"Query":      map[string]interface{}{"a": 1},
	})
	if r.IsErr() {
		t.Fatalf("Mgo count: %v", r.UnwrapErr())
	}
	if count != 5 {
		t.Errorf("count = %d, want 5", count)
	}
}

func TestList_Exec(t *testing.T) {
	ms := &mockSession{
		usable:  true,
		dbNames: []string{"db1", "db2"},
		dbs: map[string]dbProvider{
			"db1": &mockDatabase{names: []string{"c1", "c2"}},
			"db2": &mockDatabase{names: []string{"c3"}},
		},
	}
	setupMockPool(ms)
	defer teardownMockPool()

	var li map[string][]string
	r := Mgo(&li, "list", map[string]interface{}{
		"Dbs": []string{"db1", "db2"},
	})
	if r.IsErr() {
		t.Fatalf("Mgo list: %v", r.UnwrapErr())
	}
	if li["db1"][0] != "c1" || li["db1"][1] != "c2" {
		t.Errorf("db1 collections = %v", li["db1"])
	}
	if li["db2"][0] != "c3" {
		t.Errorf("db2 collections = %v", li["db2"])
	}
}

func TestList_Exec_AllDbs(t *testing.T) {
	ms := &mockSession{
		usable:  true,
		dbNames: []string{"db1"},
		dbs: map[string]dbProvider{
			"db1": &mockDatabase{names: []string{"c1"}},
		},
	}
	setupMockPool(ms)
	defer teardownMockPool()

	var li map[string][]string
	r := Mgo(&li, "list", map[string]interface{}{})
	if r.IsErr() {
		t.Fatalf("Mgo list all: %v", r.UnwrapErr())
	}
	if li["db1"][0] != "c1" {
		t.Errorf("li = %v", li)
	}
}

func TestCount_Exec(t *testing.T) {
	ms := &mockSession{
		usable: true,
		dbs: map[string]dbProvider{
			"db": &mockDatabase{
				collections: map[string]collectionProvider{
					"coll": &mockCollection{
						findQuery: &mockQuery{count: 42},
					},
				},
			},
		},
	}
	setupMockPool(ms)
	defer teardownMockPool()

	var count int
	r := Mgo(&count, "count", map[string]interface{}{
		"Database":   "db",
		"Collection": "coll",
		"Query":      map[string]interface{}{"k": "v"},
	})
	if r.IsErr() {
		t.Fatalf("Mgo count: %v", r.UnwrapErr())
	}
	if count != 42 {
		t.Errorf("count = %d, want 42", count)
	}
}

func TestCount_Exec_ObjectId(t *testing.T) {
	ms := &mockSession{
		usable: true,
		dbs: map[string]dbProvider{
			"db": &mockDatabase{
				collections: map[string]collectionProvider{
					"coll": &mockCollection{findQuery: &mockQuery{count: 1}},
				},
			},
		},
	}
	setupMockPool(ms)
	defer teardownMockPool()

	var count int
	r := Mgo(&count, "count", map[string]interface{}{
		"Database":   "db",
		"Collection": "coll",
		"Query":      map[string]interface{}{"_id": "507f1f77bcf86cd799439011"},
	})
	if r.IsErr() {
		t.Fatalf("Mgo count _id: %v", r.UnwrapErr())
	}
	if count != 1 {
		t.Errorf("count = %d, want 1", count)
	}
}

func TestCount_Exec_InvalidIdType(t *testing.T) {
	ms := &mockSession{usable: true, dbs: map[string]dbProvider{"db": &mockDatabase{}}}
	setupMockPool(ms)
	defer teardownMockPool()

	var count int
	r := Mgo(&count, "count", map[string]interface{}{
		"Database":   "db",
		"Collection": "coll",
		"Query":      map[string]interface{}{"_id": 123},
	})
	if !r.IsErr() {
		t.Error("expected error for invalid _id type")
	}
}

func TestFind_Exec(t *testing.T) {
	docs := []interface{}{map[string]interface{}{"a": 1}, map[string]interface{}{"b": 2}}
	ms := &mockSession{
		usable: true,
		dbs: map[string]dbProvider{
			"db": &mockDatabase{
				collections: map[string]collectionProvider{
					"coll": &mockCollection{
						findQuery: &mockQuery{count: 2, docs: docs},
					},
				},
			},
		},
	}
	setupMockPool(ms)
	defer teardownMockPool()

	var res map[string]interface{}
	r := Mgo(&res, "find", map[string]interface{}{
		"Database":   "db",
		"Collection": "coll",
		"Query":      map[string]interface{}{},
		"Sort":       []string{"-a"},
		"Skip":       1,
		"Limit":      10,
		"Select":     map[string]int{"name": 1},
	})
	if r.IsErr() {
		t.Fatalf("Mgo find: %v", r.UnwrapErr())
	}
	if res["Total"] != 2 {
		t.Errorf("Total = %v, want 2", res["Total"])
	}
	if len(res["Docs"].([]interface{})) != 2 {
		t.Errorf("Docs len = %d", len(res["Docs"].([]interface{})))
	}
}

func TestFind_Exec_InvalidIdType(t *testing.T) {
	ms := &mockSession{usable: true, dbs: map[string]dbProvider{"db": &mockDatabase{}}}
	setupMockPool(ms)
	defer teardownMockPool()

	var res map[string]interface{}
	r := Mgo(&res, "find", map[string]interface{}{
		"Database":   "db",
		"Collection": "coll",
		"Query":      map[string]interface{}{"_id": 123},
	})
	if !r.IsErr() {
		t.Error("expected error for invalid _id type")
	}
}

func TestInsert_Exec(t *testing.T) {
	ms := &mockSession{
		usable: true,
		dbs: map[string]dbProvider{
			"db": &mockDatabase{
				collections: map[string]collectionProvider{
					"coll": &mockCollection{},
				},
			},
		},
	}
	setupMockPool(ms)
	defer teardownMockPool()

	var ids []string
	r := Mgo(&ids, "insert", map[string]interface{}{
		"Database":   "db",
		"Collection": "coll",
		"Docs": []map[string]interface{}{
			{"name": "a"},
			{"_id": "507f1f77bcf86cd799439011", "name": "b"},
			{"_id": nil, "x": 1},
			{"_id": "", "y": 2},
			{"_id": 0, "z": 3},
		},
	})
	if r.IsErr() {
		t.Fatalf("Mgo insert: %v", r.UnwrapErr())
	}
	if len(ids) != 5 {
		t.Errorf("ids len = %d", len(ids))
	}
	if ids[1] != "507f1f77bcf86cd799439011" {
		t.Errorf("ids[1] = %s", ids[1])
	}
}

func TestInsert_Exec_NilResultPtr(t *testing.T) {
	ms := &mockSession{
		usable: true,
		dbs: map[string]dbProvider{
			"db": &mockDatabase{
				collections: map[string]collectionProvider{
					"coll": &mockCollection{},
				},
			},
		},
	}
	setupMockPool(ms)
	defer teardownMockPool()

	r := Mgo(nil, "insert", map[string]interface{}{
		"Database":   "db",
		"Collection": "coll",
		"Docs":       []map[string]interface{}{{"x": 1}},
	})
	if r.IsErr() {
		t.Fatalf("Mgo insert nil: %v", r.UnwrapErr())
	}
}

func TestRemove_Exec(t *testing.T) {
	ms := &mockSession{
		usable: true,
		dbs: map[string]dbProvider{
			"db": &mockDatabase{
				collections: map[string]collectionProvider{
					"coll": &mockCollection{},
				},
			},
		},
	}
	setupMockPool(ms)
	defer teardownMockPool()

	r := Mgo(nil, "remove", map[string]interface{}{
		"Database":   "db",
		"Collection": "coll",
		"Selector":   map[string]interface{}{"a": 1},
	})
	if r.IsErr() {
		t.Fatalf("Mgo remove: %v", r.UnwrapErr())
	}
}

func TestRemove_Exec_ObjectId(t *testing.T) {
	ms := &mockSession{
		usable: true,
		dbs: map[string]dbProvider{
			"db": &mockDatabase{
				collections: map[string]collectionProvider{
					"coll": &mockCollection{},
				},
			},
		},
	}
	setupMockPool(ms)
	defer teardownMockPool()

	r := Mgo(nil, "remove", map[string]interface{}{
		"Database":   "db",
		"Collection": "coll",
		"Selector":   map[string]interface{}{"_id": "507f1f77bcf86cd799439011"},
	})
	if r.IsErr() {
		t.Fatalf("Mgo remove _id: %v", r.UnwrapErr())
	}
}

func TestRemove_Exec_InvalidIdType(t *testing.T) {
	ms := &mockSession{usable: true, dbs: map[string]dbProvider{"db": &mockDatabase{}}}
	setupMockPool(ms)
	defer teardownMockPool()

	r := Mgo(nil, "remove", map[string]interface{}{
		"Database":   "db",
		"Collection": "coll",
		"Selector":   map[string]interface{}{"_id": 123},
	})
	if !r.IsErr() {
		t.Error("expected error for invalid _id type")
	}
}

func TestUpdate_Exec(t *testing.T) {
	ms := &mockSession{
		usable: true,
		dbs: map[string]dbProvider{
			"db": &mockDatabase{
				collections: map[string]collectionProvider{
					"coll": &mockCollection{},
				},
			},
		},
	}
	setupMockPool(ms)
	defer teardownMockPool()

	r := Mgo(nil, "update", map[string]interface{}{
		"Database":   "db",
		"Collection": "coll",
		"Selector":   map[string]interface{}{"a": 1},
		"Change":     map[string]interface{}{"$set": map[string]interface{}{"b": 2}},
	})
	if r.IsErr() {
		t.Fatalf("Mgo update: %v", r.UnwrapErr())
	}
}

func TestUpdate_Exec_InvalidIdType(t *testing.T) {
	ms := &mockSession{usable: true, dbs: map[string]dbProvider{"db": &mockDatabase{}}}
	setupMockPool(ms)
	defer teardownMockPool()

	r := Mgo(nil, "update", map[string]interface{}{
		"Database":   "db",
		"Collection": "coll",
		"Selector":   map[string]interface{}{"_id": 123},
		"Change":     map[string]interface{}{"x": 1},
	})
	if !r.IsErr() {
		t.Error("expected error for invalid _id type")
	}
}

func TestUpdateAll_Exec(t *testing.T) {
	ms := &mockSession{
		usable: true,
		dbs: map[string]dbProvider{
			"db": &mockDatabase{
				collections: map[string]collectionProvider{
					"coll": &mockCollection{
						updateAll: &mgo.ChangeInfo{Updated: 3, Removed: 0, UpsertedId: nil},
					},
				},
			},
		},
	}
	setupMockPool(ms)
	defer teardownMockPool()

	var res map[string]interface{}
	r := Mgo(&res, "update_all", map[string]interface{}{
		"Database":   "db",
		"Collection": "coll",
		"Selector":   map[string]interface{}{"a": 1},
		"Change":     map[string]interface{}{"$set": map[string]interface{}{"b": 2}},
	})
	if r.IsErr() {
		t.Fatalf("Mgo update_all: %v", r.UnwrapErr())
	}
	if res["Updated"] != 3 {
		t.Errorf("Updated = %v, want 3", res["Updated"])
	}
}

func TestUpdateAll_Exec_InvalidIdType(t *testing.T) {
	ms := &mockSession{usable: true, dbs: map[string]dbProvider{"db": &mockDatabase{}}}
	setupMockPool(ms)
	defer teardownMockPool()

	var res map[string]interface{}
	r := Mgo(&res, "update_all", map[string]interface{}{
		"Database":   "db",
		"Collection": "coll",
		"Selector":   map[string]interface{}{"_id": 123},
		"Change":     map[string]interface{}{"x": 1},
	})
	if !r.IsErr() {
		t.Error("expected error for invalid _id type")
	}
}

func TestUpsert_Exec(t *testing.T) {
	ms := &mockSession{
		usable: true,
		dbs: map[string]dbProvider{
			"db": &mockDatabase{
				collections: map[string]collectionProvider{
					"coll": &mockCollection{
						upsert: &mgo.ChangeInfo{Updated: 0, Removed: 0, UpsertedId: "abc"},
					},
				},
			},
		},
	}
	setupMockPool(ms)
	defer teardownMockPool()

	var res map[string]interface{}
	r := Mgo(&res, "upsert", map[string]interface{}{
		"Database":   "db",
		"Collection": "coll",
		"Selector":   map[string]interface{}{"a": 1},
		"Change":     map[string]interface{}{"$set": map[string]interface{}{"b": 2}},
	})
	if r.IsErr() {
		t.Fatalf("Mgo upsert: %v", r.UnwrapErr())
	}
	if res["UpsertedId"] != "abc" {
		t.Errorf("UpsertedId = %v, want abc", res["UpsertedId"])
	}
}

func TestUpsert_Exec_InvalidIdType(t *testing.T) {
	ms := &mockSession{usable: true, dbs: map[string]dbProvider{"db": &mockDatabase{}}}
	setupMockPool(ms)
	defer teardownMockPool()

	var res map[string]interface{}
	r := Mgo(&res, "upsert", map[string]interface{}{
		"Database":   "db",
		"Collection": "coll",
		"Selector":   map[string]interface{}{"_id": 123},
		"Change":     map[string]interface{}{"x": 1},
	})
	if !r.IsErr() {
		t.Error("expected error for invalid _id type")
	}
}

func TestMgoSrc_Usable(t *testing.T) {
	ms := &MgoSrc{}
	if ms.Usable() {
		t.Error("nil session should not be usable")
	}
}

func TestMgoSrc_Close(t *testing.T) {
	ms := &MgoSrc{}
	ms.Close()
}

func TestMgoSrc_Reset(t *testing.T) {
	ms := &MgoSrc{}
	ms.Reset()
}

func TestError(t *testing.T) {
	_ = Error()
}

func TestDatabaseNames(t *testing.T) {
	ms := &mockSession{
		usable:     true,
		dbNames:    []string{"db1"},
		dbNamesErr: nil,
	}
	setupMockPool(ms)
	defer teardownMockPool()

	r := DatabaseNames()
	if r.IsErr() {
		t.Fatalf("DatabaseNames: %v", r.UnwrapErr())
	}
	if len(r.Unwrap()) != 1 || r.Unwrap()[0] != "db1" {
		t.Errorf("DatabaseNames = %v", r.Unwrap())
	}
}

func TestDatabaseNames_Err(t *testing.T) {
	ms := &mockSession{
		usable:     true,
		dbNamesErr: errors.New("db names err"),
	}
	setupMockPool(ms)
	defer teardownMockPool()

	r := DatabaseNames()
	if !r.IsErr() {
		t.Error("expected DatabaseNames error")
	}
}

func TestCollectionNames_Err(t *testing.T) {
	ms := &mockSession{
		usable: true,
		dbs: map[string]dbProvider{
			"mydb": &mockDatabase{namesErr: errors.New("coll names err")},
		},
	}
	setupMockPool(ms)
	defer teardownMockPool()

	r := CollectionNames("mydb")
	if !r.IsErr() {
		t.Error("expected CollectionNames error")
	}
}

func TestCollectionNames(t *testing.T) {
	ms := &mockSession{
		usable: true,
		dbs: map[string]dbProvider{
			"mydb": &mockDatabase{names: []string{"c1", "c2"}},
		},
	}
	setupMockPool(ms)
	defer teardownMockPool()

	r := CollectionNames("mydb")
	if r.IsErr() {
		t.Fatalf("CollectionNames: %v", r.UnwrapErr())
	}
	if len(r.Unwrap()) != 2 {
		t.Errorf("CollectionNames = %v", r.Unwrap())
	}
}

func TestLen(t *testing.T) {
	ms := &mockSession{usable: true}
	setupMockPool(ms)
	defer teardownMockPool()

	if Len() != 0 {
		t.Errorf("Len = %d, want 0", Len())
	}
}

func TestClose(t *testing.T) {
	ms := &mockSession{usable: true}
	setupMockPool(ms)
	defer teardownMockPool()

	Close()
}

func TestList_DatabaseNamesErr(t *testing.T) {
	ms := &mockSession{
		usable:     true,
		dbNamesErr: errors.New("db names err"),
	}
	setupMockPool(ms)
	defer teardownMockPool()

	var li map[string][]string
	r := Mgo(&li, "list", map[string]interface{}{})
	if !r.IsErr() {
		t.Error("expected error from DatabaseNames")
	}
}

func TestList_CollectionNamesErr(t *testing.T) {
	ms := &mockSession{
		usable:  true,
		dbNames: []string{"db1"},
		dbs: map[string]dbProvider{
			"db1": &mockDatabase{namesErr: errors.New("coll err")},
		},
	}
	setupMockPool(ms)
	defer teardownMockPool()

	var li map[string][]string
	r := Mgo(&li, "list", map[string]interface{}{"Dbs": []string{"db1"}})
	if !r.IsErr() {
		t.Error("expected error from CollectionNames")
	}
}

func TestCount_QueryErr(t *testing.T) {
	ms := &mockSession{
		usable: true,
		dbs: map[string]dbProvider{
			"db": &mockDatabase{
				collections: map[string]collectionProvider{
					"coll": &mockCollection{
						findQuery: &mockQuery{countErr: errors.New("count err")},
					},
				},
			},
		},
	}
	setupMockPool(ms)
	defer teardownMockPool()

	var count int
	r := Mgo(&count, "count", map[string]interface{}{
		"Database": "db", "Collection": "coll", "Query": map[string]interface{}{},
	})
	if !r.IsErr() {
		t.Error("expected error from Count")
	}
}

func TestFind_CountErr(t *testing.T) {
	ms := &mockSession{
		usable: true,
		dbs: map[string]dbProvider{
			"db": &mockDatabase{
				collections: map[string]collectionProvider{
					"coll": &mockCollection{
						findQuery: &mockQuery{countErr: errors.New("count err")},
					},
				},
			},
		},
	}
	setupMockPool(ms)
	defer teardownMockPool()

	var res map[string]interface{}
	r := Mgo(&res, "find", map[string]interface{}{
		"Database": "db", "Collection": "coll", "Query": map[string]interface{}{},
	})
	if !r.IsErr() {
		t.Error("expected error from Find Count")
	}
}

func TestFind_AllErr(t *testing.T) {
	ms := &mockSession{
		usable: true,
		dbs: map[string]dbProvider{
			"db": &mockDatabase{
				collections: map[string]collectionProvider{
					"coll": &mockCollection{
						findQuery: &mockQuery{count: 0, allErr: errors.New("all err")},
					},
				},
			},
		},
	}
	setupMockPool(ms)
	defer teardownMockPool()

	var res map[string]interface{}
	r := Mgo(&res, "find", map[string]interface{}{
		"Database": "db", "Collection": "coll", "Query": map[string]interface{}{},
	})
	if !r.IsErr() {
		t.Error("expected error from Find All")
	}
}

func TestInsert_InsertErr(t *testing.T) {
	ms := &mockSession{
		usable: true,
		dbs: map[string]dbProvider{
			"db": &mockDatabase{
				collections: map[string]collectionProvider{
					"coll": &mockCollection{insertErr: errors.New("insert err")},
				},
			},
		},
	}
	setupMockPool(ms)
	defer teardownMockPool()

	var ids []string
	r := Mgo(&ids, "insert", map[string]interface{}{
		"Database": "db", "Collection": "coll",
		"Docs": []map[string]interface{}{{"x": 1}},
	})
	if !r.IsErr() {
		t.Error("expected error from Insert")
	}
}

func TestRemove_RemoveErr(t *testing.T) {
	ms := &mockSession{
		usable: true,
		dbs: map[string]dbProvider{
			"db": &mockDatabase{
				collections: map[string]collectionProvider{
					"coll": &mockCollection{removeErr: errors.New("remove err")},
				},
			},
		},
	}
	setupMockPool(ms)
	defer teardownMockPool()

	r := Mgo(nil, "remove", map[string]interface{}{
		"Database": "db", "Collection": "coll", "Selector": map[string]interface{}{},
	})
	if !r.IsErr() {
		t.Error("expected error from Remove")
	}
}

func TestUpdate_UpdateErr(t *testing.T) {
	ms := &mockSession{
		usable: true,
		dbs: map[string]dbProvider{
			"db": &mockDatabase{
				collections: map[string]collectionProvider{
					"coll": &mockCollection{updateErr: errors.New("update err")},
				},
			},
		},
	}
	setupMockPool(ms)
	defer teardownMockPool()

	r := Mgo(nil, "update", map[string]interface{}{
		"Database": "db", "Collection": "coll",
		"Selector": map[string]interface{}{}, "Change": map[string]interface{}{},
	})
	if !r.IsErr() {
		t.Error("expected error from Update")
	}
}

func TestUpdateAll_UpdateAllErr(t *testing.T) {
	ms := &mockSession{
		usable: true,
		dbs: map[string]dbProvider{
			"db": &mockDatabase{
				collections: map[string]collectionProvider{
					"coll": &mockCollection{updateAllErr: errors.New("updateall err")},
				},
			},
		},
	}
	setupMockPool(ms)
	defer teardownMockPool()

	var res map[string]interface{}
	r := Mgo(&res, "update_all", map[string]interface{}{
		"Database": "db", "Collection": "coll",
		"Selector": map[string]interface{}{}, "Change": map[string]interface{}{},
	})
	if !r.IsErr() {
		t.Error("expected error from UpdateAll")
	}
}

func TestUpsert_UpsertErr(t *testing.T) {
	ms := &mockSession{
		usable: true,
		dbs: map[string]dbProvider{
			"db": &mockDatabase{
				collections: map[string]collectionProvider{
					"coll": &mockCollection{upsertErr: errors.New("upsert err")},
				},
			},
		},
	}
	setupMockPool(ms)
	defer teardownMockPool()

	var res map[string]interface{}
	r := Mgo(&res, "upsert", map[string]interface{}{
		"Database": "db", "Collection": "coll",
		"Selector": map[string]interface{}{}, "Change": map[string]interface{}{},
	})
	if !r.IsErr() {
		t.Error("expected error from Upsert")
	}
}
