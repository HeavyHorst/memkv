package memkv

import (
	"reflect"
	"testing"
)

var gettests = []struct {
	key   string
	value string
	want  KVPair
}{
	{"/db/user", "admin", KVPair{"/db/user", "admin"}},
	{"/db/pass", "foo", KVPair{"/db/pass", "foo"}},
	{"/missing", "", KVPair{}},
}

func TestGet(t *testing.T) {
	for _, tt := range gettests {
		s := New()
		if tt.value != "" {
			s.Set(tt.key, tt.value)
		}
		got := s.Get(tt.key)
		if got != tt.want {
			t.Errorf("Get(%q) = %v, want %v", tt.key, got, tt.want)
		}
	}
}

var getvtests = []struct {
	key   string
	value string
	want  string
}{
	{"/db/user", "admin", "admin"},
	{"/db/pass", "foo", "foo"},
}

func TestGetValue(t *testing.T) {
	for _, tt := range getvtests {
		s := New()
		s.Set(tt.key, tt.value)

		got := s.GetValue(tt.key)
		if got != tt.want {
			t.Errorf("Get(%q) = %v, want %v", tt.key, got, tt.want)
		}
	}
}

func TestGetValueWithDefault(t *testing.T) {
	want := "defaultValue"
	s := New()
	got := s.GetValue("/db/user", "defaultValue")
	if got != want {
		t.Errorf("want %v, got %v", want, got)
	}
}

var getalltestinput = map[string]string{
	"/app/db/pass":               "foo",
	"/app/db/user":               "admin",
	"/app/port":                  "443",
	"/app/url":                   "app.example.com",
	"/app/vhosts/host1":          "app.example.com",
	"/app/upstream/host1":        "203.0.113.0.1:8080",
	"/app/upstream/host1/domain": "app.example.com",
	"/app/upstream/host2":        "203.0.113.0.2:8080",
	"/app/upstream/host2/domain": "app.example.com",
}

var getalltests = []struct {
	pattern string
	want    []KVPair
}{
	{"/app/db/*",
		[]KVPair{
			KVPair{"/app/db/pass", "foo"},
			KVPair{"/app/db/user", "admin"}}},
	{"/app/*/host1",
		[]KVPair{
			KVPair{"/app/upstream/host1", "203.0.113.0.1:8080"},
			KVPair{"/app/vhosts/host1", "app.example.com"}}},

	{"/app/upstream/*",
		[]KVPair{
			KVPair{"/app/upstream/host1", "203.0.113.0.1:8080"},
			KVPair{"/app/upstream/host2", "203.0.113.0.2:8080"}}},
	{"[]a]", nil},
}

func TestGetAll(t *testing.T) {
	s := New()
	for k, v := range getalltestinput {
		s.Set(k, v)
	}
	for _, tt := range getalltests {
		got := s.GetAll(tt.pattern)
		if !reflect.DeepEqual([]KVPair(got), []KVPair(tt.want)) {
			t.Errorf("GetAll(%q) = %v, want %v", tt.pattern, got, tt.want)
		}
	}
}

func TestDel(t *testing.T) {
	s := New()
	s.Set("/app/port", "8080")
	want := KVPair{"/app/port", "8080"}
	got := s.Get("/app/port")
	if got != want {
		t.Errorf("Get(%q) = %v, want %v", "/app/port", got, want)
	}
	s.Del("/app/port")
	want = KVPair{}
	got = s.Get("/app/port")
	if got != want {
		t.Errorf("Get(%q) = %v, want %v", "/app/port", got, want)
	}
}

func TestPurge(t *testing.T) {
	s := New()
	s.Set("/app/port", "8080")
	want := KVPair{"/app/port", "8080"}
	got := s.Get("/app/port")
	if got != want {
		t.Errorf("Get(%q) = %v, want %v", "/app/port", got, want)
	}
	s.Purge()
	want = KVPair{}
	got = s.Get("/app/port")
	if got != want {
		t.Errorf("Get(%q) = %v, want %v", "/app/port", got, want)
	}
	s.Set("/app/port", "8080")
	want = KVPair{"/app/port", "8080"}
	got = s.Get("/app/port")
	if got != want {
		t.Errorf("Get(%q) = %v, want %v", "/app/port", got, want)
	}
}

var listTestMap = map[string]string{
	"/deis/database/user":             "user",
	"/deis/database/pass":             "pass",
	"/deis/services/key":              "value",
	"/deis/services/notaservice/foo":  "bar",
	"/deis/services/srv1/node1":       "10.244.1.1:80",
	"/deis/services/srv1/node2":       "10.244.1.2:80",
	"/deis/services/srv1/node3":       "10.244.1.3:80",
	"/deis/services/srv2/node1":       "10.244.2.1:80",
	"/deis/services/srv2/node2":       "10.244.2.2:80",
	"/deis/prefix/node1":              "prefix_node1",
	"/deis/prefix/node2/leafnode":     "prefix_node2",
	"/deis/prefix/node3/leafnode":     "prefix_node3",
	"/deis/prefix_a/node4":            "prefix_a_node4",
	"/deis/prefixb/node5/leafnode":    "prefixb_node5",
	"/deis/dirprefix/node1":           "prefix_node1",
	"/deis/dirprefix/node2/leafnode":  "prefix_node2",
	"/deis/dirprefix/node3/leafnode":  "prefix_node3",
	"/deis/dirprefix_a/node4":         "prefix_a_node4",
	"/deis/dirprefixb/node5/leafnode": "prefixb_node5",
	"/deis/prefix/node2/sub1/leaf1":   "prefix_node2_sub1_leaf1",
	"/deis/prefix/node2/sub1/leaf2":   "prefix_node2_sub1_leaf2",
}

func testList(t *testing.T, want, paths []string, dir bool) {
	var got []string
	s := New()
	for k, v := range listTestMap {
		s.Set(k, v)
	}
	for _, filePath := range paths {
		if dir {
			got = s.ListDir(filePath)
		} else {
			got = s.List(filePath)
		}
		if !reflect.DeepEqual(got, want) {
			t.Errorf("List(%s) = %v, want %v", filePath, got, want)
		}
	}
}

func TestList(t *testing.T) {
	want := []string{"key", "notaservice", "srv1", "srv2"}
	paths := []string{
		"/deis/services",
		"/deis/services/",
	}
	testList(t, want, paths, false)
}

func TestListForFile(t *testing.T) {
	want := []string{}
	paths := []string{"/deis/services/key"}
	testList(t, want, paths, false)
}

func TestListDir(t *testing.T) {
	want := []string{"notaservice", "srv1", "srv2"}
	paths := []string{
		"/deis/services",
		"/deis/services/",
	}
	testList(t, want, paths, true)
}

func TestListForSamePrefix(t *testing.T) {
	want := []string{"node1", "node2", "node3"}
	paths := []string{
		"/deis/prefix",
		"/deis/prefix/",
	}
	testList(t, want, paths, false)
}

func TestListDirForSamePrefix(t *testing.T) {
	want := []string{"node2", "node3"}
	paths := []string{
		"/deis/dirprefix",
		"/deis/dirprefix/",
	}
	testList(t, want, paths, true)
}

func TestListForMixedLeafSubnodes(t *testing.T) {
	want := []string{"leaf1", "leaf2"}
	paths := []string{"/deis/prefix/node2/sub1"}
	testList(t, want, paths, false)
}
