package project

import (
	"bytes"
	"context"
	"fmt"
	"github.com/go-kratos/kratos/cmd/kratos/v2/internal/base"
	"io/fs"
	"io/ioutil"
	"os"
	"path"
	"path/filepath"
	"strings"

	"github.com/AlecAivazis/survey/v2"
	"github.com/fatih/color"
)

// Project is a project template.
type Project struct {
	Name string
	Path string
}

// New new a project from remote repo.
func (p *Project) New(ctx context.Context, dir string, layout string, branch string) error {
	to := path.Join(dir, p.Name)
	internalto := path.Join(dir, "internal", "app", p.Name)
	if _, err := os.Stat(internalto); !os.IsNotExist(err) {
		fmt.Printf("🚫 %s already exists\n", p.Name)
		override := false
		prompt := &survey.Confirm{
			Message: "📂 Do you want to override the folder ?",
			Help:    "Delete the existing folder and create the project.",
		}
		e := survey.AskOne(prompt, &override)
		if e != nil {
			return e
		}
		if !override {
			return err
		}
		os.RemoveAll(internalto)
	}
	fmt.Printf("🚀 Creating service %s, layout repo is %s, please wait a moment.\n\n", p.Name, layout)
	repo := base.NewRepo(layout, branch)
	if err := repo.CopyTo(ctx, internalto, p.Path, []string{".git", ".github", ".idea", "go.mod"}); err != nil {
		return err
	}

	err := os.MkdirAll(path.Join(dir, "api", p.Name, "v1"), 0755)
	if err != nil {
		return err
	}

	name := []byte(p.Name)
	name[0] -= 'a' - 'A'
	filepath.Walk(internalto, func(fp string, info fs.FileInfo, err error) error {
		if !info.IsDir() {
			fd, _ := ioutil.ReadFile(fp)
			ioutil.WriteFile(fp,
				bytes.ReplaceAll(
					bytes.ReplaceAll(fd, []byte("%NAME%"), name),
					[]byte("%name%"), []byte(p.Name)), 0777)
		}
		if strings.Contains(fp, "%NAME%") {
			os.Rename(fp, strings.ReplaceAll(fp, "%NAME%", p.Name))
		}
		return nil
	})

	base.Tree(to, dir)

	fmt.Printf("\n🍺 Project creation succeeded %s\n", color.GreenString(p.Name))
	fmt.Print("💻 Use the following command to start the project 👇:\n\n")

	fmt.Println(color.WhiteString("$ cd %s", p.Name))
	fmt.Println(color.WhiteString("$ go generate ./..."))
	fmt.Println(color.WhiteString("$ go build -o ./bin/ ./... "))
	fmt.Println(color.WhiteString("$ ./bin/%s -conf ./configs\n", p.Name))
	fmt.Println("			🤝 Thanks for using Kratos")
	fmt.Println("	📚 Tutorial: https://go-kratos.dev/docs/getting-started/start")
	return nil
}
