package main

import (
	"bytes"
	"fmt"
	"os/exec"

	"os"

	"strings"

	"github.com/Sirupsen/logrus"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/iam"
	_ "github.com/aws/aws-sdk-go/service/iam/iamiface"
	"gopkg.in/alecthomas/kingpin.v2"
)

var (
	stupsGit = "https://github.com/zalando-stups/"
)

var requirements = []string{"mai", "senza", "piu"}

type senzaItem struct {
	stack *string
	ip    *string
}

func main() {
	login()
	alias, err := getAlias()
	if err != nil {
		logrus.Fatalln(err)
	}
	logrus.Println(alias)
	getHosts()
	checkDep()
	app := kingpin.New("gopiu", "Wrapper for piu to save time typing shit")

	connect := app.Command("connect", "Connects to the specified IP")
	host := connect.Arg("ip", "Private IP of the host to connect").String()
	connect.Action(func(ctx *kingpin.ParseContext) error {
		fmt.Println(host)
		return nil
	})

	list := app.Command("list", "List down all instances on current profile")
	list.Action(func(ctx *kingpin.ParseContext) error {
		return nil
	})

	setDefault := app.Command("set-default", "Set the default mai profile")
	setDefault.Action(func(ctx *kingpin.ParseContext) error {
		return nil
	})

	kingpin.MustParse(app.Parse(os.Args[1:]))
}

func checkDep() {
	var err error

	_, err = exec.LookPath("senza")
	if err != nil {
		logrus.Fatalf("Please install %s. Refer to %s", "senza", stupsGit)
	}

	_, err = exec.LookPath("piu")
	if err != nil {
		logrus.Fatalf("Please install %s. Refer to %s", "piu", stupsGit)
	}

	_, err = exec.LookPath("mai")
	if err != nil {
		logrus.Fatalf("Please install %s. Refer to %s", "mai", stupsGit)
	}
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

func getHosts() ([]*string, error) {

	senza := exec.Command("senza", "inst")
	var output bytes.Buffer
	senza.Stdout = &output
	err := senza.Run()

	if err != nil {
		return nil, err
	}

	table := strings.Split(output.String(), "\n")
	table = table[1:] //skip header
	for _, row := range table {
		cols := strings.Fields(row)
		fmt.Println(cols, len(cols))
	}
	return nil, err
}
