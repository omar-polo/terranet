package main

import (
	"context"
	"flag"
	"fmt"
	"io"
	"os"
	"path"
	"strings"

	"github.com/google/go-jsonnet"
	"github.com/hashicorp/terraform-exec/tfexec"
)

func fatal(msg ...any) {
	fmt.Fprintln(os.Stderr, msg...)
	os.Exit(1)
}

func main() {
	flag.Parse()

	args := flag.Args()
	if len(args) < 1 {
		fatal("wrong usage")
	}

	cmd := args[0]

	switch cmd {
	case "version":
		fmt.Println("terranet: PoC")
		return
	case "eval", "diff", "apply":
		break
	default:
		fatal("wrong usage")
	}

	if len(args) != 2 {
		fatal("wrong usage")
	}
	main := path.Join(args[1], "main.jsonnet")
	tfdir := path.Join(args[1], ".tn2tf")

	vm := jsonnet.MakeVM()
	res, err := vm.EvaluateFile(main)
	if err != nil {
		fatal("jsonnet error:", err)
	}

	if cmd == "eval" {
		fmt.Print(res)
		return
	}

	if err := os.MkdirAll(tfdir, 0755); err != nil {
		fatal("can't mkdir", tfdir+":", err)
	}

	dst := path.Join(tfdir, "main.tf.json")
	f, err := os.Create(dst)
	if err != nil {
		fatal("couldn't create main.tf.json:", err)
	}

	rdr := strings.NewReader(res)
	if _, err := io.Copy(f, rdr); err != nil {
		fatal("failed to write main.tf.json:", err)
	}

	tf, err := tfexec.NewTerraform(tfdir, "tofu")
	if err != nil {
		fatal("newterraform failed:", err)
	}

	if err = tf.Init(context.Background(), tfexec.Upgrade(true)); err != nil {
		fatal("error running init:", err)
	}

	switch cmd {
	case "diff":
		_, err = tf.PlanJSON(context.Background(), os.Stdout)
		if err != nil {
			fatal("can't show:", err)
		}

	case "apply":
		fmt.Println("better not at this stage")
	}
}
