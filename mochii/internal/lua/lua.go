// Package lua provides Lua scripting support for mochii builders.
package lua

import (
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"

	lua "github.com/Shopify/go-lua"
)

type Executor struct {
	muEnv    map[string]string
	sources  []string
	derives  []string
	hashes   []string
	outputs  []string
	buildEnv map[string]string
}

func New() *Executor {
	return &Executor{
		muEnv:    make(map[string]string),
		sources:  []string{},
		derives:  []string{},
		hashes:   []string{},
		outputs:  []string{},
		buildEnv: make(map[string]string),
	}
}

func (e *Executor) RunScript(scriptPath string, args ...string) error {
	scriptContent, err := os.ReadFile(scriptPath)
	if err != nil {
		return fmt.Errorf("read script: %w", err)
	}

	state := lua.NewState()

	e.registerGlobals(state)
	e.registerArgs(state, args)
	e.registerEnv(state)

	if err := lua.DoString(state, string(scriptContent)); err != nil {
		return fmt.Errorf("lua error: %w", err)
	}

	return nil
}

func (e *Executor) RunShell(scriptPath string) error {
	cmd := exec.Command("sh", scriptPath)
	cmd.Env = os.Environ()
	for k, v := range e.muEnv {
		cmd.Env = append(cmd.Env, k+"="+v)
	}
	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	return cmd.Run()
}

func (e *Executor) SetEnv(key, value string) {
	e.muEnv[key] = value
	e.buildEnv[key] = value
}

func (e *Executor) GetEnv(key string) string {
	return e.muEnv[key]
}

func (e *Executor) Sources() []string { return e.sources }
func (e *Executor) Derives() []string { return e.derives }
func (e *Executor) Hashes() []string  { return e.hashes }
func (e *Executor) Outputs() []string { return e.outputs }

func (e *Executor) registerGlobals(state *lua.State) {
	state.NewTable()
	for k, v := range e.muEnv {
		state.PushString(k + "=" + v)
		state.SetField(-2, k)
	}
	state.SetGlobal("mu")

	state.NewTable()
	state.SetGlobal("derive")
	state.SetGlobal("source")
	state.SetGlobal("path")
	state.SetGlobal("hash")
	state.SetGlobal("output")
	state.SetGlobal("env")

	lua.SetFunctions(state, []lua.RegistryFunction{
		{Name: "mu", Function: e.muGlobal},
		{Name: "derive", Function: e.derive},
		{Name: "source", Function: e.source},
		{Name: "path", Function: e.pathFunc},
		{Name: "hash", Function: e.hashFunc},
		{Name: "output", Function: e.output},
		{Name: "env", Function: e.env},
		{Name: "run", Function: e.run},
		{Name: "fetchurl", Function: e.fetchurl},
		{Name: "nar", Function: e.nar},
		{Name: "unnar", Function: e.unnar},
	}, 0)
}

func (e *Executor) muGlobal(l *lua.State) int {
	l.NewTable()
	for k, v := range e.muEnv {
		l.PushString(v)
		l.SetField(-2, k)
	}
	return 1
}

func (e *Executor) derive(l *lua.State) int {
	n := l.Top()
	name := ""
	var inputs []string

	if n >= 1 {
		name = lua.CheckString(l, 1)
	}
	if n >= 2 {
		if l.IsTable(2) {
			l.PushNil()
			for l.Next(2) {
				if l.IsString(-1) {
					inputs = append(inputs, lua.CheckString(l, -1))
				}
				l.Pop(1)
			}
		}
	}

	e.derives = append(e.derives, name)
	l.PushString(fmt.Sprintf("derived-%s", name))
	return 1
}

func (e *Executor) source(l *lua.State) int {
	n := l.Top()
	if n < 1 {
		lua.ArgumentError(l, 1, "source URL expected")
		return 0
	}
	src := lua.CheckString(l, 1)
	e.sources = append(e.sources, src)
	fmt.Printf("SOURCE: %s\n", src)
	return 0
}

func (e *Executor) pathFunc(l *lua.State) int {
	n := l.Top()
	if n < 1 {
		lua.ArgumentError(l, 1, "path expected")
		return 0
	}
	path := lua.CheckString(l, 1)
	abs, err := filepath.Abs(path)
	if err != nil {
		l.PushString(path)
	} else {
		l.PushString(abs)
	}
	return 1
}

func (e *Executor) hashFunc(l *lua.State) int {
	n := l.Top()
	if n < 1 {
		lua.ArgumentError(l, 1, "hash expected")
		return 0
	}
	hash := lua.CheckString(l, 1)
	e.hashes = append(e.hashes, hash)
	fmt.Printf("HASH: %s\n", hash)
	l.PushString(hash)
	return 1
}

func (e *Executor) output(l *lua.State) int {
	n := l.Top()
	if n < 1 {
		lua.ArgumentError(l, 1, "output path expected")
		return 0
	}
	out := lua.CheckString(l, 1)
	e.outputs = append(e.outputs, out)
	fmt.Printf("OUTPUT: %s\n", out)
	return 0
}

func (e *Executor) env(l *lua.State) int {
	n := l.Top()
	if n < 2 {
		lua.ArgumentError(l, 1, "key and value expected")
		return 0
	}
	key := lua.CheckString(l, 1)
	val := lua.CheckString(l, 2)
	e.muEnv[key] = val
	e.buildEnv[key] = val
	return 0
}

func (e *Executor) run(l *lua.State) int {
	n := l.Top()
	if n < 1 {
		lua.ArgumentError(l, 1, "command expected")
		return 0
	}

	var cmdArgs []string
	if l.IsString(1) {
		cmdArgs = append(cmdArgs, lua.CheckString(l, 1))
	} else if l.IsTable(1) {
		l.PushNil()
		for l.Next(1) {
			if l.IsString(-1) {
				cmdArgs = append(cmdArgs, lua.CheckString(l, -1))
			}
			l.Pop(1)
		}
	}

	if len(cmdArgs) == 0 {
		lua.ArgumentError(l, 1, "non-empty command expected")
		return 0
	}

	cmd := exec.Command(cmdArgs[0], cmdArgs[1:]...)
	cmd.Env = os.Environ()
	for k, v := range e.muEnv {
		cmd.Env = append(cmd.Env, k+"="+v)
	}
	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	if err := cmd.Run(); err != nil {
		l.PushBoolean(false)
		l.PushString(err.Error())
		return 2
	}

	l.PushBoolean(true)
	return 1
}

func (e *Executor) fetchurl(l *lua.State) int {
	n := l.Top()
	if n < 1 {
		lua.ArgumentError(l, 1, "URL expected")
		return 0
	}
	url := lua.CheckString(l, 1)
	fmt.Printf("FETCHURL: %s\n", url)
	l.PushString(url)
	return 1
}

func (e *Executor) nar(l *lua.State) int {
	n := l.Top()
	if n < 1 {
		lua.ArgumentError(l, 1, "path expected")
		return 0
	}
	path := lua.CheckString(l, 1)

	cmd := exec.Command("tar", "cvf", "-", "-C", path, ".")
	cmd.Dir = filepath.Dir(path)

	out, err := cmd.Output()
	if err != nil {
		l.PushNil()
		l.PushString(err.Error())
		return 2
	}

	l.PushString(string(out))
	return 1
}

func (e *Executor) unnar(l *lua.State) int {
	n := l.Top()
	if n < 2 {
		lua.ArgumentError(l, 1, "archive and dest expected")
		return 0
	}
	archive := lua.CheckString(l, 1)
	dest := lua.CheckString(l, 2)

	if err := os.MkdirAll(dest, 0755); err != nil {
		l.PushBoolean(false)
		l.PushString(err.Error())
		return 2
	}

	cmd := exec.Command("tar", "xvf", archive, "-C", dest)
	if err := cmd.Run(); err != nil {
		l.PushBoolean(false)
		l.PushString(err.Error())
		return 2
	}

	l.PushBoolean(true)
	return 1
}

func (e *Executor) registerArgs(state *lua.State, args []string) {
	state.NewTable()
	for i, arg := range args {
		state.PushInteger(i)
		state.PushString(arg)
		state.SetTable(-3)
	}
	state.SetGlobal("ARGS")
}

func (e *Executor) registerEnv(state *lua.State) {
	state.NewTable()
	for _, v := range os.Environ() {
		parts := filepath.SplitList(v)
		if len(parts) == 2 {
			state.PushString(parts[0])
			state.PushString(parts[1])
			state.SetTable(-3)
		}
	}
	state.SetGlobal("ENV")
}

func DumpNar(w io.Writer, path string) error {
	cmd := exec.Command("tar", "cvf", "-", "-C", path, ".")
	cmd.Dir = path
	cmd.Stdout = w
	return cmd.Run()
}

func RestoreNar(r io.Reader, dest string) error {
	if err := os.MkdirAll(dest, 0755); err != nil {
		return err
	}
	cmd := exec.Command("tar", "xvf", "-", "-C", dest)
	cmd.Stdin = r
	return cmd.Run()
}
