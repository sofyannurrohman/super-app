class AppConstants {
  // Use localhost for Physical Device (run: adb reverse tcp:8081 tcp:8081)
  // Or use 10.0.2.2 for Android Emulator
  // Or use your specific LAN IP (e.g., 192.168.10.69) for Wi-Fi debugging
  // Use environment variable for Base URL, default to localhost for local development
  static const String baseUrl = String.fromEnvironment(
    'BASE_URL',
    defaultValue: 'http://127.0.0.1:13001/api/v1/',
  );

  static const String companyName = String.fromEnvironment(
    'COMPANY_NAME',
    defaultValue: 'Perusahaan',
  );

  static const String tokenKey = 'access_token';
  static const String refreshTokenKey = 'refresh_token';

  // Face Verification Threshold (Lower = stricter)
  // Standard range for MobileFaceNet is 0.8 - 1.2
  static const double faceVerificationThreshold = 1.7;

  static const String rememberMeKey = 'remember_me';
  static const String emailKey = 'remembered_email';
  static const String passwordKey = 'remembered_password';

  /// Builds a full face image URL with a cache-busting version parameter.
  /// The version is extracted from the filename (which includes a Unix timestamp
  /// like `<employeeID>_<timestamp>.jpg`), ensuring Flutter's image cache
  /// is bypassed whenever a new face photo is registered.
  static String buildFaceImageUrl(String faceReferenceUrl) {
    final origin = Uri.parse(baseUrl).origin;
    final normalizedPath = faceReferenceUrl.startsWith('/')
        ? faceReferenceUrl
        : '/$faceReferenceUrl';
    final fullUrl = faceReferenceUrl.startsWith('http')
        ? faceReferenceUrl
        : '$origin$normalizedPath';

    // Extract Unix timestamp embedded in the filename (e.g., abc_1234567890.jpg)
    // and use it as a cache-busting version query parameter.
    final filenamePart = fullUrl.split('/').last.split('.').first;
    final parts = filenamePart.split('_');
    final version = parts.length >= 2
        ? parts.last
        : DateTime.now().millisecondsSinceEpoch.toString();

    return '$fullUrl?v=$version';
  }

  /// Builds a full image URL for general images (stores, products, banners, etc.)
  /// Resolves local uploads and absolute URLs.
  static String buildImageUrl(String path) {
    if (path.isEmpty) return '';

    String url = path.trim();

    final origin = Uri.parse(baseUrl).origin;

    // 2. Local device absolute file paths (Android/iOS internal storage)
    if (url.startsWith('/data/') ||
        url.startsWith('/var/') ||
        url.startsWith('/storage/') ||
        url.startsWith('/private/')) {
      return url;
    }

    // 3. Server uploads (/uploads/..., /api/v1/uploads/..., etc.)
    if (url.startsWith('/uploads/') ||
        url.startsWith('uploads/') ||
        url.startsWith('/api/v1/uploads/') ||
        url.startsWith('api/v1/uploads/')) {
      final cleanLocalPath = url.startsWith('/') ? url : '/$url';
      return '$origin$cleanLocalPath';
    }

    // 4. Complete HTTP/HTTPS URLs
    String fullUrl;
    if (url.startsWith('http://') || url.startsWith('https://')) {
      fullUrl = url;
    } else {
      // Relative path - if it's a legacy branding path (e.g. "foto-toko/...", "shops/...", "banners/...")
      final stripped = url.replaceFirst(RegExp(r'^\/+'), '');
      if (stripped.startsWith('foto-toko/') ||
          stripped.startsWith('shops/') ||
          stripped.startsWith('banners/')) {
        fullUrl = '$origin/uploads/$stripped';
      } else {
        final cleanLocalPath = url.startsWith('/') ? url : '/$url';
        fullUrl = '$origin$cleanLocalPath';
      }
    }

    return fullUrl;
  }
}
