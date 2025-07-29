package config

import (
	"testing"

	"github.com/go-telegram/bot/models"
)

func TestIsWebAppLinkEnabledForPlatform(t *testing.T) {
	// Сохраняем оригинальное значение
	originalValue := conf.isWebAppLinkEnabled
	defer func() {
		conf.isWebAppLinkEnabled = originalValue
	}()

	tests := []struct {
		name           string
		globalEnabled  bool
		platform       string
		expectedResult bool
	}{
		{
			name:           "Global disabled, any platform",
			globalEnabled:  false,
			platform:       "mobile",
			expectedResult: false,
		},
		{
			name:           "Global enabled, mobile platform",
			globalEnabled:  true,
			platform:       "mobile",
			expectedResult: true,
		},
		{
			name:           "Global enabled, desktop platform",
			globalEnabled:  true,
			platform:       "desktop",
			expectedResult: false,
		},
		{
			name:           "Global enabled, unknown platform",
			globalEnabled:  true,
			platform:       "unknown",
			expectedResult: true, // использует глобальную настройку
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			conf.isWebAppLinkEnabled = tt.globalEnabled
			result := IsWebAppLinkEnabledForPlatform(tt.platform)
			if result != tt.expectedResult {
				t.Errorf("IsWebAppLinkEnabledForPlatform(%s) = %v, want %v", tt.platform, result, tt.expectedResult)
			}
		})
	}
}

func TestDetectPlatformFromUpdate(t *testing.T) {
	tests := []struct {
		name     string
		update   *models.Update
		expected string
	}{
		{
			name: "Message with WebAppData",
			update: &models.Update{
				Message: &models.Message{
					WebAppData: &models.WebAppData{
						Data: "test",
					},
				},
			},
			expected: "mobile",
		},
		{
			name: "Callback with WebApp button",
			update: &models.Update{
				CallbackQuery: &models.CallbackQuery{
					Message: &models.Message{
						ReplyMarkup: &models.InlineKeyboardMarkup{
							InlineKeyboard: [][]models.InlineKeyboardButton{
								{
									{
										Text: "Test",
										WebApp: &models.WebAppInfo{
											URL: "https://example.com",
										},
									},
								},
							},
						},
					},
				},
			},
			expected: "mobile",
		},
		{
			name: "Regular message without WebApp",
			update: &models.Update{
				Message: &models.Message{
					Text: "Hello",
				},
			},
			expected: "desktop",
		},
		{
			name: "Callback without WebApp",
			update: &models.Update{
				CallbackQuery: &models.CallbackQuery{
					Message: &models.Message{
						ReplyMarkup: &models.InlineKeyboardMarkup{
							InlineKeyboard: [][]models.InlineKeyboardButton{
								{
									{
										Text:        "Test",
										CallbackData: "test",
									},
								},
							},
						},
					},
				},
			},
			expected: "desktop",
		},
		{
			name:     "Nil update",
			update:   nil,
			expected: "desktop",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := DetectPlatformFromUpdate(tt.update)
			if result != tt.expected {
				t.Errorf("DetectPlatformFromUpdate() = %v, want %v", result, tt.expected)
			}
		})
	}
} 