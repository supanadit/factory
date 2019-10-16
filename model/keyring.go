package model

import (
	"fmt"
	"github.com/99designs/keyring"
	"github.com/naoina/toml"
	"io/ioutil"
	"log"
	"strings"
)

type Keyring struct {
	Username string
	Host     string
	Port     string
	Password string `toml:"-"`
}

func (keyringModel Keyring) GetSystemKeyringName() string {
	return keyringModel.Host + "-" + keyringModel.Username
}

func (keyringModel Keyring) GetPasswordFromSystem() string {
	ring, _ := keyring.Open(keyring.Config{
		ServiceName: SystemName,
	})
	password, _ := ring.Get(keyringModel.GetSystemKeyringName())
	return string(password.Data)
}

func (keyringModel Keyring) SaveFull(configuration Configuration) {
	keyringModel.SaveToConfiguration(configuration)
	keyringModel.SaveToSystem()
}

func (keyringModel Keyring) SaveToSystem() {
	ring, _ := keyring.Open(keyring.Config{
		ServiceName: SystemName,
	})

	_ = ring.Set(keyring.Item{
		Key:  keyringModel.GetSystemKeyringName(),
		Data: []byte(keyringModel.Password),
	})
}

func (keyringModel Keyring) SaveToConfiguration(configuration Configuration) {
	allKeyring := GetAllKeyringConfiguration(configuration)

	allKeyring.Keyring = append(allKeyring.Keyring, keyringModel)

	dataToml, _ := toml.Marshal(&allKeyring)
	err := ioutil.WriteFile(configuration.GetKeyringConfigFilePath(), dataToml, 0644)
	if err != nil {
		if DEBUG {
			log.Print(err)
		} else {
			fmt.Printf("Cannot save keyring configuration")
		}
	}
}

func (keyringModel Keyring) RemoveFromSystem() {
	ring, _ := keyring.Open(keyring.Config{
		ServiceName: SystemName,
	})

	_ = ring.Remove(keyringModel.GetSystemKeyringName())
}

func (keyringModel Keyring) RemoveFromConfiguration(configuration Configuration) {
	allKeyring := GetAllKeyringConfiguration(configuration)
	var changedKeyRing KeyringConfiguration
	for _, element := range allKeyring.Keyring {
		if element.Host == keyringModel.Host {
			if element.Username != keyringModel.Username {
				changedKeyRing.Keyring = append(changedKeyRing.Keyring, element)
			}
		} else {
			changedKeyRing.Keyring = append(changedKeyRing.Keyring, element)
		}
	}

	dataToml, _ := toml.Marshal(&changedKeyRing)
	err := ioutil.WriteFile(configuration.GetKeyringConfigFilePath(), dataToml, 0644)
	if err != nil {
		if DEBUG {
			log.Print(err)
		} else {
			fmt.Printf("Failed to remove keyring from configuration")
		}
	}
}

func (keyringModel Keyring) RemoveFromAll(configuration Configuration) {
	keyringModel.RemoveFromConfiguration(configuration)
	keyringModel.RemoveFromSystem()
}

func (keyringModel Keyring) Exist(configuration Configuration) bool {
	allKeyring := GetAllKeyringConfiguration(configuration)
	found := false
	for _, element := range allKeyring.Keyring {
		if element.Host == keyringModel.Host && element.Username == keyringModel.Username {
			found = true
		}
	}
	return found
}

func (keyringModel Keyring) GetHostPort() string {
	var port string
	if keyringModel.Port == "" {
		port = "22"
	} else {
		port = keyringModel.Port
	}
	return keyringModel.Host + ":" + port
}

func GetKeyringFromString(data string) Keyring {
	var keyringData Keyring
	hostPort := strings.Split(data, ":")
	if len(hostPort) > 1 {
		keyringData.Port = hostPort[1]
	}

	userHost := strings.Split(hostPort[0], "@")
	if len(userHost) > 1 {
		keyringData.Username = userHost[0]
		keyringData.Host = userHost[1]
	} else {
		keyringData.Host = data
	}
	return keyringData
}
