package tasks

import (
	"context"
	"errors"
	"fmt"
	"path/filepath"
	"testing"
	"time"

	"github.com/joho/godotenv"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/tim8842/tender-data-loader/internal/config"
	"github.com/tim8842/tender-data-loader/internal/mocks"
	"github.com/tim8842/tender-data-loader/internal/model"
	"github.com/tim8842/tender-data-loader/internal/repository"
	subtasks "github.com/tim8842/tender-data-loader/internal/tasks/sub_tasks"
	baseutils "github.com/tim8842/tender-data-loader/internal/util/base_utils"
	"github.com/tim8842/tender-data-loader/internal/util/wrappers"
	"go.uber.org/zap"
	"go.uber.org/zap/zaptest"
)

type RetErr struct {
	Return any
	Err    error
}

func mockFuncWrapperFactory(results map[string]RetErr) func(ctx context.Context, logger *zap.Logger, maxRetries int, delay time.Duration, fn wrappers.FuncInWrapp) (any, error) {
	return func(ctx context.Context, logger *zap.Logger, maxRetries int, delay time.Duration, fn wrappers.FuncInWrapp) (any, error) {
		switch fn.(type) {
		case *subtasks.GetVariableBackToNowAgreementById:
			r := results["GetVariable"]
			return r.Return, r.Err
		case *subtasks.GetRequest:
			r := results["GetProxy"]
			return r.Return, r.Err
		case *subtasks.GetPage:
			r := results["GetPage"]
			return r.Return, r.Err
		case *subtasks.ParseData:
			r := results["ParseIDs"]
			return r.Return, r.Err
		case *subtasks.SBtnaManyRequests:
			r := results["Btna"]
			return r.Return, r.Err
		default:
			return nil, fmt.Errorf("unexpected subtask type: %T", fn)
		}
	}
}
func TestBackToNowAgreementTask_Process(t *testing.T) {
	date, err := baseutils.ParseDate("11.06.2025")
	date2, err2 := baseutils.ParseDate("10.06.2025")
	// mockRepo.On("GetByID", mock.Anything, "id123").Return(tt.mockData, tt.mockError) просто пример
	assert.NoError(t, err)
	assert.NoError(t, err2)
	tests := []struct {
		name       string
		results    map[string]RetErr
		needErr    bool
		now        func() time.Time
		mockAgRE   error
		mockVaU1Re error
		mockVaU2Re error
		mockCuRE   error
	}{
		{
			mockAgRE:   nil,
			mockVaU1Re: nil,
			mockVaU2Re: nil,
			mockCuRE:   nil,
			name:       "test endData",
			results: map[string]RetErr{
				"GetVariable": {&model.VariableBackToNowAgreement{ID: "back_to_now_agreement",
					Vars: model.VarsBackToNowAgreement{Page: 1, SignedAt: date},
				}, nil},
				"GetProxy": {nil, nil},
				"GetPage":  {"dsadas", nil}, "ParseIDs": {[]string{"123", "32"}, nil},
				"Btna": {[]any{nil, nil}, nil}},
			needErr: false,
			now:     func() time.Time { return date },
		},
		{
			mockAgRE:   nil,
			mockVaU1Re: nil,
			mockVaU2Re: nil,
			mockCuRE:   nil,
			name:       "Test err get proxy",
			results: map[string]RetErr{
				"GetVariable": {model.VariableBackToNowAgreement{ID: "1",
					Vars: model.VarsBackToNowAgreement{Page: 1, SignedAt: date},
				}, errors.New("")},
				"GetProxy": {nil, nil},
				"GetPage":  {"dsadas", nil}, "ParseIDs": {[]string{"123", "32"}, nil},
				"Btna": {[]any{nil, nil}, nil}},
			needErr: true,
			now:     func() time.Time { return date },
		},
		{
			mockAgRE:   nil,
			mockVaU1Re: nil,
			mockVaU2Re: nil,
			mockCuRE:   nil,
			name:       "Bad parse type *model.VariableBackToNowAgreement",
			results: map[string]RetErr{
				"GetVariable": {model.Variable{ID: "1",
					Vars: map[string]interface{}{"dasads": 3},
				}, nil},
				"GetProxy": {nil, nil},
				"GetPage":  {"dsadas", nil}, "ParseIDs": {[]string{"123", "32"}, nil},
				"Btna": {[]any{nil, nil}, nil}},
			needErr: true,
			now:     func() time.Time { return date },
		},
		{
			mockAgRE:   nil,
			mockVaU1Re: nil,
			mockVaU2Re: nil,
			mockCuRE:   nil,
			name:       "test bad Get Proxy",
			results: map[string]RetErr{
				"GetVariable": {&model.VariableBackToNowAgreement{ID: "back_to_now_agreement",
					Vars: model.VarsBackToNowAgreement{Page: 1, SignedAt: date2},
				}, nil},
				"GetProxy": {nil, errors.New("")},
				"GetPage":  {"dsadas", nil}, "ParseIDs": {[]string{"123", "32"}, nil},
				"Btna": {[]any{nil, nil}, nil}},
			needErr: true,
			now:     func() time.Time { return date },
		},
		{
			mockAgRE:   nil,
			mockVaU1Re: nil,
			mockVaU2Re: nil,
			mockCuRE:   nil,
			name:       "test bad Parse Proxy",
			results: map[string]RetErr{
				"GetVariable": {&model.VariableBackToNowAgreement{ID: "back_to_now_agreement",
					Vars: model.VarsBackToNowAgreement{Page: 1, SignedAt: date2},
				}, nil},
				"GetProxy": {1, nil},
				"GetPage":  {"dsadas", nil}, "ParseIDs": {[]string{"123", "32"}, nil},
				"Btna": {[]any{nil, nil}, nil}},
			needErr: true,
			now:     func() time.Time { return date },
		},
		{
			mockAgRE:   nil,
			mockVaU1Re: nil,
			mockVaU2Re: nil,
			mockCuRE:   nil,
			name:       "test bad get Page",
			results: map[string]RetErr{
				"GetVariable": {&model.VariableBackToNowAgreement{ID: "back_to_now_agreement",
					Vars: model.VarsBackToNowAgreement{Page: 1, SignedAt: date2},
				}, nil},
				"GetProxy": {&model.UserAgentResponse{}, nil},
				"GetPage":  {1, nil}, "ParseIDs": {[]string{"123", "32"}, errors.New("")},
				"Btna": {[]any{nil, nil}, nil}},
			needErr: true,
			now:     func() time.Time { return date },
		},
		{
			mockAgRE:   nil,
			mockVaU1Re: nil,
			mockVaU2Re: nil,
			mockCuRE:   nil,
			name:       "test bad type parse Page",
			results: map[string]RetErr{
				"GetVariable": {&model.VariableBackToNowAgreement{ID: "back_to_now_agreement",
					Vars: model.VarsBackToNowAgreement{Page: 1, SignedAt: date2},
				}, nil},
				"GetProxy": {&model.UserAgentResponse{}, nil},
				"GetPage":  {1, nil}, "ParseIDs": {[]string{"123", "32"}, nil},
				"Btna": {[]any{nil, nil}, nil}},
			needErr: true,
			now:     func() time.Time { return date },
		},
		{
			mockAgRE:   nil,
			mockVaU1Re: nil,
			mockVaU2Re: nil,
			mockCuRE:   nil,
			name:       "err parse ids",
			results: map[string]RetErr{
				"GetVariable": {&model.VariableBackToNowAgreement{ID: "back_to_now_agreement",
					Vars: model.VarsBackToNowAgreement{Page: 1, SignedAt: date2},
				}, nil},
				"GetProxy": {&model.UserAgentResponse{}, nil},
				"GetPage":  {[]byte{10, 12}, nil}, "ParseIDs": {[]string{"32", "d32"}, errors.New("")},
				"Btna": {[]any{nil, nil}, nil}},
			needErr: true,
			now:     func() time.Time { return date },
		},
		{
			mockAgRE:   nil,
			mockVaU1Re: nil,
			mockVaU2Re: nil,
			mockCuRE:   nil,
			name:       "Bad parse type ids",
			results: map[string]RetErr{
				"GetVariable": {&model.VariableBackToNowAgreement{ID: "back_to_now_agreement",
					Vars: model.VarsBackToNowAgreement{Page: 1, SignedAt: date2},
				}, nil},
				"GetProxy": {&model.UserAgentResponse{}, nil},
				"GetPage":  {[]byte{10, 12}, nil}, "ParseIDs": {[]int{32, 33}, nil},
				"Btna": {[]any{nil, nil}, nil}},
			needErr: true,
			now:     func() time.Time { return date },
		},
		{
			mockAgRE:   nil,
			mockVaU1Re: errors.New(""),
			mockVaU2Re: nil,
			mockCuRE:   nil,
			name:       "Bad update after ids",
			results: map[string]RetErr{
				"GetVariable": {&model.VariableBackToNowAgreement{ID: "1",
					Vars: model.VarsBackToNowAgreement{Page: 1, SignedAt: date2},
				}, nil},
				"GetProxy": {&model.UserAgentResponse{}, nil},
				"GetPage":  {[]byte{10, 12}, nil}, "ParseIDs": {[]string{}, nil},
				"Btna": {[]any{nil, nil}, nil}},
			needErr: true,
			now:     func() time.Time { return date },
		},
		{
			mockAgRE:   nil,
			mockVaU1Re: nil,
			mockVaU2Re: nil,
			mockCuRE:   nil,
			name:       "Bad many Requ",
			results: map[string]RetErr{
				"GetVariable": {&model.VariableBackToNowAgreement{ID: "back_to_now_agreement",
					Vars: model.VarsBackToNowAgreement{Page: 1, SignedAt: date2},
				}, nil},
				"GetProxy": {&model.UserAgentResponse{}, nil},
				"GetPage":  {[]byte{10, 12}, nil}, "ParseIDs": {[]string{"123", "3123"}, nil},
				"Btna": {nil, errors.New("")}},
			needErr: true,
			now:     func() time.Time { return date },
		},
		{
			mockAgRE:   nil,
			mockVaU1Re: nil,
			mockVaU2Re: nil,
			mockCuRE:   nil,
			name:       "Bad type Parse many request ",
			results: map[string]RetErr{
				"GetVariable": {&model.VariableBackToNowAgreement{ID: "back_to_now_agreement",
					Vars: model.VarsBackToNowAgreement{Page: 1, SignedAt: date2},
				}, nil},
				"GetProxy": {&model.UserAgentResponse{}, nil},
				"GetPage":  {[]byte{10, 12}, nil}, "ParseIDs": {[]string{"123", "3123"}, nil},
				"Btna": {nil, nil}},
			needErr: true,
			now:     func() time.Time { return date },
		},
		{
			mockAgRE:   errors.New(""),
			mockVaU1Re: nil,
			mockVaU2Re: nil,
			mockCuRE:   nil,
			name:       "Bad AgreementRepo.BulkMergeMany ",
			results: map[string]RetErr{
				"GetVariable": {&model.VariableBackToNowAgreement{ID: "back_to_now_agreement",
					Vars: model.VarsBackToNowAgreement{Page: 1, SignedAt: date2},
				}, nil},
				"GetProxy": {&model.UserAgentResponse{}, nil},
				"GetPage":  {[]byte{10, 12}, nil}, "ParseIDs": {[]string{"123", "3123"}, nil},
				"Btna": {[]*model.AgreementParesedData{
					{ID: "1", Customer: &model.Customer{ID: "1"}},
					{ID: "2", Customer: &model.Customer{ID: "2"}},
				}, nil}},
			needErr: true,
			now:     func() time.Time { return date },
		},
		{
			mockAgRE:   nil,
			mockVaU1Re: nil,
			mockVaU2Re: nil,
			mockCuRE:   errors.New(""),
			name:       "Bad CustomerRepo.BulkMergeMany",
			results: map[string]RetErr{
				"GetVariable": {&model.VariableBackToNowAgreement{ID: "back_to_now_agreement",
					Vars: model.VarsBackToNowAgreement{Page: 1, SignedAt: date2},
				}, nil},
				"GetProxy": {&model.UserAgentResponse{}, nil},
				"GetPage":  {[]byte{10, 12}, nil}, "ParseIDs": {[]string{"123", "3123"}, nil},
				"Btna": {[]*model.AgreementParesedData{
					{ID: "1", Customer: &model.Customer{ID: "1"}},
					{ID: "2", Customer: &model.Customer{ID: "2"}},
				}, nil}},
			needErr: true,
			now:     func() time.Time { return date },
		},
		{
			mockAgRE:   nil,
			mockVaU1Re: nil,
			mockVaU2Re: errors.New(""),
			mockCuRE:   nil,
			name:       "Bad VarRepo.Update2 ",
			results: map[string]RetErr{
				"GetVariable": {&model.VariableBackToNowAgreement{ID: "back_to_now_agreement",
					Vars: model.VarsBackToNowAgreement{Page: 1, SignedAt: date2},
				}, nil},
				"GetProxy": {&model.UserAgentResponse{}, nil},
				"GetPage":  {[]byte{10, 12}, nil}, "ParseIDs": {[]string{"123", "3123"}, nil},
				"Btna": {[]*model.AgreementParesedData{
					{ID: "1", Customer: &model.Customer{ID: "1"}},
					{ID: "2", Customer: &model.Customer{ID: "2"}},
				}, nil}},
			needErr: true,
			now:     func() time.Time { return date },
		},
	}
	oldFuncWrapper := funcWrapper
	oldNow := Now
	envPath, err := filepath.Abs("../../.env.test")
	if err != nil {
		panic(err)
	}

	if err := godotenv.Load(envPath); err != nil {
		panic("Error loading .env file")
	}
	logger := zaptest.NewLogger(t)
	cfg, _ := config.LoadConfig()
	for _, tt := range tests {
		ctx, cancelled := context.WithTimeout(context.Background(), 20*time.Microsecond)
		Now = tt.now
		mockAgRepo := new(mocks.MockGenericRepository[*model.Agreement])
		mockVaRepo := new(mocks.MockGenericRepository[*model.Variable])
		mockCuRepo := new(mocks.MockGenericRepository[*model.Customer])
		mockAgRepo.On("BulkMergeMany", mock.Anything, mock.Anything).Return(tt.mockAgRE)
		mockVaRepo.On("Update", mock.Anything, "1", mock.Anything).Return(tt.mockVaU1Re)
		mockCuRepo.On("BulkMergeMany", mock.Anything, mock.Anything).Return(tt.mockCuRE)
		mockVaRepo.On("Update", mock.Anything, "back_to_now_agreement", mock.Anything).Return(tt.mockVaU2Re)
		reps := &repository.Repositories{
			AgreementRepo: mockAgRepo,
			CustomerRepo:  mockCuRepo,
			VarRepo:       mockVaRepo,
		}
		back := NewBackToNowAgreementTask(
			cfg, reps)
		funcWrapper = mockFuncWrapperFactory(tt.results)
		err = back.Process(ctx, logger)
		if tt.needErr {
			assert.Error(t, err)
		} else {
			assert.NoError(t, err)
		}
		defer func() {
			funcWrapper = oldFuncWrapper
			Now = oldNow
			cancelled()
		}()

	}
}
