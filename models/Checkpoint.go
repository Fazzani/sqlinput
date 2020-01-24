package models

import (
	"crypto/sha1"
	"encoding/base64"
	"io/ioutil"
	"path"
)

// CheckpointService checkpoint handler
type CheckpointService interface {
	GetCheckpointPath() (pathCP string)
	SaveCheckpoint(data string) (err error)
	GetLastCheckpointValue(defaultValue string) (lastID string, err error)
}

// GetCheckpointPath return the file path where the id will be saved
func (c ConfigInput) GetCheckpointPath() (pathCP string) {
	hasher := sha1.New()
	hasher.Write([]byte(c.Configuration.Stanza.Name))
	sha := base64.URLEncoding.EncodeToString(hasher.Sum(nil))
	pathCP = path.Join(c.CheckpointDir, sha)
	Logf(LogDebug, "GetCheckpointPath =>>> %s", pathCP)
	return
}

// SaveCheckpoint save checkpoint id to the file
func (c ConfigInput) SaveCheckpoint(cpID string) (err error) {
	err = ioutil.WriteFile(c.GetCheckpointPath(), []byte(cpID), 0777)
	return
}

// GetLastCheckpointValue extract the last saved id
func (c ConfigInput) GetLastCheckpointValue(defaultValue string) (lastID string, err error) {
	data, err := ioutil.ReadFile(c.GetCheckpointPath())
	Logf(LogDebug, "configInput =>>> %v", data)
	lastID = string(data)
	if lastID == "" {
		lastID = defaultValue
	}
	Logf(LogDebug, "lastID =>>> %s", lastID)
	return
}
