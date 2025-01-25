package conf

import (
	"github.com/spf13/viper"
	"github.com/sqc157400661/util"
	"time"
)

const (
	// 环境
	EnvDebug = "debug" // 开启了debug
	EnvDev   = "dev"   // 开发环境
	EnvTest  = "sit"   // 测试环境
	EnvProd  = "prod"  // 生产环境
)

var localConfig, commonConfig *viper.Viper

func Setup(path string, common ...string) {
	localConfig = viper.New()
	var err error
	err = loadFileConfig(localConfig, path)
	if err != nil {
		return
	}
	if len(common) > 0 {
		commonConfig = viper.New()
		err = loadFileConfig(commonConfig, common[0])
	}
	if err != nil {
		util.PrintFatalError(err)
	}
	return
}

func IsDev() bool {
	return GetEnv() == EnvDev
}

func IsProd() bool {
	return GetEnv() == EnvProd
}

func IsDebug() bool {
	return GetEnv() == EnvDebug
}

func GetEnv() string {
	return GetString("runMode")
}

func GetString(s string) string {
	return getConfigSource(s).GetString(s)
}

func Confer() *viper.Viper {
	return localConfig
}

// GetStringD GetString with a default value
func GetStringD(s string, d string) string {
	if GetString(s) == "" {
		return d
	}
	return GetString(s)
}

func UnmarshalKey(s string, rawVal interface{}) error {
	return getConfigSource(s).UnmarshalKey(s, rawVal)
}

func GetInt(s string) int {
	return getConfigSource(s).GetInt(s)
}

// GetIntD GetInt with a default value
func GetIntD(s string, d int) int {
	if GetInt(s) == 0 {
		return d
	}
	return GetInt(s)
}

func GetDuration(s string) time.Duration {
	return getConfigSource(s).GetDuration(s)
}

// GetDurationD GetDuration with a default value
func GetDurationD(s string, d time.Duration) time.Duration {
	if GetDuration(s) == 0 {
		return d
	}
	return GetDuration(s)
}

func GetStringSlice(s string) []string {
	return getConfigSource(s).GetStringSlice(s)
}

func GetStringMapString(s string) map[string]string {
	return getConfigSource(s).GetStringMapString(s)
}

func GetBool(s string) bool {
	return getConfigSource(s).GetBool(s)
}

func getConfigSource(s string) *viper.Viper {
	if localConfig.IsSet(s) {
		return localConfig
	}
	if commonConfig != nil {
		return commonConfig
	}
	return localConfig
}

func loadFileConfig(config *viper.Viper, file string) (err error) {
	config.SetConfigFile(file)
	//config.SetConfigType("yaml")
	err = config.ReadInConfig()
	if err != nil {
		return
	} else {
		// 注释掉了，热加载配置的逻辑
		//config.WatchConfig()
		//config.OnConfigChange(func(in fsnotify.Event) {
		//})
	}
	return
}
