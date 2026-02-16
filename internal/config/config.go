package config

import "os"

type Config struct {
	HTTPAddr     string
	MySQLDSN     string
	RedisAddr    string
	RedisPass    string
	KafkaBrokers []string
	KafkaTopic   string
}

func Load() (Config, error) {
	cfg := Config{
		HTTPAddr:     getenv("HTTP_ADDR", ":8080"),
		MySQLDSN:     os.Getenv("MYSQL_DSN"),
		RedisAddr:    getenv("REDIS_ADDR", "127.0.0.1:6379"),
		RedisPass:    os.Getenv("REDIS_PASS"),
		KafkaTopic:   getenv("KAFKA_TOPIC", "mahjong.room.events"),
		KafkaBrokers: splitCSV(getenv("KAFKA_BROKERS", "127.0.0.1:9092")),
	}
	if cfg.MySQLDSN == "" {
		cfg.MySQLDSN = "inmemory"
	}
	return cfg, nil
}

func getenv(k, def string) string {
	v := os.Getenv(k)
	if v == "" {
		return def
	}
	return v
}

func splitCSV(raw string) []string {
	out := []string{}
	cur := ""
	for _, r := range raw {
		if r == ',' {
			if cur != "" {
				out = append(out, cur)
			}
			cur = ""
			continue
		}
		if r != ' ' {
			cur += string(r)
		}
	}
	if cur != "" {
		out = append(out, cur)
	}
	if len(out) == 0 {
		out = []string{"127.0.0.1:9092"}
	}
	return out
}
