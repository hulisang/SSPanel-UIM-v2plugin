package utility

import (
	"crypto/md5"
	"fmt"
	"github.com/shirou/gopsutil/host"
	"github.com/shirou/gopsutil/load"
	"math/rand"
	"time"
)

func InStr(s string, list []string) bool {
	for _, v := range list {
		if v == s {
			return true
		}
	}
	return false
}

func GetSystemLoad() string {
	stat, err := load.Avg()
	if err != nil {
		return "0.00 0.00 0.00"
	}

	return fmt.Sprintf("%.2f %.2f %.2f", stat.Load1, stat.Load5, stat.Load15)
}
func GetSystemUptime() string {
	time, err := host.Uptime()
	if err != nil {
		return ""
	}
	return fmt.Sprint(time)

}
func MD5(text string) []byte {
	ctx := md5.New()
	ctx.Write([]byte(text))
	return ctx.Sum(nil)
}

func GetRandomString(len1 int) string {
	str := "0123456789abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ"
	bytes := []byte(str)
	result := []byte{}
	r := rand.New(rand.NewSource(time.Now().UnixNano()))
	for i := 0; i < len1; i++ {
		result = append(result, bytes[r.Intn(len(bytes))])
	}
	return string(result)
}
