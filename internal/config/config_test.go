package config

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
)

func setEnvVars() map[string]string {
	envVars := map[string]string{
		"MONGO_USER":     "test_user",
		"MONGO_PASSWORD": "test_password",
		"MONGO_HOST":     "localhost",
		"MONGO_PORT":     "27017",
		"URL_GET_PROXY":  "http://proxy.example.com",
		"URL_ZAKUPKI_AGREEMENT_GET_NUMBERS_FIRST":        "https://example.com/api/v1/first",
		"URL_ZAKUPKI_AGREEMENT_GET_NUMBERS_SECOND":       "https://example.com/api/v1/second",
		"URL_ZAKUPKI_AGREEMENT_GET_NUMBERS_THIRD":        "https://example.com/api/v1/third",
		"URL_ZAKUPKI_AGREEMENT_GET_NUMBERS_FORTH":        "https://example.com/api/v1/fourth",
		"URL_ZAKUPKI_AGREEMENT_GET_AGREEGMENT_WEB":       "https://example.com/web/agreement",
		"URL_ZAKUPKI_AGREEMENT_GET_AGREEGMENT_SHOW_HTML": "https://example.com/show/html",
		"URL_ZAKUPKI_AGREEMENT_GET_CUSTOMER_WEB":         "https://example.com/customer",
		"URL_ZAKUPKI_CONTRACT_GET_NUMBERS":               "https://example.com/customer",
		"URL_ZUKUPKI_CONTRACT_GET_WEB":                   "https://example.com/customer",
		"URL_ZUKUPKI_CONTRACT_GET_HTML":                  "https://example.com/customer",
		"URL_ZUKUPKI_CONTRACT_GET_CUSTOMER_WEB":          "https://example.com/customer",
		"URL_ZUKUPKI_CONTRACT_GET_CUSTOMER_WEB_ADD_INFO": "https://example.com/customer",
		"URL_PATCH_PROXY_USERS":                          "https://example.com/customer",
	}
	for k, v := range envVars {
		os.Setenv(k, v)
	}
	return envVars
}

// Очистка окружения после завершения теста
func clearEnvVars(keys []string) {
	for _, k := range keys {
		os.Unsetenv(k)
	}
}

// Тестируем успешную загрузку всех необходимых переменных
func TestLoadConfig_Successful(t *testing.T) {
	envVars := setEnvVars()
	defer clearEnvVars([]string{"MONGO_USER", "MONGO_PASSWORD", "MONGO_HOST", "MONGO_PORT", "URL_GET_PROXY"})
	envPath, err := filepath.Abs("../../configs/.env.test")
	if err != nil {
		panic(err)
	}
	cfg, err := LoadConfig(envPath)
	assert.Nil(t, err)
	assert.NotNil(t, cfg)

	// Проверим значения некоторых важных полей
	assert.Equal(t, envVars["MONGO_USER"], cfg.MongoUser)
	assert.Equal(t, envVars["MONGO_PASSWORD"], cfg.MongoPassword)
	assert.Equal(t, envVars["MONGO_HOST"], cfg.MongoHost)
	assert.Equal(t, envVars["MONGO_PORT"], cfg.MongoPort)
	assert.Equal(t, envVars["URL_GET_PROXY"], cfg.UrlGetProxy)
}

// Тестируем значение по умолчанию для MONGO_DB и PORT
func TestLoadConfig_DefaultValues(t *testing.T) {
	_ = setEnvVars()
	clearEnvVars([]string{"MONGO_DB", "PORT"})
	defer clearEnvVars([]string{"MONGO_USER", "MONGO_PASSWORD", "MONGO_HOST", "MONGO_PORT", "URL_GET_PROXY"})
	envPath, err := filepath.Abs("../../configs/.env.test")
	if err != nil {
		panic(err)
	}
	cfg, err := LoadConfig(envPath)
	assert.Nil(t, err)
	assert.Equal(t, "tenderdb", cfg.MongoDB)
	assert.Equal(t, "8080", cfg.Port)
}

// Тестируем случай, когда отсутствуют необходимые переменные
func TestLoadConfig_MissingRequiredVar(t *testing.T) {
	_, err := LoadConfig(".env.test.without.req")
	assert.Error(t, err)
	assert.True(t, strings.Contains(err.Error(), "отсутствует обязательная переменная"))
}
