package config

type HTTP struct {
	APIPrefix             string   `yaml:"apiPrefix,omitempty"`
	AccessLogFile         string   `yaml:"accessLogFile,omitempty"`
	LimitRateByIPEnable   bool     `yaml:"limitRateByIPEnable,omitempty"`
	LimitRateByIPInterval int64    `yaml:"limitRateByIPInterval,omitempty"`
	LimitRateByIPMaxNum   uint     `yaml:"limitRateByIPMaxNum,omitempty"`
	CORSEnable            bool     `yaml:"corsEnable,omitempty"`
	CORSAllowAllOrigins   bool     `yaml:"corsAllowAllOrigins,omitempty"`
	CORSAllowOrigins      []string `yaml:"corsAllowOrigins,omitempty"`
	CORSAllowMethods      []string `yaml:"corsAllowMethods,omitempty"`
	CORSAllowHeaders      []string `yaml:"corsAllowHeaders,omitempty"`
	CORSExposeHeaders     []string `yaml:"corsExposeHeaders,omitempty"`
	CORSAllowCredentials  bool     `yaml:"corsAllowCredentials,omitempty"`
	CORSMaxAge            int64    `yaml:"corsMaxAge,omitempty"`
}
