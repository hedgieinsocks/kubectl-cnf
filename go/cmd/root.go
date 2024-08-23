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

	"github.com/josestg/getenv"
	fzf "github.com/junegunn/fzf/src"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"github.com/tiagomelo/go-clipboard/clipboard"
)

type Options struct {
	KubeConfigDir   string
	SelectionHeight string
	NoVerboseFlag   bool
	NoShellFlag     bool
	copyClipFlag    bool
}

var opts Options

var rootCmd = &cobra.Command{
	Use:  "kubectl-cnf",
	Long: "kubectl cnf helps switch between current-contexts in multiple kubeconfigs",
	Annotations: map[string]string{
		cobra.CommandDisplayNameAnnotation: "kubectl cnf",
	},
	Version: "v0.0.6",
	Run:     main,
}

func Execute() {
	if err := rootCmd.Execute(); err != nil {
		os.Exit(1)
	}
}

func init() {
	log.SetFlags(0)
	log.SetPrefix("error: ")
	setOpts()
	rootCmd.Flags().StringVarP(&opts.KubeConfigDir, "dir", "d", opts.KubeConfigDir, "directory with kubeconfigs")
	rootCmd.Flags().StringVarP(&opts.SelectionHeight, "height", "H", opts.SelectionHeight, "selection menu height")
	rootCmd.Flags().BoolVarP(&opts.NoVerboseFlag, "no-verbose", "V", opts.NoVerboseFlag, "do not print auxiliary messages")
	rootCmd.Flags().BoolVarP(&opts.NoShellFlag, "no-shell", "S", opts.NoShellFlag, "do not launch a subshell, instead print 'export KUBECONFIG=PATH' to stdout")
	rootCmd.Flags().BoolVarP(&opts.copyClipFlag, "clip", "c", opts.copyClipFlag, "when --no-shell is provided, copy 'export KUBECONFIG=PATH' to clipboard instead of printing to stdout")
	rootCmd.Flags().SortFlags = false
}

func setOpts() {
	opts.KubeConfigDir = getenv.String("KCNF_DIR", expandHomeDir("~/.kube/configs"))
	opts.SelectionHeight = getenv.String("KCNF_DIR_HEIGHT", "40%")
	opts.NoVerboseFlag = getenv.Bool("KCNF_NO_VERBOSE", false)
	opts.NoShellFlag = getenv.Bool("KCNF_NO_SHELL", false)
	opts.copyClipFlag = getenv.Bool("KCNF_COPY_CLIP", false)
}

func expandHomeDir(path string) string {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		log.Fatalf("failed to expand homedir: %v", err)
	}
	return filepath.Join(homeDir, path[2:])
}

func copyToClipboard(text string) {
	clipBoard := clipboard.New()
	if err := clipBoard.CopyText(text); err != nil {
		log.Fatalf("failed to copy data to clipboard: %v", err)
	}
}

func getPreviewCmd() string {
	if _, err := exec.LookPath("bat"); err != nil {
		return "cat"
	}
	return "bat --style=plain --color=always --language=yaml"
}

func getCurrentContext(file string) string {
	viper.SetConfigFile(file)
	if err := viper.ReadInConfig(); err != nil {
		return ""
	}
	return viper.GetString("current-context")
}

func getKubeConfigs(directory string) ([]string, error) {
	var kubeConfigs []string
	viper.SetConfigType("yaml")
	err := filepath.WalkDir(directory, func(path string, info os.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if !info.IsDir() {
			if currentContext := getCurrentContext(path); currentContext != "" {
				kubeConfigs = append(kubeConfigs, currentContext+"\t"+path)
			}
		}
		return nil
	})
	return kubeConfigs, err
}

func launchSubShell(kubeConfig, kubeContext string) {
	os.Setenv("KUBECONTEXT", kubeContext)
	os.Setenv("KUBECONFIG", kubeConfig)
	subShell := exec.Command(os.Getenv("SHELL"))
	subShell.Stdin = os.Stdin
	subShell.Stdout = os.Stdout
	subShell.Stderr = os.Stderr
	if !opts.NoVerboseFlag {
		fmt.Println("⇲ " + kubeContext)
	}
	subShell.Run()
	if !opts.NoVerboseFlag {
		fmt.Println("⇱ " + kubeContext)
	}
}

func processSelection(selection string) {
	selectedKubeConfig := strings.Split(selection, "\t")
	kubeContext, kubeConfig := selectedKubeConfig[0], selectedKubeConfig[1]
	if opts.NoShellFlag {
		if !opts.NoVerboseFlag {
			fmt.Println("⮺ " + kubeContext)
		}
		exportCmd := "export KUBECONFIG='" + kubeConfig + "'"
		if opts.copyClipFlag {
			copyToClipboard(exportCmd)
		} else {
			fmt.Println(exportCmd)
		}
	} else {
		launchSubShell(kubeConfig, kubeContext)
	}
}

func main(cmd *cobra.Command, args []string) {
	var wg sync.WaitGroup

	kubeConfigs, err := getKubeConfigs(opts.KubeConfigDir)
	if err != nil {
		log.Fatalf("failed to parse kubeconfigs: %v", err)
	}
	if kubeConfigs == nil {
		log.Fatalf("no valid kubeconfigs found in: %s", opts.KubeConfigDir)
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
	wg.Add(1)
	go func() {
		defer wg.Done()
		for s := range outputChan {
			processSelection(s)
		}
	}()

	query := strings.Join(args, " ")
	previewCmd := getPreviewCmd()
	fzfOptions, err := fzf.ParseOptions(
		true,
		[]string{
			"--layout=reverse",
			"--height=" + opts.SelectionHeight,
			"--delimiter=\t",
			"--with-nth=1",
			"--query=" + query,
			"--bind=tab:toggle-preview",
			"--preview-window=hidden,wrap,75%",
			"--preview={ echo '# {2}'; kubectl config view --kubeconfig {2}; } | " + previewCmd,
		},
	)
	if err != nil {
		log.Fatalf("failed to parse fzf options: %v", err)
	}

	fzfOptions.Input = inputChan
	fzfOptions.Output = outputChan

	if _, err := fzf.Run(fzfOptions); err != nil {
		log.Fatalf("failed to run fzf selection: %v", err)
	}

	close(outputChan)
	wg.Wait()
}
