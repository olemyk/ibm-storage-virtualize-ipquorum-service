package jar

import (
	"archive/zip"
	"bytes"
	"io"
	"strings"
	"testing"

	"github.com/olemyk/ipquorum-emulator/internal/config"
)

// templateJar builds a small zip that looks like the fake jar template.
func templateJar(t *testing.T, extra ...string) []byte {
	t.Helper()
	var buf bytes.Buffer
	zw := zip.NewWriter(&buf)
	entries := append([]string{"META-INF/MANIFEST.MF", "ipquorum/FakeIpQuorum.class"}, extra...)
	for _, name := range entries {
		w, err := zw.Create(name)
		if err != nil {
			t.Fatal(err)
		}
		io.WriteString(w, "content of "+name)
	}
	if err := zw.Close(); err != nil {
		t.Fatal(err)
	}
	return buf.Bytes()
}

func cluster(name string, addrs ...string) config.Cluster {
	cl := config.Cluster{Name: name}
	for _, a := range addrs {
		cl.Nodes = append(cl.Nodes, config.Node{Address: a})
	}
	return cl
}

func readEntry(t *testing.T, data []byte, name string) (string, bool) {
	t.Helper()
	zr, err := zip.NewReader(bytes.NewReader(data), int64(len(data)))
	if err != nil {
		t.Fatalf("output is not a valid zip: %v", err)
	}
	for _, f := range zr.File {
		if f.Name == name {
			rc, _ := f.Open()
			defer rc.Close()
			b, _ := io.ReadAll(rc)
			return string(b), true
		}
	}
	return "", false
}

func TestBuildPerClusterJARsDiffer(t *testing.T) {
	b, err := NewBuilder(templateJar(t))
	if err != nil {
		t.Fatal(err)
	}
	a, err := b.Build(cluster("flash-a", "10.0.0.1:1260", "10.0.0.2:1260"))
	if err != nil {
		t.Fatal(err)
	}
	c, err := b.Build(cluster("flash-b", "10.9.9.1:1260"))
	if err != nil {
		t.Fatal(err)
	}
	if bytes.Equal(a, c) {
		t.Fatal("jars for different clusters must differ")
	}

	propsA, ok := readEntry(t, a, ConfigEntry)
	if !ok || !strings.Contains(propsA, "nodes=10.0.0.1:1260,10.0.0.2:1260") || !strings.Contains(propsA, "system=flash-a") {
		t.Errorf("flash-a properties: %q", propsA)
	}
	propsB, _ := readEntry(t, c, ConfigEntry)
	if strings.Contains(propsB, "10.0.0.1") || !strings.Contains(propsB, "nodes=10.9.9.1:1260") {
		t.Errorf("flash-b properties: %q", propsB)
	}
}

func TestBuildKeepsTemplateEntriesInOrder(t *testing.T) {
	b, _ := NewBuilder(templateJar(t))
	out, err := b.Build(cluster("a", "h:1"))
	if err != nil {
		t.Fatal(err)
	}
	zr, _ := zip.NewReader(bytes.NewReader(out), int64(len(out)))
	var names []string
	for _, f := range zr.File {
		names = append(names, f.Name)
	}
	want := []string{"META-INF/MANIFEST.MF", "ipquorum/FakeIpQuorum.class", ConfigEntry}
	if strings.Join(names, ",") != strings.Join(want, ",") {
		t.Errorf("entries = %v, want %v (manifest must stay first)", names, want)
	}
	if got, _ := readEntry(t, out, "ipquorum/FakeIpQuorum.class"); got != "content of ipquorum/FakeIpQuorum.class" {
		t.Errorf("template content altered: %q", got)
	}
}

func TestBuildCachesAndDoesNotMutateTemplate(t *testing.T) {
	tmpl := templateJar(t)
	orig := append([]byte(nil), tmpl...)
	b, _ := NewBuilder(tmpl)
	first, _ := b.Build(cluster("a", "h:1"))
	second, _ := b.Build(cluster("a", "h:1"))
	if !bytes.Equal(first, second) {
		t.Error("same cluster must give identical jars")
	}
	if !bytes.Equal(tmpl, orig) {
		t.Error("template was modified")
	}
}

func TestNewBuilderRejectsBadTemplates(t *testing.T) {
	if _, err := NewBuilder([]byte("not a zip")); err == nil {
		t.Error("expected error for non-zip template")
	}
	if _, err := NewBuilder(templateJar(t, ConfigEntry)); err == nil {
		t.Error("expected error when template already has quorum.properties")
	}
}
