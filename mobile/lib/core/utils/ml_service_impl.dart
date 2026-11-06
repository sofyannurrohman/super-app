import 'dart:io';
import 'dart:math';
import 'package:flutter/foundation.dart';
import 'package:flutter/services.dart';
import 'package:image/image.dart' as imglib;
import 'package:flutter_litert/flutter_litert.dart';
import 'package:hris_app/core/utils/ml_config.dart';

class MLService {
  late Interpreter _interpreter;
  bool _isInitialized = false;

  // Dynamic shapes
  late List<int> _inputShape;
  late List<int> _outputShape;
  late int _inputSize; // assumed square Width/Height
  late int _outputSize;

  String? _lastError;

  Future<void> initialize() async {
    if (kIsWeb || !kUseMlKit) return; // Simulator: skip TFLite init
    if (_isInitialized) return;

    _lastError = null;
    try {
      final data = await rootBundle.load('assets/mobilefacenet.tflite');
      final bytes = data.buffer.asUint8List();

      // Strategy 1: Attempt with GPU Delegate (opsional, boleh gagal)
      if (!kIsWeb && Platform.isAndroid) {
        try {
          debugPrint('MLService: Trying GPU delegate (Android)...');
          final delegate = GpuDelegateV2(
            options: GpuDelegateOptionsV2(
              isPrecisionLossAllowed: false,
            ),
          );
          final options = InterpreterOptions()..addDelegate(delegate);
          _interpreter = Interpreter.fromBuffer(bytes, options: options);
          _extractShapes();
          _isInitialized = true;
          debugPrint(
              'MLService: GPU delegate OK. Input: $_inputShape, Output: $_outputShape');
          return;
        } catch (gpuErr) {
          debugPrint(
              'MLService: GPU delegate gagal (normal di beberapa device): $gpuErr');
          // Lanjut ke CPU fallback
        }
      } else if (!kIsWeb && Platform.isIOS) {
        try {
          debugPrint('MLService: Trying Metal delegate (iOS)...');
          final delegate = GpuDelegate();
          final options = InterpreterOptions()..addDelegate(delegate);
          _interpreter = Interpreter.fromBuffer(bytes, options: options);
          _extractShapes();
          _isInitialized = true;
          debugPrint(
              'MLService: Metal delegate OK. Input: $_inputShape, Output: $_outputShape');
          return;
        } catch (metalErr) {
          debugPrint('MLService: Metal delegate gagal: $metalErr');
        }
      }

      // Strategy 2: CPU — selalu bisa berjalan di semua device Android
      debugPrint('MLService: Menggunakan CPU (kompatibel semua device)...');
      final cpuOptions = InterpreterOptions()..threads = 2;
      _interpreter = Interpreter.fromBuffer(bytes, options: cpuOptions);
      _extractShapes();
      _isInitialized = true;
      debugPrint(
          'MLService: CPU OK. Input: $_inputShape, Output: $_outputShape');
    } catch (e, stack) {
      _lastError = e.toString();
      debugPrint('MLService: Gagal load model: $e\n$stack');
      _isInitialized = false;
    }
  }

  void _extractShapes() {
    final inputTensors = _interpreter.getInputTensors();
    final outputTensors = _interpreter.getOutputTensors();

    _inputShape = inputTensors.first.shape;
    _outputShape = outputTensors.first.shape;

    // Most MobileFaceNet models use [1, 112, 112, 3] or [1, height, width, 3]
    _inputSize = _inputShape[1];
    _outputSize = _outputShape[1];
  }

  bool get isInitialized => _isInitialized;
  String? get lastError => _lastError;

  List<double> predict(imglib.Image image) {
    if (kIsWeb || !kUseMlKit) return List.filled(192, 0.0);
    if (!_isInitialized) {
      debugPrint('MLService: predict dipanggil sebelum init selesai');
      throw Exception('MLService is not initialized');
    }

    // Resize image to model's expected input shape (likely 112x112)
    imglib.Image resizedImage =
        imglib.copyResizeCropSquare(image, size: _inputSize);

    // Prepare input: Shape [1, Size, Size, 3]
    var input = _imageTo4DList(resizedImage, _inputSize, 127.5, 127.5);

    // Prepare output: Shape [1, OutputSize]
    var output = List.generate(1, (index) => List.filled(_outputSize, 0.0));

    try {
      _interpreter.run(input, output);
    } catch (e) {
      debugPrint("MLService: Interpreter run error: $e");
      rethrow;
    }

    // Normalize and return the output vector
    List<double> rawOutput = List<double>.from(output[0]);
    return _normalize(rawOutput);
  }

  // Efficient build of 4D list via Float32List
  List<dynamic> _imageTo4DList(
      imglib.Image image, int size, double mean, double std) {
    var rawBytes = Float32List(1 * size * size * 3);
    int pixelIndex = 0;

    for (int y = 0; y < size; y++) {
      for (int x = 0; x < size; x++) {
        final pixel = image.getPixelSafe(x, y);
        rawBytes[pixelIndex++] = (pixel.r - mean) / std;
        rawBytes[pixelIndex++] = (pixel.g - mean) / std;
        rawBytes[pixelIndex++] = (pixel.b - mean) / std;
      }
    }

    // reshape returns List<dynamic> (which contains nested lists)
    return rawBytes.reshape([1, size, size, 3]);
  }

  List<double> _normalize(List<double> input) {
    double sum = 0.0;
    for (int i = 0; i < input.length; i++) {
      sum += input[i] * input[i];
    }
    double length = sqrt(sum);
    if (length == 0) return input; // guard division by zero
    for (int i = 0; i < input.length; i++) {
      input[i] /= length;
    }
    return input;
  }

  double calculateEuclideanDistance(List<double> e1, List<double> e2) {
    if (e1.length != e2.length) {
      throw Exception(
          'Embeddings must have same length: ${e1.length} vs ${e2.length}');
    }
    double sum = 0.0;
    for (int i = 0; i < e1.length; i++) {
      sum += pow((e1[i] - e2[i]), 2);
    }
    return sqrt(sum);
  }

  void dispose() {
    if (kIsWeb || !kUseMlKit) return;
    if (_isInitialized) {
      _interpreter.close();
      _isInitialized = false;
    }
  }
}
