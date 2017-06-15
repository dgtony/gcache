package utils

import (
	"github.com/BurntSushi/toml"
	"os"
)

type Config struct {
	General     GeneralSettings     `toml:"general"`
	Storage     StorageSettings     `toml:"storage`
	Replication ReplicationSettings `toml:"replication"`

	//General    GeneralInfo    `toml:"general"`
	//Swagger    SwaggerInfo    `toml:"swagger"`
	//HTTP       HTTPInfo       `toml:"HTTP"`
	//MQTT       MQTTInfo       `toml:"MQTT"`
	//Monitoring MonitoringInfo `toml:"monitoring"`
}

type GeneralSettings struct {
	LogLevel  string `toml:"log_level"`
	LogFormat string `toml:"log_format"`
	LogOut    string `toml:"log_out"`
}

type StorageSettings struct {
	NumShards               int `toml:"shards`
	ExpiredKeyCheckInterval int `toml:"key_exp_check_interval`
}

type ReplicationSettings struct {
	RestoreCacheFromFile bool   `toml:"restore_from_file"`
	SaveCacheToFile      bool   `toml:"save_to_file"`
	CacheFile            string `toml:"cache_file"`
	FileWritePeriod      int    `toml:"file_write_period"`
	DumpUpdatePeriod     int    `toml:"dump_period"`
	MasterAddr           string `toml:"master_addr"`
	MasterSecret         string `toml:"master_secret"`
}

/*

type StorageInfo struct {
    NumShards          int    `toml:"shards"`
    ContentServerAddr  string `toml:"content_addr"`
    UserDeviceEndpoint string `toml:"user_device_endpoint"`
    DeviceInfoEndpoint string `toml:"device_info_endpoint"`
    DeviceInfoLifeTime int    `toml:"device_info_lifetime"`
    SessionCleanPeriod int    `toml:"session_clean_period"`
    SessionLifeTime    int    `toml:"session_lifetime"`
}

type SwaggerInfo struct {
    Expose bool   `toml:"expose"`
    Route  string `toml:"route"`
    File   string `toml:"file"`
}

type HTTPInfo struct {
    Addr        string `toml:"address"`
    Port        int16  `toml:"port"`
    RoutePrefix string `toml:"prefix"`
    Timeout     int    `toml:"timeout"`
}

type MQTTInfo struct {
    BrokerAddr string `toml:"broker_addr"`
    PortMqtt   int    `toml:"port_mqtt"`
    PortMqtts  int    `toml:"port_mqtts"`
    UseMqtts   bool   `toml:"use_mqtts"`
    Username   string `toml:"username"`
    Password   string `toml:"password"`
    Keepalive  int    `toml:"keepalive"`
    Topic      string `toml:"topic"`
}

type MonitoringInfo struct {
    ExposeStats     bool   `toml:"expose_stats"`
    PrometheusRoute string `toml:"route"`
}

*/

func ReadConfig(configFile string) (*Config, error) {
	_, err := os.Stat(configFile)
	if err != nil {
		//return &Config{}, err
		return nil, err
	}

	var config Config
	if _, err := toml.DecodeFile(configFile, &config); err != nil {
		return &Config{}, err
	}
	return &config, nil
}
