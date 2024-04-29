package configstruct

import (
	"os"
	"strings"
	"testing"
	"time"
)

var multilineVar1 = `-----BEGIN PUBLIC KEY-----
MFkwEwYHKoZIzj0CAQYIKoZIzj0DAQcDQgAEUOu0Nc9/EiVSyBKyfvv38MlteRWA
+6S8jpRIOC2eMn2kYSv1RCc7uejvLVc0EYn2spObZjsMv4qvNz0XxYduDQ==
-----END PUBLIC KEY-----`

var multilineVar2 = `-----BEGIN PRIVATE KEY-----
MFkwEwYHKoZIzj0CAQYIKoZIzj0DAQcDQgAEUOu0Nc9/EiVSyBKyfvv38MlteRWA
+6S8jpRIOC2eMn2kYSv1RCc7uejvLVc0EYn2spObZjsMv4qvNz0XxYduDQ==
-----END PRIVATE KEY-----`

func TestLoadAndParse(t *testing.T) {
	t.Run("file_in_same_dir", func(t *testing.T) {
		cleanEnv()
		type MyConfig struct {
			Foo string `env:"FOO" default:"bar"`
		}
		myConfig := MyConfig{}
		err := Parse(&myConfig, ".env")
		if err != nil {
			t.Error(err)
		}
		if myConfig.Foo != "root" {
			t.Errorf("Foo should be root, got %s", myConfig.Foo)
		}
	})

	t.Run("file_in_subdir", func(t *testing.T) {
		cleanEnv()
		type MyConfig struct {
			Foo  string `env:"FOO"`
			Env  string `env:"ENV"`
			Port int    `env:"PORT"`
		}
		myConfig := MyConfig{}
		err := Parse(&myConfig, "./testdata/.env")
		if err != nil {
			t.Error(err)
		}
		if myConfig.Foo != "nested" {
			t.Errorf("Foo should be nested, got %s", myConfig.Foo)
		}
		if myConfig.Env != "testing" {
			t.Errorf("Env should be testing, got %s", myConfig.Env)
		}
		if myConfig.Port != 3000 {
			t.Errorf("Port should be 3000, got %d", myConfig.Port)
		}
	})

	t.Run("file_in_subdir_with_different_name", func(t *testing.T) {
		cleanEnv()
		type MyConfig struct {
			Foo string `env:"FOO" default:"bar"`
		}
		myConfig := MyConfig{}
		err := Parse(&myConfig, "./testdata/.env.local")
		if err != nil {
			t.Error(err)
		}
	})

	t.Run("env_file_does_not_override_existing_vars", func(t *testing.T) {
		cleanEnv()
		type MyConfig struct {
			Foo string `env:"FOO"`
		}
		myConfig := MyConfig{}
		os.Setenv("FOO", "PreExistingValue")
		err := Parse(&myConfig, ".env")
		if err != nil {
			t.Error(err)
		}
		if myConfig.Foo != "PreExistingValue" {
			t.Errorf("Foo should be PreExistingValue, got %s", myConfig.Foo)
		}
	})

	t.Run("more_complex_values", func(t *testing.T) {
		cleanEnv()
		type MyConfig struct {
			NoQuotes     string        `env:"NO_QUOTES"`
			DoubleQuotes string        `env:"DOUBLE_QUOTES"`
			SingleQuotes string        `env:"SINGLE_QUOTES"`
			Timeout      time.Duration `env:"TIMEOUT"`
			SomeOther    string        `env:"SOME_OTHER" default:"hello"`
			PublKey      string        `env:"PUBLIC_KEY"`
			PrivKey      string        `env:"PRIVATE_KEY"`
			Port         int           `env:"PORT"`
		}
		myConfig := MyConfig{}

		os.Setenv("TIMEOUT", "3s")

		err := Parse(&myConfig, "testdata/.env.complex")
		if err != nil {
			t.Error(err)
		}

		want := "https://search.brave.com/search?q=somesearch&source=desktop"
		if myConfig.NoQuotes != want {
			t.Errorf("want %s, got %s", want, myConfig.NoQuotes)
		}

		want = "mongodb://127.0.0.1:27017/test?directConnection=true"
		if myConfig.DoubleQuotes != want {
			t.Errorf("want %s, got %s", want, myConfig.DoubleQuotes)
		}

		want = "https://api.github.com/"
		if myConfig.SingleQuotes != want {
			t.Errorf("want %s, got %s", want, myConfig.SingleQuotes)
		}

		// The value in the file should not overwrite the value in the env
		if myConfig.Timeout != 3*time.Second {
			t.Errorf("Timeout should be 3s, got %s", myConfig.Timeout)
		}

		// The default should be used as it is not set in the env or the file
		if myConfig.SomeOther != "hello" {
			t.Errorf("SomeOther should be hello, got %s", myConfig.SomeOther)
		}

		if myConfig.PublKey != multilineVar1 {
			t.Errorf("Error parsing multiline value for public key, got %s", myConfig.PublKey)
		}

		if myConfig.PrivKey != multilineVar2 {
			t.Errorf("Error parsing multiline value for private key, got %s", myConfig.PrivKey)
		}

		if myConfig.Port != 5432 {
			t.Errorf("Port should be 5432, got %d", myConfig.Port)
		}
	})
}

func TestParse(t *testing.T) {
	t.Run("all_defaults", func(t *testing.T) {
		cleanEnv()
		type MyConfig struct {
			Env         string        `env:"ENV" default:"dev"`
			Port        int           `env:"PORT" default:"8080"`
			EnableDebug bool          `env:"DEBUG" default:"false"`
			Timeout     time.Duration `env:"TIMEOUT" default:"5s"`
		}
		myConfig := MyConfig{}
		err := Parse(&myConfig, "testdata/.env.empty")
		if err != nil {
			t.Error(err)
		}
		if myConfig.Env != "dev" {
			t.Errorf("Env should be dev, got %s", myConfig.Env)
		}
		if myConfig.Port != 8080 {
			t.Errorf("Port should be 8080, got %d", myConfig.Port)
		}
		if myConfig.EnableDebug != false {
			t.Errorf("EnableDebug should be false, got %t", myConfig.EnableDebug)
		}
		if myConfig.Timeout != 5*time.Second {
			t.Errorf("Timeout should be 5s, got %s", myConfig.Timeout)
		}
	})

	t.Run("from_env", func(t *testing.T) {
		cleanEnv()
		type MyConfig struct {
			Env         string `env:"ENV"`
			EnableDebug bool   `env:"DEBUG"`
		}
		myConfig := MyConfig{}
		os.Setenv("ENV", "production")
		os.Setenv("DEBUG", "true")
		err := Parse(&myConfig, "testdata/.env.empty")
		if err != nil {
			t.Error(err)
		}
		if myConfig.Env != "production" {
			t.Errorf("Env should be production, got %s", myConfig.Env)
		}
		if !myConfig.EnableDebug {
			t.Errorf("EnableDebug should be true, got %t", myConfig.EnableDebug)
		}
	})

	t.Run("parses_durations", func(t *testing.T) {
		cleanEnv()
		type MyConfig struct {
			TimeoutS time.Duration `env:"TIMEOUT_SECS"`
			TimeoutM time.Duration `env:"TIMEOUT_MINS"`
		}
		myConfig := MyConfig{}
		os.Setenv("TIMEOUT_SECS", "23s")
		os.Setenv("TIMEOUT_MINS", "45m")
		err := Parse(&myConfig, "testdata/.env.empty")
		if err != nil {
			t.Error(err)
		}
		if myConfig.TimeoutS != 23*time.Second {
			t.Errorf("TimeoutSecs should be 23s, got %s", myConfig.TimeoutS)
		}
		if myConfig.TimeoutM != 45*time.Minute {
			t.Errorf("TimeoutMins should be 45m, got %s", myConfig.TimeoutM)
		}
	})

	t.Run("env_overrides_default", func(t *testing.T) {
		cleanEnv()
		type MyConfig struct {
			Env string `env:"ENV" default:"dev"`
		}
		myConfig := MyConfig{}
		os.Setenv("ENV", "production")
		err := Parse(&myConfig, "testdata/.env.empty")
		if err != nil {
			t.Error(err)
		}
		if myConfig.Env != "production" {
			t.Errorf("Env should be production, got %s", myConfig.Env)
		}
	})

	t.Run("errors_on_missing_env", func(t *testing.T) {
		cleanEnv()
		type MyConfig struct {
			Env  string `env:"ENV" default:"dev"`
			Port int    `env:"PORT"`
		}
		myConfig := MyConfig{}
		err := Parse(&myConfig, "testdata/.env.empty")
		want := "missing env var PORT (no default provided)"
		if err.Error() != want {
			t.Errorf("want error %q, got %q", want, err)
		}
	})

	t.Run("errors_on_wrong_type_int", func(t *testing.T) {
		cleanEnv()
		type MyConfig struct {
			Port int `env:"PORT"`
		}
		myConfig := MyConfig{}
		os.Setenv("PORT", "hello")
		err := Parse(&myConfig, "testdata/.env.empty")
		want := "cannot parse hello as int"
		if !strings.Contains(err.Error(), want) {
			t.Errorf("want error to contain %q, got %q", want, err)
		}
	})

	t.Run("errors_on_wrong_type_bool", func(t *testing.T) {
		cleanEnv()
		type MyConfig struct {
			EnableDebug bool `env:"DEBUG"`
		}
		myConfig := MyConfig{}
		os.Setenv("DEBUG", "hello")
		err := Parse(&myConfig, "testdata/.env.empty")
		want := "cannot parse hello as bool"
		if !strings.Contains(err.Error(), want) {
			t.Errorf("want error to contain %q, got %q", want, err)
		}
	})

	t.Run("errors_on_wrong_type_duration", func(t *testing.T) {
		cleanEnv()
		type MyConfig struct {
			Timeout time.Duration `env:"TIMEOUT" default:"5s"`
		}
		myConfig := MyConfig{}
		os.Setenv("TIMEOUT", "hello")
		err := Parse(&myConfig, "testdata/.env.empty")
		want := "cannot parse hello as time.Duration"
		if !strings.Contains(err.Error(), want) {
			t.Errorf("want error to contain %q, got %q", want, err)
		}
	})
}

// cleanEnv removes all env vars used for testing.
func cleanEnv() {
	os.Unsetenv("FOO")
	os.Unsetenv("ENV")
	os.Unsetenv("PORT")
	os.Unsetenv("TIMEOUT")
}
