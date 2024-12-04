package main

import (
	"github.com/spf13/viper"
	_ "github.com/spf13/viper/remote"
)

func main() {

	initViper()
	server := InitWebServer()
	err := server.Run(":8080")
	if err != nil {
		return
	}
}

// func initViper() {
// 	viper.SetConfigName("dev")
// 	viper.AddConfigPath("config")
// 	viper.SetConfigType("yaml")
// 	viper.ReadInConfig()
// }

func initViper() {
	viper.SetConfigFile("config/dev.yaml")
	viper.ReadInConfig()
}

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

// func initViper() {
// 	viper.SetConfigType("yaml")
// 	cfgStr := `
// db:
//   dsn: "root:root@tcp(127.0.0.1:3306)/db?charset=utf8mb4&parseTime=True&loc=Local"
// redis:
//   addr: "127.0.0.1:6379"
// `
// 	err := viper.ReadConfig(strings.NewReader(cfgStr))
// 	if err != nil {
// 		panic(err)
// 	}
// }

// func initViper() {
// 	viper.SetDefault("db.dsn", "root:root@tcp(127.0.0.1:3306)/db?charset=utf8mb4&parseTime=True&loc=Local")
// 	viper.SetDefault("redis.addr", "127.0.0.1:6379")
// 	viper.SetConfigType("yaml")
// 	cfgStr := `
// `
// 	err := viper.ReadConfig(strings.NewReader(cfgStr))
// 	if err != nil {
// 		panic(err)
// 	}
// }

// func initViper() {
// 	pflagCfg := pflag.String("config", "config/dev.yaml", "配置文件路径")
// 	pflag.Parse()
// 	viper.SetConfigFile(*pflagCfg)
// 	err := viper.ReadInConfig()
// 	if err != nil {
// 		panic(err)
// 	}
// }

// func initViper() {
// 	err := viper.AddRemoteProvider("etcd3", "http://127.0.0.1:12379", "/webook")
// 	if err != nil {
// 		panic(err)
// 	}
// 	viper.SetConfigType("yaml")
// 	viper.ReadRemoteConfig()
// }

// func initViper() {
// 	viper.SetConfigFile("config/dev.yaml")
// 	viper.ReadInConfig()
// 	viper.WatchConfig()
// 	viper.OnConfigChange(func(in fsnotify.Event) {
// 		fmt.Println("配置文件发生变化")
// 		fmt.Println(in.Name)
// 		fmt.Println(in.Op)
// 	})
// }

// func initViper() {
// 	err := viper.AddRemoteProvider("etcd3", "http://127.0.0.1:12379", "/webook")
// 	if err != nil {
// 		panic(err)
// 	}
// 	viper.SetConfigType("yaml")
// 	err = viper.WatchRemoteConfig()
// 	if err != nil {
// 		panic(err)
// 	}
// 	fmt.Println("All settings")
// 	fmt.Println(viper.AllSettings())
// 	viper.OnConfigChange(func(in fsnotify.Event) {
// 		fmt.Println("配置文件发生变化")
// 		fmt.Println(in.Name)
// 		fmt.Println(in.Op)
// 	})
// 	go func() {
// 		err = viper.WatchRemoteConfig()
// 		if err != nil {
// 			panic(err)
// 		}
// 		err = viper.ReadRemoteConfig()
// 		if err != nil {
// 			panic(err)
// 		}
// 	}()

// }
