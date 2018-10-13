package main

import (
	"encoding/json"
	"io/ioutil"
	"os"
	"path/filepath"
)

type AppDir string

func (d AppDir) Init() error {
	if err := os.MkdirAll(string(d), 0755); err != nil {
		return err
	}
	return nil
}

func (d AppDir) PathOfDataJSON() string {
	return filepath.Join(string(d), "data.json")
}

func (d AppDir) LoadData() (*AppData, error) {
	path := d.PathOfDataJSON()
	if _, err := os.Stat(path); err != nil {
		return &AppData{}, nil
	}
	data, err := ioutil.ReadFile(path)
	if err != nil {
		return nil, err
	}
	appData := &AppData{}
	if err := json.Unmarshal(data, appData); err != nil {
		return nil, err
	}
	return appData, nil
}

func (d AppDir) SaveData(appData *AppData) error {
	data, err := json.Marshal(appData)
	if err != nil {
		return nil
	}
	if err := ioutil.WriteFile(d.PathOfDataJSON(), data, 0666); err != nil {
		return err
	}
	return nil
}
