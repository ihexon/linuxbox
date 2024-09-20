package env

import "sync"

var (
	CustomHomeEnv     string
	CustomHomeEnvOnce sync.Once
)

func InitCustomHomeEnv(value string) {
	CustomHomeEnvOnce.Do(func() {
		CustomHomeEnv = value
	})
}
