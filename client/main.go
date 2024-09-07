package main

import (
	"fmt"
	"os"
	"strings"
	"time"
	"encoding/csv"

	"github.com/op/go-logging"
	"github.com/pkg/errors"
	"github.com/spf13/viper"

	"github.com/7574-sistemas-distribuidos/docker-compose-init/client/common"
)

var log = logging.MustGetLogger("log")

// InitConfig Function that uses viper library to parse configuration parameters.
// Viper is configured to read variables from both environment variables and the
// config file ./config.yaml. Environment variables takes precedence over parameters
// defined in the configuration file. If some of the variables cannot be parsed,
// an error is returned
func InitConfig() (*viper.Viper, error) {
	v := viper.New()

	// Configure viper to read env variables with the CLI_ prefix
	v.AutomaticEnv()
	v.SetEnvPrefix("cli")
	v.SetEnvKeyReplacer(strings.NewReplacer(".", "_"))

	// Add env variables supported
	v.BindEnv("id")
	v.BindEnv("server", "address")
	v.BindEnv("loop", "period")
	v.BindEnv("loop", "amount")
	v.BindEnv("log", "level")
	v.BindEnv("batch", "maxAmount")

	v.SetConfigFile("./config.yaml")
	if err := v.ReadInConfig(); err != nil {
		fmt.Printf("Configuration could not be read from config file. Using env variables instead")
	}

	durationStr := v.GetString("loop.period")
	if durationStr == "" {
		return nil, errors.New("loop.period is not set")
	}

	if _, err := time.ParseDuration(durationStr); err != nil {
		return nil, errors.Wrapf(err, "Could not parse CLI_LOOP_PERIOD env var as time.Duration.")
	}

	return v, nil
}

// InitLogger Receives the log level to be set in go-logging as a string. This method
// parses the string and set the level to the logger. If the level string is not
// valid an error is returned
func InitLogger(logLevel string) error {
	baseBackend := logging.NewLogBackend(os.Stdout, "", 0)
	format := logging.MustStringFormatter(
		`%{time:2006-01-02 15:04:05} %{level:.5s}     %{message}`,
	)
	backendFormatter := logging.NewBackendFormatter(baseBackend, format)

	backendLeveled := logging.AddModuleLevel(backendFormatter)
	logLevelCode, err := logging.LogLevel(logLevel)
	if err != nil {
		return err
	}
	backendLeveled.SetLevel(logLevelCode, "")

	logging.SetBackend(backendLeveled)
	return nil
}

// PrintConfig Print all the configuration parameters of the program.
// For debugging purposes only
func PrintConfig(v *viper.Viper) {
	log.Infof("action: config | result: success | client_id: %s | server_address: %s | loop_amount: %v | loop_period: %v | log_level: %s",
		v.GetString("id"),
		v.GetString("server.address"),
		v.GetInt("loop.amount"),
		v.GetDuration("loop.period"),
		v.GetString("log.level"),
	)
}

func ReadBets(fileName string, id string) func(int) ([]map[string]string, error) {
	file, err := os.Open(fileName)
	if err != nil {
		log.Fatalf("Could not open file %s: %s", fileName, err)
	}
	log.Debugf("action: open_file %v | result: success", fileName)
	reader := csv.NewReader(file)

	return func(batchSize int) ([]map[string]string, error) {
		var bets []map[string]string
		for i := 0; i < batchSize; i++ {
			record, err := reader.Read()
			if err != nil {
				if err.Error() == "EOF" {
					return bets, nil
				}
				return nil, errors.Wrap(err, "error reading CSV file")
			}

			if len(record) != 5 {
				log.Warningf("Skipping invalid record: %v", record)
				continue
			}

			bet := map[string]string{
				"NOMBRE":    record[0],
				"APELLIDO":  record[1],
				"DOCUMENTO": record[2],
				"NACIMIENTO": record[3],
				"NUMERO":    record[4],
			}
			bets = append(bets, bet)
		}
		return bets, nil
	}
}

func NotifyEndOfBets(client *common.Client, config common.ClientConfig) error {
    if err := client.NotifyBetsFinished(); err != nil {
        log.Errorf("action: notify_bets_finished | result: fail | client_id: %v | error: %v", config.ID, err)
        return err
    }
    log.Infof("action: notify_bets_finished | result: success | client_id: %v", config.ID)
    return nil
}


func GetWinners(client *common.Client, config common.ClientConfig) (string, error) {
    winners := client.GetWinners()
    return winners, nil
}


func main() {
	v, err := InitConfig()
	if err != nil {
		log.Criticalf("%s", err)
		os.Exit(1)
	}

	if err := InitLogger(v.GetString("log.level")); err != nil {
		log.Criticalf("%s", err)
		os.Exit(1)
	}

	PrintConfig(v)

	clientConfig := common.ClientConfig{
		ServerAddress: v.GetString("server.address"),
		ID:            v.GetString("id"),
		LoopAmount:    v.GetInt("loop.amount"),
		LoopPeriod:    v.GetDuration("loop.period"),
		BatchMaxSize:  v.GetInt("batch.maxAmount"),
	}

	client := common.NewClient(clientConfig)

	// Get the function to read bets from the CSV file
	betDataRead := ReadBets(v.GetString("data.file"), clientConfig.ID)

	for {
		batch, err := betDataRead(clientConfig.BatchMaxSize)
		if err != nil {
			log.Criticalf("Error reading bets: %v", err)
			os.Exit(1)
		}
		if len(batch) == 0 {
			break
		}

		err = client.SendBets(batch)
		if err != nil {
			log.Criticalf("Error sending bets: %v", err)
			os.Exit(1)
		}

	}

	if err := NotifyEndOfBets(client, clientConfig); err != nil {
		os.Exit(1)
	}

	var winners string
	for {
		winners, err = GetWinners(client, clientConfig)
		if err != nil {
			log.Errorf("action: consulta_ganadores | result: fail | error: %v", err)
			time.Sleep(5 * time.Second)
			continue
		}

		if strings.Contains(winners, "Sorteo no realizado") {
			log.Infof("Sorteo no realizado aún, reintentando...")
			time.Sleep(5 * time.Second)
		} else {
			break
		}
	}
	log.Infof("action: consulta_ganadores | result: success | cant_ganadores: %d", len(strings.Split(winners, "\n")))
	os.Exit(0)
}
