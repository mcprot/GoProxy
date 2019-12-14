package main
func getDestAddr(addr string) string {
	if addr == "127.0.0.1" {
		return "127.0.0.1:25566"
	}
	if addr == "localtest.foxorsomething.net" {
		return "127.0.0.1:25567"
	}
	return ""
}
