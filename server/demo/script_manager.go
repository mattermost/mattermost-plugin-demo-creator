package demo

import (
	"errors"
	"fmt"
	"github.com/go-yaml/yaml"
	"github.com/mattermost/mattermost-server/plugin"
	"io/ioutil"
	"math/rand"
	"path"
	"sync"
	"time"
)

func NewScriptManager(api plugin.API, scriptDir string) (*ScriptManager, error) {
	sm := &ScriptManager{}
	sm.api = api
	sm.scriptDir = scriptDir
	sm.scripts = make(map[string]Script)
	sm.runner = make(map[string]*ScriptRunner)
	err := sm.loadScriptsFromDir()
	if err != nil {
		return nil, err
	}

	sm.runnerLock = &sync.RWMutex{}

	rand.Seed(time.Now().Unix())

	return sm, nil
}

type ScriptManager struct {
	api        plugin.API
	botId      string
	scriptDir  string
	scripts    map[string]Script
	runner     map[string]*ScriptRunner
	runnerLock *sync.RWMutex
}

func (sm *ScriptManager) SetBotId(id string) {
	sm.botId = id
}

func (sm *ScriptManager) GetScript(id string) (Script, error) {
	script, ok := sm.scripts[id]

	if !ok {
		return Script{}, errors.New("script not found")
	}

	return script, nil
}

func (sm *ScriptManager) GetScripts() map[string]Script {
	return sm.scripts
}

func (sm *ScriptManager) GetScriptCount() int {
	return len(sm.scripts)
}

func (sm *ScriptManager) StartScript(teamId, userId, scriptId string) {

	script, err := sm.GetScript(scriptId)

	if err != nil {
		sm.api.LogError(fmt.Sprintf("Error getting script %s", scriptId))
		return
	}

	runner, err := NewScriptRunner(sm.api, script, sm.botId, teamId, userId)

	if err != nil {
		sm.api.LogError(fmt.Sprintf("Error creating runner for script %s: %s", scriptId, err))
		return
	}

	sm.runnerLock.Lock()
	sm.runner[runner.GetChannelId()] = runner
	sm.runnerLock.Unlock()

	err = runner.Start()

	if err != nil {
		sm.api.LogError(fmt.Sprintf("Error running script %s: %s", scriptId, err))
		return
	}

}

func (sm *ScriptManager) TriggerResponse(responseId, channelId, userId string) error{
	sm.runnerLock.RLock()
	runner, ok := sm.runner[channelId]

	if !ok {
		return errors.New(fmt.Sprintf("No runner found for given channel id: %s", channelId))
	}

	return runner.TriggerResponse(responseId, userId)
}

func (sm *ScriptManager) loadScriptsFromDir() error {

	files, err := ioutil.ReadDir(sm.scriptDir)
	if err != nil {
		return err
	}

	var script Script

	for _, f := range files {
		data, err := ioutil.ReadFile(path.Join(sm.scriptDir, f.Name()))

		if err != nil {
			return errors.New(fmt.Sprintf("error reading file %s: %s", f.Name(), err))
		}

		err = yaml.Unmarshal(data, &script)

		if err != nil {
			return errors.New(fmt.Sprintf("error parsing file %s: %s", f.Name(), err))
		}

		sm.scripts[script.Id] = script
	}

	return nil
}
