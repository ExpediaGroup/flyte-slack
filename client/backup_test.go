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
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"io/ioutil"
	"os"
	"path/filepath"
	"github.com/HotelsDotCom/go-logger/loggertest"
	"testing"
)

var BackupTmpDir string

func BeforeBackup() {

	loggertest.Init("DEBUG")
	BackupTmpDir = filepath.Join(os.TempDir(), "test-flyte-slack")
	os.MkdirAll(BackupTmpDir, 0755)
}

func AfterBackup() {

	os.RemoveAll(BackupTmpDir)
	loggertest.Reset()
}

func TestFileBackupSavesChannels(t *testing.T) {

	BeforeBackup()
	defer AfterBackup()

	channelIds := []string{"id-1", "id-2"}
	backup := NewFileBackup(BackupTmpDir)

	backup.Backup(channelIds)

	loadedChannels := backup.Load()
	require.Equal(t, 2, len(loadedChannels))
	assert.Equal(t, channelIds, loadedChannels)
}

func TestNewBackupWithNonExistingPath(t *testing.T) {

	BeforeBackup()
	defer AfterBackup()

	assert.Panics(t, func() { NewFileBackup("/path-does-not-exist") })
}

func TestLoadInvalidJsonFromBackup(t *testing.T) {

	BeforeBackup()
	defer AfterBackup()

	backup := NewFileBackup(BackupTmpDir)
	backupPath := filepath.Join(BackupTmpDir, "channels.json")
	err := ioutil.WriteFile(backupPath, []byte(`--- invalid json ---`), 0644)
	require.Nil(t, err)

	assert.Panics(t, func() { backup.Load() })
}
