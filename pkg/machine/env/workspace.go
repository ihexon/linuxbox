package env

import "sync"

var (
	CustomHomeEnv     string
	CustomHomeEnvOnce sync.Once
)

func InitCustomHomeEnvOnce(value string) {
	CustomHomeEnvOnce.Do(func() {
		CustomHomeEnv = value
	})
}
