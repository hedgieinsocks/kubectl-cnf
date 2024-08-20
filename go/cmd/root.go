package cmd

import (
	"fmt"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"sort"
	"strings"
	"sync"

	"github.com/josestg/getenv"
	fzf "github.com/junegunn/fzf/src"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

type Options struct {
	KubeConfigDir   string
	SelectionHeight string
	NoVerboseFlag   bool
	NoShellFlag     bool
	copyClipFlag    bool
}

var (
	opts        Options
	previewCmd  string
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
	Run:     main,
}

func Execute() {
	if err := rootCmd.Execute(); err != nil {
		os.Exit(1)
	}
}

func init() {
	initConfig()
	setPreviewCmd()
	rootCmd.Flags().StringVarP(&opts.KubeConfigDir, "dir", "d", opts.KubeConfigDir, "directory with kubeconfigs")
	rootCmd.Flags().StringVarP(&opts.SelectionHeight, "height", "H", opts.SelectionHeight, "selection menu height")
	rootCmd.Flags().BoolVarP(&opts.NoVerboseFlag, "no-verbose", "V", opts.NoVerboseFlag, "do not print auxiliary messages")
	rootCmd.Flags().BoolVarP(&opts.NoShellFlag, "no-shell", "S", opts.NoShellFlag, "do not launch a subshell, instead print 'export KUBECONFIG=PATH' to stdout")
	rootCmd.Flags().BoolVarP(&opts.copyClipFlag, "clip", "c", opts.copyClipFlag, "when --no-shell is provided, copy 'export KUBECONFIG=PATH' to clipboard instead of printing to stdout")
	rootCmd.Flags().SortFlags = false
}

func initConfig() {
	opts.KubeConfigDir = getenv.String("KCNF_DIR", expandHomeDir("~/.kube/configs"))
	opts.SelectionHeight = getenv.String("KCNF_DIR_HEIGHT", "40%")
	opts.NoVerboseFlag = getenv.Bool("KCNF_NO_VERBOSE", false)
	opts.NoShellFlag = getenv.Bool("KCNF_NO_SHELL", false)
	opts.copyClipFlag = getenv.Bool("KCNF_COPY_CLIP", false)
}

func setPreviewCmd() {
	if _, err := exec.LookPath("bat"); err != nil {
		previewCmd = "cat"
	} else {
		previewCmd = "bat --style=plain --color=always --language=yaml"
	}
}

func expandHomeDir(path string) string {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		log.Fatalf("Error: failed to expand homedir: %v", err)
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
	err := filepath.WalkDir(directory, func(path string, info os.DirEntry, err error) error {
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

func launchSubShell(kubeconfig, kubecontext string) {
	os.Setenv("KUBECONTEXT", kubecontext)
	os.Setenv("KUBECONFIG", kubeconfig)
	subShell := exec.Command(os.Getenv("SHELL"))
	subShell.Stdin = os.Stdin
	subShell.Stdout = os.Stdout
	subShell.Stderr = os.Stderr
	if !opts.NoVerboseFlag {
		fmt.Println("⇲ " + kubecontext)
	}
	subShell.Run()
	if !opts.NoVerboseFlag {
		fmt.Println("⇱ " + kubecontext)
	}
}

func copyToClipboard(kubeconfig string) {
	var clipBin string
	var clipArg []string
	platform := runtime.GOOS
	switch platform {
	case "linux":
		session := getenv.String("XDG_SESSION_TYPE", "x11")
		switch session {
		case "x11":
			clipBin = "xsel"
			clipArg = []string{"--input", "--clipboard", "--trim"}
		case "wayland":
			clipBin = "wl-copy"
			clipArg = []string{"--trim-newline"}
		default:
			log.Fatalf("Error: clipboard copy is not supported on this session: %s", session)
		}
	case "darwin":
		clipBin = "pbcopy"
		clipArg = []string{}
	default:
		log.Fatalf("Error: clipboard copy is not supported on this platform: %s", platform)
	}
	if _, err := exec.LookPath(clipBin); err != nil {
		log.Fatalf("Error: failed to locate clipboard binary: %v", err)
	}
	clipCopy := exec.Command(clipBin, clipArg...)
	clipCopy.Stdin = strings.NewReader("export KUBECONFIG='" + kubeconfig + "'")
	if err := clipCopy.Run(); err != nil {
		log.Fatalf("Error: failed to copy data to clipboard: %v", err)
	}
}

func processSelection(selection string) {
	selectedKubeConfig := strings.Split(selection, "\t")
	kubecontext, kubeconfig := selectedKubeConfig[0], selectedKubeConfig[1]
	if !opts.NoShellFlag {
		launchSubShell(kubeconfig, kubecontext)
	} else {
		if !opts.NoVerboseFlag {
			fmt.Println("⮺ " + kubecontext)
		}
		if !opts.copyClipFlag {
			fmt.Println("export KUBECONFIG='" + kubeconfig + "'")
		} else {
			copyToClipboard(kubeconfig)
		}
	}
}

func main(cmd *cobra.Command, args []string) {
	query := strings.Join(args, " ")

	if err := getKubeConfigs(opts.KubeConfigDir); err != nil {
		log.Fatalf("Error: failed to get kubeconfigs: %v", err)
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
			processSelection(s)
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
			"--preview={echo '# {2}'; kubectl config view --kubeconfig {2};} | " + previewCmd,
		},
	)
	if err != nil {
		log.Fatalf("Error: failed to parse fzf options: %v", err)
	}

	options.Input = inputChan
	options.Output = outputChan

	if _, err := fzf.Run(options); err != nil {
		log.Fatalf("Error: failed to run fzf selection: %v", err)
	}

	close(outputChan)
	wg.Wait()
}
