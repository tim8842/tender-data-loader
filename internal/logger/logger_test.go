package logger

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestInitLogger(t *testing.T) {
	// Создаём временную директорию для логов
	tempDir := t.TempDir()

	// Инициализируем логгер
	log, closer, err := InitLogger(tempDir, 100, 7, 30, true)
	assert.NoError(t, err)
	assert.NotNil(t, log)

	// Логируем тестовое сообщение
	log.Info("test log message")

	// Закрываем логгер, чтобы сбросить буферы
	err = log.Sync()
	assert.NoError(t, err)
	err = closer.Close()
	assert.NoError(t, err)
	// Проверяем, что файл лога был создан
	logFilePath := filepath.Join(tempDir, "app.log")
	_, err = os.Stat(logFilePath)
	assert.NoError(t, err, "log file should exist")

	// Можно дополнительно прочитать файл и проверить наличие записи (по желанию)
	content, err := os.ReadFile(logFilePath)
	assert.NoError(t, err)
	assert.Contains(t, string(content), "test log message")
}
