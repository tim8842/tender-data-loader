package task

import (
	"context"
	"errors"
	"fmt"
	"path/filepath"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/tim8842/tender-data-loader/internal/agreement"
	"github.com/tim8842/tender-data-loader/internal/config"
	"github.com/tim8842/tender-data-loader/internal/customer"
	inmock "github.com/tim8842/tender-data-loader/internal/mock"
	agreementt "github.com/tim8842/tender-data-loader/internal/task/agreement"
	uagentt "github.com/tim8842/tender-data-loader/internal/task/uagent"
	variablet "github.com/tim8842/tender-data-loader/internal/task/variable"
	"github.com/tim8842/tender-data-loader/internal/uagent"
	"github.com/tim8842/tender-data-loader/internal/variable"
	"github.com/tim8842/tender-data-loader/pkg"
	"github.com/tim8842/tender-data-loader/pkg/parser"
	"go.uber.org/zap"
	"go.uber.org/zap/zaptest"
)

type RetErr struct {
	Return any
	Err    error
}

func mockFuncWrapperFactory(results map[string]RetErr) func(ctx context.Context, logger *zap.Logger, maxRetries int, delay time.Duration, fn pkg.FuncInWrapp) (any, error) {
	return func(ctx context.Context, logger *zap.Logger, maxRetries int, delay time.Duration, fn pkg.FuncInWrapp) (any, error) {
		switch fn.(type) {
		case *variablet.GetVariableBackToNowAgreementById:
			r := results["GetVariable"]
			return r.Return, r.Err
		case *uagentt.GetRequest:
			r := results["GetProxy"]
			return r.Return, r.Err
		case *uagentt.GetPage:
			r := results["GetPage"]
			return r.Return, r.Err
		case *agreementt.ParseData:
			r := results["ParseIDs"]
			return r.Return, r.Err
		case *SBtnaManyRequests:
			r := results["Btna"]
			return r.Return, r.Err
		default:
			return nil, fmt.Errorf("unexpected subtask type: %T", fn)
		}
	}
}
func TestBackToNowAgreementTask_Process(t *testing.T) {
	date, err := parser.ParseFromDateToTime("11.06.2025")
	date2, err2 := parser.ParseFromDateToTime("10.06.2025")
	// mockRepo.On("GetByID", mock.Anything, "id123").Return(tt.mockData, tt.mockError) просто пример
	assert.NoError(t, err)
	assert.NoError(t, err2)
	tests := []struct {
		name        string
		results     map[string]RetErr
		needErr     bool
		now         func() time.Time
		mockAgRE    error
		mockVaU1Re  error
		mockVaU2Re  error
		mockCuRE    error
		staticProxy bool
	}{
		{
			mockAgRE:    nil,
			mockVaU1Re:  nil,
			mockVaU2Re:  nil,
			mockCuRE:    nil,
			staticProxy: false,
			name:        "test endData",
			results: map[string]RetErr{
				"GetVariable": {&variable.VariableBackToNowAgreement{ID: "back_to_now_agreement",
					Vars: variable.VarsBackToNowAgreement{Page: 1, SignedAt: date},
				}, nil},
				"GetProxy": {nil, nil},
				"GetPage":  {"dsadas", nil}, "ParseIDs": {[]string{"123", "32"}, nil},
				"Btna": {[]any{nil, nil}, nil}},
			needErr: false,
			now:     func() time.Time { return date },
		},
		{
			mockAgRE:    nil,
			mockVaU1Re:  nil,
			mockVaU2Re:  nil,
			mockCuRE:    nil,
			staticProxy: false,
			name:        "Test err get proxy",
			results: map[string]RetErr{
				"GetVariable": {variable.VariableBackToNowAgreement{ID: "1",
					Vars: variable.VarsBackToNowAgreement{Page: 1, SignedAt: date},
				}, errors.New("")},
				"GetProxy": {nil, nil},
				"GetPage":  {"dsadas", nil}, "ParseIDs": {[]string{"123", "32"}, nil},
				"Btna": {[]any{nil, nil}, nil}},
			needErr: true,
			now:     func() time.Time { return date },
		},
		{
			mockAgRE:    nil,
			mockVaU1Re:  nil,
			mockVaU2Re:  nil,
			mockCuRE:    nil,
			staticProxy: false,
			name:        "Bad parse type *variable.VariableBackToNowAgreement",
			results: map[string]RetErr{
				"GetVariable": {variable.Variable{ID: "1",
					Vars: map[string]interface{}{"dasads": 3},
				}, nil},
				"GetProxy": {nil, nil},
				"GetPage":  {"dsadas", nil}, "ParseIDs": {[]string{"123", "32"}, nil},
				"Btna": {[]any{nil, nil}, nil}},
			needErr: true,
			now:     func() time.Time { return date },
		},
		{
			mockAgRE:    nil,
			mockVaU1Re:  nil,
			mockVaU2Re:  nil,
			mockCuRE:    nil,
			staticProxy: false,
			name:        "test bad Get Proxy",
			results: map[string]RetErr{
				"GetVariable": {&variable.VariableBackToNowAgreement{ID: "back_to_now_agreement",
					Vars: variable.VarsBackToNowAgreement{Page: 1, SignedAt: date2},
				}, nil},
				"GetProxy": {nil, errors.New("")},
				"GetPage":  {"dsadas", nil}, "ParseIDs": {[]string{"123", "32"}, nil},
				"Btna": {[]any{nil, nil}, nil}},
			needErr: true,
			now:     func() time.Time { return date },
		},
		{
			mockAgRE:    nil,
			mockVaU1Re:  nil,
			mockVaU2Re:  nil,
			mockCuRE:    nil,
			staticProxy: false,
			name:        "test bad Parse Proxy",
			results: map[string]RetErr{
				"GetVariable": {&variable.VariableBackToNowAgreement{ID: "back_to_now_agreement",
					Vars: variable.VarsBackToNowAgreement{Page: 1, SignedAt: date2},
				}, nil},
				"GetProxy": {1, nil},
				"GetPage":  {"dsadas", nil}, "ParseIDs": {[]string{"123", "32"}, nil},
				"Btna": {[]any{nil, nil}, nil}},
			needErr: true,
			now:     func() time.Time { return date },
		},
		{
			mockAgRE:    nil,
			mockVaU1Re:  nil,
			mockVaU2Re:  nil,
			mockCuRE:    nil,
			staticProxy: false,
			name:        "test bad get Page",
			results: map[string]RetErr{
				"GetVariable": {&variable.VariableBackToNowAgreement{ID: "back_to_now_agreement",
					Vars: variable.VarsBackToNowAgreement{Page: 1, SignedAt: date2},
				}, nil},
				"GetProxy": {&uagent.UserAgentResponse{}, nil},
				"GetPage":  {1, nil}, "ParseIDs": {[]string{"123", "32"}, errors.New("")},
				"Btna": {[]any{nil, nil}, nil}},
			needErr: true,
			now:     func() time.Time { return date },
		},
		{
			mockAgRE:    nil,
			mockVaU1Re:  nil,
			mockVaU2Re:  nil,
			mockCuRE:    nil,
			staticProxy: false,
			name:        "test bad type parse Page",
			results: map[string]RetErr{
				"GetVariable": {&variable.VariableBackToNowAgreement{ID: "back_to_now_agreement",
					Vars: variable.VarsBackToNowAgreement{Page: 1, SignedAt: date2},
				}, nil},
				"GetProxy": {&uagent.UserAgentResponse{}, nil},
				"GetPage":  {1, nil}, "ParseIDs": {[]string{"123", "32"}, nil},
				"Btna": {[]any{nil, nil}, nil}},
			needErr: true,
			now:     func() time.Time { return date },
		},
		{
			mockAgRE:    nil,
			mockVaU1Re:  nil,
			mockVaU2Re:  nil,
			mockCuRE:    nil,
			staticProxy: false,
			name:        "err parse ids",
			results: map[string]RetErr{
				"GetVariable": {&variable.VariableBackToNowAgreement{ID: "back_to_now_agreement",
					Vars: variable.VarsBackToNowAgreement{Page: 1, SignedAt: date2},
				}, nil},
				"GetProxy": {&uagent.UserAgentResponse{}, nil},
				"GetPage":  {[]byte{10, 12}, nil}, "ParseIDs": {[]string{"32", "d32"}, errors.New("")},
				"Btna": {[]any{nil, nil}, nil}},
			needErr: true,
			now:     func() time.Time { return date },
		},
		{
			mockAgRE:    nil,
			mockVaU1Re:  nil,
			mockVaU2Re:  nil,
			mockCuRE:    nil,
			staticProxy: false,
			name:        "Bad parse type ids",
			results: map[string]RetErr{
				"GetVariable": {&variable.VariableBackToNowAgreement{ID: "back_to_now_agreement",
					Vars: variable.VarsBackToNowAgreement{Page: 1, SignedAt: date2},
				}, nil},
				"GetProxy": {&uagent.UserAgentResponse{}, nil},
				"GetPage":  {[]byte{10, 12}, nil}, "ParseIDs": {[]int{32, 33}, nil},
				"Btna": {[]any{nil, nil}, nil}},
			needErr: true,
			now:     func() time.Time { return date },
		},
		{
			mockAgRE:    nil,
			mockVaU1Re:  errors.New(""),
			mockVaU2Re:  nil,
			staticProxy: false,
			mockCuRE:    nil,
			name:        "Bad update after ids",
			results: map[string]RetErr{
				"GetVariable": {&variable.VariableBackToNowAgreement{ID: "1",
					Vars: variable.VarsBackToNowAgreement{Page: 1, SignedAt: date2},
				}, nil},
				"GetProxy": {&uagent.UserAgentResponse{}, nil},
				"GetPage":  {[]byte{10, 12}, nil}, "ParseIDs": {[]string{}, nil},
				"Btna": {[]any{nil, nil}, nil}},
			needErr: true,
			now:     func() time.Time { return date },
		},
		{
			mockAgRE:    nil,
			mockVaU1Re:  nil,
			mockVaU2Re:  nil,
			mockCuRE:    nil,
			staticProxy: false,
			name:        "Bad many Requ",
			results: map[string]RetErr{
				"GetVariable": {&variable.VariableBackToNowAgreement{ID: "back_to_now_agreement",
					Vars: variable.VarsBackToNowAgreement{Page: 1, SignedAt: date2},
				}, nil},
				"GetProxy": {&uagent.UserAgentResponse{}, nil},
				"GetPage":  {[]byte{10, 12}, nil}, "ParseIDs": {[]string{"123", "3123"}, nil},
				"Btna": {nil, errors.New("")}},
			needErr: true,
			now:     func() time.Time { return date },
		},
		{
			mockAgRE:    nil,
			mockVaU1Re:  nil,
			mockVaU2Re:  nil,
			staticProxy: false,
			mockCuRE:    nil,
			name:        "Bad type Parse many request ",
			results: map[string]RetErr{
				"GetVariable": {&variable.VariableBackToNowAgreement{ID: "back_to_now_agreement",
					Vars: variable.VarsBackToNowAgreement{Page: 1, SignedAt: date2},
				}, nil},
				"GetProxy": {&uagent.UserAgentResponse{}, nil},
				"GetPage":  {[]byte{10, 12}, nil}, "ParseIDs": {[]string{"123", "3123"}, nil},
				"Btna": {nil, nil}},
			needErr: true,
			now:     func() time.Time { return date },
		},
		{
			mockAgRE:    errors.New(""),
			mockVaU1Re:  nil,
			mockVaU2Re:  nil,
			staticProxy: false,
			mockCuRE:    nil,
			name:        "Bad AgreementRepo.BulkMergeMany ",
			results: map[string]RetErr{
				"GetVariable": {&variable.VariableBackToNowAgreement{ID: "back_to_now_agreement",
					Vars: variable.VarsBackToNowAgreement{Page: 1, SignedAt: date2},
				}, nil},
				"GetProxy": {&uagent.UserAgentResponse{}, nil},
				"GetPage":  {[]byte{10, 12}, nil}, "ParseIDs": {[]string{"123", "3123"}, nil},
				"Btna": {[]*agreement.AgreementParesedData{
					{ID: "1", Customer: &customer.Customer{ID: "1"}},
					{ID: "2", Customer: &customer.Customer{ID: "2"}},
				}, nil}},
			needErr: true,
			now:     func() time.Time { return date },
		},
		{
			mockAgRE:    nil,
			mockVaU1Re:  nil,
			staticProxy: false,
			mockVaU2Re:  nil,
			mockCuRE:    errors.New(""),
			name:        "Bad CustomerRepo.BulkMergeMany",
			results: map[string]RetErr{
				"GetVariable": {&variable.VariableBackToNowAgreement{ID: "back_to_now_agreement",
					Vars: variable.VarsBackToNowAgreement{Page: 1, SignedAt: date2},
				}, nil},
				"GetProxy": {&uagent.UserAgentResponse{}, nil},
				"GetPage":  {[]byte{10, 12}, nil}, "ParseIDs": {[]string{"123", "3123"}, nil},
				"Btna": {[]*agreement.AgreementParesedData{
					{ID: "1", Customer: &customer.Customer{ID: "1"}},
					{ID: "2", Customer: &customer.Customer{ID: "2"}},
				}, nil}},
			needErr: true,
			now:     func() time.Time { return date },
		},
		{
			mockAgRE:    nil,
			mockVaU1Re:  nil,
			staticProxy: false,
			mockVaU2Re:  errors.New(""),
			mockCuRE:    nil,
			name:        "Bad VarRepo.Update2 ",
			results: map[string]RetErr{
				"GetVariable": {&variable.VariableBackToNowAgreement{ID: "back_to_now_agreement",
					Vars: variable.VarsBackToNowAgreement{Page: 1, SignedAt: date2},
				}, nil},
				"GetProxy": {&uagent.UserAgentResponse{}, nil},
				"GetPage":  {[]byte{10, 12}, nil}, "ParseIDs": {[]string{"123", "3123"}, nil},
				"Btna": {[]*agreement.AgreementParesedData{
					{ID: "1", Customer: &customer.Customer{ID: "1"}},
					{ID: "2", Customer: &customer.Customer{ID: "2"}},
				}, nil}},
			needErr: true,
			now:     func() time.Time { return date },
		},
	}
	oldFuncWrapper := funcWrapper
	oldNow := Now
	envPath, err := filepath.Abs("../../configs/.env.test")
	if err != nil {
		panic(err)
	}
	logger := zaptest.NewLogger(t)
	cfg, _ := config.LoadConfig(envPath)
	for _, tt := range tests {
		ctx, cancelled := context.WithTimeout(context.Background(), 20*time.Microsecond)
		Now = tt.now
		mockAgRepo := new(inmock.MockGenericRepository[*agreement.Agreement])
		mockVaRepo := new(inmock.MockGenericRepository[*variable.Variable])
		mockCuRepo := new(inmock.MockGenericRepository[*customer.Customer])
		mockAgRepo.On("BulkMergeMany", mock.Anything, mock.Anything).Return(tt.mockAgRE)
		mockVaRepo.On("Update", mock.Anything, "1", mock.Anything).Return(tt.mockVaU1Re)
		mockCuRepo.On("BulkMergeMany", mock.Anything, mock.Anything).Return(tt.mockCuRE)
		mockVaRepo.On("Update", mock.Anything, "back_to_now_agreement", mock.Anything).Return(tt.mockVaU2Re)
		back := NewBackToNowAgreementTask(
			cfg, mockAgRepo, mockVaRepo, mockCuRepo, tt.staticProxy)
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
