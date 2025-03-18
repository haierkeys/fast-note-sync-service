package util

import (
	"math/rand"
	"reflect"
	"regexp"
	"strconv"
	"strings"
	"time"
	"unicode"

	"github.com/pkg/errors"
)

const (
	firstIndex = 0
)

func GetIndexSlice(target int, slice []int) (index int) {
	position := -1
	if len(slice) > 0 {
		num := firstIndex
		for _, v := range slice {
			if v == target {
				position = num
			}
			num++
		}
	}
	return position
}

func Inarray(val interface{}, array interface{}) (exists bool, index int) {
	exists = false
	index = -1
	switch reflect.TypeOf(array).Kind() {
	case reflect.Slice:
		s := reflect.ValueOf(array)
		for i := 0; i < s.Len(); i++ {
			if reflect.DeepEqual(val, s.Index(i).Interface()) == true {
				index = i
				exists = true
				return
			}
		}
	}
	return
}

func IsValidEmail(email string) bool {
	// 定义邮箱的正则表达式，支持 Gmail 和其他常见邮箱格式
	var emailRegex = `^[a-zA-Z0-9._%+-]+@[a-zA-Z0-9.-]+\.[a-zA-Z]{2,}$`

	// 编译正则表达式
	re := regexp.MustCompile(emailRegex)

	// 使用正则表达式匹配邮箱
	return re.MatchString(email)
}

func IsValidUsername(username string) bool {
	if len(username) < 3 || len(username) > 15 {
		return false // 用户名长度不符合
	}
	// 检查每个字符是否是字母、数字或下划线
	for _, char := range username {
		if !(unicode.IsLetter(char) || unicode.IsDigit(char) || char == '_') {
			return false // 包含不合法字符
		}
	}
	return true
}

func GetLastDateOfNextMonth(d time.Time) time.Time {
	return GetFirstDateOfMonth(d).AddDate(0, 2, -1)
}

func ArrayUnique(a interface{}) (ret []interface{}) {
	va := reflect.ValueOf(a)
	for i := 0; i < va.Len(); i++ {
		if i > 0 && reflect.DeepEqual(va.Index(i-1).Interface(), va.Index(i).Interface()) {
			continue
		}
		ret = append(ret, va.Index(i).Interface())
	}
	return ret
}

// 生成若干个不重复的随机数
func GenerateRandomNumber(start int, end int, count int) (nums map[int]int) {
	//范围检查
	if end < start || (end-start) < count {
		return nil
	}
	nums = make(map[int]int)
	//随机数生成器，加入时间戳保证每次生成的随机数不一样
	r := rand.New(rand.NewSource(time.Now().UnixNano()))
	for len(nums) < count {
		//生成随机数
		num := r.Intn(end-start) + start
		//查重
		if _, ok := nums[num]; !ok {
			nums[num] = num
		}
	}
	return nums
}

// 生成一个随机数
func GenerateRandomSingleNumber(start int, end int, count int) (num int) {
	//范围检查
	if end < start || (end-start) < count {
		return num
	}

	//随机数生成器，加入时间戳保证每次生成的随机数不一样
	r := rand.New(rand.NewSource(time.Now().UnixNano()))
	num = r.Intn(end-start) + start
	return num
}

func GenerateRandom(start int, end int, count int) []int {
	//范围检查
	if end < start || (end-start) < count {
		return nil
	}

	//存放结果的slice
	nums := make([]int, 0)
	//随机数生成器，加入时间戳保证每次生成的随机数不一样
	r := rand.New(rand.NewSource(time.Now().UnixNano()))
	for len(nums) < count {
		//生成随机数
		num := r.Intn(end-start) + start

		//查重
		exist := false
		for _, v := range nums {
			if v == num {
				exist = true
				break
			}
		}

		if !exist {
			nums = append(nums, num)
		}
	}

	return nums
}

func Wait(num float32) {
	tmpTime := time.Duration(num * 1000000000)
	time.Sleep(tmpTime)
}

func StrToMap(s string) (out map[int]int) {
	out = make(map[int]int)
	tmpList := strings.Split(s, ",")
	for _, v := range tmpList {
		tmpid, _ := strconv.Atoi(v)
		out[tmpid] = tmpid
	}
	return out
}

func StrToInt(s string) (out []int) {
	tmpList := strings.Split(s, ",")
	for _, v := range tmpList {
		tmpid, _ := strconv.Atoi(v)
		out = append(out, tmpid)
	}
	return out
}

func IntSliceToStringSlice(target []int) (out []string) {
	for _, v := range target {
		tmp := strconv.Itoa(v)
		out = append(out, tmp)
	}
	return out
}

func StringToInt64(str string) (out []int64) {
	tmpList := strings.Split(str, ",")
	for _, v := range tmpList {
		tmpId, _ := strconv.ParseInt(v, 10, 64)
		out = append(out, tmpId)
	}
	return out
}

func InArray(search interface{}, haystack interface{}) (exists bool, index int, err error) {
	exists = false
	index = -1
	err = errors.New("unsupported type")
	switch reflect.TypeOf(haystack).Kind() {
	case reflect.Slice:
		err = nil
		s := reflect.ValueOf(haystack)
		for i := 0; i < s.Len(); i++ {
			if reflect.DeepEqual(search, s.Index(i).Interface()) == true {
				index = i
				exists = true
				break
			}
		}
	}
	return exists, index, err
}

func IntSliceToStrSlice(data interface{}) (out []string, err error) {
	err = errors.New("unsupported type")
	switch reflect.TypeOf(data).Kind() {
	case reflect.Slice:
		err = nil
		s := reflect.ValueOf(data)
		for i := 0; i < s.Len(); i++ {
			var (
				tmpData int64
				val     = s.Index(i).Interface()
			)
			switch val.(type) {
			case int8:
				tmpData = int64(val.(int8))
			case int16:
				tmpData = int64(val.(int16))
			case int32:
				tmpData = int64(val.(int32))
			case int64:
				tmpData = val.(int64)
			case int:
				tmpData = int64(val.(int))
			}
			tmpStr := strconv.FormatInt(tmpData, 10)
			out = append(out, tmpStr)
		}
	}
	return out, err
}

func RemoveDuplicate(list []int) (out []int) {
	for _, i := range list {
		if len(out) == 0 {
			out = append(out, i)
		} else {
			for k, v := range out {
				if i == v {
					break
				}
				if k == len(out)-1 {
					out = append(out, i)
				}
			}
		}
	}
	return out
}

// GetFirstDateOfMonth 获取传入的时间所在月份的第一天，即某月第一天的0点。如传入time.Now(), 返回当前月份的第一天0点时间。
func GetFirstDateOfMonth(d time.Time) time.Time {
	d = d.AddDate(0, 0, -d.Day()+1)
	return GetZeroTime(d)
}

// GetLastDateOfMonth 获取传入的时间所在月份的最后一天，即某月最后一天的0点。如传入time.Now(), 返回当前月份的最后一天0点时间。
func GetLastDateOfMonth(d time.Time) time.Time {
	return GetFirstDateOfMonth(d).AddDate(0, 1, -1)
}

// GetZeroTime 获取某一天的0点时间
func GetZeroTime(d time.Time) time.Time {
	return time.Date(d.Year(), d.Month(), d.Day(), 0, 0, 0, 0, d.Location())
}

// GetEndTime 获取某一天的0点时间
func GetEndTime(d time.Time) time.Time {
	return time.Date(d.Year(), d.Month(), d.Day(), 23, 59, 59, 0, d.Location())
}

// IntersectionInt 获取两个slice的交集
func IntersectionInt(slice1 []int, slice2 []int) (outSlice []int) {
	if len(slice1) == 0 || len(slice2) == 0 {
		return outSlice
	}
	for _, li1 := range slice1 {
		for _, li2 := range slice2 {
			if li1 == li2 {
				outSlice = append(outSlice, li1)
				continue
			}
		}
	}
	return outSlice
}

// TimeParse 时间日期格式化
func TimeParse(layout string, in string) time.Time {
	local, _ := time.LoadLocation("Local")
	timer, _ := time.ParseInLocation(layout, in, local)
	return timer
}
