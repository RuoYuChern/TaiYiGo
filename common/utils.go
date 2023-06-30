package common

import (
	"bytes"
	"crypto/md5"
	"encoding/base64"
	"io"
	"strings"
	"time"
)

var (
	YYYYMMDD = "20060102"
)

func GetDay(f string, t time.Time) string {
	return t.Format(f)
}

func ToDay(f string, tstr string) (time.Time, error) {
	t, err := time.Parse(f, tstr)
	return t, err
}

func GetNextDay(tstr string) (string, error) {
	t, err := ToDay(YYYYMMDD, tstr)
	if err != nil {
		return "", err
	}
	t = t.Add(24 * time.Hour)
	return GetDay(YYYYMMDD, t), nil
}

func MD5Sign(sault string, content string, time string) string {
	wr := md5.New()
	buff := bytes.Buffer{}
	buff.WriteString(sault)
	buff.WriteString(time)
	buff.WriteString(content)
	io.WriteString(wr, buff.String())
	sign := base64.StdEncoding.EncodeToString(wr.Sum(nil))
	return sign
}

func FillteST(name string) bool {
	return strings.HasPrefix(name, "*ST") || strings.HasPrefix(name, "ST")
}

func GetTodayNHour() int {
	now := time.Now()
	return now.Hour()
}
