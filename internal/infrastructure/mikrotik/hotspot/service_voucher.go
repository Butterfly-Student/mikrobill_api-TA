package hotspot

import (
	"fmt"
	"mikrobill/internal/infrastructure/mikrotik/model"
	"strconv"
	"strings"
	"time"
)

// ========== VOUCHER GENERATION ==========

// GenerateVouchers generate voucher codes
func (s *Service) GenerateVouchers(config model.VoucherRequest) (*model.VoucherResponse, error) {
	// Create comment
	now := time.Now()
	dateStr := fmt.Sprintf("%02d.%02d.%s",
		now.Month(),
		now.Day(),
		strconv.Itoa(now.Year())[2:])
	fullComment := fmt.Sprintf("%s-%s-%s-%s",
		config.UserType,
		config.GenCode,
		dateStr,
		config.Comment)

	// Calculate data limit
	var dataLimitBytes int64
	if config.DataLimit != "" {
		var err error
		dataLimitBytes, err = calculateDataLimit(config.DataLimit)
		if err != nil {
			return &model.VoucherResponse{
				Message: "error",
				Data: model.VoucherResponseData{
					Error: err.Error(),
				},
			}, err
		}
	}

	prefixLength := len(config.Prefix)
	actualRandomLength := config.UserLength - prefixLength
	if actualRandomLength < 1 {
		actualRandomLength = 1
	}

	var generatedUsers []model.GeneratedVoucher

	// Generate vouchers
	for i := 0; i < config.Qty; i++ {
		var username, password string

		if config.UserType == model.UserTypeUP {
			// User + Password mode
			username = config.Prefix + generateRandomString(actualRandomLength, config.CharType)
			password = generateRandomString(config.UserLength, config.CharType)
		} else if config.UserType == model.UserTypeVC {
			// Voucher Code mode
			if config.CharType == model.CharTypeNum {
				code := generateRandomString(actualRandomLength, model.CharTypeNum)
				username = config.Prefix + code
				password = username
			} else {
				codeLength := actualRandomLength
				numLength := 2

				// Adjust lengths
				if actualRandomLength >= 6 && actualRandomLength <= 7 {
					codeLength = actualRandomLength - 3
					numLength = 3
				} else if actualRandomLength >= 8 {
					codeLength = actualRandomLength - 4
					numLength = 4
				} else if actualRandomLength >= 4 && actualRandomLength <= 5 {
					codeLength = actualRandomLength - 2
					numLength = 2
				} else if actualRandomLength < 4 {
					codeLength = actualRandomLength
					numLength = 0
				}

				if (config.CharType == model.CharTypeLower1 ||
					config.CharType == model.CharTypeUpper1 ||
					config.CharType == model.CharTypeUppLow1) && numLength > 0 {
					// Remove "1" suffix from char type
					baseCharType := model.CharType(strings.TrimSuffix(string(config.CharType), "1"))
					charPart := generateRandomString(codeLength, baseCharType)
					numPart := generateRandomString(numLength, model.CharTypeNum)
					username = config.Prefix + charPart + numPart
					password = username
				} else {
					username = config.Prefix + generateRandomString(actualRandomLength, config.CharType)
					password = username
				}
			}
		}

		generatedUsers = append(generatedUsers, model.GeneratedVoucher{
			Username: username,
			Password: password,
		})

		// Add user to MikroTik
		args := []string{
			"=name=" + username,
			"=password=" + password,
			"=profile=" + config.Profile,
			"=comment=" + fullComment,
		}

		if config.Server != "" {
			args = append(args, "=server="+config.Server)
		}
		if config.TimeLimit != "" {
			args = append(args, "=limit-uptime="+config.TimeLimit)
		}
		if dataLimitBytes > 0 {
			args = append(args, "=limit-bytes-total="+strconv.FormatInt(dataLimitBytes, 10))
		}

		sentence := append([]string{"/ip/hotspot/user/add"}, args...)
		_, err := s.client.RunArgs(sentence)
		if err != nil {
			return &model.VoucherResponse{
				Message: "error",
				Data: model.VoucherResponseData{
					Error: err.Error(),
				},
			}, err
		}
	}

	// Get created users by comment
	reply, err := s.client.Run("/ip/hotspot/user/print", "?comment="+fullComment)
	if err == nil && len(reply.Re) > 0 {
		users := make([]model.UserData, len(reply.Re))
		for i, re := range reply.Re {
			users[i] = model.UserData{
				ID:       re.Map[".id"],
				Name:     re.Map["name"],
				Password: re.Map["password"],
				Profile:  re.Map["profile"],
				Comment:  re.Map["comment"],
			}
		}

		return &model.VoucherResponse{
			Message: "success",
			Data: model.VoucherResponseData{
				Count:   len(users),
				Comment: fullComment,
				Profile: config.Profile,
				Users:   users,
			},
		}, nil
	}

	return &model.VoucherResponse{
		Message: "success",
		Data: model.VoucherResponseData{
			Count:   config.Qty,
			Comment: fullComment,
			Profile: config.Profile,
		},
	}, nil
}
