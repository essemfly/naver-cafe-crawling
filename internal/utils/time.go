package utils

import (
	"math/rand"
	"time"
)

// 요청 간 랜덤 지연을 위한 함수 (재사용)
func RandomSleep() {
	sleepTime := time.Duration(rand.Intn(2000)+1000) * time.Millisecond // 1-3초로 줄임
	time.Sleep(sleepTime)
}
