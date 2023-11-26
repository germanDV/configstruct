# Config Struct

`configstruct` is a zero-dependency package that facilitates working with environment variables in Go applications.

## How It Works

You define a struct with the fields you want to load from environment variables, and map them using `env` struct tags.

All fields are considered required, a missing environment variable will cause an error unless a default value is provided.
To provide defaults, you can use the `default` struct tag.

The package exposes two functions:

- `Parse` -> reads environment variables and populates the config struct.
- `LoadAndParse` -> reads values from an **env** file and sets them in the environment (without overwriting existing variables), then calls `Parse`.

## ENV files

The path to the env file must be relative to the _root_ of the project.
By _root_ we mean the directory where a go.mod file is present.
The package will start at the current directory and walk upwards until it finds the root directory.

If a variable provided in the env file is already present in the environment, it will be skipped; in other words, the pre-existing value has precedence.

## Usage

Define a struct with the fields you want to load from environment variables:

```go
type MyConfig struct {
    NoQuotes     string        `env:"NO_QUOTES"`
    DoubleQuotes string        `env:"DOUBLE_QUOTES"`
    SingleQuotes string        `env:"SINGLE_QUOTES"`
    Timeout      time.Duration `env:"TIMEOUT"`
    SomeOther    string        `env:"SOME_OTHER" default:"hello"`
    PublKey      string        `env:"PUBLIC_KEY"`
    Port         int           `env:"PORT"`
    IsFoo        bool          `env:"IS_FOO" default:"false"`
}
```

Considering the following .env file:

```
# Comments are supported but only when they are on their own line

# It is not necessary to wrap values in quotes
NO_QUOTES=https://search.brave.com/search?q=somesearch&source=desktop

# However, double quotes are supported
DOUBLE_QUOTES="mongodb://127.0.0.1:27017/test?directConnection=true"

# And so are single quotes
SINGLE_QUOTES='https://api.github.com/'

# We use `time.ParseDuration`, which means things like 3m and 500ms are supported.
TIMEOUT=10s

# Numbers are casted to `int`
PORT=8080

# We use `strconv.ParseBool`, which means you can use multiple options, such as:
#     1, t, T, true, True, TRUE
#     0, f, F, false, False, FALSE
IS_FOO=true

# There's also support for multi-line variables. They must be sorrounded by double quotes
PUBLIC_KEY="-----BEGIN PUBLIC KEY-----
MFkwEwYHKoZIzj0CAQYIKoZIzj0DAQcDQgAEUOu0Nc9/EiVSyBKyfvv38MlteRWA
+6S8jpRIOC2eMn2kYSv1RCc7uejvLVc0EYn2spObZjsMv4qvNz0XxYduDQ==
-----END PUBLIC KEY-----"
```

Load values into the config struct:

```go
myConfig := MyConfig{}
err := configstruct.LoadAndParse(&myConfig, "./.env")
if err != nil {
    panic(err)
}
```

This will result in:

- `myConfig.NoQuotes` -> "https://search.brave.com/search?q=somesearch&source=desktop"
- `DoubleQuotes` -> "mongodb://127.0.0.1:27017/test?directConnection=true"
- `SingleQuotes` -> "https://api.github.com/"
- `myConfig.Timeout` -> time.Duration(10 * time.Second)
- `myConfig.SomeOther` -> "hello" (the default value)
- `myConfig.PublKey` -> "..."
- `myConfig.Port` -> 8080
- `myConfig.IsFoo` -> true

Considering that `LoadAndParse` will error if the env file doesn't exist, and that you may not have one of those in cloud environments and only use them for local development, you could check the environment to decide whether to call `LoadAndParse` or `Parse`:

```go
var myConfig = MyConfig{}
var err error

if os.Getenv("ENV") == "local" {
    err = configstruct.LoadAndParse(&myConfig, "./.env")
} else {
    err = configstruct.Parse(&myConfig)
}

if err != nil {
    panic(err)
}

// myConfig is now populated
```

## Supported Types

- string
- int
- bool
- time.Duration

## Not Supported

Types not listed above are not supported, for example, you cannot have a `float64` in your config struct.

Nested structs are not supported.

In-line comments in env files.
