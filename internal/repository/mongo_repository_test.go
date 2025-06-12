package repository_test

import (
	"context"
	"path/filepath"
	"testing"
	"time"

	"github.com/joho/godotenv"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/tim8842/tender-data-loader/internal/config"
	"github.com/tim8842/tender-data-loader/internal/model"
	"github.com/tim8842/tender-data-loader/internal/repository"
	"github.com/tim8842/tender-data-loader/internal/setup"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.uber.org/zap"
)

type TestEntity struct {
	ID    string `bson:"_id" json:"id"`
	Value string `bson:"value" json:"value"`
}

func (t TestEntity) GetID() any {
	return t.ID
}

func setupTestRepo(t *testing.T) (*mongo.Client, *mongo.Database, *repository.GenericRepository[*model.Variable]) {
	logger, _ := zap.NewDevelopment()
	ctx, _ := context.WithTimeout(context.Background(), 10*time.Second)
	envPath, err := filepath.Abs("../../.env.test")
	if err != nil {
		panic(err)
	}

	if err := godotenv.Load(envPath); err != nil {
		panic("Error loading .env file")
	}
	cfg, _ := config.LoadConfig()
	// Подключение к тестовой БД (замени на реальный вызов подключения)
	client, db, err := setup.SetupTestMongo(ctx, logger, &setup.MongoConfig{
		User: cfg.MongoUser, Password: cfg.MongoPassword,
		Host: cfg.MongoHost, Port: cfg.MongoPort,
		DBName: cfg.MongoDB,
	})
	require.NoError(t, err)
	// Очистка тестовой БД по окончании теста
	t.Cleanup(func() {
		err := db.Drop(ctx)
		assert.NoError(t, err)
		client.Disconnect(ctx)
	})

	coll := db.Collection("variables")
	repo := repository.NewGenericRepository[*model.Variable](coll, logger)

	// Очистить коллекцию перед каждым тестом
	require.NoError(t, coll.Drop(ctx))

	return client, db, repo
}

func TestGenericRepository_GetByID(t *testing.T) {
	ctx := context.Background()
	_, _, repo := setupTestRepo(t)

	tests := []struct {
		name        string
		setupDoc    *model.Variable
		searchID    string
		wantErr     bool
		expectEmpty bool
	}{
		{
			name: "found document",
			setupDoc: &model.Variable{
				ID:   "id123",
				Vars: map[string]interface{}{"foo": "bar"},
			},
			searchID: "id123",
			wantErr:  false,
		},
		{
			name:        "not found",
			setupDoc:    nil,
			searchID:    "nonexistent",
			wantErr:     true,
			expectEmpty: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Перед тестом вставляем документ, если есть
			if tt.setupDoc != nil {
				err := repo.Create(ctx, tt.setupDoc)
				require.NoError(t, err)
			}
			res, err := repo.GetByID(ctx, tt.searchID)
			if tt.wantErr {
				assert.Error(t, err)
				if tt.expectEmpty {
					assert.Empty(t, res)
				}
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.searchID, res.ID)
			}
		})
	}
}

func TestGenericRepository_Create(t *testing.T) {
	ctx := context.Background()
	_, _, repo := setupTestRepo(t)

	tests := []struct {
		name    string
		doc     *model.Variable
		wantErr bool
	}{
		{
			name: "successful insert",
			doc: &model.Variable{
				ID:   "doc1",
				Vars: map[string]interface{}{"key": "value"},
			},
			wantErr: false,
		},
		{
			name:    "nil doc", // на всякий случай, если метод не умеет обрабатывать nil — может вызвать панику
			doc:     nil,
			wantErr: true,
		},
		{
			name: "empty id",
			doc: &model.Variable{
				ID:   "",
				Vars: map[string]interface{}{"key": "value"},
			},
			wantErr: false, // MongoDB позволяет вставлять документ без _id — он сам сгенерирует ObjectID
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := repo.Create(ctx, tt.doc)
			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestGenericRepository_CreateMany(t *testing.T) {
	ctx := context.Background()
	_, _, repo := setupTestRepo(t)

	tests := []struct {
		name      string
		docs      []*model.Variable
		wantErr   bool
		checkFunc func(t *testing.T, repo *repository.GenericRepository[*model.Variable], docs []*model.Variable)
	}{
		{
			name:    "empty slice",
			docs:    []*model.Variable{},
			wantErr: false,
			checkFunc: func(t *testing.T, repo *repository.GenericRepository[*model.Variable], docs []*model.Variable) {
				// Проверяем, что коллекция пуста
				count, err := repo.CountDocuments(ctx, bson.D{})
				require.NoError(t, err)
				assert.Equal(t, int64(0), count)
			},
		},
		{
			name: "successful insert",
			docs: []*model.Variable{
				{ID: "var1", Vars: map[string]interface{}{"foo": "bar"}},
				{ID: "var2", Vars: map[string]interface{}{"hello": "world"}},
			},
			wantErr: false,
			checkFunc: func(t *testing.T, repo *repository.GenericRepository[*model.Variable], docs []*model.Variable) {
				count, err := repo.CountDocuments(ctx, bson.D{})
				require.NoError(t, err)
				assert.Equal(t, int64(len(docs)), count)

				// Проверяем, что конкретные документы есть
				for _, doc := range docs {
					v, err := repo.GetByID(ctx, doc.ID)
					assert.NoError(t, err)
					assert.Equal(t, doc.ID, v.ID)
				}
			},
		},
		{
			name: "duplicate insert error",
			docs: []*model.Variable{
				{ID: "var3", Vars: map[string]interface{}{"foo": "bar"}},
			},
			wantErr: false,
			checkFunc: func(t *testing.T, repo *repository.GenericRepository[*model.Variable], docs []*model.Variable) {
				err := repo.CreateMany(ctx, docs)
				assert.Error(t, err)
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {

			err := repo.CreateMany(ctx, tt.docs)
			if tt.wantErr {
				assert.Error(t, err)
				return
			}
			assert.NoError(t, err)
			if tt.checkFunc != nil {
				tt.checkFunc(t, repo, tt.docs)
			}
		})
	}
}

func TestGenericRepository_BulkCreateOrUpdateMany(t *testing.T) {
	ctx := context.Background()
	_, _, repo := setupTestRepo(t)

	tests := []struct {
		name    string
		docs    []*model.Variable
		wantErr bool
		setup   func(t *testing.T, repo *repository.GenericRepository[*model.Variable])
		check   func(t *testing.T, repo *repository.GenericRepository[*model.Variable], docs []*model.Variable)
	}{
		{
			name:    "empty slice",
			docs:    []*model.Variable{},
			wantErr: false,
			check: func(t *testing.T, repo *repository.GenericRepository[*model.Variable], docs []*model.Variable) {
				count, err := repo.CountDocuments(ctx, bson.D{})
				require.NoError(t, err)
				assert.Equal(t, int64(0), count)
			},
		},
		{
			name: "insert new documents",
			docs: []*model.Variable{
				{ID: "new1", Vars: map[string]interface{}{"a": int32(1)}},
				{ID: "new2", Vars: map[string]interface{}{"b": int32(2)}},
			},
			wantErr: false,
			check: func(t *testing.T, repo *repository.GenericRepository[*model.Variable], docs []*model.Variable) {
				count, err := repo.CountDocuments(ctx, bson.D{})
				require.NoError(t, err)
				assert.Equal(t, int64(len(docs)), count)
				for _, d := range docs {
					got, err := repo.GetByID(ctx, d.ID)
					require.NoError(t, err)
					assert.Equal(t, d.ID, got.ID)
					assert.Equal(t, d.Vars, got.Vars)
				}
			},
		},
		{
			name: "update existing documents",
			docs: []*model.Variable{
				{ID: "upd1", Vars: map[string]interface{}{"val": "old"}},
				{ID: "upd2", Vars: map[string]interface{}{"val": "old"}},
			},
			setup: func(t *testing.T, repo *repository.GenericRepository[*model.Variable]) {
				// Вставляем старые документы
				err := repo.CreateMany(ctx, []*model.Variable{
					{ID: "upd1", Vars: map[string]interface{}{"val": "old"}},
					{ID: "upd2", Vars: map[string]interface{}{"val": "old"}},
				})
				require.NoError(t, err)
			},
			wantErr: false,
			check: func(t *testing.T, repo *repository.GenericRepository[*model.Variable], docs []*model.Variable) {
				for _, d := range docs {
					got, err := repo.GetByID(ctx, d.ID)
					require.NoError(t, err)
					assert.Equal(t, d.Vars, got.Vars)
				}
			},
		},
		{
			name: "mixed insert and update",
			docs: []*model.Variable{
				{ID: "upd3", Vars: map[string]interface{}{"val": "updated"}},
				{ID: "new3", Vars: map[string]interface{}{"val": "new"}},
			},
			setup: func(t *testing.T, repo *repository.GenericRepository[*model.Variable]) {
				// Вставляем старый документ upd3
				err := repo.CreateMany(ctx, []*model.Variable{
					{ID: "upd3", Vars: map[string]interface{}{"val": "old"}},
				})
				require.NoError(t, err)
			},
			wantErr: false,
			check: func(t *testing.T, repo *repository.GenericRepository[*model.Variable], docs []*model.Variable) {
				for _, d := range docs {
					got, err := repo.GetByID(ctx, d.ID)
					require.NoError(t, err)
					assert.Equal(t, d.Vars, got.Vars)
				}
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Очистить коллекцию перед каждым тестом
			err := repo.ReturnCollection().Drop(ctx)
			require.NoError(t, err)

			if tt.setup != nil {
				tt.setup(t, repo)
			}

			err = repo.BulkCreateOrUpdateMany(ctx, tt.docs)
			if tt.wantErr {
				assert.Error(t, err)
				return
			}
			assert.NoError(t, err)

			if tt.check != nil {
				tt.check(t, repo, tt.docs)
			}
		})
	}
}

func TestGenericRepository_Update(t *testing.T) {
	ctx := context.Background()
	// setup реального репозитория с тестовой БД
	_, _, repo := setupTestRepo(t)

	tests := []struct {
		name    string
		id      string
		update  *model.Variable
		wantErr bool
	}{
		{
			name:    "successful update",
			id:      "var1",
			update:  &model.Variable{ID: "var1", Vars: map[string]interface{}{"foo": "bar"}},
			wantErr: false,
		},
		{
			name:    "update non-existent",
			id:      "nonexistent",
			update:  &model.Variable{ID: "nonexistent", Vars: map[string]interface{}{"foo": "baz"}},
			wantErr: false, // Mongo не выдаст ошибку, просто не обновит ничего
		},
	}

	// Перед запуском успешного кейса вставим исходный документ
	err := repo.CreateMany(ctx, []*model.Variable{
		{ID: "var1", Vars: map[string]interface{}{"old": "data"}},
	})
	require.NoError(t, err)

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := repo.Update(ctx, tt.id, tt.update)
			if tt.wantErr {
				assert.Error(t, err)
				return
			}
			assert.NoError(t, err)

			// Проверка что обновление произошло для успешного кейса
			if !tt.wantErr && tt.id != "nonexistent" {
				got, err := repo.GetByID(ctx, tt.id)
				require.NoError(t, err)
				assert.Equal(t, tt.update.Vars, got.Vars)
			}
		})
	}
}

func TestGenericRepository_List(t *testing.T) {
	ctx := context.Background()
	_, _, repo := setupTestRepo(t)

	// Подготовим данные для теста
	docs := []*model.Variable{
		{ID: "var1", Vars: map[string]interface{}{"a": 1}},
		{ID: "var2", Vars: map[string]interface{}{"b": 2}},
	}
	err := repo.CreateMany(ctx, docs)
	require.NoError(t, err)

	tests := []struct {
		name    string
		filter  interface{}
		wantLen int
		wantErr bool
		setup   func()
	}{
		{
			name:    "empty collection",
			filter:  bson.M{"_id": "nonexistent"},
			wantLen: 0,
			wantErr: false,
			setup:   nil,
		},
		{
			name:    "multiple documents",
			filter:  bson.M{},
			wantLen: len(docs),
			wantErr: false,
			setup:   nil,
		},
		{
			name:    "multiple documents",
			filter:  bson.D{{Key: "$invalidOperator", Value: 1}},
			wantLen: len(docs),
			wantErr: true,
			setup:   nil,
		},
		{
			name:    "multiple documents",
			filter:  bson.D{{Key: "$invalidOperator", Value: 1}},
			wantLen: len(docs),
			wantErr: true,
			setup: func() {
				_, err = repo.ReturnCollection().InsertOne(ctx, map[string]any{"_id": 1231, "vars": "1dsa"})
				assert.NoError(t, err)
			},
		},
		// Далее кейсы с ошибками лучше покрывать моками, пример ниже для наглядности
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.setup != nil {
				tt.setup()
			}
			got, err := repo.List(ctx, tt.filter)
			if tt.wantErr {
				assert.Error(t, err)
				return
			}
			assert.NoError(t, err)
			assert.Len(t, got, tt.wantLen)
		})
	}
}
