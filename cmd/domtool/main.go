package main

import (
	"flag"
	"fmt"
	"os"

	"github.com/Cloud-Foundations/Dominator/lib/constants"
	"github.com/Cloud-Foundations/Dominator/lib/flags/loadflags"
	"github.com/Cloud-Foundations/Dominator/lib/flagutil"
	"github.com/Cloud-Foundations/Dominator/lib/srpc"
	"github.com/Cloud-Foundations/Dominator/lib/srpc/setupclient"
)

var (
	cpuPercent = flag.Uint("cpuPercent", 0,
		"CPU speed as percentage of capacity (default 50)")
	networkSpeedPercent = flag.Uint("networkSpeedPercent",
		constants.DefaultNetworkSpeedPercent,
		"Network speed as percentage of capacity")
	scanExcludeList  flagutil.StringList = constants.ScanExcludeList
	scanSpeedPercent                     = flag.Uint("scanSpeedPercent",
		constants.DefaultScanSpeedPercent,
		"Scan speed as percentage of capacity")
	domHostname = flag.String("domHostname", "localhost",
		"Hostname of dominator")
	domPortNum = flag.Uint("domPortNum", constants.DominatorPortNumber,
		"Port number of dominator")
)

func init() {
	flag.Var(&scanExcludeList, "scanExcludeList",
		"Comma separated list of patterns to exclude from scanning")
}

func printSubcommands(subcommands []subcommand) {
	for _, subcommand := range subcommands {
		if subcommand.args == "" {
			fmt.Fprintln(os.Stderr, " ", subcommand.command)
		} else {
			fmt.Fprintln(os.Stderr, " ", subcommand.command, subcommand.args)
		}
	}
}

func printUsage() {
	fmt.Fprintln(os.Stderr,
		"Usage: domtool [flags...] command")
	fmt.Fprintln(os.Stderr, "Common flags:")
	flag.PrintDefaults()
	fmt.Fprintln(os.Stderr, "Commands:")
	printSubcommands(subcommands)
}

type commandFunc func(*srpc.Client, []string)

type subcommand struct {
	command string
	args    string
	numArgs int
	cmdFunc commandFunc
}

var subcommands = []subcommand{
	{"clear-safety-shutoff", "sub", 1, clearSafetyShutoffSubcommand},
	{"configure-subs", "", 0, configureSubsSubcommand},
	{"disable-updates", "reason", 1, disableUpdatesSubcommand},
	{"enable-updates", "reason", 1, enableUpdatesSubcommand},
	{"get-default-image", "", 0, getDefaultImageSubcommand},
	{"get-subs-configuration", "", 0, getSubsConfigurationSubcommand},
	{"set-default-image", "", 1, setDefaultImageSubcommand},
}

func main() {
	if err := loadflags.LoadForCli("domtool"); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
	flag.Usage = printUsage
	flag.Parse()
	if flag.NArg() < 1 {
		printUsage()
		os.Exit(2)
	}
	if err := setupclient.SetupTls(true); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
	clientName := fmt.Sprintf("%s:%d", *domHostname, *domPortNum)
	client, err := srpc.DialHTTP("tcp", clientName, 0)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error dialing\t%s\n", err)
		os.Exit(1)
	}
	for _, subcommand := range subcommands {
		if flag.Arg(0) == subcommand.command {
			if flag.NArg()-1 != subcommand.numArgs {
				printUsage()
				os.Exit(2)
			}
			subcommand.cmdFunc(client, flag.Args()[1:])
			os.Exit(3)
		}
	}
	printUsage()
	os.Exit(2)
}
