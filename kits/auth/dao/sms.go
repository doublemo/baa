package dao

import (
	"context"
	"crypto/rand"
	"errors"
	"fmt"
	"io"
	"strconv"
	"time"
)

const defaultSMSKey = "sms"

// GenerateSMSCode 创建验证码
func GenerateSMSCode(phone string, codeLen int, expire time.Duration) (string, error) {
	code, vcode, err := generateSMSVerificationCode(codeLen, expire)
	if err != nil {
		return "", err
	}

	namer := RDBNamer(defaultSMSKey, phone)
	ctx, cancel := context.WithTimeout(context.Background(), time.Second*10)
	defer cancel()
	ret := rdb.Set(ctx, namer, vcode, expire)
	err = ret.Err()
	if err != nil {
		return "", err
	}
	return code, nil
}

// RemoveSMSCode 删除验证码
func RemoveSMSCode(phone string) error {
	namer := RDBNamer(defaultSMSKey, phone)
	ctx, cancel := context.WithTimeout(context.Background(), time.Second*10)
	defer cancel()
	ret := rdb.Del(ctx, namer)
	return ret.Err()
}

// GetSMSCode 获取验证码
func GetSMSCode(phone string) (string, error) {
	namer := RDBNamer(defaultSMSKey, phone)
	ctx, cancel := context.WithTimeout(context.Background(), time.Second*10)
	defer cancel()
	ret := rdb.Get(ctx, namer)
	err := ret.Err()
	if err != nil {
		return "", err
	}

	return ret.Val(), nil
}

func generateSMSVerificationCode(max int, expire time.Duration) (string, string, error) {
	var table = [...]byte{'1', '2', '3', '4', '5', '6', '7', '8', '9', '0'}

	b := make([]byte, max)
	n, err := io.ReadAtLeast(rand.Reader, b, max)
	if n != max {
		return "", "", err
	}

	for i := 0; i < len(b); i++ {
		b[i] = table[int(b[i])%len(table)]
	}

	code := string(b)
	return code, fmt.Sprintf("%s%d", string(b), time.Now().Add(expire).Unix()), nil
}

// ParseSMSVerificationCode 解析验证码Code
func ParseSMSVerificationCode(s string, max int) (string, int64, error) {
	if len(s) < max+10 {
		return "", 0, errors.New("InvalidVerificationCode")
	}

	code := s[0:max]
	expires := s[max:]
	m, err := strconv.ParseInt(expires, 10, 64)
	if err != nil {
		return "", 0, err
	}
	return code, m, nil
}
