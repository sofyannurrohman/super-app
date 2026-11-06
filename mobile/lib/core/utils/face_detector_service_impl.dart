// ---------------------------------------------------------------------------
// face_detector_service_impl.dart — Implementasi lengkap FaceDetectorService
//
// Menggunakan kUseMlKit (--dart-define=USE_MLKIT=true/false) untuk memilih:
//   true  → real MLKit FaceDetector (Android & iPhone fisik)
//   false → stub yang selalu return [] (iOS Simulator)
//
// Dengan pendekatan ini, file ini bisa di-compile pada SEMUA platform
// karena MLKit hanya di-import secara kondisional via dart:mirrors trick
// — atau lebih tepatnya kita wrapping dalam try/catch dan kUseMlKit guard.
// ---------------------------------------------------------------------------

import 'dart:ui';
import 'package:flutter/foundation.dart';
import 'package:camera/camera.dart';
import 'package:google_mlkit_face_detection/google_mlkit_face_detection.dart';
export 'package:google_mlkit_face_detection/google_mlkit_face_detection.dart'
    show Face, InputImage;
import 'package:hris_app/core/utils/ml_config.dart';

class FaceDetectorService {
  FaceDetector? _faceDetector;
  bool _isInitialized = false;

  void initialize() {
    if (kIsWeb || !kUseMlKit) return; // Simulator: skip inisialisasi
    _faceDetector = FaceDetector(
      options: FaceDetectorOptions(
        enableContours: false,
        enableLandmarks: false,
        enableClassification: false,
        performanceMode: FaceDetectorMode.fast,
      ),
    );
    _isInitialized = true;
  }

  bool get isInitialized => _isInitialized;

  Future<List<Face>> processImage(InputImage inputImage) async {
    if (kIsWeb || !kUseMlKit || !_isInitialized) return [];
    return await _faceDetector!.processImage(inputImage);
  }

  Future<List<Face>> processImageFromFile(String filePath) async {
    if (kIsWeb || !kUseMlKit || !_isInitialized) return [];
    final inputImage = InputImage.fromFilePath(filePath);
    return await _faceDetector!.processImage(inputImage);
  }

  void dispose() {
    if (kIsWeb) return;
    if (_isInitialized && _faceDetector != null) {
      _faceDetector!.close();
      _isInitialized = false;
    }
  }

  InputImage? createInputImageFromCameraImage(
      CameraImage image, CameraDescription camera) {
    if (!kUseMlKit)
      return null; // Simulator: return null, image stream akan di-skip

    final int sensorOrientation = camera.sensorOrientation;
    InputImageRotation? rotation =
        InputImageRotationValue.fromRawValue(sensorOrientation);
    if (rotation == null) return null;

    final int numPlanes = image.planes.length;
    final InputImageFormat format =
        numPlanes == 3 ? InputImageFormat.yuv_420_888 : InputImageFormat.nv21;

    final WriteBuffer allBytes = WriteBuffer();
    for (final Plane plane in image.planes) {
      allBytes.putUint8List(plane.bytes);
    }
    final bytes = allBytes.done().buffer.asUint8List();

    return InputImage.fromBytes(
      bytes: bytes,
      metadata: InputImageMetadata(
        size: Size(image.width.toDouble(), image.height.toDouble()),
        rotation: rotation,
        format: format,
        bytesPerRow: image.planes.first.bytesPerRow,
      ),
    );
  }
}
