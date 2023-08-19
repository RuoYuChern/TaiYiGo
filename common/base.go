package common

func BaseInit(path string) {
	loadTaoConf(path)
	initLogger(Conf)
}
