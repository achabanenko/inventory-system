# Mobile App Information for CLAUDE.md

## Mobile App Architecture & Features

### Mobile App (Flutter)
- **Framework**: Flutter 3.7.2+
- **State Management**: Provider pattern with ChangeNotifier
- **HTTP Client**: HTTP package with Dio for enhanced features
- **Authentication**: JWT tokens stored in SharedPreferences with auto-refresh
- **Barcode Scanning**: flutter_barcode_scanner & qr_code_scanner packages
- **Material Design**: Material 3 with custom theme (blue primary, green secondary)

### Mobile App Structure
```
mobile_app/             # Flutter mobile app
   lib/
      main.dart        # App entry point with MultiProvider setup
      models/          # Data models
         user.dart     # User model with JSON serialization
         item.dart     # Item model with price parsing
      screens/         # UI screens
         login_screen.dart      # Login with email/password validation
         home_screen.dart       # Bottom navigation with 4 tabs
         items_screen.dart      # Item search/list with details sheet
         barcode_scanner_screen.dart # Camera barcode scanning
         chat_screen.dart       # AI chat with message history
         settings_screen.dart   # User profile and app preferences
      services/        # Business logic services
         auth_service.dart      # JWT auth with token refresh
         api_service.dart       # REST API client with error handling
      theme/           # App styling
         app_theme.dart # Light/dark theme with Material 3
   android/            # Android platform files
   ios/                # iOS platform files  
   pubspec.yaml        # Flutter dependencies
```

## Mobile App Features Implemented

### Authentication & Security
- JWT authentication with access/refresh token pattern
- Automatic token refresh on 401 responses  
- Secure token storage using SharedPreferences
- Login form with email validation and password visibility toggle
- User session persistence across app restarts

### Core Functionality
- **Item Management**: Search items by SKU, name, or barcode with detailed view
- **Barcode Scanner**: Camera-based barcode scanning with item lookup
- **AI Chat**: OpenAI-compatible chat interface for inventory queries
- **Settings**: User profile, server configuration, theme selection, and preferences

### User Interface
- Modern Material 3 design with custom blue/green color scheme
- Bottom navigation with 4 main sections (Items, Scanner, AI Chat, Settings)
- Responsive layouts with proper loading states and error handling
- Light/dark theme support with system preference detection

### API Integration
- Full integration with inventory backend API
- Support for all item CRUD operations
- Inventory level queries by item ID
- AI chat endpoint integration (placeholder)
- Automatic retry logic for failed requests

### Key Dependencies
```yaml
dependencies:
  flutter: sdk: flutter
  cupertino_icons: ^1.0.8
  http: ^1.5.0                    # HTTP client
  provider: ^6.1.5+1              # State management
  shared_preferences: ^2.5.3      # Local storage
  qr_code_scanner: ^1.0.1         # QR scanning
  dio: ^5.9.0                     # Enhanced HTTP client
  flutter_barcode_scanner: ^2.0.0  # Barcode scanning
```

## Mobile App Current Status

### ✅ Implemented
- Complete authentication flow with JWT
- Item search and display functionality
- Barcode scanning with camera integration
- AI chat interface (UI ready, backend integration pending)
- Settings management with preferences persistence
- Responsive Material 3 UI design
- Error handling and loading states

### 🚧 In Progress / Placeholder
- AI chat backend integration (API endpoint exists but needs implementation)
- Add/edit item functionality (UI shows "coming soon" message)
- Advanced inventory operations (stock levels, transfers, adjustments)
- Push notifications for inventory alerts
- Offline mode with local caching

### Mobile-Specific Features
- Cross-platform support (iOS/Android)
- Camera access for barcode scanning
- Local data persistence
- Mobile-optimized navigation patterns
- Touch-friendly UI components
- Device-specific styling (iOS/Android)

The mobile app provides a complete inventory management experience optimized for mobile devices, with emphasis on barcode scanning and AI-powered assistance for field operations.