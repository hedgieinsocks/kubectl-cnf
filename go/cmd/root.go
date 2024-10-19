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
	quietFlag       bool
	printFlag       bool
	clipboardFlag   bool
	symlinkFlag     bool
}

var opts options

var rootCmd = &cobra.Command{
	Use:  "kubectl-cnf",
	Long: "kubectl cnf helps switch between current-contexts in multiple kubeconfigs",
	Annotations: map[string]string{
		cobra.CommandDisplayNameAnnotation: "kubectl cnf",
	},
	Version: "v0.0.7",
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
	rootCmd.Flags().BoolVarP(&opts.quietFlag, "quiet", "q", opts.quietFlag, "do not print auxiliary messages")
	rootCmd.Flags().BoolVarP(&opts.printFlag, "print", "p", opts.printFlag, "print 'export KUBECONFIG=PATH' to stdout instead of launching a subshell")
	rootCmd.Flags().BoolVarP(&opts.clipboardFlag, "clipboard", "c", opts.clipboardFlag, "copy 'export KUBECONFIG=PATH' to clipboard instead of launching a subshell")
	rootCmd.Flags().BoolVarP(&opts.symlinkFlag, "link", "l", opts.symlinkFlag, "symlink selected kubeconfig to '~/.kube/config' instead of launching a subshell")
	rootCmd.MarkFlagsMutuallyExclusive("print", "clipboard", "link")
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
	opts.quietFlag = getenv.Bool("KCNF_NO_VERBOSE", false)
	opts.printFlag = getenv.Bool("KCNF_NO_SHELL", false)
	opts.clipboardFlag = getenv.Bool("KCNF_COPY_CLIP", false)
	opts.symlinkFlag = getenv.Bool("KCNF_SYMLINK", false)
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

func symlinkKubeconfig(kubeconfig string) error {
	home, err := os.UserHomeDir()
	if err != nil {
		return err
	}
	symlink := filepath.Join(home, ".kube", "config")
	if _, err := os.Lstat(symlink); err == nil {
		if err = os.Remove(symlink); err != nil {
			return err
		}
	}
	if err := os.Symlink(kubeconfig, symlink); err != nil {
		return err
	}
	return nil
}

func launchShell(shell, kubecontext, kubeconfig string) error {
	os.Setenv("KUBECONTEXT", kubecontext)
	os.Setenv("KUBECONFIG", kubeconfig)
	cmd := exec.Command(shell)
	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	if !opts.quietFlag {
		fmt.Printf("⇲ %s\n", kubecontext)
	}
	if err := cmd.Run(); err != nil {
		if !strings.HasPrefix(err.Error(), "exit status") {
			return err
		}
	}
	if !opts.quietFlag {
		fmt.Printf("⇱ %s\n", kubecontext)
	}
	return nil
}

func processSelection(selection string) error {
	selectionSplit := strings.Split(selection, "\t")
	kubecontext, kubeconfig := selectionSplit[0], selectionSplit[1]
	if opts.symlinkFlag || opts.printFlag || opts.clipboardFlag {
		if !opts.quietFlag {
			fmt.Printf("⮺ %s\n", kubecontext)
		}
	}
	if opts.printFlag {
		fmt.Println(fmt.Sprintf("export KUBECONFIG='%s'", kubeconfig))
		return nil
	}
	if opts.clipboardFlag {
		cb := clipboard.New()
		if err := cb.CopyText(fmt.Sprintf("export KUBECONFIG='%s'", kubeconfig)); err != nil {
			return err
		}
		return nil
	}
	if opts.symlinkFlag {
		if err := symlinkKubeconfig(kubeconfig); err != nil {
			return err
		}
		return nil
	}
	shell := getenv.String("SHELL", "bash")
	if err := launchShell(shell, kubecontext, kubeconfig); err != nil {
		return err
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
