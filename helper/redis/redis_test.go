package redis

import (
	"goil/util/encoding"
	"testing"
)

type Bean struct {
	Name string
	Age  int
}

func TestRedis(t *testing.T) {
	client := GetRedisClient("redis:127.0.0.1:6379/1")
	b := Bean{"Jim", 20}
	byts, _ := encoding.GobEncode(b)
	err := client.HSet("h1", "bean", string(byts))
	if err != nil {
		t.Error(err)
	}
	client.Expire("h1", 15*60)

	b1, _ := client.HGet("h1", "bean")

	b2, _ := encoding.GobDecode([]byte(b1))
	t.Error(b2)

}
