package wordbomb

func parseUserIDMap(data []byte) map[string]string {
	result := make(map[string]string)
	for i := 0; i < len(data); i++ {
		if data[i] == 0xA9 && i+9 < len(data) {
			userID := string(data[i+1 : i+10])
			for j := i + 10; j < len(data); j++ {
				if data[j] >= 0xA0 && data[j] <= 0xBF {
					strLen := int(data[j] & 0x1F)
					if j+1+strLen <= len(data) {
						username := string(data[j+1 : j+1+strLen])
						result[userID] = username
						i = j + strLen
						break
					}
				}
			}
		}
	}
	return result
}
