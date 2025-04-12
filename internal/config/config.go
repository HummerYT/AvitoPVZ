package config

import (
	"context"
	"flag"
	"fmt"
	"net/url"
	"os"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/pkg/errors"

	"github.com/golang-migrate/migrate/v4"
	_ "github.com/golang-migrate/migrate/v4/database/postgres"
	_ "github.com/golang-migrate/migrate/v4/source/file"
	"github.com/ilyakaznacheev/cleanenv"
)

type Config struct {
	App      App      `yaml:"app"`
	Postgres Postgres `yaml:"postgres"`
	JWT      JWT      `yaml:"jwt"`
}

type App struct {
	Port string `yaml:"port"`
	Host string `yaml:"host"`
}

type Postgres struct {
	Host     string `yaml:"host"`
	Port     int    `yaml:"port"`
	User     string `yaml:"user"`
	Password string `yaml:"password"`
	DBName   string `yaml:"dbname"`
}

type JWT struct {
	Secret string `yaml:"secret"`
}

func New() *Config {
	return &Config{
		App:      App{},
		Postgres: Postgres{},
	}
}

func MustConfig(p *string) *Config {
	var path string
	if p == nil {
		path = fetchConfigPath()
	} else {
		path = *p
	}

	if path == "" {
		path = "./config.yml"
	}

	if _, ok := os.Stat(path); os.IsNotExist(ok) {
		panic("Config file does not exist: " + path)
	}

	cfg := New()

	if err := cleanenv.ReadConfig(path, cfg); err != nil {
		panic("failed to read config: " + err.Error())
	}

	return cfg
}

func NewPostgres(ctx context.Context, cfg Postgres) *pgxpool.Pool {
	pool, err := pgxpool.New(ctx, cfg.String())
	fmt.Println("config database: ", cfg.String())
	if err != nil {
		panic("no connect to database")
	}

	return pool
}

func fetchConfigPath() string {
	var res string

	// --config="path/to/config.yaml"
	flag.StringVar(&res, "config", "", "path to config")
	flag.Parse()

	if res == "" {
		res = os.Getenv("CONFIG_PATH")
	}

	return res
}

func (f App) String() string {
	return fmt.Sprintf("%s:%s", f.Host, f.Port)
}

func (p *Postgres) String() string {
	u := url.URL{
		Scheme: "postgres",
		User:   url.UserPassword(p.User, p.Password),
		Host:   fmt.Sprintf("%s:%d", p.Host, p.Port),
		Path:   p.DBName,
	}

	q := u.Query()
	q.Set("sslmode", "disable")

	u.RawQuery = q.Encode()

	return u.String()
}

func (p *Postgres) MigrationsUp(url ...string) error {
	var sourceURL string
	if url == nil {
		sourceURL = "file://internal/migrations/up"
	} else {
		sourceURL = url[0]
	}
	fmt.Println(p.String())
	m, err := migrate.New(sourceURL, p.String())
	if err != nil {
		return fmt.Errorf("migration failed: %w", err)
	}
	if err = m.Up(); err != nil {
		if errors.Is(err, migrate.ErrNoChange) {
			return nil
		}
		return errors.Wrap(err, "migration failed")
	}

	return nil
}
