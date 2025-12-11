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
	// Сохраняем оригинальное значение
	originalValue := conf.isWebAppLinkEnabled
	defer func() {
		conf.isWebAppLinkEnabled = originalValue
	}()

	tests := []struct {
		name           string
		update         *models.Update
		globalEnabled  bool
		expected       string
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
			globalEnabled: false,
			expected:      "mobile",
		},
		{
			name: "Callback with WebApp button",
			update: &models.Update{
				CallbackQuery: &models.CallbackQuery{
					Message: models.MaybeInaccessibleMessage{
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
			},
			globalEnabled: false,
			expected:      "mobile",
		},
		{
			name: "Regular message without WebApp, global disabled",
			update: &models.Update{
				Message: &models.Message{
					Text: "Hello",
				},
			},
			globalEnabled: false,
			expected:      "desktop",
		},
		{
			name: "Message via bot (Web App), global disabled",
			update: &models.Update{
				Message: &models.Message{
					Text: "Hello",
					ViaBot: &models.User{
						ID: 123456789,
					},
				},
			},
			globalEnabled: false,
			expected:      "desktop", // При отключенной глобальной настройке возвращаем desktop
		},
		{
			name: "Regular message without WebApp, global enabled",
			update: &models.Update{
				Message: &models.Message{
					Text: "Hello",
				},
			},
			globalEnabled: true,
			expected:      "mobile", // Когда глобальная настройка включена, предполагаем мобильное устройство
		},
		{
			name: "Message via bot (Web App), global enabled",
			update: &models.Update{
				Message: &models.Message{
					Text: "Hello",
					ViaBot: &models.User{
						ID: 123456789,
					},
				},
			},
			globalEnabled: true,
			expected:      "mobile", // Сообщение через бота указывает на Web App
		},
		{
			name: "Callback without WebApp, global disabled",
			update: &models.Update{
				CallbackQuery: &models.CallbackQuery{
					Message: models.MaybeInaccessibleMessage{
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
			},
			globalEnabled: false,
			expected:      "desktop",
		},
		{
			name: "Callback without WebApp, global enabled",
			update: &models.Update{
				CallbackQuery: &models.CallbackQuery{
					Message: models.MaybeInaccessibleMessage{
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
			},
			globalEnabled: true,
			expected:      "mobile", // Когда глобальная настройка включена, предполагаем мобильное устройство
		},
		{
			name:          "Nil update, global disabled",
			update:        nil,
			globalEnabled: false,
			expected:      "unknown", // nil update возвращает unknown
		},
		{
			name:          "Nil update, global enabled",
			update:        nil,
			globalEnabled: true,
			expected:      "unknown", // nil update возвращает unknown
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			conf.isWebAppLinkEnabled = tt.globalEnabled
			result := DetectPlatformFromUpdate(tt.update)
			if result != tt.expected {
				t.Errorf("DetectPlatformFromUpdate() = %v, want %v (global enabled: %v)", result, tt.expected, tt.globalEnabled)
			}
		})
	}
} 