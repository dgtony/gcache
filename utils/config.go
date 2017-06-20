package utils

import (
	"github.com/BurntSushi/toml"
	"os"
)

type Config struct {
	General     GeneralSettings     `toml:"general"`
	Storage     StorageSettings     `toml:"storage"`
	Replication ReplicationSettings `toml:"replication"`
	ClientHTTP  ClientHTTPSettings  `toml:"client-HTTP"`
}

type GeneralSettings struct {
	LogLevel  string `toml:"log_level"`
	LogFormat string `toml:"log_format"`
	LogOut    string `toml:"log_out"`
}

type StorageSettings struct {
	NumShards               int `toml:"shards"`
	ExpiredKeyCheckInterval int `toml:"key_exp_check_interval"`
}

type ReplicationSettings struct {
	NodeRole             string `toml:"node_role"`
	RestoreCacheFromFile bool   `toml:"restore_from_file"`
	SaveCacheToFile      bool   `toml:"save_to_file"`
	CacheFile            string `toml:"cache_file"`
	FileWritePeriod      int    `toml:"file_write_period"`
	DumpUpdatePeriod     int    `toml:"dump_update_period"`
	MasterAddr           string `toml:"master_addr"`
	MasterSecret         string `toml:"master_secret"`
}

type ClientHTTPSettings struct {
	Addr        string `toml:"address"`
	Port        string `toml:"port"`
	RoutePrefix string `toml:"prefix"`
	IdleTimeout int    `toml:"idle_timeout"`
}

func ReadConfig(configFile string) (*Config, error) {
	_, err := os.Stat(configFile)
	if err != nil {
		return nil, err
	}

	var config Config
	if _, err := toml.DecodeFile(configFile, &config); err != nil {
		return &Config{}, err
	}
	return &config, nil
}
