package config

import (
	"flag"
	"log"
	"os"

	"github.com/ilyakaznacheev/cleanenv"
)

// creating a separate struct for nested value
type HTTPServer struct {
	Address string `yaml:"address" env-required:"true"`
}

type Config struct {
	Env         string `yaml:"env" env:"ENV" env-required:"true"`
	StoragePath string `yaml:"storage_path" env:"STORAGE_PATH" env-required:"true"`
	HTTPServer  `yaml:"http_server"`
	JWTSecret   string `yaml:"jwt_secret" env:"JWT_SECRET" env-required:"true"`
}

// using a right naming convention, a Must function does not return an error
// it kills the program on failure.The prefix is a warning to the caller:
// "you get no error to handle, so only call me somewhere that dying is acceptable."
// return the pointer to the Config struct
func MustLoad() *Config {
	//taking a variable
	var configPath string

	//using the built-in package(os) and using Getenv function, it takes CONFIG_PATHas a string and return the value in it as a string
	configPath = os.Getenv("CONFIG_PATH")

	//Checking if configPath does not have any value
	if configPath == "" {
		//we'll use the flag package so we can manually set the config path in command line?
		configPathFlag := flag.String("config", "", "path to the configuration file")
		flag.Parse()

		//writting it as *configPathFlag as they are the pointer to the variable
		configPath = *configPathFlag

		//if it is still empty throw Fatal error
		if configPath == "" {
			log.Fatal("Config Path is not Set")
		}
	}

	//using os.Stat to get info about the 'configPath' and using IsNotExist to check if the err is all about the file doesn't even exists?
	if _, err := os.Stat(configPath); os.IsNotExist(err) {
		log.Fatalf("config file does not exist %s", configPath)
	}

	//Now after getting through all of the above checks this code will execute

	//taking a variable named 'cfg' on the struct Config
	var cfg Config

	//reading the configuration file using cleanenv.
	//Passing &cfg hands over the address so the function can write into the variable
	err := cleanenv.ReadConfig(configPath, &cfg)

	//if an error occurred log it
	if err != nil {
		log.Fatalf("can not read config file: %s", err.Error())
	}

	return &cfg
}
