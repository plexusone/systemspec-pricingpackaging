package schema_test

import (
	"bytes"
	"encoding/json"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"

	"github.com/grokify/oscompat/fs"
	"github.com/grokify/oscompat/testutil"
	jsonschema "github.com/santhosh-tekuri/jsonschema/v6"

	"github.com/plexusone/systemspec-pricingpackaging/schema"
)

func compile(t *testing.T) *jsonschema.Schema {
	t.Helper()
	doc, err := jsonschema.UnmarshalJSON(bytes.NewReader(schema.SpecJSON))
	if err != nil {
		t.Fatal(err)
	}
	c := jsonschema.NewCompiler()
	if err := c.AddResource("pricing.schema.json", doc); err != nil {
		t.Fatal(err)
	}
	sch, err := c.Compile("pricing.schema.json")
	if err != nil {
		t.Fatal(err)
	}
	return sch
}

func validate(t *testing.T, sch *jsonschema.Schema, data []byte) error {
	t.Helper()
	inst, err := jsonschema.UnmarshalJSON(bytes.NewReader(data))
	if err != nil {
		t.Fatal(err)
	}
	return sch.Validate(inst)
}

func TestExampleValidatesAgainstTheSchema(t *testing.T) {
	data, err := os.ReadFile("../testdata/valid.json")
	if err != nil {
		t.Fatal(err)
	}
	if err := validate(t, compile(t), data); err != nil {
		t.Fatalf("testdata/valid.json: %v", err)
	}
}

func TestExamplesValidateAgainstTheSchema(t *testing.T) {
	files, err := filepath.Glob("../examples/*.json")
	if err != nil || len(files) < 2 {
		t.Fatalf("examples: %v (%d files)", err, len(files))
	}
	sch := compile(t)
	for _, f := range files {
		data, err := os.ReadFile(f)
		if err != nil {
			t.Fatal(err)
		}
		if err := validate(t, sch, data); err != nil {
			t.Errorf("%s: %v", f, err)
		}
	}
}

func TestSchemaRejectsWhatTheTypesReject(t *testing.T) {
	sch := compile(t)
	data, _ := os.ReadFile("../testdata/valid.json")
	for name, bad := range map[string]string{
		"extra field":    strings.Replace(string(data), `"currency": "usd",`, `"currency": "usd", "surprise": 1,`, 1),
		"bad period":     strings.Replace(string(data), `"period": "month"`, `"period": "week"`, 1),
		"bad interval":   strings.Replace(string(data), `"interval": "year"`, `"interval": "week"`, 1),
		"zero amount":    strings.Replace(string(data), `"amountCents": 1900`, `"amountCents": 0`, 1),
		"fee over 10000": strings.Replace(string(data), `"platformFeeBps": 1000`, `"platformFeeBps": 10001`, 1),
		"format":         strings.Replace(string(data), `"formatVersion": "1"`, `"formatVersion": "2"`, 1),
	} {
		if err := validate(t, sch, []byte(bad)); err == nil {
			t.Errorf("%s: schema accepted it", name)
		}
	}
}

func TestSchemaIsCurrent(t *testing.T) {
	// Generate into a temp dir, then compare with the embedded copy. The
	// generator runs from the module so its imports resolve.
	out := t.TempDir()
	cmd := exec.Command("go", "run", filepath.Join("schema", "gen", "main.go"), out) //nolint:gosec // fixed generator path; out is this test's temp dir
	cmd.Dir = ".."
	if b, err := cmd.CombinedOutput(); err != nil {
		t.Fatalf("run the generator: %v\n%s", err, b)
	}
	got := testutil.ReadFileLF(t, filepath.Join(out, "pricing.schema.json"))
	if !bytes.Equal(got, fs.NormalizeLineEndingsBytes(schema.SpecJSON)) {
		t.Fatal("schema/pricing.schema.json is out of date; run go generate ./schema/...")
	}
	var v any
	if err := json.Unmarshal(got, &v); err != nil {
		t.Fatal(err)
	}
}
