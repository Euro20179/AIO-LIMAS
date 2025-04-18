package main

import (
	"errors"
	"flag"
	"fmt"
	"net/http"
	"os"

	"aiolimas/accounts"
	api "aiolimas/api"
	lua_api "aiolimas/lua-api"
	"aiolimas/webservice/dynamic"
)

func setupAIODir() string {
	dir, envExists := os.LookupEnv("AIO_DIR")
	if !envExists {
		dataDir, envExists := os.LookupEnv("XDG_DATA_HOME")
		if !envExists {
			home, envEenvExists := os.LookupEnv("HOME")
			if !envEenvExists {
				panic("Could not setup aio directory, $HOME does not exist")
			}
			dataDir = fmt.Sprintf("%s/.local/share", home)
		}
		dir = fmt.Sprintf("%s/aio-limas", dataDir)
	}

	if _, err := os.Stat(dir); errors.Is(err, os.ErrNotExist) {
		os.MkdirAll(dir, 0o755)
	} else if err != nil {
		panic(fmt.Sprintf("Could not create directory %s\n%s", dir, err.Error()))
	}
	return dir
}

type EndPointMap map[string]func(http.ResponseWriter, *http.Request)

func startServer() {
	const apiRoot = "/api/v1"

	for k, v := range api.Endpoints {
		println(apiRoot + k)
		api.MakeEndPointsFromList(apiRoot + k, v)
	}

	api.MakeEndPointsFromList("/account", api.AccountEndPoints)

	http.HandleFunc("/docs", api.MainDocs.Listener)

	http.HandleFunc("/html/", dynamic.HtmlEndpoint)
	// http.HandleFunc("/", webservice.Root)

	port := os.Getenv("AIO_PORT")
	if port == "" {
		port = "8080"
	}

	http.ListenAndServe(fmt.Sprintf(":%s", port), nil)
}

func setEnvOrPanic(name string, val string) {
	if err := os.Setenv(name, val); err != nil {
		panic(err.Error())
	}
}

func initConfig(aioPath string) {
	configPath := aioPath + "/config.json"
	setEnvOrPanic("AIO_CONFIG_FILE", configPath)
	if _, err := os.Stat(configPath); err == nil {
		return
	}
	file, err := os.OpenFile(configPath, os.O_CREATE|os.O_WRONLY, 0o644)
	if err != nil {
		panic("Failed to create config file")
	}
	file.Write([]byte("{}"))
	if err := file.Close(); err != nil {
		panic("Failed to create config file, writing {}")
	}
}

func main() {
	aioPath := setupAIODir()
	setEnvOrPanic("AIO_DIR", aioPath)

	initConfig(aioPath)

	accounts.InitAccountsDb(aioPath)

	flag.Parse()

	inst, err := lua_api.InitGlobalLuaInstance("./lua-extensions/init.lua")
	if err != nil {
		panic("Could not initialize global lua instance")
	}
	lua_api.GlobalLuaInstance = inst

	startServer()
}
