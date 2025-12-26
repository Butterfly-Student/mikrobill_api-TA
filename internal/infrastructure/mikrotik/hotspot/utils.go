package hotspot

import (
	"fmt"
	"math/rand"
	"mikrobill/internal/infrastructure/mikrotik/model"
	"regexp"
	"strconv"
	"strings"
	"time"
)

// sanitizeName membersihkan nama dari karakter tidak valid
func sanitizeName(name string) string {
	// Ganti spasi dengan dash
	name = strings.ReplaceAll(name, " ", "-")

	// Hanya izinkan alphanumeric, dash, dan underscore
	reg := regexp.MustCompile(`[^a-zA-Z0-9\-_]`)
	return reg.ReplaceAllString(name, "")
}

// generateRandomString menghasilkan string random
func generateRandomString(length int, charType model.CharType) string {
	const (
		lowerChars = "abcdefghijklmnopqrstuvwxyz"
		upperChars = "ABCDEFGHIJKLMNOPQRSTUVWXYZ"
		numbers    = "0123456789"
	)

	var chars string
	switch charType {
	case model.CharTypeLower, model.CharTypeLower1:
		chars = lowerChars
	case model.CharTypeUpper, model.CharTypeUpper1:
		chars = upperChars
	case model.CharTypeUppLow, model.CharTypeUppLow1:
		chars = lowerChars + upperChars
	case model.CharTypeMix:
		chars = lowerChars + numbers
	case model.CharTypeMix1:
		chars = upperChars + numbers
	case model.CharTypeMix2:
		chars = lowerChars + upperChars + numbers
	case model.CharTypeNum:
		chars = numbers
	default:
		chars = lowerChars + upperChars + numbers
	}

	rand.Seed(time.Now().UnixNano())
	result := make([]byte, length)
	for i := range result {
		result[i] = chars[rand.Intn(len(chars))]
	}
	return string(result)
}

// calculateDataLimit menghitung data limit dalam bytes
func calculateDataLimit(dataLimit string) (int64, error) {
	if dataLimit == "" {
		return 0, nil
	}

	lastChar := strings.ToLower(dataLimit[len(dataLimit)-1:])
	value := dataLimit[:len(dataLimit)-1]

	if lastChar != "m" && lastChar != "g" {
		// Jika tidak ada unit, coba parse sebagai number
		num, err := strconv.ParseInt(dataLimit, 10, 64)
		if err != nil {
			return 0, nil
		}
		return num, nil
	}

	numValue, err := strconv.ParseFloat(value, 64)
	if err != nil {
		return 0, nil
	}

	var multiplier int64
	if lastChar == "g" {
		multiplier = 1073741824 // GB in bytes
	} else {
		multiplier = 1048576 // MB in bytes
	}

	return int64(numValue * float64(multiplier)), nil
}

// generateOnLoginScript menghasilkan script on-login untuk profile
func generateOnLoginScript(config model.ProfileRequest) string {
	expMode := config.ExpMode
	price := config.Price
	if price == "" {
		price = "0"
	}
	validity := config.Validity
	if validity == "" {
		validity = ""
	}
	sellingPrice := config.SellingPrice
	if sellingPrice == "" {
		sellingPrice = "0"
	}
	lockUser := config.LockUser
	lockServer := config.LockServer
	name := config.Name

	var mode string
	if expMode == model.ExpireModeNTF || expMode == model.ExpireModeNTFC {
		mode = "N"
	} else if expMode == model.ExpireModeREM || expMode == model.ExpireModeREMC {
		mode = "X"
	}

	var lockScript string
	if lockUser == model.LockEnable {
		lockScript = "; [:local mac $\"mac-address\"; /ip hotspot user set mac-address=$mac [find where name=$user]]"
	}

	var serverLockScript string
	if lockServer == model.LockEnable {
		serverLockScript = "; [:local mac $\"mac-address\"; :local srv [/ip hotspot host get [find where mac-address=\"$mac\"] server]; /ip hotspot user set server=$srv [find where name=$user]]"
	}

	recordScript := fmt.Sprintf("; :local mac $\"mac-address\"; :local time [/system clock get time ]; /system script add name=\"$date-|-$time-|-$user-|-%s-|-$address-|-$mac-|-%s-|-%s-|-$comment\" owner=\"$month$year\" source=$date comment=mikhmon",
		price, validity, name)

	var onLoginScript string

	baseExpScript := fmt.Sprintf("{:local date [ /system clock get date ];:local year [ :pick $date 7 11 ];:local month [ :pick $date 0 3 ];:local comment [ /ip hotspot user get [/ip hotspot user find where name=\"$user\"] comment]; :local ucode [:pic $comment 0 2]; :if ($ucode = \"vc\" or $ucode = \"up\" or $comment = \"\") do={ /sys sch add name=\"$user\" disable=no start-date=$date interval=\"%s\"; :delay 2s; :local exp [ /sys sch get [ /sys sch find where name=\"$user\" ] next-run]; :local getxp [len $exp]; :if ($getxp = 15) do={ :local d [:pic $exp 0 6]; :local t [:pic $exp 7 16]; :local s (\"/\"); :local exp (\"$d$s$year $t\"); /ip hotspot user set comment=\"$exp %s\" [find where name=\"$user\"];}; :if ($getxp = 8) do={ /ip hotspot user set comment=\"$date $exp %s\" [find where name=\"$user\"];}; :if ($getxp > 15) do={ /ip hotspot user set comment=\"$exp %s\" [find where name=\"$user\"];}; /sys sch remove [find where name=\"$user\"]%s%s}}",
		validity, mode, mode, mode, lockScript, serverLockScript)

	switch expMode {
	case model.ExpireModeREM:
		onLoginScript = fmt.Sprintf(":put (\",%s,%s,%s,%s,,%s,%s,\"); :local mode \"%s\"; %s",
			expMode, price, validity, sellingPrice, lockUser, lockServer, mode, baseExpScript)
	case model.ExpireModeNTF:
		onLoginScript = fmt.Sprintf(":put (\",%s,%s,%s,%s,,%s,%s,\"); :local mode \"%s\"; %s",
			expMode, price, validity, sellingPrice, lockUser, lockServer, mode, baseExpScript)
	case model.ExpireModeREMC:
		scriptWithRecord := strings.Replace(baseExpScript, lockScript, recordScript+lockScript, 1)
		onLoginScript = fmt.Sprintf(":put (\",%s,%s,%s,%s,,%s,%s,\"); :local mode \"%s\"; %s",
			expMode, price, validity, sellingPrice, lockUser, lockServer, mode, scriptWithRecord)
	case model.ExpireModeNTFC:
		scriptWithRecord := strings.Replace(baseExpScript, lockScript, recordScript+lockScript, 1)
		onLoginScript = fmt.Sprintf(":put (\",%s,%s,%s,%s,,%s,%s,\"); :local mode \"%s\"; %s",
			expMode, price, validity, sellingPrice, lockUser, lockServer, mode, scriptWithRecord)
	case model.ExpireModeNone:
		if price != "" && price != "0" {
			onLoginScript = fmt.Sprintf(":put (\",%s,,%s,,noexp,%s,%s,\")%s%s",
				price, sellingPrice, lockUser, lockServer, lockScript, serverLockScript)
		}
	}

	return onLoginScript
}

// validateProfileConfig memvalidasi konfigurasi profile
func validateProfileConfig(config model.ProfileRequest) error {
	if strings.TrimSpace(config.Name) == "" {
		return fmt.Errorf("profile name is required")
	}

	if config.SharedUsers != nil {
		if *config.SharedUsers < 1 || *config.SharedUsers > 100 {
			return fmt.Errorf("shared users must be between 1 and 100")
		}
	}

	if config.Price != "" {
		price, err := strconv.ParseFloat(config.Price, 64)
		if err == nil && price < 0 {
			return fmt.Errorf("price cannot be negative")
		}
	}

	if config.SellingPrice != "" {
		sellingPrice, err := strconv.ParseFloat(config.SellingPrice, 64)
		if err == nil && sellingPrice < 0 {
			return fmt.Errorf("selling price cannot be negative")
		}
	}

	return nil
}

// boolToYesNo converts bool to "yes" or "no"
func boolToYesNo(b bool) string {
	if b {
		return "yes"
	}
	return "no"
}

// yesNoToBool converts "yes" or "no" to bool
func yesNoToBool(s string) bool {
	return strings.ToLower(s) == "yes" || s == "true"
}
