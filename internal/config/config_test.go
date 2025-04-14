package config_test

import (
	"AvitoPVZ/internal/config"
	"github.com/stretchr/testify/suite"
	"os"
	"testing"
)

type ConfigSuite struct {
	suite.Suite
	tempConfigPath string
}

func (s *ConfigSuite) SetupSuite() {
	content := []byte(`
app:
  port: "8080"
  host: "localhost"
postgres:
  host: "localhost"
  port: 5432
  user: "user"
  password: "password"
  dbname: "dbname"
jwt:
  secret: "supersecret"
`)
	tmpFile, err := os.CreateTemp("", "config_test_*.yml")
	s.Require().NoError(err)
	_, err = tmpFile.Write(content)
	s.Require().NoError(err)
	err = tmpFile.Close()
	s.Require().NoError(err)

	s.tempConfigPath = tmpFile.Name()
}

func (s *ConfigSuite) TearDownSuite() {
	_ = os.Remove(s.tempConfigPath)
}

func (s *ConfigSuite) TestMustConfig_WithValidPath() {
	cfg := config.MustConfig(&s.tempConfigPath)

	s.Equal("8080", cfg.App.Port)
	s.Equal("localhost", cfg.App.Host)
	s.Equal("localhost", cfg.Postgres.Host)
	s.Equal(5432, cfg.Postgres.Port)
	s.Equal("user", cfg.Postgres.User)
	s.Equal("password", cfg.Postgres.Password)
	s.Equal("dbname", cfg.Postgres.DBName)
	s.Equal("supersecret", cfg.JWT.Secret)
}

func (s *ConfigSuite) TestPostgres_String() {
	pg := config.Postgres{
		Host:     "localhost",
		Port:     5432,
		User:     "user",
		Password: "pass",
		DBName:   "testdb",
	}

	connStr := pg.String()

	s.Contains(connStr, "postgres://user:pass@localhost:5432/testdb")
	s.Contains(connStr, "sslmode=disable")
}

func TestConfigSuite(t *testing.T) {
	suite.Run(t, new(ConfigSuite))
}
