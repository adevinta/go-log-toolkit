package log_test

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"testing"
	"time"

	log "github.com/adevinta/go-log-toolkit"
	gologr "github.com/go-logr/logr"
	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
)

type timeSetterHook struct {
}

func (t timeSetterHook) Levels() []logrus.Level {
	return logrus.AllLevels
}

func (t timeSetterHook) Fire(entry *logrus.Entry) error {
	entry.Time = time.Date(2020, 03, 13, 14, 00, 0, 0, time.UTC)
	return nil
}

func initTestLogrus(level logrus.Level) (gologr.Logger, *bytes.Buffer) {
	b := new(bytes.Buffer)
	l := logrus.New()
	l.AddHook(timeSetterHook{})
	l.SetLevel(level)
	l.SetFormatter(&logrus.JSONFormatter{})
	l.SetOutput(b)
	return log.NewLogr(l), b
}

func TestError(t *testing.T) {
	tested, b := initTestLogrus(logrus.TraceLevel)
	tested.Error(errors.New("testError"), "this is a test", "some-context", "help")
	assert.JSONEq(t, `{"error":"testError","some-context": "help","level":"error","msg":"this is a test","time":"2020-03-13T14:00:00Z"}`, b.String())
}

func TestWithName(t *testing.T) {
	tested, b := initTestLogrus(logrus.TraceLevel)
	tested.V(0).WithName("pkg").WithName("method").Info("hello world")
	assert.JSONEq(t, `{"level":"info","msg":"hello world", "name": "pkg.method","time":"2020-03-13T14:00:00Z"}`, b.String())
}

func TestLogrLevelsWithLogrusLevels(t *testing.T) {
	tests := []struct {
		logrusLevel       logrus.Level
		logrLevel         int
		expectedInfoLevel string
	}{
		{
			logrusLevel:       logrus.TraceLevel,
			logrLevel:         0,
			expectedInfoLevel: "info",
		},
		{
			logrusLevel:       logrus.TraceLevel,
			logrLevel:         1,
			expectedInfoLevel: "debug",
		},
		{
			logrusLevel:       logrus.TraceLevel,
			logrLevel:         2,
			expectedInfoLevel: "trace",
		},
		{
			logrusLevel:       logrus.TraceLevel,
			logrLevel:         3,
			expectedInfoLevel: "trace",
		},
		{
			logrusLevel:       logrus.DebugLevel,
			logrLevel:         0,
			expectedInfoLevel: "info",
		},
		{
			logrusLevel:       logrus.DebugLevel,
			logrLevel:         1,
			expectedInfoLevel: "debug",
		},
		{
			logrusLevel:       logrus.DebugLevel,
			logrLevel:         2,
			expectedInfoLevel: "<DROPPED_LOG>",
		},
		{
			logrusLevel:       logrus.InfoLevel,
			logrLevel:         0,
			expectedInfoLevel: "info",
		},
		{
			logrusLevel:       logrus.InfoLevel,
			logrLevel:         1,
			expectedInfoLevel: "<DROPPED_LOG>",
		},
	}

	for _, tt := range tests {
		t.Run(fmt.Sprintf("V(%d) with Logrus %s level", tt.logrLevel, tt.logrusLevel.String()), func(t *testing.T) {
			tested, b := initTestLogrus(tt.logrusLevel)
			tested.V(tt.logrLevel).Info("hello world")

			if tt.expectedInfoLevel == "<DROPPED_LOG>" {
				assert.Empty(t, b.String())
			} else {
				expectedLog := fmt.Sprintf(`{"level":"%s","msg":"hello world","time":"2020-03-13T14:00:00Z"}`, tt.expectedInfoLevel)
				assert.JSONEq(t, expectedLog, b.String())
			}

			b.Reset()
			tested.V(tt.logrLevel).Error(errors.New("testError"), "this is a test", "some-context", "help")
			assert.JSONEq(t, `{"error":"testError","some-context": "help","level":"error","msg":"this is a test","time":"2020-03-13T14:00:00Z"}`, b.String())
		})
	}
}

func TestEnabled(t *testing.T) {
	tested, _ := initTestLogrus(logrus.TraceLevel)
	assert.True(t, tested.Enabled())

	tested = log.NewLogr(nil)
	assert.False(t, tested.Enabled())
}

func TestNilDoesNotCrash(t *testing.T) {
	tested := log.NewLogr(nil)

	tested.Info("should not crash")
	tested.Error(errors.New("err"), "should not crash")
	tested.WithName("some")
	tested.WithValues("key", "value")
	tested.V(9).Info("should not crash")
}

func TestContextualizeLogr(t *testing.T) {
	ctx := context.Background()
	ctx = log.AddLogFieldsToContext(ctx, logrus.Fields{"key": "value"})

	l := log.New()
	b := bytes.Buffer{}
	l.Out = &b

	logger := log.NewLogr(l)

	log.ContextualizeLogr(logger, ctx).Info("hello world")
	loggedData := map[string]interface{}{}
	json.NewDecoder(&b).Decode(&loggedData)

	assert.Equal(t, "value", loggedData["key"])
}
