package fit

import (
	"context"
	"fmt"
	"golang.org/x/oauth2"
	"golang.org/x/oauth2/google"
	"google.golang.org/api/fitness/v1"
	"google.golang.org/api/option"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"strconv"
	"weight-interceptor-http/parser"
	"weight-interceptor-http/storage"
)

const credentialsPath = "data/credentials.json"

func IsAvailable() bool {
	_, err := os.Stat(credentialsPath)
	return err == nil
}

func Authenticate(user string) error {
	path := fmt.Sprintf("data/%s.json", user)
	token, err := storage.ReadTokenFromFile(path)
	if err != nil {
		config, err := getConfig()
		if err != nil {
			return err
		}
		token, err = getTokenFromWeb(config)
		if err != nil {
			return err
		}
		err = storage.SaveTokenToFile(path, token)
		if err != nil {
			return err
		}
	}
	return nil
}

func AddWeight(weight parser.Weight) error {
	user, err := storage.FindUser(weight.GetWeight())
	if err != nil {
		return err
	}
	service, err := getService(user)
	if err != nil {
		return err
	}
	dataSource, err := getOrCreateDataSource(service, weight.Uid)
	if err != nil {
		return err
	}
	return createDataSet(service, dataSource.DataStreamId, weight)
}

func getService(user string) (*fitness.Service, error) {
	config, err := getConfig()
	if err != nil {
		return nil, err
	}
	client, err := getClient(config, user)
	if err != nil {
		return nil, err
	}
	ctx := context.Background()
	service, err := fitness.NewService(ctx, option.WithHTTPClient(client))
	if err != nil {
		return nil, fmt.Errorf("unable to create fitness client: %w", err)
	}
	return service, nil
}

func getConfig() (*oauth2.Config, error) {
	bytes, err := ioutil.ReadFile(credentialsPath)
	if err != nil {
		return nil, fmt.Errorf("unable to read client secret file: %w", err)
	}
	config, err := google.ConfigFromJSON(bytes, fitness.FitnessBodyWriteScope)
	if err != nil {
		return nil, fmt.Errorf("unable to parse client secret file to config: %w", err)
	}
	return config, nil
}

func getClient(config *oauth2.Config, user string) (*http.Client, error) {
	path := fmt.Sprintf("data/%s.json", user)
	token, err := storage.ReadTokenFromFile(path)
	if err != nil {
		return nil, err
	}
	return config.Client(context.Background(), token), nil
}

func getTokenFromWeb(config *oauth2.Config) (*oauth2.Token, error) {
	authURL := config.AuthCodeURL("state-token", oauth2.AccessTypeOffline)
	fmt.Printf("Open the following URL in your browser: \n%v\n", authURL)

	var authCode string
	if _, err := fmt.Scan(&authCode); err != nil {
		return nil, fmt.Errorf("unable to read authorization code: %w", err)
	}

	token, err := config.Exchange(context.Background(), authCode)
	if err != nil {
		return nil, fmt.Errorf("unable to retrieve token from web: %w", err)
	}
	return token, nil
}

func getOrCreateDataSource(service *fitness.Service, uid string) (*fitness.DataSource, error) {
	path := fmt.Sprintf("data/%s.txt", uid)
	id, err := storage.ReadDataSourceIdFromFile(path)
	if err != nil {
		dataSource, err := getDataSourceIdFromWeb(service, uid)
		if err != nil {
			return nil, err
		}
		err = storage.SaveDataSourceIdToFile(path, dataSource.DataStreamId)
		if err != nil {
			return nil, err
		}
		return dataSource, nil
	}
	return service.Users.DataSources.Get("me", id).Do()
}

func getDataSourceIdFromWeb(service *fitness.Service, uid string) (*fitness.DataSource, error) {
	dataSources, err := service.Users.DataSources.List("me").Do()
	if err != nil {
		log.Printf("Unable to list data sources: %v\n", err)
	}
	for _, dataSource := range dataSources.DataSource {
		if dataSource.DataStreamName == "WeightInterceptor" && dataSource.Device.Uid == uid {
			return dataSource, nil
		}
	}

	dataSource, err := service.Users.DataSources.Create("me", &fitness.DataSource{
		Application: &fitness.Application{
			Name:    "Weight Interceptor",
			Version: "1",
		},
		DataStreamName: "WeightInterceptor",
		DataType: &fitness.DataType{
			Name: "com.google.weight",
		},
		Device: &fitness.Device{
			Manufacturer: "Soehnle",
			Model:        "Web Connect",
			Type:         "scale",
			Uid:          uid,
		},
		Type: "raw",
	}).Do()
	if err != nil {
		return nil, fmt.Errorf("unable to create data source: %w", err)
	}
	return dataSource, nil
}

func createDataSet(service *fitness.Service, dataSourceId string, weight parser.Weight) error {
	nanoTime := weight.Time.UnixNano()
	dataSetId := strconv.FormatInt(nanoTime, 10) + "-" + strconv.FormatInt(nanoTime, 10)
	_, err := service.Users.DataSources.Datasets.Patch("me", dataSourceId, dataSetId, &fitness.Dataset{
		DataSourceId:   dataSourceId,
		MaxEndTimeNs:   nanoTime,
		MinStartTimeNs: nanoTime,
		Point: []*fitness.DataPoint{{
			DataTypeName:   "com.google.weight",
			EndTimeNanos:   nanoTime,
			StartTimeNanos: nanoTime,
			Value: []*fitness.Value{{
				FpVal: weight.GetWeight(),
			}},
		}},
	}).Do()
	if err != nil {
		return fmt.Errorf("unable to create data set: %w", err)
	}
	return nil
}
