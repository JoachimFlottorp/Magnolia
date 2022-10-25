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
	RabbitMQ struct {
		URI string `json:"uri"`
	} `json:"rmq"`
	Markov struct {
		HealthAddress string `json:"health_address"`
		HealthBind    int    `json:"health_bind"`
	} `json:"markov"`
	Http struct {
		Port int `json:"port"`
	} `json:"http"`
}
