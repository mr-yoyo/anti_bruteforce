package cmd

import (
	"encoding/json"
	"fmt"
	"log"
	"net"
	"net/http"
	"strconv"

	"github.com/go-resty/resty/v2"
	"github.com/spf13/cobra"
)

var (
	client       = resty.New()
	listKind     string
	host         string
	port         int
	login        string
	IP           net.IP
	allowedKinds = map[string]bool{
		"blacklist": true,
		"whitelist": true,
	}
)

var ctlCommand = &cobra.Command{Use: "ctl", Short: "Control application"}

func init() {
	ctlCommand.Flags().StringVarP(&host, "host", "H", "localhost", "app hostname")
	ctlCommand.Flags().IntVarP(&port, "port", "p", 8080, "app port")
	ctlCommand.Flags().StringVarP(&listKind, "kind", "k", "", "list kind: blacklist or whitelist")
	ctlCommand.MarkFlagRequired("kind")

	resetBucketCommand.Flags().StringVarP(&login, "login", "l", "", "login to reset in bucket")
	resetBucketCommand.Flags().IPVar(&IP, "ip", net.IP{}, "ip to reset in bucket")
	resetBucketCommand.Flags().StringVarP(&host, "host", "H", "localhost", "app hostname")
	resetBucketCommand.Flags().IntVarP(&port, "port", "p", 8080, "app port")
	ctlCommand.AddCommand(addListCommand, deleteListCommand, resetBucketCommand)
}

var addListCommand = &cobra.Command{
	Use:   "add",
	Short: "Add whitelist/blacklist",
	Args:  cobra.MinimumNArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		initClient()

		if !allowedKinds[listKind] {
			log.Fatalf("invalid kind: %s", listKind)
		}

		for _, v := range args {
			_, ip4Net, err := net.ParseCIDR(v)
			if err != nil {
				log.Fatalf("invalid ip range: %s", err)
			}

			makeAddIpNetToList(ip4Net, listKind)
		}

		return nil
	},
}

var deleteListCommand = &cobra.Command{
	Use:   "delete",
	Short: "Delete whitelist/blacklist",
	Args:  cobra.MinimumNArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		initClient()

		if !allowedKinds[listKind] {
			log.Fatalf("invalid kind: %s", listKind)
		}

		for _, v := range args {
			_, ip4Net, err := net.ParseCIDR(v)
			if err != nil {
				log.Fatalf("invalid ip range: %s", err)
			}

			makeDeleteIpNetToList(ip4Net, listKind)
		}

		return nil
	},
}

var resetBucketCommand = &cobra.Command{
	Use:   "reset",
	Short: "Reset bucket",
	RunE: func(cmd *cobra.Command, args []string) error {
		initClient()

		if login == "" && IP == nil {
			log.Fatalf("login or ip must be set")
		}

		makeResetBucket(login, IP)

		return nil
	},
}

func makeAddIpNetToList(ip4Net *net.IPNet, kind string) {
	body, _ := json.Marshal(struct {
		Subnet string `json:"subnet"`
	}{
		Subnet: ip4Net.String(),
	})
	r, err := client.R().
		SetHeader("Content-Type", "application/json").
		SetBody(string(body)).
		Post("/" + kind)
	if err != nil {
		log.Fatalf("error while executing request: %s", err.Error())
	}

	if r.StatusCode() != http.StatusOK {
		log.Fatalf("error: %s", string(r.Body()))
	} else {
		fmt.Printf("%s successfully added to %s\n", ip4Net.String(), kind)
	}
}

func makeDeleteIpNetToList(ip4Net *net.IPNet, kind string) {
	body, _ := json.Marshal(struct {
		Subnet string `json:"subnet"`
	}{
		Subnet: ip4Net.String(),
	})
	r, err := client.R().
		SetHeader("Content-Type", "application/json").
		SetBody(string(body)).
		Delete("/" + kind)
	if err != nil {
		log.Fatalf("error while executing request: %s", err.Error())
	}

	switch r.StatusCode() {
	case http.StatusOK:
		fmt.Printf("%s successfully deleted from %s\n", ip4Net.String(), kind)
	case http.StatusNoContent:
		fmt.Printf("%s not found in %s\n", ip4Net.String(), kind)
	default:
		log.Fatalf("error: %s", string(r.Body()))
	}
}

func makeResetBucket(login string, IP net.IP) {
	body, _ := json.Marshal(struct {
		Login string `json:"login"`
		IP    net.IP `json:"ip"`
	}{
		Login: login,
		IP:    IP,
	})

	r, err := client.R().
		SetHeader("Content-Type", "application/json").
		SetBody(string(body)).
		Delete("/bucket")
	if err != nil {
		log.Fatalf("error while executing request: %s", err.Error())
	}

	switch r.StatusCode() {
	case http.StatusOK:
		fmt.Println("bucket successfully deleted")
	case http.StatusNoContent:
		fmt.Println("bucket not found")
	default:
		log.Fatalf("error: %s", string(r.Body()))
	}
}

func initClient() {
	client.SetBaseURL("http://" + host + ":" + strconv.Itoa(port))
}
