package common

import (
	"bytes"
	"container/list"
	"crypto/md5"
	"encoding/base64"
	"encoding/hex"
	"io"
	"io/fs"
	"math"
	"path/filepath"
	"strconv"
	"strings"
	"time"
)

var (
	YYYYMMDD = "20060102"
)

func GetDay(f string, t time.Time) string {
	return t.Format(f)
}

func GetYear(t time.Time) string {
	return strconv.Itoa(t.Year())
}

func GetYearMon(t time.Time) string {
	day := GetDay(YYYYMMDD, t)
	return SubString(day, 0, 6)
}

func ToDay(f string, tstr string) (time.Time, error) {
	t, err := time.Parse(f, tstr)
	return t, err
}

func IsDayBeforN(f string, tstr string, days int) bool {
	day, err := ToDay(f, tstr)
	if err != nil {
		Logger.Infof("%s toDay failed:%s", tstr, err)
		return true
	}
	diff := time.Since(day)
	return (int(diff.Hours()/24) >= days)
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

func GetMd5(content string, noice string) string {
	wr := md5.New()
	buff := bytes.Buffer{}
	buff.WriteString(noice)
	buff.WriteString(content)
	io.WriteString(wr, buff.String())
	sign := hex.EncodeToString(wr.Sum(nil))
	return sign
}

func TodayIsWeek() bool {
	t := time.Now()
	wkd := t.Weekday()
	return (wkd == time.Sunday) || (wkd == time.Saturday)
}

func FillteST(name string) bool {
	return strings.HasPrefix(name, "*ST") || strings.HasPrefix(name, "ST")
}

func GetTodayNHour() int {
	now := time.Now()
	return now.Hour()
}

func GetTodayHAndM() (int, int) {
	now := time.Now()
	return now.Hour(), now.Minute()
}

func FFloat(f float64, decimal int) float64 {
	dd := float64(1)
	if decimal > 0 {
		dd = math.Pow10(decimal)
	}
	res := strconv.FormatFloat(math.Trunc(f*dd)/dd, 'f', -1, 64)
	fv, _ := strconv.ParseFloat(res, 64)
	return fv
}

func StrToF32(value string) float32 {
	f, _ := strconv.ParseFloat(value, 32)
	return float32(f)
}

func FloatToStr(f float64, decimal int) string {
	dd := float64(1)
	if decimal > 0 {
		dd = math.Pow10(decimal)
	}
	res := strconv.FormatFloat(math.Trunc(f*dd)/dd, 'f', -1, 64)
	return res
}

func SubString(source string, start int, end int) string {
	length := len(source)
	if (start == 0) && (end == length) {
		return source
	}

	var r = []rune(source)
	return string(r[start:end])
}

func PreSubString(source string, start int) string {
	if start == 0 {
		return source
	}
	var r = []rune(source)
	return string(r[start:])
}

func GetFileList(dir string, sufix string, exclude string, size int) (*list.List, error) {
	lhp := NewLp(size+1, func(a1, a2 any) int {
		v1 := a1.(string)
		v2 := a2.(string)
		return strings.Compare(v1, v2)
	})
	err := filepath.WalkDir(dir, func(p string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if d.IsDir() {
			return nil
		}
		if !strings.HasSuffix(d.Name(), sufix) || strings.HasSuffix(d.Name(), exclude) {
			return nil
		}
		lhp.Add(d.Name())
		return nil
	})
	if err != nil {
		return nil, err
	}
	fsList := list.New()
	for {
		v := lhp.Top()
		if v == nil {
			break
		}
		fsList.PushBack(v)
	}
	return fsList, nil
}
