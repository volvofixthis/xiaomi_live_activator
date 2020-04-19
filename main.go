package main

import "bitbucket.org/laki9/xiaomi_live_activator/client"

func main() {
	client := client.YiCameraClient{Address: "192.168.1.88:7878"}
	client.Connect()
	client.Authenticate()
	client.Stream(0)
}
