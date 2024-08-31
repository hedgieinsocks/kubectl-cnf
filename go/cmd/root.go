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
	viper.SetConfigType("yaml")
	if err := setOptsFromEnv(); err != nil {
		log.Fatalf("failed to get homedir: %v", err)
	}
	rootCmd.Flags().StringVarP(&opts.kubeconfigsDir, "directory", "d", opts.kubeconfigsDir, "directory with kubeconfigs")
	rootCmd.Flags().StringVarP(&opts.selectionHeight, "height", "H", opts.selectionHeight, "selection menu height")
	rootCmd.Flags().BoolVarP(&opts.noVerboseFlag, "no-verbose", "V", opts.noVerboseFlag, "do not print auxiliary messages")
	rootCmd.Flags().BoolVarP(&opts.noShellFlag, "no-shell", "S", opts.noShellFlag, "do not launch a subshell, instead print 'export KUBECONFIG=PATH' to stdout")
	rootCmd.Flags().BoolVarP(&opts.copyClipFlag, "clipboard", "c", opts.copyClipFlag, "when --no-shell is provided, copy 'export KUBECONFIG=PATH' to clipboard instead of printing to stdout")
	rootCmd.Flags().SortFlags = false
}

func getKubeconfigsDir() (string, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(home, ".kube", "configs"), nil
}

func setOptsFromEnv() error {
	dir, err := getKubeconfigsDir()
	if err != nil {
		return err
	}
	opts.kubeconfigsDir = getenv.String("KCNF_DIR", dir)
	opts.selectionHeight = getenv.String("KCNF_DIR_HEIGHT", "40%")
	opts.noVerboseFlag = getenv.Bool("KCNF_NO_VERBOSE", false)
	opts.noShellFlag = getenv.Bool("KCNF_NO_SHELL", false)
	opts.copyClipFlag = getenv.Bool("KCNF_COPY_CLIP", false)
	return nil
}

func getCurrentContext(file string) string {
	viper.SetConfigFile(file)
	if err := viper.ReadInConfig(); err != nil {
		return ""
	}
	return viper.GetString("current-context")
}

func getKubeconfigs(directory string) ([][]string, error) {
	var kubeconfigs [][]string
	err := filepath.WalkDir(directory, func(path string, info os.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if !info.IsDir() {
			if currentContext := getCurrentContext(path); currentContext != "" {
				kubeconfigs = append(kubeconfigs, []string{currentContext, path})
			}
		}
		return nil
	})
	return kubeconfigs, err
}

func getPreviewCmd() string {
	baseCmd := "{ echo '# {2}'; kubectl config view --kubeconfig {2}; }"
	if _, err := exec.LookPath("bat"); err != nil {
		return fmt.Sprintf("%s | cat", baseCmd)
	}
	return fmt.Sprintf("%s | bat --style=plain --color=always --language=yaml", baseCmd)
}

func configureFzf(height, query, preview string) (*fzf.Options, error) {
	return fzf.ParseOptions(
		true,
		[]string{
			"--layout=reverse",
			"--delimiter=\t",
			"--with-nth=1",
			"--bind=tab:toggle-preview",
			"--preview-window=hidden,wrap,75%",
			fmt.Sprintf("--height=%s", height),
			fmt.Sprintf("--query=%s", query),
			fmt.Sprintf("--preview=%s", preview),
		},
	)
}

func launchShell(shell, kubecontext, kubeconfig string) error {
	os.Setenv("KUBECONTEXT", kubecontext)
	os.Setenv("KUBECONFIG", kubeconfig)
	cmd := exec.Command(shell)
	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	if !opts.noVerboseFlag {
		fmt.Printf("⇲ %s\n", kubecontext)
	}
	if err := cmd.Run(); err != nil {
		if !strings.HasPrefix(err.Error(), "exit status") {
			return err
		}
	}
	if !opts.noVerboseFlag {
		fmt.Printf("⇱ %s\n", kubecontext)
	}
	return nil
}

func processSelection(selection string) error {
	selectionSplit := strings.Split(selection, "\t")
	kubecontext, kubeconfig := selectionSplit[0], selectionSplit[1]
	if opts.noShellFlag {
		if !opts.noVerboseFlag {
			fmt.Printf("⮺ %s\n", kubecontext)
		}
		exportCmd := fmt.Sprintf("export KUBECONFIG='%s'", kubeconfig)
		if opts.copyClipFlag {
			cb := clipboard.New()
			if err := cb.CopyText(exportCmd); err != nil {
				return err
			}
		} else {
			fmt.Println(exportCmd)
		}
	} else {
		shell := getenv.String("SHELL", "bash")
		if err := launchShell(shell, kubecontext, kubeconfig); err != nil {
			return err
		}
	}
	return nil
}

func main(cmd *cobra.Command, args []string) {
	var wg sync.WaitGroup

	kubeconfigs, err := getKubeconfigs(opts.kubeconfigsDir)
	if err != nil {
		log.Fatalf("failed to retrieve kubeconfigs: %v", err)
	}
	if kubeconfigs == nil {
		log.Fatalf("no valid kubeconfigs found in: %s", opts.kubeconfigsDir)
	}
	sort.Slice(kubeconfigs, func(i, j int) bool {
		return kubeconfigs[i][0] < kubeconfigs[j][0]
	})

	inputChan := make(chan string)
	go func() {
		for _, s := range kubeconfigs {
			inputChan <- fmt.Sprintf("%s\t%s", s[0], s[1])
		}
		close(inputChan)
	}()

	outputChan := make(chan string, 1)
	wg.Add(1)
	go func() {
		defer wg.Done()
		for s := range outputChan {
			if err := processSelection(s); err != nil {
				log.Fatalf("failed to process selection: %v", err)
			}
		}
	}()

	query := strings.Join(args, " ")
	preview := getPreviewCmd()
	fzfOptions, err := configureFzf(opts.selectionHeight, query, preview)
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
