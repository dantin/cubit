package storage

import (
	"testing"

	"github.com/dantin/cubit/storage/mysql"
	"github.com/stretchr/testify/require"
	yaml "gopkg.in/yaml.v2"
)

func TestStorageConfig(t *testing.T) {
	cfg := Config{}

	memCfg := `
  type: memory
`
	err := yaml.Unmarshal([]byte(memCfg), &cfg)
	require.Nil(t, err)
	require.Equal(t, Memory, cfg.Type)

	mySQLCfg := `
  type: mysql
  mysql:
    host: 127.0.0.1
    user: username
    password: password
    database: db
    pool_size: 16
`

	err = yaml.Unmarshal([]byte(mySQLCfg), &cfg)
	require.Nil(t, err)
	require.Equal(t, MySQL, cfg.Type)
	require.Equal(t, "127.0.0.1", cfg.MySQL.Host)
	require.Equal(t, "username", cfg.MySQL.User)
	require.Equal(t, "password", cfg.MySQL.Password)
	require.Equal(t, "db", cfg.MySQL.Database)
	require.Equal(t, 16, cfg.MySQL.PoolSize)

	mySQLCfg2 := `
  type: mysql
  mysql:
    host: 127.0.0.1
    user: username
    password: password
    database: db
`

	err = yaml.Unmarshal([]byte(mySQLCfg2), &cfg)
	require.Nil(t, err)
	require.Equal(t, MySQL, cfg.Type)
	require.Equal(t, mysql.DefaultPoolSize, cfg.MySQL.PoolSize)

	invalidMySQLCfg := `
  type: mysql
`

	err = yaml.Unmarshal([]byte(invalidMySQLCfg), &cfg)
	require.NotNil(t, err)
}

func TestStorageBadConfig(t *testing.T) {
	cfg := Config{}

	invalidCfg := `
  type: invalid`

	err := yaml.Unmarshal([]byte(invalidCfg), &cfg)
	require.NotNil(t, err)

	badCfg := `
  type`

	err = yaml.Unmarshal([]byte(badCfg), &cfg)
	require.NotNil(t, err)
}
