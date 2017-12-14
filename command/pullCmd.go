package command

import (
	"bufio"
	"bytes"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/everfore/exc"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"github.com/toukii/goutils"
)

var (
	Command = &cobra.Command{
		Use:   "pull",
		Short: "git pull repo",
		Long:  ``,
		Run: func(cmd *cobra.Command, args []string) {
			Excute(args)
		},
	}
)

func init() {

	Command.PersistentFlags().BoolP("ssh", "s", false, "git clone use ssh")
	Command.PersistentFlags().BoolP("install", "i", false, "install gopkg after pull")
	Command.PersistentFlags().BoolP("reclone", "c", false, "if repo exists, re clone")
	viper.BindPFlag("ssh", Command.PersistentFlags().Lookup("ssh"))
	viper.BindPFlag("install", Command.PersistentFlags().Lookup("install"))
	viper.BindPFlag("reclone", Command.PersistentFlags().Lookup("reclone"))
}

func Excute(args []string) error {
	newArgs := make([]string, 0, len(args))
	for _, arg := range args {
		if strings.HasPrefix(arg, "-") {
			continue
		}
		newArgs = append(newArgs, arg)
	}
	pull(strings.Join(newArgs, ";"), viper.GetBool("ssh"), viper.GetBool("reclone"))
	return nil
}

func pull(input string, writable bool, reinstall bool) {
	fmt.Println(input, writable, reinstall)
	var user, repo, branch, input_1 /*,target*/ string
	if len(input) <= 0 {
		tips := "[user/]repo[:branch]  > $"
		fmt.Print(tips)
		fmt.Scanf("%s", &input)
	}
	start := time.Now()
	if strings.Contains(input, "/") {
		if strings.HasPrefix(input, "github.com/") {
			input = input[11:]
		}
		inputs := strings.Split(input, "/")
		user = inputs[0]
		input_1 = inputs[1]
	} else if len(input) <= 0 {
		branch := currentBranch()
		exc.NewCMD(fmt.Sprintf("git pull origin %s:%s", branch, branch)).Debug().Execute()
		return
	} else {
		pwd, _ := os.Getwd()
		user = filepath.Base(pwd)
		input_1 = input
	}

	if strings.Contains(input_1, ":") {
		input_1s := strings.Split(input_1, ":")
		repo = input_1s[0]
		branch = input_1s[1]
	} else {
		repo = input_1
		branch = "master"
	}
	fmt.Printf("%s/%s:%s\n", user, repo, branch)

	codeload_uri := ""
	if writable {
		codeload_uri = fmt.Sprintf("git clone --progress --depth 1 git@github.com:%s/%s.git", user, repo)
	} else {
		codeload_uri = fmt.Sprintf("git clone --progress --depth 1 git://github.com/%s/%s", user, repo)
	}
	GOPATH := os.Getenv("GOPATH")
	target := filepath.Join(GOPATH, "src", "github.com", user, repo)
	if pathExists(target) {
		if reinstall {
			exc.NewCMD("rm -rf " + repo).Env("GOPATH").Cd("src/github.com/").Cd(user).Debug().Execute()
		} else {
			fmt.Println("repo already exists, try to use falg: -c")
			return
		}
	}
	os.MkdirAll(target, 0777)
	exc.NewCMD(codeload_uri).Env("GOPATH").Cd("src/github.com/").Cd(user).Wd().Debug().Execute()

	fmt.Printf("cost time:%v\n", time.Now().Sub(start))
}

func pathExists(path string) bool {
	_, err := os.Stat(path)
	if err == nil || os.IsExist(err) {
		return true
	}
	return false
}

func cloneLoop(bs []byte) {
	br := bytes.NewReader(bs)
	bufr := bufio.NewReader(br)
	for {
		bs, err := bufr.ReadSlice('\n')
		if err != nil {
			break
		}
		bs_str := goutils.ToString(bs)
		if strings.Contains(bs_str, "cannot find package") {
			splts := strings.Split(bs_str, "\"")
			if len(splts) > 1 {
				repo := splts[1]
				// go func(repo string) {
				user_repos := strings.Split(repo, "/")
				if len(user_repos) > 2 {
					if !strings.EqualFold(user_repos[0], "github.com") {
						exc.NewCMD(fmt.Sprintf("go get -u %s", repo))
						continue
					}
					user_repo := fmt.Sprintf("%s/%s", user_repos[1], user_repos[2])
					fmt.Println(user_repo)
					if strings.EqualFold(user_repos[1], "toukii") || strings.EqualFold(user_repos[1], "everfore") || strings.EqualFold(user_repos[1], "datc") {
						pull(user_repo, true, false)
					} else {
						pull(user_repo, false, false)
					}
				}
				// }(splts[1])
				// time.Sleep(1e9)
			}
		}
	}

}

func currentBranch() string {
	bs, err := exc.NewCMD("git rev-parse --abbrev-ref HEAD").DoNoTime()
	if err != nil {
		panic(err)
	}
	cb := string(bs[:len(bs)-1])
	fmt.Printf("* %s\n", cb)
	return cb
}
