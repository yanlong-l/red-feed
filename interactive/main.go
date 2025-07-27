package main

import "github.com/spf13/viper"

func main() {
	initViper()
	app := InitAPP()
	for _, c := range app.consumers {
		err := c.Start()
		if err != nil {
			panic(err)
		}
	}
	err := app.server.Serve()
	if err != nil {
		panic(err)
	}
}

func initViper() {
	viper.SetConfigFile("config/dev.yaml")
	err := viper.ReadInConfig()
	if err != nil {
		panic(err)
	}
}
