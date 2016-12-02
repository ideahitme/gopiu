package main

import (
	"bytes"
	"errors"
	"fmt"
	"log"
	"os"
	"os/exec"
	"strings"

	"github.com/Sirupsen/logrus"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/iam"
	"gopkg.in/alecthomas/kingpin.v2"
)

var (
	stupsGit = "https://github.com/zalando-stups/"
)

var requirements = []string{"mai", "senza", "piu"}

type senzaItem struct {
	stack string
	ip    string
}

func (item *senzaItem) String() string {
	return item.stack + " - " + item.ip
}

func main() {
	var accountName string

	if err := login(); err != nil {
		log.Fatalln(err)
	}

	if err := checkDep(); err != nil {
		log.Fatalln(err)
	}

	if alias, err := getAlias(); err != nil {
		logrus.Fatalln(err)
	} else {
		accountName = alias
	}

	logrus.Infoln("Account: ", accountName)
	app := kingpin.New("gopiu", "Wrapper for piu to save time typing shit")

	connect := app.Command("connect", "Connects to the specified IP")
	host := connect.Arg("ip", "Private IP of the host to connect").String()
	connect.Action(func(ctx *kingpin.ParseContext) error {
		item := &senzaItem{
			stack: "",
			ip:    *host,
		}
		return item.connect(accountName)
	})

	list := app.Command("list", "List down all running instances on current profile")
	list.Action(func(ctx *kingpin.ParseContext) error {
		hosts, err := getHosts()
		if err != nil {
			return err
		}

		for _, host := range hosts {
			logrus.Infoln(host)
		}
		return nil
	})

	kingpin.MustParse(app.Parse(os.Args[1:]))
}

func checkDep() error {
	var err error

	for _, require := range requirements {
		_, err = exec.LookPath(require)
		if err != nil {
			return fmt.Errorf("Please install %s. Refer to %s", require, stupsGit)
		}
	}

	return nil
}

func login() error {
	mai := exec.Command("mai", "list")

	err := mai.Run()
	if err != nil {
		return err
	}
	return nil
}

func getAlias() (string, error) {
	sess := session.New(nil)

	client := iam.New(sess)
	input := &iam.ListAccountAliasesInput{}

	output, err := client.ListAccountAliases(input)

	if err != nil {
		return "", err
	}
	return strings.Split(*output.AccountAliases[0], "-")[1], nil

}

func getHosts() ([]*senzaItem, error) {
	senza := exec.Command("senza", "inst")
	var output bytes.Buffer
	senza.Stdout = &output
	err := senza.Run()
	if err != nil {
		return nil, err
	}

	hosts := make([]*senzaItem, 0)
	table := strings.Split(output.String(), "\n")
	table = table[1:] //skip header
	for _, row := range table {
		cols := strings.Fields(row)
		if stack, privateIP := extractIP(cols); !isOddServer(cols) && privateIP != "" {
			hosts = append(hosts, &senzaItem{
				stack: stack,
				ip:    privateIP,
			})
		}
	}

	if len(hosts) == 0 {
		return nil, errors.New("No available instances are found")
	}
	return hosts, err
}

func isOddServer(row []string) bool {
	for _, col := range row {
		if strings.Index(col, "OddServer") != -1 {
			return true
		}
	}
	return false
}

func extractIP(row []string) (string, string) {
	for i := range row {
		if row[i] == "RUNNING" && i > 0 {
			return row[0] + "-" + row[1], row[i-1]
		}
	}
	return "", ""
}

func (item *senzaItem) connect(accountName string) error {
	logrus.Infof("Connecting to %s at %s", item.stack, item.ip)
	piu := exec.Command("piu", item.ip, "debug", "-O", fmt.Sprintf("odd-eu-central-1.%s.zalan.do", accountName))
	var buf, bufError bytes.Buffer
	piu.Stdout = &buf
	piu.Stderr = &bufError
	fmt.Println(buf.String(), bufError.String())
	return piu.Run()
}
