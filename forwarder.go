package main

import (
	"fmt"
	"github.com/go-cmd/cmd"
	"io"
	"log"
	"os"
	"os/exec"
	"os/signal"
	"strings"
	"syscall"
	"time"
)

type outCounter struct {
	int
}

type errCounter struct {
	int
}

type counterPair struct {
	outC outCounter
	errC errCounter
}

type lastOutLineCounter interface {
	count() int
}

func (o outCounter) count() int {
	return o.int
}

func (e errCounter) count() int {
	return e.int
}

func portForwardAll(args []forwardArgs, out io.Writer) error {
	kubeExec, err := exec.LookPath(kubectlCmdName)
	if err != nil {
		return err
	}

	agg := make(chan cmd.Status)
	var kubeCmds []*cmd.Cmd
	lastLineMap := make(map[string]counterPair)
	for _, fArgs := range args {

		kubeCmd := cmd.NewCmd(kubeExec, fArgs...)
		kubeCmds = append(kubeCmds, kubeCmd)
		statusChan := kubeCmd.Start()
		go func(c <-chan cmd.Status) {
			for msg := range c {
				agg <- msg
			}
		}(statusChan)
		kubeCmdKey := getKubeCmdKey(kubeCmd)
		lastLineMap[kubeCmdKey] = counterPair{}
		_, err = fmt.Fprintf(out, "Running %q\n", kubeCmdKey)
		if err != nil {
			logPrintToOutErr(err)
		}
	}

	ticker := time.NewTicker(100 * time.Millisecond)

	go func() {
		for range ticker.C {
			for _, kubeCmd := range kubeCmds {
				kubeCmdKey := getKubeCmdKey(kubeCmd)
				pair, ok := lastLineMap[kubeCmdKey]
				if !ok {
					_, err := fmt.Fprintf(
						out,
						"An error while trying to print the status: %v",
						fmt.Errorf("could not find the %s in the last line map: %w", kubeCmd.Name, err),
					)
					if err != nil {
						log.Fatal(err)
					}
				}
				lastLineMap[kubeCmdKey] = printlnStatusLines(kubeCmd.Status(), pair, out)
			}
		}
	}()

	sig := make(chan os.Signal, 1)
	signal.Notify(sig, syscall.SIGINT, syscall.SIGTERM, syscall.SIGKILL)

	select {
	case rec := <-agg:
		if err != nil {
			logPrintToOutErr(err)
		}
		err = stopCmds(kubeCmds)
		if err != nil {
			return fmt.Errorf("an error while stopping the cmds: %w", err)
		}
		return fmt.Errorf("%w: %s", ErrSignal, fmt.Sprintf("%d", rec.Exit))
	case <-sig:
		signal.Stop(sig)
		return nil
	}
}

func stopCmds(cmds []*cmd.Cmd) error {
	for _, aCmd := range cmds {
		if !aCmd.Status().Complete {
			err := aCmd.Stop()
			if err != nil {
				return err
			}
		}
	}
	return nil
}

func getKubeCmdKey(kubeCmd *cmd.Cmd) string {
	return strings.Join(kubeCmd.Args, " ")
}

func printlnStatusLines(finalStatus cmd.Status, cPair counterPair, out io.Writer) counterPair {
	readOutLinesCounter := printlnLinesAndReturnTheirCount(finalStatus.Stdout, cPair.outC, out)
	readErrLinesCounter := printlnLinesAndReturnTheirCount(finalStatus.Stderr, cPair.outC, out)

	return counterPair{
		outC: outCounter{cPair.outC.count() + readOutLinesCounter},
		errC: errCounter{cPair.errC.count() + readErrLinesCounter},
	}
}

func printlnLinesAndReturnTheirCount(lines []string, aLastOutLineCounter lastOutLineCounter, out io.Writer) int {
	if len(lines) == 0 {
		return 0
	}

	var counter int

	for _, line := range lines[aLastOutLineCounter.count()+1:] {
		_, err := fmt.Fprintln(out, line)
		if err != nil {
			logPrintToOutErr(err)
		}
		counter++
	}

	return counter
}

func logPrintToOutErr(err error) {
	log.Fatalf("An error while printing to the provided out: %v", err)
}
