package schema_test

import (
	"bytes"
	"encoding/json"
	"os"
	"os/exec"
	"strings"
	"testing"

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
	cmd := exec.Command("go", "run", "gen/main.go")
	cmd.Dir = t.TempDir()
	// Generate into a temp dir, then compare with the embedded copy.
	src, err := os.Getwd()
	if err != nil {
		t.Fatal(err)
	}
	cmd.Args = []string{"go", "run", src + "/gen/main.go"}
	cmd.Dir = t.TempDir()
	if out, err := cmd.CombinedOutput(); err != nil {
		t.Skipf("cannot run the generator here: %v\n%s", err, out)
	}
	got, err := os.ReadFile(cmd.Dir + "/pricing.schema.json")
	if err != nil {
		t.Fatal(err)
	}
	if !bytes.Equal(got, schema.SpecJSON) {
		t.Fatal("schema/pricing.schema.json is out of date; run go generate ./schema/...")
	}
	var v any
	if err := json.Unmarshal(got, &v); err != nil {
		t.Fatal(err)
	}
}
