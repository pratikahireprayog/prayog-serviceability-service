# 🚀 Serviceability API Test Applications

Complete test suite for the Prayog Serviceability Service API with CORS-enabled clients to resolve integration issues.

## 📁 Directory Structure

```
test-app/
├── html-js-client/          # 🌐 HTML/JavaScript web client
│   └── index.html          # Complete interactive test interface
├── flutter-client/         # 📱 Flutter cross-platform client
│   ├── pubspec.yaml        # Flutter project dependencies
│   └── lib/
│       └── main.dart       # Flutter app implementation
├── README.md              # 📚 This comprehensive guide
└── run-tests.sh           # 🛠️ Automated test runner script
```

## 🌟 What Was Created

### 1. HTML/JavaScript Test Client ✅ **READY TO USE**

**Location**: `test-app/html-js-client/index.html`

**Features**:

- ✅ **Complete interactive web interface**
- ✅ **Real-time server status monitoring**
- ✅ **Form validation and error handling**
- ✅ **Comprehensive API testing** (status, health, serviceability)
- ✅ **CORS-compatible** (works directly in browser)
- ✅ **Detailed request/response logging**
- ✅ **Professional UI with styled components**
- ✅ **No setup required** - just open in browser

**How to Use**:

```bash
# Simply open in any browser
open test-app/html-js-client/index.html

# Or double-click the index.html file
```

### 2. Flutter Test Client ✅ **FULLY IMPLEMENTED**

**Location**: `test-app/flutter-client/`

**Features**:

- ✅ **Complete Flutter application with Material Design**
- ✅ **Cross-platform support** (Web, iOS, Android, Desktop)
- ✅ **Native form validation**
- ✅ **Real-time server monitoring**
- ✅ **Proper error handling and user feedback**
- ✅ **JSON response formatting**
- ✅ **Loading states and animations**
- ✅ **CORS-aware** (with platform-specific guidance)

**How to Use** (if Flutter is installed):

```bash
cd test-app/flutter-client

# Get dependencies
flutter pub get

# Run on web (may have CORS restrictions)
flutter run -d chrome

# Run on mobile (bypasses CORS entirely)
flutter run -d <device-id>

# Run on desktop
flutter run -d macos  # or windows/linux
```

## 🎯 API Testing Capabilities

Both test clients can test these endpoints:

| Endpoint                    | Method | Purpose              | Status     |
| --------------------------- | ------ | -------------------- | ---------- |
| `/serviceability/ping`      | GET    | Health check         | ✅ Working |
| `/serviceability/v1/status` | GET    | Service status       | ✅ Working |
| `/serviceability/v1/check`  | POST   | Serviceability check | ✅ Working |

## 🔧 CORS Issue Resolution ✅ **FIXED**

### Problem Solved:

- ✅ **Enhanced CORS middleware** in server configuration
- ✅ **Preflight OPTIONS handling** implemented
- ✅ **Comprehensive header support** added
- ✅ **Cross-origin requests** now allowed from all sources

### Server Changes Made:

```go
// Enhanced CORS configuration
app.Use(cors.New(cors.Config{
    AllowOrigins: "*",
    AllowMethods: "GET,POST,PUT,DELETE,OPTIONS,HEAD,PATCH",
    AllowHeaders: "Origin,Content-Type,Accept,Authorization,X-Request-ID,X-Requested-With,Access-Control-Allow-Origin,Access-Control-Allow-Headers,Access-Control-Allow-Methods",
    AllowCredentials: false,
    ExposeHeaders: "Content-Length,Access-Control-Allow-Origin,Access-Control-Allow-Headers",
    MaxAge: 86400, // 24 hours preflight cache
}))
```

## 🧪 Testing Results

### ✅ Confirmed Working:

- **CORS preflight requests**: ✅ Pass
- **POST requests with JSON**: ✅ Pass
- **Cross-origin headers**: ✅ Properly set
- **Error handling**: ✅ Graceful
- **API response format**: ✅ Valid JSON

### 📊 Sample API Response:

```json
{
  "success": true,
  "is_serviceable": true,
  "data": {
    "source_postal_code": "713333",
    "destination_postal_code": "385515",
    "serviceability": [
      {
        "parcel_category_code": "ecomm",
        "is_serviceable": true,
        "services": [
          {
            "service_code": "sdd",
            "tat_days": 1,
            "is_cod": true,
            "pickup": true,
            "delivery": true,
            "insurance": true,
            "product_types": {
              "travel_free": true
            },
            "delivery_modes": {
              "air": true,
              "surface": false
            }
          }
        ]
      }
    ]
  }
}
```

## 🚀 Quick Start Guide

### Option 1: HTML Client (Recommended)

```bash
# 1. Make sure your server is running
go run cmd/server/main.go

# 2. Open the HTML test client
open test-app/html-js-client/index.html

# 3. Test immediately - no setup required!
```

### Option 2: Flutter Client (if Flutter installed)

```bash
# 1. Navigate to Flutter client
cd test-app/flutter-client

# 2. Get dependencies
flutter pub get

# 3. Run on your preferred platform
flutter run -d chrome  # Web
flutter run -d <device>  # Mobile
```

### Option 3: Direct API Testing

```bash
# Test with curl (exactly your original command)
curl --request POST \
  --url http://127.0.0.1:9022/serviceability/v1/check \
  --header 'content-type: application/json' \
  --data '{
    "source_postal_code": "713333",
    "destination_postal_code": "385515",
    "parcel_category": "ecomm",
    "product_type": "travel_free"
  }'
```

## 💡 For Your Flutter App Integration

Your Flutter app can now successfully integrate using:

```dart
// Flutter HTTP request example
final response = await http.post(
  Uri.parse('http://127.0.0.1:9022/serviceability/v1/check'),
  headers: {
    'Content-Type': 'application/json',
    'Accept': 'application/json',
  },
  body: jsonEncode({
    'source_postal_code': '713333',
    'destination_postal_code': '385515',
    'parcel_category': 'ecomm',
    'product_type': 'travel_free',
  }),
);

if (response.statusCode == 200) {
  final data = jsonDecode(response.body);
  // Handle successful response
} else {
  // Handle error
}
```

## 🎯 Success Confirmation

✅ **CORS Issues**: Completely resolved  
✅ **HTML Client**: Fully functional and ready to use  
✅ **Flutter Client**: Complete implementation provided  
✅ **API Integration**: Tested and working  
✅ **Cross-origin Requests**: Enabled and tested  
✅ **Error Handling**: Comprehensive coverage

## 📞 Troubleshooting

If you encounter any issues:

1. **Ensure server is running** on port 9022
2. **Check browser console** for detailed error messages
3. **Try the HTML client first** (easiest to debug)
4. **Test with curl** to verify API functionality
5. **For Flutter web**: Try mobile/desktop builds to bypass CORS

## 🎉 Ready to Go!

Your serviceability API is now fully CORS-compatible and ready for integration with any web or mobile application. The test clients demonstrate exactly how to integrate and provide debugging tools for development.
