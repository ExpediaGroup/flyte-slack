/*
Copyright (C) 2018 Expedia Group.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package client

import (
	"encoding/json"
	"github.com/HotelsDotCom/go-logger"
	"io/ioutil"
	"os"
	"path/filepath"
)

type Backup interface {
	// saves/backs up channel ids
	Backup(channelIds []string)
	// loads channel ids from backup
	Load() []string
}

type fileBackup struct {
	path string
}

// returns file backup, channels will be saved to the supplied directory - {backupDir}/channels.json
func NewFileBackup(backupDir string) Backup {

	backupPath := createBackupFile(backupDir, "channels.json")
	return fileBackup{backupPath}
}

// returns file backup, channels will be saved to $TMPDIR/channels.json
// (env. variable for temporary dir is different on different OS)
func NewTmpFileBackup() Backup {

	backupDir := createDefaultBackupDir()
	return NewFileBackup(backupDir)
}

func (f fileBackup) Backup(channelIds []string) {

	b, err := json.Marshal(channelIds)
	if err != nil {
		logger.Fatalf("cannot backup channels=%v to file=%s, error marshaling: %v", channelIds, f.path, err)
		return
	}

	if err := ioutil.WriteFile(f.path, b, 0644); err != nil {
		logger.Fatalf("cannot save channels=%v to file=%s: %v", channelIds, f.path, err)
	}
}

func (f fileBackup) Load() []string {

	b, err := ioutil.ReadFile(f.path)
	if err != nil {
		logger.Fatalf("cannot read channels from backup file=%s: %v", f.path, err)
	}

	channels := []string{}
	if err := json.Unmarshal(b, &channels); err != nil {
		logger.Fatalf("cannot unmarshal channels from backup file=%s content=%s: %v", f.path, string(b), err)
	}
	return channels
}

func createDefaultBackupDir() string {

	dir := filepath.Join(os.TempDir(), "flyte-slack")
	if err := os.MkdirAll(dir, 0755); err != nil {
		logger.Fatalf("cannot create backup dir=%s: %v", dir, err)
	}
	return dir
}

func createBackupFile(dir, fileName string) string {

	path := filepath.Join(dir, fileName)
	if _, err := os.Stat(path); os.IsNotExist(err) {
		if _, err := os.Create(path); err != nil {
			logger.Fatalf("cannot create backup path=%s: %v", path, err)
		}
		if err = ioutil.WriteFile(path, []byte(`[]`), 0644); err != nil {
			logger.Fatalf("cannot write initial backup file=%s: %v", path, err)
		}
	}
	return path
}
