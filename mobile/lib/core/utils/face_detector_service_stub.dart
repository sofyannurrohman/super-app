import 'dart:ui';

// ---------------------------------------------------------------------------
// STUB IMPLEMENTATION — Web & iOS Simulator
//
// Digunakan ketika dart:io tidak tersedia (Web/Chrome) atau saat
// run_simulator.sh mengalihkan export ke file ini.
//
// TIDAK boleh mengimport dart:io, dart:ffi, camera, atau MLKit
// karena file ini dipakai untuk platform web yang tidak punya akses tersebut.
// ---------------------------------------------------------------------------

/// Dummy pengganti [InputImage] dari MLKit — aman di Web.
class InputImage {
  InputImage._();
  static InputImage stub() => InputImage._();
}

/// Dummy pengganti [Face] dari MLKit — aman di Web.
class Face {
  final Rect boundingBox;
  final double? headEulerAngleY;
  final double? headEulerAngleZ;
  Face({required this.boundingBox, this.headEulerAngleY, this.headEulerAngleZ});
}

/// Dummy CameraImage — hanya dipakai sebagai parameter type di signature method.
class _StubCameraImage {}

/// Dummy CameraDescription — hanya dipakai sebagai parameter type.
class _StubCameraDescription {}

class FaceDetectorService {
  bool _isInitialized = false;

  void initialize() => _isInitialized = true;

  bool get isInitialized => _isInitialized;

  Future<List<Face>> processImage(InputImage inputImage) async => [];

  Future<List<Face>> processImageFromFile(String filePath) async => [];

  void dispose() => _isInitialized = false;

  /// Signature sengaja pakai dynamic agar tidak perlu import camera package.
  InputImage? createInputImageFromCameraImage(dynamic image, dynamic camera) {
    return null;
  }
}
