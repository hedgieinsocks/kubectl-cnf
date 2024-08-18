package cmd

import (
	"fmt"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"sort"
	"strings"
	"sync"

	fzf "github.com/junegunn/fzf/src"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

type Options struct {
	KubeConfigDir   string
	SelectionHeight string
	NoShellFlag     bool
	NoVerboseFlag   bool
}

var (
	opts        Options
	kubeConfigs []string
	wg          sync.WaitGroup
)

var rootCmd = &cobra.Command{
	Use:  "kubectl-cnf",
	Long: "kubectl cnf helps switch between current-contexts in multiple kubeconfigs",
	Annotations: map[string]string{
		cobra.CommandDisplayNameAnnotation: "kubectl cnf",
	},
	Version: "v0.1.0",
	Run:     selectKubeConfig,
}

func Execute() {
	if err := rootCmd.Execute(); err != nil {
		os.Exit(1)
	}
}

func init() {
	initConfig()
	rootCmd.PersistentFlags().StringVarP(&opts.KubeConfigDir, "dir", "d", opts.KubeConfigDir, "directory with kubeconfigs")
	rootCmd.PersistentFlags().StringVarP(&opts.SelectionHeight, "height", "H", opts.SelectionHeight, "selection menu height")
	rootCmd.PersistentFlags().BoolVarP(&opts.NoShellFlag, "no-shell", "S", opts.NoShellFlag, "do not launch subshell")
	rootCmd.PersistentFlags().BoolVarP(&opts.NoVerboseFlag, "no-verbose", "V", opts.NoVerboseFlag, "supress subshell notifications")
}

func initConfig() {
	viper.SetEnvPrefix("kcnf")
	viper.AutomaticEnv()
	replacer := strings.NewReplacer("-", "_")
	viper.SetEnvKeyReplacer(replacer)
	viper.SetDefault("dir", expandHomeDir("~/.kube/configs"))
	viper.SetDefault("height", "40%")
	opts.KubeConfigDir = viper.GetString("dir")
	opts.SelectionHeight = viper.GetString("height")
	opts.NoShellFlag = viper.GetBool("no-shell")
	opts.NoVerboseFlag = viper.GetBool("no-verbose")
}

func expandHomeDir(path string) string {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		log.Fatalf("Failed to expand homedir: %v", err)
	}
	return filepath.Join(homeDir, path[2:])
}

func getCurrentContext(filename string) string {
	viper.SetConfigFile(filename)
	if err := viper.ReadInConfig(); err != nil {
		return ""
	}
	return viper.GetString("current-context")
}

func getKubeConfigs(directory string) error {
	viper.SetConfigType("yaml")
	err := filepath.Walk(directory, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if !info.IsDir() {
			currentContext := getCurrentContext(path)
			if currentContext != "" {
				kubeConfigs = append(kubeConfigs, currentContext+"\t"+path)
			}
		}
		return nil
	})
	return err
}

func launchSubShell(kubeconfig string, kubecontext string) {
	os.Setenv("KUBECONTEXT", kubecontext)
	os.Setenv("KUBECONFIG", kubeconfig)
	subShell := exec.Command(os.Getenv("SHELL"))
	subShell.Stdin = os.Stdin
	subShell.Stdout = os.Stdout
	subShell.Stderr = os.Stderr
	if !opts.NoVerboseFlag {
		fmt.Println("⇱ entered subshell with context: " + kubecontext)
	}
	subShell.Run()
	if !opts.NoVerboseFlag {
		fmt.Println("⇱ exited subshell with context: " + kubecontext)
	}
}

func selectKubeConfig(cmd *cobra.Command, args []string) {
	query := strings.Join(args, " ")

	if err := getKubeConfigs(opts.KubeConfigDir); err != nil {
		log.Fatalf("Failed to get kubeconfigs: %v", err)
	}
	sort.Strings(kubeConfigs)

	inputChan := make(chan string)
	go func() {
		for _, s := range kubeConfigs {
			inputChan <- s
		}
		close(inputChan)
	}()

	outputChan := make(chan string, 1)
	go func() {
		wg.Add(1)
		defer wg.Done()
		for s := range outputChan {
			selectedKubeConfig := strings.Split(s, "\t")
			kubecontext, kubeconfig := selectedKubeConfig[0], selectedKubeConfig[1]
			if !opts.NoShellFlag {
				launchSubShell(kubeconfig, kubecontext)
			} else {
				fmt.Println("export KUBECONFIG='" + kubeconfig + "'")
			}
		}
	}()

	options, err := fzf.ParseOptions(
		true,
		[]string{
			"--layout=reverse",
			"--height=" + opts.SelectionHeight,
			"--delimiter=\t",
			"--with-nth=1",
			"--query=" + query,
			"--bind=tab:toggle-preview",
			"--preview-window=hidden,wrap,75%",
			"--preview=echo '# {2}' && kubectl config view --kubeconfig {2}",
		},
	)
	if err != nil {
		log.Fatalf("Failed to parse fzf options: %v", err)
	}

	options.Input = inputChan
	options.Output = outputChan

	fzf.Run(options)

	close(outputChan)
	wg.Wait()
}
