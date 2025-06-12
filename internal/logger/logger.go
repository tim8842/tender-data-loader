package logger

import (
	"io"
	"os"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"gopkg.in/natefinch/lumberjack.v2"
)

func InitLogger(logDir string) (*zap.Logger, io.Closer, error) {
	// Настраиваем lumberjack — файл логов + ротация
	lumberjackLogger := &lumberjack.Logger{
		Filename:   logDir + "/app.log", // путь к файлу
		MaxSize:    100,                 // Мб до ротации
		MaxBackups: 7,                   // сколько резервных файлов хранить
		MaxAge:     30,                  // дней хранить логи
		Compress:   true,                // сжимать старые логи
	}

	// Создаём Writer, который пишет в lumberjack
	fileWriteSyncer := zapcore.AddSync(lumberjackLogger)

	// Создаём Writer для консоли (stdout)
	consoleWriteSyncer := zapcore.AddSync(os.Stdout)

	// Настраиваем Encoder
	encoderConfig := zap.NewProductionEncoderConfig()
	encoderConfig.TimeKey = "timestamp"
	encoderConfig.EncodeTime = zapcore.ISO8601TimeEncoder

	// Создаём Encoder для JSON (файловый лог)
	jsonEncoder := zapcore.NewJSONEncoder(encoderConfig)

	// Создаём Encoder для Console (консольный лог)
	consoleEncoderConfig := zap.NewProductionEncoderConfig() // Можно использовать другую конфигурацию для консоли
	consoleEncoderConfig.EncodeTime = zapcore.ISO8601TimeEncoder
	consoleEncoder := zapcore.NewConsoleEncoder(consoleEncoderConfig) // Используем ConsoleEncoder для более читаемого вывода

	// Создаём cores (ядра логирования) для файла и консоли
	fileCore := zapcore.NewCore(jsonEncoder, fileWriteSyncer, zap.DebugLevel)          // Логи в файл
	consoleCore := zapcore.NewCore(consoleEncoder, consoleWriteSyncer, zap.DebugLevel) // Логи в консоль

	// Объединяем cores в один Tee (разветвитель)
	core := zapcore.NewTee(fileCore, consoleCore)

	logger := zap.New(core)
	return logger, lumberjackLogger, nil
}
