package cmd

import (
	"bytes"
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"sort"
	"strings"
	"text/template"

	"github.com/saveio/edge/cmd/flags"

	"github.com/urfave/cli"
)

// AppHelpTemplate is the test template for the default, global app help topic.
var (
	AppHelpTemplate = `NAME:
  {{.App.Name}} - {{.App.Usage}}

	Edge CLI is an Edge node command line Client for starting and managing Edge nodes,
	managing user wallets, sending transactions, deploying and invoking contract, and so on.

USAGE:
  {{.App.HelpName}} [options]{{if .App.Commands}} command [command options] {{end}}{{if .App.ArgsUsage}}{{.App.ArgsUsage}}{{else}}[arguments...]{{end}}
  {{if .App.Version}}
VERSION:
  {{.App.Version}}
  {{end}}{{if len .App.Authors}}
AUTHOR(S):
  {{range .App.Authors}}{{ . }}{{end}}
  {{end}}{{if .App.Commands}}
COMMANDS:
  {{range .App.Commands}}{{join .Names ", "}}{{ "  " }}{{.Usage}}
  {{end}}{{end}}{{if .FlagGroups}}
{{range .FlagGroups}}{{.Name}} OPTIONS:
  {{range .Flags}}{{.}}
  {{end}}
{{end}}{{end}}{{if .App.Copyright }}COPYRIGHT: 
  {{.App.Copyright}}
{{end}}
`
	SubcommandHelpTemplate = `NAME:
   {{.HelpName}} - {{if .Description}}{{.Description}}{{else}}{{.Usage}}{{end}}

USAGE:
   {{if .UsageText}}{{.UsageText}}{{else}}{{.HelpName}} command{{if .VisibleFlags}} [command options]{{end}} {{if .ArgsUsage}}{{.ArgsUsage}}{{else}}[arguments...]{{end}}{{end}}

COMMANDS:
  {{range .Commands}}{{join .Names ", "}}{{ "  " }}{{.Usage}}
  {{end}}{{if .VisibleFlags}}
OPTIONS:
   {{range .VisibleFlags}}{{.}}
   {{end}}{{end}}
`
	CommandHelpTemplate = `
USAGE:
	{{if .cmd.UsageText}}{{.cmd.UsageText}}{{else}}{{.cmd.HelpName}}{{if .cmd.VisibleFlags}} [command options]{{end}} 
	{{if .cmd.ArgsUsage}}{{.cmd.ArgsUsage}}{{else}}[arguments...]{{end}}{{end}}{{if .cmd.Description}}

DESCRIPTION:
	{{.cmd.Description}}
	{{end}}{{if .cmd.Subcommands}}
SUBCOMMANDS:
	{{range .cmd.Subcommands}}{{.Name}}{{with .ShortName}}, {{.}}{{end}}{{ "\t" }}{{.Usage}}
	{{end}}{{end}}{{if .categorizedFlags}}
{{range $idx, $categorized := .categorizedFlags}}{{$categorized.Name}} OPTIONS:
{{range $categorized.Flags}}{{"\t"}}{{.}}
{{end}}
{{end}}{{end}}`
)

//flagGroup is a collection of flags belonging to a single topic.
type flagGroup struct {
	Name  string
	Flags []cli.Flag
}

var AppHelpFlagGroups = []flagGroup{
	{
		Name: "DSP",
		Flags: []cli.Flag{
			flags.ConfigFlag,
			flags.LogStderrFlag,
			flags.LogLevelFlag,
		},
	},
	{
		Name:  "ACCOUNT",
		Flags: []cli.Flag{},
	},
}

// byCategory sorts flagGroup by Name in in the order of AppHelpFlagGroups.
type byCategory []flagGroup

func (a byCategory) Len() int      { return len(a) }
func (a byCategory) Swap(i, j int) { a[i], a[j] = a[j], a[i] }
func (a byCategory) Less(i, j int) bool {
	iCat, jCat := a[i].Name, a[j].Name
	iIdx, jIdx := len(AppHelpFlagGroups), len(AppHelpFlagGroups) // ensure non categorized flags come last

	for i, group := range AppHelpFlagGroups {
		if iCat == group.Name {
			iIdx = i
		}
		if jCat == group.Name {
			jIdx = i
		}
	}

	return iIdx < jIdx
}

func flagCategory(flag cli.Flag) string {
	for _, category := range AppHelpFlagGroups {
		for _, flg := range category.Flags {
			if flg.String() == flag.String() {
				return category.Name
			}
		}
	}
	return "MISC"
}

type cusHelpData struct {
	App        interface{}
	FlagGroups []flagGroup
}

func PrintErrorMsg(format string, a ...interface{}) {
	format = fmt.Sprintf("\033[31m[ERROR] %s\033[0m\n", format) //Print error msg with red color
	fmt.Printf(format, a...)
}

func PrintWarnMsg(format string, a ...interface{}) {
	format = fmt.Sprintf("\033[33m[WARN] %s\033[0m\n", format) //Print error msg with yellow color
	fmt.Printf(format, a...)
}

func PrintInfoMsg(format string, a ...interface{}) {
	fmt.Printf(format+"\n", a...)
}

func PrintJsonData(data []byte) {
	var out bytes.Buffer
	err := json.Indent(&out, data, "", "   ")
	if err != nil {
		PrintErrorMsg("json.Indent error:%s", err)
		return
	}
	PrintInfoMsg(out.String())
}

func PrintJsonObject(obj interface{}) {
	data, err := json.Marshal(obj)
	if err != nil {
		PrintErrorMsg("json.Marshal error:%s", err)
		return
	}
	PrintJsonData(data)
}
func init() {
	//Using customize AppHelpTemplate
	cli.AppHelpTemplate = AppHelpTemplate
	cli.CommandHelpTemplate = CommandHelpTemplate
	cli.SubcommandHelpTemplate = SubcommandHelpTemplate

	//Override the default global app help printer
	cli.HelpPrinter = cusHelpPrinter
}

type FmtFlag struct {
	name  string
	usage string
}

func (this *FmtFlag) GetName() string {
	return this.name
}

func (this *FmtFlag) String() string {
	return this.usage
}

func (this *FmtFlag) Apply(*flag.FlagSet) {}

func formatCommand(cmds []cli.Command) []cli.Command {
	maxWidth := 0
	for _, cmd := range cmds {
		if len(cmd.Name) > maxWidth {
			maxWidth = len(cmd.Name)
		}
	}
	formatter := "%-" + fmt.Sprintf("%d", maxWidth) + "s"
	newCmds := make([]cli.Command, 0, len(cmds))
	for _, cmd := range cmds {
		name := cmd.Name
		if len(cmd.Aliases) != 0 {
			for _, aliase := range cmd.Aliases {
				name += ", " + aliase
			}
			cmd.Aliases = nil
		}
		cmd.Name = fmt.Sprintf(formatter, name)
		newCmds = append(newCmds, cmd)
	}
	return newCmds
}

func formatFlags(flags []cli.Flag) []cli.Flag {
	maxWidth := 0
	fmtFlagStrs := make(map[string][]string)
	for _, flag := range flags {
		fs := strings.Split(flag.String(), "\t")
		if len(fs[0]) > maxWidth {
			maxWidth = len(fs[0])
		}
		fmtFlagStrs[flag.GetName()] = fs
	}
	formatter := "%-" + fmt.Sprintf("%d", maxWidth) + "s   %s"
	fmtFlags := make([]cli.Flag, 0, len(fmtFlagStrs))
	for _, flag := range flags {
		flagStrs := fmtFlagStrs[flag.GetName()]

		fmtFlags = append(fmtFlags, &FmtFlag{
			name:  flag.GetName(),
			usage: fmt.Sprintf(formatter, flagStrs[0], flagStrs[1]),
		})
	}
	return fmtFlags
}

func cusHelpPrinter(w io.Writer, tmpl string, data interface{}) {
	if tmpl == AppHelpTemplate {
		cliApp := data.(*cli.App)
		cliApp.Commands = formatCommand(cliApp.Commands)
		cliApp.Flags = formatFlags(cliApp.Flags)
		categorized := make(map[string][]cli.Flag)
		for _, flag := range data.(*cli.App).Flags {
			_, ok := categorized[flag.String()]
			if !ok {
				gName := flagCategory(flag)
				categorized[gName] = append(categorized[gName], flag)
			}
		}
		sorted := make([]flagGroup, 0, len(categorized))
		for cat, flgs := range categorized {
			sorted = append(sorted, flagGroup{cat, flgs})
		}
		sort.Sort(byCategory(sorted))
		cusData := &cusHelpData{
			App:        cliApp,
			FlagGroups: sorted,
		}
		data = cusData
	} else if tmpl == SubcommandHelpTemplate {
		cliApp := data.(*cli.App)
		cliApp.Commands = formatCommand(cliApp.Commands)
		data = cliApp
	} else if tmpl == CommandHelpTemplate {
		cliCmd := data.(cli.Command)
		cliCmd.Flags = formatFlags(cliCmd.Flags)
		categorized := make(map[string][]cli.Flag)
		for _, flag := range cliCmd.Flags {
			_, ok := categorized[flag.String()]
			if !ok {
				categorized[flagCategory(flag)] = append(categorized[flagCategory(flag)], flag)
			}
		}
		sorted := make([]flagGroup, 0, len(categorized))
		for cat, flgs := range categorized {
			sorted = append(sorted, flagGroup{cat, flgs})
		}
		sort.Sort(byCategory(sorted))
		data = map[string]interface{}{
			"cmd":              cliCmd,
			"categorizedFlags": sorted,
		}
	}

	funcMap := template.FuncMap{"join": strings.Join}
	t := template.Must(template.New("help").Funcs(funcMap).Parse(tmpl))
	err := t.Execute(w, data)
	if err != nil {
		// If the writer is closed, t.Execute will fail, and there's nothing we can do to recover.
		PrintErrorMsg("CLI TEMPLATE ERROR: %#v\n", err)
		return
	}
}
