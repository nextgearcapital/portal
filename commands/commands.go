// commands is where all cli logic is, including starting portal as a server.
package commands

import (
	"crypto/tls"
	"fmt"
	"io"
	"net/http"
	"os"
	"os/signal"
	"syscall"

	"github.com/jcelliott/lumber"
	"github.com/spf13/cobra"

	"github.com/nanopack/portal/api"
	"github.com/nanopack/portal/balance"
	"github.com/nanopack/portal/cluster"
	"github.com/nanopack/portal/config"
	"github.com/nanopack/portal/core"
	"github.com/nanopack/portal/database"
	"github.com/nanopack/portal/proxymgr"
)

var (
	runServer bool
	Portal    = &cobra.Command{
		Use:   "portal",
		Short: "portal - load balancer/proxy",
		Long:  ``,

		Run: startPortal,
	}
)

func init() {
	Portal.PersistentFlags().BoolVarP(&config.Insecure, "insecure", "i", config.Insecure, "Disable tls key checking (client) and listen on http (server)")
	Portal.PersistentFlags().StringVarP(&config.ApiToken, "api-token", "t", config.ApiToken, "Token for API Access")
	Portal.PersistentFlags().StringVarP(&config.ApiHost, "api-host", "H", config.ApiHost, "Listen address for the API")
	Portal.PersistentFlags().StringVarP(&config.ApiPort, "api-port", "P", config.ApiPort, "Listen address for the API")
	Portal.PersistentFlags().StringVarP(&config.ConfigFile, "conf", "c", config.ConfigFile, "Configuration file to load")

	Portal.Flags().StringVarP(&config.ApiKey, "api-key", "k", config.ApiKey, "SSL key for the api")
	Portal.Flags().StringVarP(&config.ApiCert, "api-cert", "C", config.ApiCert, "SSL cert for the api")
	Portal.Flags().StringVarP(&config.ApiKeyPassword, "api-key-password", "p", config.ApiKeyPassword, "Password for the SSL key")
	Portal.Flags().StringVarP(&config.DatabaseConnection, "db-connection", "d", config.DatabaseConnection, "Database connection string")
	Portal.Flags().StringVarP(&config.ClusterConnection, "cluster-connection", "r", config.ClusterConnection, "Cluster connection string (redis://127.0.0.1:6379)")
	Portal.Flags().StringVarP(&config.ClusterToken, "cluster-token", "T", config.ClusterToken, "Cluster security token")
	Portal.Flags().StringVarP(&config.LogLevel, "log-level", "l", config.LogLevel, "Log level to output")
	Portal.Flags().StringVarP(&config.LogFile, "log-file", "L", config.LogFile, "Log file to write to")

	Portal.Flags().BoolVarP(&config.Server, "server", "s", config.Server, "Run in server mode")

	Portal.AddCommand(serviceAddCmd)
	Portal.AddCommand(serviceRemoveCmd)
	Portal.AddCommand(serviceShowCmd)
	Portal.AddCommand(servicesShowCmd)
	Portal.AddCommand(servicesSetCmd)
	Portal.AddCommand(serviceSetCmd)

	Portal.AddCommand(serverAddCmd)
	Portal.AddCommand(serverRemoveCmd)
	Portal.AddCommand(serverShowCmd)
	Portal.AddCommand(serversShowCmd)
	Portal.AddCommand(serversSetCmd)

	Portal.AddCommand(routeAddCmd)
	Portal.AddCommand(routesSetCmd)
	Portal.AddCommand(routesShowCmd)
	Portal.AddCommand(routeRemoveCmd)

	Portal.AddCommand(certAddCmd)
	Portal.AddCommand(certsSetCmd)
	Portal.AddCommand(certsShowCmd)
	Portal.AddCommand(certRemoveCmd)
}

func startPortal(ccmd *cobra.Command, args []string) {
	if err := config.LoadConfigFile(); err != nil {
		config.Log.Fatal("Failed to read config - %v", err)
		os.Exit(1)
	}

	if !config.Server {
		ccmd.HelpFunc()(ccmd, args)
		return
	}

	if config.LogFile == "" {
		config.Log = lumber.NewConsoleLogger(lumber.LvlInt(config.LogLevel))
	} else {
		var err error
		config.Log, err = lumber.NewFileLogger(config.LogFile, lumber.LvlInt(config.LogLevel), lumber.ROTATE, 5000, 9, 100)
		if err != nil {
			config.Log.Fatal("File logger init failed - %v", err)
			os.Exit(1)
		}
	}
	// initialize database
	err := database.Init()
	if err != nil {
		config.Log.Fatal("Database init failed - %v", err)
		os.Exit(1)
	}
	// initialize balancer
	err = balance.Init()
	if err != nil {
		config.Log.Fatal("Balancer init failed - %v", err)
		os.Exit(1)
	}
	// initialize proxymgr
	err = proxymgr.Init()
	if err != nil {
		config.Log.Fatal("Proxymgr init failed - %v", err)
		os.Exit(1)
	}
	// initialize cluster
	err = cluster.Init()
	if err != nil {
		config.Log.Fatal("Cluster init failed - %v", err)
		os.Exit(1)
	}

	go sigHandle()

	// start api
	err = api.StartApi()
	if err != nil {
		config.Log.Fatal("Api start failed - %v", err)
		os.Exit(1)
	}
	return
}

func sigHandle() {
	sigs := make(chan os.Signal, 1)
	signal.Notify(sigs, syscall.SIGINT, syscall.SIGTERM)
	go func() {
		switch <-sigs {
		default:
			// clear balancer rules - (stop balancing if we are offline)
			balance.SetServices(make([]core.Service, 0, 0))
			fmt.Println()
			os.Exit(0)
		}
	}()
}

func rest(path string, method string, body io.Reader) (*http.Response, error) {
	var client *http.Client
	client = http.DefaultClient
	uri := fmt.Sprintf("https://%s:%s/%s", config.ApiHost, config.ApiPort, path)

	if config.Insecure {
		http.DefaultTransport.(*http.Transport).TLSClientConfig = &tls.Config{InsecureSkipVerify: true}
		uri = fmt.Sprintf("http://%s:%s/%s", config.ApiHost, config.ApiPort, path)
	}

	req, err := http.NewRequest(method, uri, body)
	if err != nil {
		panic(err)
	}
	req.Header.Add("X-NANOBOX-TOKEN", config.ApiToken)
	res, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	if res.StatusCode == 401 {
		return nil, fmt.Errorf("401 Unauthorized. Please specify api token (-t 'token')")
	}
	return res, nil
}

func fail(format string, args ...interface{}) {
	fmt.Printf(fmt.Sprintf("%v\n", format), args...)
	os.Exit(1)
}
