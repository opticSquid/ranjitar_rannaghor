package meals

import "github.com/google/uuid"

func utilsGenerateUUID() string {
	return uuid.New().String()
}
