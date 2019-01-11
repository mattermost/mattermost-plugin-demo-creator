package demo

func NewScriptRunner(script Script) *ScriptRunner {
	s := &ScriptRunner {
		script: &script,
	}

	return s
}

type ScriptRunner struct {
	script *Script
}