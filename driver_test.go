package migratebroccoli

import (
	"testing"

	st "github.com/golang-migrate/migrate/v4/source/testing"
)

func Test(t *testing.T) {
	config := Config{
		Broccoli: broccoli,
		Dir:      "examples/migrations",
	}
	d, err := WithInstance(config)
	if err != nil {
		t.Fatal(err)
	}
	st.Test(t, d)
}
