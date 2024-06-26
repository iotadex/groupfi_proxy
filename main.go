package main

import (
	"gproxy/api"
	"gproxy/config"
	"gproxy/daemon"
	"gproxy/gl"
	"gproxy/model"
	"gproxy/service"
	"gproxy/tools"
	"gproxy/wallet"
	"os"
)

func main() {
	if os.Args[len(os.Args)-1] != "-d" {
		tools.Aes.Input()
		os.Args = append(os.Args, "-d")
	}
	daemon.Background("./out.log", true)

	config.Load()
	setSeeds()

	gl.CreateLogFiles()

	model.ConnectToMysql(config.Db.Host, config.Db.Port, config.Db.DbName, config.Db.Usr, config.Db.Pwd)

	api.StartHttpServer(config.HttpPort)

	service.Start()

	daemon.WaitForKill()

	api.StopHttpServer()
}

func setSeeds() {
	seeds := tools.Aes.ReadRand()
	// set model's seeeds
	model.SetSeeds(seeds)
	// set service's seeds
	wallet.SetSeeds(seeds)
}
