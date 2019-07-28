package trace

import (
	"errors"
	"net"
	"os"
)

const (
	localIPEnvironName = "TRACE_LOCAL_IP"
)

var (
	localIP          net.IP
	errFailToGuessIP error
)

func init() {
	var err error
	defer func() {
		if localIP == nil {
			localIP = net.IPv4(127, 0, 0, 1)

			if err == nil {
				errFailToGuessIP = errors.New("fail to guess local ip")
			}
		}
	}()

	// 优先使用环境变量里面设置的 IP。
	if ipstr, ok := os.LookupEnv(localIPEnvironName); ok {
		ip := net.ParseIP(ipstr)

		if ip != nil && ip.To4() != nil {
			localIP = ip.To4()
			return
		}
	}

	addrs, err := net.InterfaceAddrs()

	if err != nil {
		errFailToGuessIP = err
		return
	}

	for _, addr := range addrs {
		if ip, ok := addr.(*net.IPNet); ok && !ip.IP.IsLoopback() {
			if ipv4 := ip.IP.To4(); ipv4 != nil {
				localIP = ip.IP.To4()
				break
			}
		}
	}
}

// GuessIP 猜测本机 IP 地址。
func GuessIP() (net.IP, error) {
	return localIP, errFailToGuessIP
}
