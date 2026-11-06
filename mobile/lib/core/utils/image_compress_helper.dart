import 'dart:convert';
import 'dart:io';
import 'package:flutter/foundation.dart';
import 'package:image/image.dart' as img;
import 'package:path_provider/path_provider.dart';

class ImageCompressParams {
  final Uint8List bytes;
  final int quality;
  final int maxDimension;

  ImageCompressParams({
    required this.bytes,
    this.quality = 75,
    this.maxDimension = 1024,
  });
}

Uint8List _compressImageIsolate(ImageCompressParams params) {
  final image = img.decodeImage(params.bytes);
  if (image == null) {
    return params.bytes;
  }

  // Bake orientation
  final oriented = img.bakeOrientation(image);

  img.Image finalImage = oriented;
  if (oriented.width > params.maxDimension ||
      oriented.height > params.maxDimension) {
    int targetWidth;
    int targetHeight;
    if (oriented.width > oriented.height) {
      targetWidth = params.maxDimension;
      targetHeight = (oriented.height * params.maxDimension) ~/ oriented.width;
    } else {
      targetHeight = params.maxDimension;
      targetWidth = (oriented.width * params.maxDimension) ~/ oriented.height;
    }
    finalImage = img.copyResize(
      oriented,
      width: targetWidth,
      height: targetHeight,
      interpolation: img.Interpolation.linear,
    );
  }

  final jpgBytes = img.encodeJpg(finalImage, quality: params.quality);
  return Uint8List.fromList(jpgBytes);
}

class RawRgbaCompressParams {
  final Uint8List rgbaBytes;
  final int width;
  final int height;
  final int quality;

  RawRgbaCompressParams({
    required this.rgbaBytes,
    required this.width,
    required this.height,
    this.quality = 75,
  });
}

Uint8List _encodeRawRgbaIsolate(RawRgbaCompressParams params) {
  final image = img.Image.fromBytes(
    width: params.width,
    height: params.height,
    bytes: params.rgbaBytes.buffer,
    bytesOffset: params.rgbaBytes.offsetInBytes,
    numChannels: 4,
    order: img.ChannelOrder.rgba,
  );
  final jpgBytes = img.encodeJpg(image, quality: params.quality);
  return Uint8List.fromList(jpgBytes);
}

class ImageCompressHelper {
  /// Compresses input image bytes to JPEG with given quality (default 75%) and maxDimension (default 1024px).
  /// Runs in a background isolate (compute) to prevent freezing the UI.
  static Future<Uint8List> compressToJpg(
    Uint8List bytes, {
    int quality = 75,
    int maxDimension = 1024,
  }) async {
    return await compute(
      _compressImageIsolate,
      ImageCompressParams(
        bytes: bytes,
        quality: quality,
        maxDimension: maxDimension,
      ),
    );
  }

  /// Directly encodes raw RGBA bytes (from ui.ImageByteFormat.rawRgba) to JPEG in a background isolate.
  /// Avoids the massive overhead of encoding to PNG on the main thread and decoding PNG again in Dart.
  static Future<Uint8List> encodeRawRgbaToJpg(
    Uint8List rgbaBytes, {
    required int width,
    required int height,
    int quality = 75,
  }) async {
    return await compute(
      _encodeRawRgbaIsolate,
      RawRgbaCompressParams(
        rgbaBytes: rgbaBytes,
        width: width,
        height: height,
        quality: quality,
      ),
    );
  }

  /// Compresses image bytes and saves directly as a file in application documents directory.
  /// Returns the absolute local file path string to avoid storing large Base64 in SQLite.
  static Future<String> saveCompressedToFile(
    Uint8List bytes, {
    String subDir = 'media',
    String prefix = 'img',
    int quality = 70,
    int maxDimension = 1024,
  }) async {
    final compressedBytes = await compressToJpg(
      bytes,
      quality: quality,
      maxDimension: maxDimension,
    );

    if (kIsWeb) {
      // Web fallback: return base64 data URI
      return 'data:image/jpeg;base64,${base64Encode(compressedBytes)}';
    }

    try {
      final appDocDir = await getApplicationDocumentsDirectory();
      final targetDir = Directory('${appDocDir.path}/$subDir');
      if (!await targetDir.exists()) {
        await targetDir.create(recursive: true);
      }
      final filePath =
          '${targetDir.path}/${prefix}_${DateTime.now().millisecondsSinceEpoch}.jpg';
      final file = File(filePath);
      await file.writeAsBytes(compressedBytes, flush: true);
      return filePath;
    } catch (e) {
      debugPrint('[ImageCompressHelper] Error saving image to file: $e');
      // Fallback
      return 'data:image/jpeg;base64,${base64Encode(compressedBytes)}';
    }
  }

  /// Helper specifically for canvas signatures: converts PNG bytes to compressed JPG file.
  static Future<String> saveSignatureToFile(
    Uint8List bytes, {
    String prefix = 'sig',
  }) async {
    return await saveCompressedToFile(
      bytes,
      subDir: 'signatures',
      prefix: prefix,
      quality: 80,
      maxDimension: 800,
    );
  }
}
