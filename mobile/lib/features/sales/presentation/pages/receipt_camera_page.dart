import 'dart:io';
import 'dart:convert';
import 'package:flutter/material.dart';
import 'package:flutter/foundation.dart';
import 'package:google_fonts/google_fonts.dart';
import 'package:camera/camera.dart';
import 'package:image/image.dart' as img;
import 'package:path_provider/path_provider.dart';
import 'package:image_picker/image_picker.dart';
import 'package:hris_app/features/sales/data/models/store_model.dart';
import 'package:hris_app/features/sales/presentation/pages/visit_checkout_page.dart';

class ReceiptImageProcessResult {
  final bool isBlurry;
  final Uint8List processedBytes;

  ReceiptImageProcessResult(
      {required this.isBlurry, required this.processedBytes});
}

Future<ReceiptImageProcessResult> _processReceiptImage(Uint8List bytes) async {
  final image = img.decodeImage(bytes);
  if (image == null) {
    return ReceiptImageProcessResult(isBlurry: false, processedBytes: bytes);
  }

  // Bake orientation to ensure proper rotation before any processing
  final orientedImage = img.bakeOrientation(image);

  // Laplacian variance for blur detection on the oriented image
  double sum = 0;
  double sumSq = 0;
  int count = 0;

  // Optimize performance on large image dimensions by stepping pixels
  final step = orientedImage.width > 2000 ? 3 : 2;

  for (int y = 1; y < orientedImage.height - 1; y += step) {
    for (int x = 1; x < orientedImage.width - 1; x += step) {
      final px = img.getLuminance(orientedImage.getPixel(x, y));
      final top = img.getLuminance(orientedImage.getPixel(x, y - 1));
      final bottom = img.getLuminance(orientedImage.getPixel(x, y + 1));
      final left = img.getLuminance(orientedImage.getPixel(x - 1, y));
      final right = img.getLuminance(orientedImage.getPixel(x + 1, y));
      final lap = (top + bottom + left + right - 4 * px).abs();
      sum += lap;
      sumSq += lap * lap;
      count++;
    }
  }

  bool isBlurry = false;
  if (count > 0) {
    final mean = sum / count;
    final variance = (sumSq / count) - (mean * mean);
    isBlurry = variance < 80;
  }

  // Downscale image if it exceeds 1280px in width or height
  img.Image finalImage = orientedImage;
  const int maxDimension = 1280;
  if (orientedImage.width > maxDimension ||
      orientedImage.height > maxDimension) {
    int targetWidth;
    int targetHeight;
    if (orientedImage.width > orientedImage.height) {
      targetWidth = maxDimension;
      targetHeight =
          (orientedImage.height * maxDimension) ~/ orientedImage.width;
    } else {
      targetHeight = maxDimension;
      targetWidth =
          (orientedImage.width * maxDimension) ~/ orientedImage.height;
    }
    finalImage = img.copyResize(orientedImage,
        width: targetWidth,
        height: targetHeight,
        interpolation: img.Interpolation.average);
  }

  // Compress to JPEG with quality 90
  final processedBytes = img.encodeJpg(finalImage, quality: 90);

  return ReceiptImageProcessResult(
    isBlurry: isBlurry,
    processedBytes: Uint8List.fromList(processedBytes),
  );
}

class ReceiptCameraPage extends StatefulWidget {
  final StoreModel store;
  final String selfiePath;
  final String jobPositionTitle;
  final String? companyId;
  final String? companyName;

  const ReceiptCameraPage({
    super.key,
    required this.store,
    required this.selfiePath,
    required this.jobPositionTitle,
    this.companyId,
    this.companyName,
  });

  @override
  State<ReceiptCameraPage> createState() => _ReceiptCameraPageState();
}

class _ReceiptCameraPageState extends State<ReceiptCameraPage> {
  CameraController? _controller;
  bool _isInitialized = false;
  bool _isTaking = false;
  String? _capturedPath;
  bool _isBlurry = false;

  @override
  void initState() {
    super.initState();
    _initCamera();
  }

  Future<void> _initCamera() async {
    final cameras = await availableCameras();
    if (cameras.isEmpty) return;
    // Prefer back camera
    final back = cameras.firstWhere(
      (c) => c.lensDirection == CameraLensDirection.back,
      orElse: () => cameras.first,
    );
    _controller =
        CameraController(back, ResolutionPreset.high, enableAudio: false);
    await _controller!.initialize();
    if (mounted) setState(() => _isInitialized = true);
  }

  Future<void> _capturePhoto() async {
    if (_controller == null || !_isInitialized || _isTaking) return;
    setState(() {
      _isTaking = true;
      _isBlurry = false;
    });

    try {
      final xfile = await _controller!.takePicture();
      final bytes = await xfile.readAsBytes();

      final result = await compute(_processReceiptImage, bytes);

      String pathOrBase64 = '';
      if (kIsWeb) {
        pathOrBase64 =
            'data:image/jpeg;base64,${base64Encode(result.processedBytes)}';
      } else {
        final tempDir = await getTemporaryDirectory();
        final outPath =
            '${tempDir.path}/receipt_${DateTime.now().millisecondsSinceEpoch}.jpg';
        await File(outPath).writeAsBytes(result.processedBytes);
        pathOrBase64 = outPath;
      }

      setState(() {
        _capturedPath = pathOrBase64;
        _isBlurry = result.isBlurry;
      });
    } catch (e) {
      debugPrint('Camera error: $e');
    } finally {
      if (mounted) setState(() => _isTaking = false);
    }
  }

  Future<void> _pickFromGallery() async {
    if (_isTaking) return;
    setState(() {
      _isTaking = true;
      _isBlurry = false;
    });
    try {
      final picker = ImagePicker();
      final pickedFile = await picker.pickImage(source: ImageSource.gallery);
      if (pickedFile != null) {
        final bytes = await pickedFile.readAsBytes();
        final result = await compute(_processReceiptImage, bytes);

        String pathOrBase64 = '';
        if (kIsWeb) {
          pathOrBase64 =
              'data:image/jpeg;base64,${base64Encode(result.processedBytes)}';
        } else {
          final tempDir = await getTemporaryDirectory();
          final outPath =
              '${tempDir.path}/receipt_${DateTime.now().millisecondsSinceEpoch}.jpg';
          await File(outPath).writeAsBytes(result.processedBytes);
          pathOrBase64 = outPath;
        }

        setState(() {
          _capturedPath = pathOrBase64;
          _isBlurry = result.isBlurry;
        });
      }
    } catch (e) {
      debugPrint('Gallery error: $e');
    } finally {
      if (mounted) setState(() => _isTaking = false);
    }
  }

  @override
  void dispose() {
    _controller?.dispose();
    super.dispose();
  }

  @override
  Widget build(BuildContext context) {
    if (_capturedPath != null) return _buildPreview();
    return _buildCamera();
  }

  Widget _buildCamera() {
    return Scaffold(
      backgroundColor: Colors.black,
      body: Stack(
        children: [
          if (_isInitialized && _controller != null)
            SizedBox.expand(child: CameraPreview(_controller!))
          else
            const Center(child: CircularProgressIndicator(color: Colors.white)),

          // Top info
          Positioned(
            top: 0,
            left: 0,
            right: 0,
            child: Container(
              padding: const EdgeInsets.fromLTRB(16, 60, 16, 24),
              decoration: BoxDecoration(
                  gradient: LinearGradient(
                      begin: Alignment.topCenter,
                      end: Alignment.bottomCenter,
                      colors: [
                    Colors.black.withOpacity(0.75),
                    Colors.transparent
                  ])),
              child: Column(
                children: [
                  Text(widget.selfiePath.isEmpty ? 'FOTO NOTA' : 'FOTO NOTA',
                      style: GoogleFonts.outfit(
                          color: Colors.blueAccent,
                          fontSize: 11,
                          fontWeight: FontWeight.w900,
                          letterSpacing: 1)),
                  const SizedBox(height: 4),
                  Text('Posisikan nota dalam bingkai kuning',
                      style: GoogleFonts.outfit(
                          color: Colors.white,
                          fontSize: 15,
                          fontWeight: FontWeight.w700)),
                  Text('Pastikan teks nota terlihat JELAS dan tidak buram.',
                      style: GoogleFonts.outfit(
                          color: Colors.white70, fontSize: 12)),
                ],
              ),
            ),
          ),

          // Guide frame
          Center(
            child: Container(
              width: MediaQuery.of(context).size.width * 0.85,
              height: MediaQuery.of(context).size.height * 0.50,
              decoration: BoxDecoration(
                border: Border.all(color: Colors.yellowAccent, width: 3),
                borderRadius: BorderRadius.circular(20),
              ),
              child: Center(
                child: Text(
                  'POSISIKAN NOTA DI SINI',
                  style: GoogleFonts.outfit(
                      color: Colors.yellowAccent.withOpacity(0.6),
                      fontWeight: FontWeight.w900,
                      fontSize: 13,
                      letterSpacing: 1),
                ),
              ),
            ),
          ),

          // Capture and Gallery buttons
          Positioned(
            bottom: 40,
            left: 0,
            right: 0,
            child: Row(
              mainAxisAlignment: MainAxisAlignment.spaceEvenly,
              children: [
                GestureDetector(
                  onTap: _pickFromGallery,
                  child: Container(
                    width: 64,
                    height: 64,
                    decoration: BoxDecoration(
                      shape: BoxShape.circle,
                      color: Colors.white24,
                      border: Border.all(color: Colors.white, width: 2),
                    ),
                    child: const Icon(Icons.photo_library_rounded,
                        size: 28, color: Colors.white),
                  ),
                ),
                GestureDetector(
                  onTap: _capturePhoto,
                  child: Container(
                    width: 80,
                    height: 80,
                    decoration: BoxDecoration(
                      shape: BoxShape.circle,
                      color: Colors.yellowAccent,
                      border: Border.all(
                          color: Colors.yellowAccent.withOpacity(0.4),
                          width: 6),
                      boxShadow: [
                        BoxShadow(
                            color: Colors.yellowAccent.withOpacity(0.4),
                            blurRadius: 20)
                      ],
                    ),
                    child: _isTaking
                        ? const CircularProgressIndicator(color: Colors.black)
                        : const Icon(Icons.camera_alt_rounded,
                            size: 36, color: Colors.black),
                  ),
                ),
                const SizedBox(width: 64), // Balance the row layout
              ],
            ),
          ),
        ],
      ),
    );
  }

  Widget _buildPreview() {
    return Scaffold(
      backgroundColor: Colors.black,
      body: Stack(
        children: [
          SizedBox.expand(
              child: _capturedPath!.startsWith('data:')
                  ? Image.memory(base64Decode(_capturedPath!.split(',').last),
                      fit: BoxFit.contain)
                  : kIsWeb
                      ? const Center(
                          child:
                              Text('Preview not available on Web (Non-Base64)'))
                      : Image.file(File(_capturedPath!), fit: BoxFit.contain)),

          // Blur warning
          if (_isBlurry)
            Positioned(
              top: 0,
              left: 0,
              right: 0,
              child: Container(
                padding: const EdgeInsets.fromLTRB(20, 60, 20, 24),
                color: Colors.red.withOpacity(0.85),
                child: Column(
                  children: [
                    const Icon(Icons.warning_amber_rounded,
                        color: Colors.white, size: 36),
                    const SizedBox(height: 8),
                    Text('FOTO BURAM!',
                        style: GoogleFonts.outfit(
                            color: Colors.white,
                            fontWeight: FontWeight.w900,
                            fontSize: 20,
                            letterSpacing: 1)),
                    const SizedBox(height: 4),
                    Text(
                        'Harap foto ulang agar nota dapat dibaca jelas oleh admin.',
                        textAlign: TextAlign.center,
                        style: GoogleFonts.outfit(
                            color: Colors.white70, fontSize: 13)),
                  ],
                ),
              ),
            )
          else
            Positioned(
              top: 0,
              left: 0,
              right: 0,
              child: Container(
                padding: const EdgeInsets.fromLTRB(20, 60, 20, 24),
                decoration: BoxDecoration(
                    gradient: LinearGradient(
                        begin: Alignment.topCenter,
                        end: Alignment.bottomCenter,
                        colors: [
                      Colors.black.withOpacity(0.7),
                      Colors.transparent
                    ])),
                child: Column(
                  children: [
                    const Icon(Icons.check_circle_rounded,
                        color: Colors.greenAccent, size: 32),
                    const SizedBox(height: 6),
                    Text('FOTO JELAS!',
                        style: GoogleFonts.outfit(
                            color: Colors.greenAccent,
                            fontWeight: FontWeight.w900,
                            fontSize: 18,
                            letterSpacing: 1)),
                    Text('Nota dapat dibaca jelas oleh admin.',
                        style: GoogleFonts.outfit(
                            color: Colors.white70, fontSize: 13)),
                  ],
                ),
              ),
            ),

          // Buttons
          Positioned(
            bottom: 40,
            left: 20,
            right: 20,
            child: Row(
              children: [
                Expanded(
                  child: OutlinedButton.icon(
                    onPressed: () => setState(() {
                      _capturedPath = null;
                      _isBlurry = false;
                    }),
                    icon: const Icon(Icons.replay_rounded, color: Colors.white),
                    label: Text(_isBlurry ? 'FOTO ULANG' : 'ULANGI',
                        style: GoogleFonts.outfit(
                            color: Colors.white, fontWeight: FontWeight.w800)),
                    style: OutlinedButton.styleFrom(
                      side: BorderSide(
                          color: _isBlurry ? Colors.red : Colors.white,
                          width: 2),
                      shape: RoundedRectangleBorder(
                          borderRadius: BorderRadius.circular(16)),
                      padding: const EdgeInsets.symmetric(vertical: 16),
                    ),
                  ),
                ),
                if (!_isBlurry) ...[
                  const SizedBox(width: 16),
                  Expanded(
                    child: ElevatedButton.icon(
                      onPressed: () => Navigator.push(
                          context,
                          MaterialPageRoute(
                            builder: (_) => VisitCheckoutPage(
                              store: widget.store,
                              selfiePath: widget.selfiePath,
                              receiptPath: _capturedPath,
                              jobPositionTitle: widget.jobPositionTitle,
                              companyId: widget.companyId,
                              companyName: widget.companyName,
                            ),
                          )),
                      icon: const Icon(Icons.check_circle_rounded,
                          color: Colors.white),
                      label: Text('LANJUT',
                          style: GoogleFonts.outfit(
                              color: Colors.white,
                              fontWeight: FontWeight.w900,
                              fontSize: 16)),
                      style: ElevatedButton.styleFrom(
                        backgroundColor: Colors.green,
                        shape: RoundedRectangleBorder(
                            borderRadius: BorderRadius.circular(16)),
                        padding: const EdgeInsets.symmetric(vertical: 16),
                      ),
                    ),
                  ),
                ],
              ],
            ),
          ),
        ],
      ),
    );
  }
}
