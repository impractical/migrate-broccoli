package migratebroccoli

import (
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"

	"aletheia.icu/broccoli/fs"
	"github.com/golang-migrate/migrate/v4/source"
)

//go:generate broccoli -src examples/migrations -o broccoli.gen.go -var broccoli

func init() {
	source.Register("broccoli", &Driver{})
}

var (
	ErrInvalidInstance = errors.New("Invalid parameter for WithInstance; must be a migratebroccoli.Config")
	ErrNilBroccoli     = errors.New("migratebroccoli.Config must have its Broccoli property set to the variable output by broccoli")
)

type Driver struct {
	path       string
	migrations *source.Migrations
	br         *fs.Broccoli
	root       string
}

type Config struct {
	Broccoli *fs.Broccoli
	Dir      string
}

func WithInstance(instance interface{}) (source.Driver, error) {
	conf, ok := instance.(Config)
	if !ok {
		return nil, ErrInvalidInstance
	}
	if conf.Broccoli == nil {
		return nil, ErrNilBroccoli
	}
	d := &Driver{
		path:       "<broccoli>",
		migrations: source.NewMigrations(),
		br:         conf.Broccoli,
		root:       conf.Dir,
	}
	err := conf.Broccoli.Walk(conf.Dir, func(path string, info os.FileInfo, err error) error {
		if path != conf.Dir && info.IsDir() {
			return nil
		}

		m, err := source.DefaultParse(filepath.Base(path))
		if err != nil {
			return nil
		}

		if !d.migrations.Append(m) {
			return fmt.Errorf("unable to parse file %v", path)
		}
		return nil
	})
	if err != nil {
		return nil, err
	}
	return d, nil
}

func (d *Driver) Open(url string) (source.Driver, error) {
	return nil, fmt.Errorf("not yet implemented")
}

func (d *Driver) Close() error {
	return nil
}

func (d *Driver) First() (version uint, err error) {
	v, ok := d.migrations.First()
	if !ok {
		return 0, &os.PathError{Op: "first", Path: d.path, Err: os.ErrNotExist}
	}
	return v, nil
}

func (d *Driver) Prev(version uint) (prevVersion uint, err error) {
	v, ok := d.migrations.Prev(version)
	if !ok {
		return 0, &os.PathError{Op: fmt.Sprintf("prev for version %d", version), Path: d.path, Err: os.ErrNotExist}
	}
	return v, nil
}

func (d *Driver) Next(version uint) (nextVersion uint, err error) {
	v, ok := d.migrations.Next(version)
	if !ok {
		return 0, &os.PathError{Op: fmt.Sprintf("next for version %d", version), Path: d.path, Err: os.ErrNotExist}
	}
	return v, nil
}

func (d *Driver) ReadUp(version uint) (r io.ReadCloser, identifier string, err error) {
	m, ok := d.migrations.Up(version)
	if !ok {
		return nil, "", &os.PathError{Op: fmt.Sprintf("read version %v", version), Path: d.path, Err: os.ErrNotExist}
	}
	file, err := d.br.Open(filepath.Join(d.root, m.Raw))
	if err != nil {
		return nil, "", err
	}
	return file, m.Identifier, nil
}

func (d *Driver) ReadDown(version uint) (r io.ReadCloser, identifier string, err error) {
	m, ok := d.migrations.Down(version)
	if !ok {
		return nil, "", &os.PathError{Op: fmt.Sprintf("read version %v", version), Path: d.path, Err: os.ErrNotExist}
	}
	file, err := d.br.Open(filepath.Join(d.root, m.Raw))
	if err != nil {
		return nil, "", err
	}
	return file, m.Identifier, nil
}
