package main

import (
	"github.com/spf13/viper"
)

func main() {
	// server := InitWebServer()
	// err := server.Run(":8080")
	// if err != nil {
	// 	return
	// }
	initViper()
	// fmt.Println(viper.GetString("redis.addr"))
	// fmt.Println(viper.GetString("db.dsn"))
}

// func initViper() {
// 	viper.SetConfigName("dev")
// 	viper.AddConfigPath("config")
// 	viper.SetConfigType("yaml")
// 	viper.ReadInConfig()
// }

// func initViper() {
// 	viper.SetConfigFile("config/dev.yaml")
// 	viper.ReadInConfig()
// }

// func initViper() {
// 	viper.SetConfigFile("config/dev.yaml")
// 	viper.ReadInConfig()
// 	type DBConfig struct {
// 		DSN string `yaml:"dsn"`
// 	}

// 	var db DBConfig
// 	viper.UnmarshalKey("db", &db)
// 	fmt.Println(db)
// }

func initViper() {
	viper.SetConfigFile("config/dev.yaml")
	viper.ReadInConfig()
}
