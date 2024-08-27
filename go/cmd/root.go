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

type options struct {
	kubeconfigsDir  string
	selectionHeight string
	noVerboseFlag   bool
	noShellFlag     bool
	copyClipFlag    bool
}

var opts options

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
	log.SetPrefix("Error: ")
	setOpts()
	rootCmd.Flags().StringVarP(&opts.kubeconfigsDir, "directory", "d", opts.kubeconfigsDir, "directory with kubeconfigs")
	rootCmd.Flags().StringVarP(&opts.selectionHeight, "height", "H", opts.selectionHeight, "selection menu height")
	rootCmd.Flags().BoolVarP(&opts.noVerboseFlag, "no-verbose", "V", opts.noVerboseFlag, "do not print auxiliary messages")
	rootCmd.Flags().BoolVarP(&opts.noShellFlag, "no-shell", "S", opts.noShellFlag, "do not launch a subshell, instead print 'export KUBECONFIG=PATH' to stdout")
	rootCmd.Flags().BoolVarP(&opts.copyClipFlag, "clipboard", "c", opts.copyClipFlag, "when --no-shell is provided, copy 'export KUBECONFIG=PATH' to clipboard instead of printing to stdout")
	rootCmd.Flags().SortFlags = false
}

func setOpts() {
	opts.kubeconfigsDir = getenv.String("KCNF_DIR", getKubeconfigsDir())
	opts.selectionHeight = getenv.String("KCNF_DIR_HEIGHT", "40%")
	opts.noVerboseFlag = getenv.Bool("KCNF_NO_VERBOSE", false)
	opts.noShellFlag = getenv.Bool("KCNF_NO_SHELL", false)
	opts.copyClipFlag = getenv.Bool("KCNF_COPY_CLIP", false)
}

func getKubeconfigsDir() string {
	home, err := os.UserHomeDir()
	if err != nil {
		log.Fatalf("failed to get homedir: %v", err)
	}
	return filepath.Join(home, ".kube", "configs")
}

func copyToClipboard(text string) {
	cb := clipboard.New()
	if err := cb.CopyText(text); err != nil {
		log.Fatalf("failed to copy text to clipboard: %v", err)
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

func getKubeconfigs(directory string) ([]string, error) {
	var kubeconfigs []string
	viper.SetConfigType("yaml")
	err := filepath.WalkDir(directory, func(path string, info os.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if !info.IsDir() {
			if currentContext := getCurrentContext(path); currentContext != "" {
				kubeconfigs = append(kubeconfigs, currentContext+"\t"+path)
			}
		}
		return nil
	})
	return kubeconfigs, err
}

func launchSubShell(kubeconfig, kubecontext string) {
	os.Setenv("KUBECONTEXT", kubecontext)
	os.Setenv("KUBECONFIG", kubeconfig)
	shell := exec.Command(os.Getenv("SHELL"))
	shell.Stdin = os.Stdin
	shell.Stdout = os.Stdout
	shell.Stderr = os.Stderr
	if !opts.noVerboseFlag {
		fmt.Printf("⇲ %s\n", kubecontext)
	}
	shell.Run()
	if !opts.noVerboseFlag {
		fmt.Printf("⇱ %s\n", kubecontext)
	}
}

func processSelection(selection string) {
	selectionSplit := strings.Split(selection, "\t")
	kubecontext, kubeconfig := selectionSplit[0], selectionSplit[1]
	if opts.noShellFlag {
		if !opts.noVerboseFlag {
			fmt.Printf("⮺ %s\n", kubecontext)
		}
		exportCmd := fmt.Sprintf("export KUBECONFIG='%s'", kubeconfig)
		if opts.copyClipFlag {
			copyToClipboard(exportCmd)
		} else {
			fmt.Println(exportCmd)
		}
	} else {
		launchSubShell(kubeconfig, kubecontext)
	}
}

func main(cmd *cobra.Command, args []string) {
	var wg sync.WaitGroup

	kubeconfigs, err := getKubeconfigs(opts.kubeconfigsDir)
	if err != nil {
		log.Fatalf("failed to parse kubeconfigs: %v", err)
	}
	if kubeconfigs == nil {
		log.Fatalf("no valid kubeconfigs found in: %s", opts.kubeconfigsDir)
	}
	sort.Strings(kubeconfigs)

	inputChan := make(chan string)
	go func() {
		for _, s := range kubeconfigs {
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
			"--delimiter=\t",
			"--with-nth=1",
			"--bind=tab:toggle-preview",
			"--preview-window=hidden,wrap,75%",
			fmt.Sprintf("--height=%s", opts.selectionHeight),
			fmt.Sprintf("--query=%s", query),
			fmt.Sprintf("--preview={ echo '# {2}'; kubectl config view --kubeconfig {2}; } | %s", previewCmd),
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
