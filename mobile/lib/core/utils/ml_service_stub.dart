import 'package:image/image.dart' as imglib;

class MLService {
  bool get isInitialized => false;
  String? get lastError => null;

  Future<void> initialize() async {
    // No-op for web or platforms without tflite_flutter support
  }

  List<double> predict(imglib.Image image) {
    // Return dummy vector for web/unsupported platforms
    return List.filled(192, 0.0);
  }

  double calculateEuclideanDistance(List<double> e1, List<double> e2) {
    return 0.0;
  }

  void dispose() {}
}
