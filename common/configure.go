package common

import (
	"log"
	"os"

	"gopkg.in/yaml.v3"
)

type httpConf struct {
	Port   int    `yaml:"port"`
	Prefix string `yaml:"path-prefix"`
	Token  string `yaml:"token"`
	Jwt    string `yaml:"jwt"`
	Avator string `yaml:"avator"`
}

type loggerConf struct {
	Level      string `yaml:"level"`
	File       string `yaml:"log_file"`
	MaxSize    int    `yaml:"max_size"`
	MaxBackups int    `yaml:"max_backups"`
	MaxAge     int    `yaml:"max_age"`
	Env        string `yaml:"env"`
}

type infraConf struct {
	FsDir string `yaml:"fs_dir"`
}

type quotesConf struct {
	TuToken    string `yaml:"tu.share.token"`
	Sault      string `yaml:"taiji.sault"`
	HistoryEnd string `yaml:"history.end"`
}

type brainConf struct {
	StfPriceCount int `yaml:"stf-price-count"`
}

type TaoConf struct {
	Infra  infraConf  `yaml:"infra"`
	Log    loggerConf `yaml:"logger"`
	Http   httpConf   `yaml:"http"`
	Quotes quotesConf `yaml:"quotes"`
	Brain  brainConf  `yaml:"brain"`
}

var Conf *TaoConf

func loadTaoConf(path string) {
	Conf = &TaoConf{}
	ymlFile, err := os.ReadFile(path)
	if err != nil {
		log.Printf("ReadFile failed:%s", err.Error())
		panic(err)
	}
	err = yaml.Unmarshal(ymlFile, Conf)
	if err != nil {
		log.Printf("Unmarshal failed:%s", err.Error())
		panic(err)
	}
}
