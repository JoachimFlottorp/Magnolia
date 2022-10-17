package config

type Config struct {
	Redis struct {
		Username string `json:"username"`
		Password string `json:"password"`
		Database int    `json:"database"`
		Address  string `json:"address"`
	} `json:"redis"`
	Mongo struct {
		Username string `json:"username"`
		Password string `json:"password"`
		Address  string `json:"address"`
		SRV      bool   `json:"srv"`
		DB       string `json:"db"`
	}
	Http struct {
		Port int `json:"port"`
	} `json:"http"`
}
