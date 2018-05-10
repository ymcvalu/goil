package session

var ExpireDuration int64 = 15 * 60 //s
var GCDuration int64 = 5 * 60      //s

var RedisAddr string = "redis:127.0.0.1:6379"
var ClientSidTag string = "goil_sid"
