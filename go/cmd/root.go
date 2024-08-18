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

var (
	optKubeConfigDir   string
	optSelectionHeight string
	optNoShell         bool
	optNoVerbose       bool
	wg                 sync.WaitGroup
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
	err := rootCmd.Execute()
	if err != nil {
		os.Exit(1)
	}
}

func init() {
	optKubeConfigDir = os.Getenv("KCNF_DIR")
	if optKubeConfigDir == "" {
		optKubeConfigDir = expandHomeDir("~/.kube/configs")
	}
	optSelectionHeight = os.Getenv("KCNF_HEIGHT")
	if optSelectionHeight == "" {
		optSelectionHeight = "40%"
	}
	if val := os.Getenv("KCNF_NO_SHELL"); val == "" {
		optNoShell = false
	} else {
		optNoShell = true
	}
	if val := os.Getenv("KCNF_NO_VERBOSE"); val == "" {
		optNoVerbose = false
	} else {
		optNoVerbose = true
	}
	rootCmd.PersistentFlags().StringVarP(&optKubeConfigDir, "directory", "d", optKubeConfigDir, "directory with kubeconfigs")
	rootCmd.PersistentFlags().StringVarP(&optSelectionHeight, "height", "H", optSelectionHeight, "selection menu height")
	rootCmd.PersistentFlags().BoolVarP(&optNoShell, "no-shell", "S", optNoShell, "do not launch subshell")
	rootCmd.PersistentFlags().BoolVarP(&optNoVerbose, "no-verbose", "V", optNoVerbose, "supress subshell notifications")
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
	err := viper.ReadInConfig()
	if err != nil {
		return ""
	}
	return viper.GetString("current-context")
}

func getKubeConfigs(kubeConfigDir string) ([]string, error) {
	viper.SetConfigType("yaml")
	var kubeConfigs []string
	err := filepath.Walk(kubeConfigDir, func(path string, info os.FileInfo, err error) error {
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
	return kubeConfigs, err
}

func launchSubShell(kubeconfig string, kubecontext string) {
	os.Setenv("KUBECONTEXT", kubecontext)
	os.Setenv("KUBECONFIG", kubeconfig)
	subShell := exec.Command(os.Getenv("SHELL"))
	subShell.Stdin = os.Stdin
	subShell.Stdout = os.Stdout
	subShell.Stderr = os.Stderr
	if !optNoVerbose {
		fmt.Println("⇱ entered subshell with context: " + kubecontext)
	}
	subShell.Run()
	if !optNoVerbose {
		fmt.Println("⇱ exited subshell with context: " + kubecontext)
	}
}

func selectKubeConfig(cmd *cobra.Command, args []string) {
	query := strings.Join(args, " ")

	kubeConfigs, err := getKubeConfigs(optKubeConfigDir)
	if err != nil {
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
			if !optNoShell {
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
			"--height=" + optSelectionHeight,
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
