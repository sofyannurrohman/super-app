import 'dart:math';
import 'dart:typed_data';
import 'package:flutter_test/flutter_test.dart';
import 'package:hris_app/core/utils/image_compress_helper.dart';
import 'package:image/image.dart' as img;

void main() {
  TestWidgetsFlutterBinding.ensureInitialized();

  test('ImageCompressHelper compresses raw large image bytes to JPG', () async {
    // Generate a 400x400 image with noise (raw uncompressed pixels = ~640KB)
    final dummy = img.Image(width: 400, height: 400);
    final rnd = Random(42);
    for (int y = 0; y < dummy.height; y++) {
      for (int x = 0; x < dummy.width; x++) {
        dummy.setPixelRgb(x, y, rnd.nextInt(256), rnd.nextInt(256), rnd.nextInt(256));
      }
    }
    final rawBmpBytes = Uint8List.fromList(img.encodeBmp(dummy));

    final compressed = await ImageCompressHelper.compressToJpg(
      rawBmpBytes,
      quality: 70,
      maxDimension: 200,
    );

    expect(compressed, isNotEmpty);
    // Compressed 200px JPG should be significantly smaller than raw 400px BMP
    expect(compressed.length, lessThan(rawBmpBytes.length ~/ 4));
  });

  test('ImageCompressHelper.encodeRawRgbaToJpg directly encodes raw RGBA buffer to valid JPG', () async {
    const int width = 300;
    const int height = 200;
    final rgbaBytes = Uint8List(width * height * 4);
    for (int i = 0; i < rgbaBytes.length; i += 4) {
      rgbaBytes[i] = 20;     // R
      rgbaBytes[i + 1] = 120; // G
      rgbaBytes[i + 2] = 220; // B
      rgbaBytes[i + 3] = 255; // A
    }

    final jpgBytes = await ImageCompressHelper.encodeRawRgbaToJpg(
      rgbaBytes,
      width: width,
      height: height,
      quality: 75,
    );

    expect(jpgBytes, isNotEmpty);
    // Verify JPEG magic header (FF D8)
    expect(jpgBytes[0], equals(0xFF));
    expect(jpgBytes[1], equals(0xD8));

    final decoded = img.decodeJpg(jpgBytes);
    expect(decoded, isNotNull);
    expect(decoded!.width, equals(width));
    expect(decoded.height, equals(height));
  });
}
