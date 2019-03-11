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
	sm.runner = sync.Map{}
	err := sm.loadScriptsFromDir()
	if err != nil {
		return nil, err
	}

	rand.Seed(time.Now().Unix())

	return sm, nil
}

type ScriptManager struct {
	api       plugin.API
	botId     string
	scriptDir string
	scripts   map[string]Script
	runner    sync.Map
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
	sm.api.LogInfo(fmt.Sprintf("Starting new runner for script id %s for team %s and user %s", scriptId, teamId, userId))

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

	sm.runner.Store(runner.GetChannelId(), runner)

	err = runner.Start()

	if err != nil {
		sm.api.LogError(fmt.Sprintf("Error running script %s: %s", scriptId, err))
		return
	}

	sm.api.LogDebug(fmt.Sprintf("Stopping runner for script id %s for team %s and user %s", scriptId, teamId, userId))
}

func (sm *ScriptManager) TriggerResponse(responseId, channelId, userId string) error {
	sm.api.LogDebug(fmt.Sprintf("Starting response trigger for reaction id %s for team %s and user %s", responseId, channelId, userId))

	data, ok := sm.runner.Load(channelId)

	if !ok {
		return errors.New(fmt.Sprintf("No runner found for given channel id: %s", channelId))
	}

	runner := data.(*ScriptRunner)

	sm.api.LogDebug(fmt.Sprintf("Stopping response trigger for reaction id %s for team %s and user %s", responseId, channelId, userId))

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
