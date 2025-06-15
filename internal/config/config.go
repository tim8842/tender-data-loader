package config

import (
	"fmt"
	"os"

	"github.com/joho/godotenv"
)

type Config struct {
	MongoUser                                string
	MongoPassword                            string
	MongoHost                                string
	MongoPort                                string
	MongoDB                                  string
	Port                                     string
	UrlGetProxy                              string
	UrlZakupkiAgreementGetNumbersFirst       string
	UrlZakupkiAgreementGetNumbersSecond      string
	UrlZakupkiAgreementGetNumbersThird       string
	UrlZakupkiAgreementGetNumbersForth       string
	UrlZakupkiAgreementGetAgreegmentWeb      string
	UrlZakupkiAgreementGetAgreegmentShowHtml string
	UrlZakupkiAgreementGetCustomerWeb        string
}

func LoadConfig(fileToEnv string) (*Config, error) {
	err := godotenv.Load(fileToEnv)
	if err != nil {
		panic("can not initialize logger")
	}

	cfg := &Config{
		MongoUser:                                os.Getenv("MONGO_USER"),
		MongoPassword:                            os.Getenv("MONGO_PASSWORD"),
		MongoHost:                                os.Getenv("MONGO_HOST"),
		MongoPort:                                os.Getenv("MONGO_PORT"),
		MongoDB:                                  getOrDefault("MONGO_DB", "tenderdb"),
		Port:                                     getOrDefault("PORT", "8080"),
		UrlGetProxy:                              os.Getenv("URL_GET_PROXY"),
		UrlZakupkiAgreementGetNumbersFirst:       os.Getenv("URL_ZAKUPKI_AGREEMENT_GET_NUMBERS_FIRST"),
		UrlZakupkiAgreementGetNumbersSecond:      os.Getenv("URL_ZAKUPKI_AGREEMENT_GET_NUMBERS_SECOND"),
		UrlZakupkiAgreementGetNumbersThird:       os.Getenv("URL_ZAKUPKI_AGREEMENT_GET_NUMBERS_THIRD"),
		UrlZakupkiAgreementGetNumbersForth:       os.Getenv("URL_ZAKUPKI_AGREEMENT_GET_NUMBERS_FORTH"),
		UrlZakupkiAgreementGetAgreegmentWeb:      os.Getenv("URL_ZAKUPKI_AGREEMENT_GET_AGREEGMENT_WEB"),
		UrlZakupkiAgreementGetAgreegmentShowHtml: os.Getenv("URL_ZAKUPKI_AGREEMENT_GET_AGREEGMENT_SHOW_HTML"),
		UrlZakupkiAgreementGetCustomerWeb:        os.Getenv("URL_ZAKUPKI_AGREEMENT_GET_CUSTOMER_WEB"),
	}

	required := map[string]string{
		"MONGO_USER":     cfg.MongoUser,
		"MONGO_PASSWORD": cfg.MongoPassword,
		"MONGO_HOST":     cfg.MongoHost,
		"MONGO_PORT":     cfg.MongoPort,
		"URL_GET_PROXY":  cfg.UrlGetProxy,
		"URL_ZAKUPKI_AGREEMENT_GET_NUMBERS_FIRST":        cfg.UrlZakupkiAgreementGetNumbersFirst,
		"URL_ZAKUPKI_AGREEMENT_GET_NUMBERS_SECOND":       cfg.UrlZakupkiAgreementGetNumbersSecond,
		"URL_ZAKUPKI_AGREEMENT_GET_NUMBERS_THIRD":        cfg.UrlZakupkiAgreementGetNumbersThird,
		"URL_ZAKUPKI_AGREEMENT_GET_NUMBERS_FORTH":        cfg.UrlZakupkiAgreementGetNumbersForth,
		"URL_ZAKUPKI_AGREEMENT_GET_AGREEGMENT_WEB":       cfg.UrlZakupkiAgreementGetAgreegmentWeb,
		"URL_ZAKUPKI_AGREEMENT_GET_AGREEGMENT_SHOW_HTML": cfg.UrlZakupkiAgreementGetAgreegmentShowHtml,
		"URL_ZAKUPKI_AGREEMENT_GET_CUSTOMER_WEB":         cfg.UrlZakupkiAgreementGetCustomerWeb,
	}

	for key, val := range required {
		if val == "" {
			return nil, fmt.Errorf("отсутствует обязательная переменная окружения: %s", key)
		}
	}

	return cfg, nil
}

func getOrDefault(key, fallback string) string {
	v := os.Getenv(key)
	if v == "" {
		return fallback
	}
	return v
}
