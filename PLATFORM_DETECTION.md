# Platform Detection Feature

## Overview

The bot now automatically detects the user's platform (mobile vs desktop) and adjusts the subscription link display accordingly.

## How It Works

### Platform Detection Logic

The bot uses several methods to determine if a user is on a mobile device:

1. **Web App Data**: If the user interacts with a Web App, they are automatically detected as mobile
2. **Web App Buttons**: If the user clicks on a button with Web App functionality, they are detected as mobile
3. **Default Behavior**: If no mobile indicators are found, the user is assumed to be on desktop

### Configuration

Set the `IS_WEB_APP_LINK` environment variable to control this feature:

```bash
# Enable platform-dependent behavior
IS_WEB_APP_LINK=true

# Disable platform-dependent behavior (default)
IS_WEB_APP_LINK=false
```

## Behavior by Platform

### Mobile Devices
- Subscription links are displayed as **Web App buttons**
- Links open directly within Telegram
- No text links are shown in the message body
- Optimal for mobile VPN applications

### Desktop Clients
- Subscription links are displayed as **regular text links**
- Links can be copied and used in external applications
- Compatible with desktop VPN clients
- Full link visibility for manual configuration

## Code Implementation

### New Functions

```go
// Detect platform from Telegram update
platform := config.DetectPlatformFromUpdate(update)

// Check if Web App should be enabled for this platform
if config.IsWebAppLinkEnabledForPlatform(platform) {
    // Show Web App button
} else {
    // Show regular link
}
```

### Example Usage

```go
func (h Handler) ConnectCallbackHandler(ctx context.Context, b *bot.Bot, update *models.Update) {
    // ... existing code ...
    
    platform := config.DetectPlatformFromUpdate(update)
    if config.IsWebAppLinkEnabledForPlatform(platform) {
        // Create Web App button for mobile
        markup = append(markup, []models.InlineKeyboardButton{{
            Text: "Connect",
            WebApp: &models.WebAppInfo{URL: subscriptionLink},
        }})
    }
    
    // ... rest of the code ...
}
```

## Benefits

1. **Better UX**: Mobile users get native Web App experience
2. **Desktop Compatibility**: Desktop users can copy links for external apps
3. **Automatic Detection**: No manual configuration needed
4. **Backward Compatibility**: Existing functionality preserved

## Migration

Existing installations will continue to work as before. The new platform detection is opt-in through the `IS_WEB_APP_LINK` environment variable.

To enable the new behavior:

1. Set `IS_WEB_APP_LINK=true` in your environment
2. Restart the bot
3. The bot will automatically detect platforms and adjust behavior

## Technical Details

### Detection Methods

- **WebAppData**: Detects when users interact with Web Apps
- **InlineKeyboard Analysis**: Checks for Web App buttons in message markup
- **Fallback**: Defaults to desktop if no mobile indicators found

### Performance

- Detection is lightweight and happens on each message/callback
- No additional API calls required
- Minimal memory overhead

### Error Handling

- Graceful fallback to desktop behavior if detection fails
- Maintains existing functionality even if detection logic changes
- No breaking changes to existing API 