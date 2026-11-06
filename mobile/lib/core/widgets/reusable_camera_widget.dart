import 'dart:io';
import 'package:camera/camera.dart';
import 'package:flutter/foundation.dart';
import 'package:flutter/material.dart';

class ReusableCameraWidget extends StatefulWidget {
  final Function(String imagePath) onCapture;
  final VoidCallback? onCancel;
  final String instructionText;

  const ReusableCameraWidget({
    super.key,
    required this.onCapture,
    this.onCancel,
    this.instructionText = "Mohon posisikan wajah Anda di dalam area",
  });

  @override
  State<ReusableCameraWidget> createState() => _ReusableCameraWidgetState();
}

class _ReusableCameraWidgetState extends State<ReusableCameraWidget>
    with WidgetsBindingObserver {
  CameraController? _controller;
  List<CameraDescription> _cameras = [];
  bool _isInit = false;
  int _selectedCameraIndex = -1;

  @override
  void initState() {
    super.initState();
    WidgetsBinding.instance.addObserver(this);
    _initCamera();
  }

  Future<void> _initCamera() async {
    try {
      _cameras = await availableCameras();
      if (_cameras.isEmpty) {
        if (mounted) {
          ScaffoldMessenger.of(context).showSnackBar(
            const SnackBar(
                content: Text("Tidak menemukan kamera di perangkat")),
          );
        }
        return;
      }

      // Prioritize front camera
      _selectedCameraIndex = _cameras
          .indexWhere((c) => c.lensDirection == CameraLensDirection.front);

      if (_selectedCameraIndex == -1) {
        _selectedCameraIndex = 0; // fallback to the first camera
      }

      await _startCamera(_cameras[_selectedCameraIndex]);
    } on CameraException catch (e) {
      if (mounted) {
        ScaffoldMessenger.of(context).showSnackBar(
          SnackBar(content: Text("Error Kamera: ${e.description}")),
        );
      }
    }
  }

  Future<void> _startCamera(CameraDescription cameraDescription) async {
    final prevController = _controller;
    final newController = CameraController(
      cameraDescription,
      ResolutionPreset.high,
      enableAudio: false,
      imageFormatGroup: !kIsWeb && Platform.isAndroid
          ? ImageFormatGroup.nv21
          : ImageFormatGroup.bgra8888,
    );

    _controller = newController;

    try {
      if (prevController != null) {
        await prevController.dispose();
      }
      await newController.initialize();
      if (mounted) {
        setState(() {
          _isInit = true;
        });
      }
    } on CameraException catch (e) {
      if (mounted) {
        ScaffoldMessenger.of(context).showSnackBar(
          SnackBar(content: Text("Gagal memuat kamera: ${e.description}")),
        );
      }
    }
  }

  Future<void> _switchCamera() async {
    if (_cameras.length < 2) return;

    _selectedCameraIndex = (_selectedCameraIndex + 1) % _cameras.length;
    setState(() {
      _isInit = false;
    });

    await _startCamera(_cameras[_selectedCameraIndex]);
  }

  Future<void> _takePicture() async {
    if (_controller == null || !_controller!.value.isInitialized) return;
    if (_controller!.value.isTakingPicture) return;

    try {
      final XFile picture = await _controller!.takePicture();
      widget.onCapture(picture.path);
    } on CameraException catch (e) {
      if (mounted) {
        ScaffoldMessenger.of(context).showSnackBar(
          SnackBar(content: Text("Gagal mengambil foto: ${e.description}")),
        );
      }
    }
  }

  @override
  void dispose() {
    WidgetsBinding.instance.removeObserver(this);
    _controller?.dispose();
    super.dispose();
  }

  @override
  void didChangeAppLifecycleState(AppLifecycleState state) {
    if (_controller == null || !_controller!.value.isInitialized) return;

    if (state == AppLifecycleState.inactive) {
      _controller?.dispose();
    } else if (state == AppLifecycleState.resumed) {
      if (_selectedCameraIndex != -1) {
        _startCamera(_cameras[_selectedCameraIndex]);
      }
    }
  }

  @override
  Widget build(BuildContext context) {
    if (!_isInit || _controller == null) {
      return const Scaffold(
        backgroundColor: Colors.black,
        body: Center(child: CircularProgressIndicator(color: Colors.white)),
      );
    }

    final size = MediaQuery.of(context).size;
    var scale = size.aspectRatio * _controller!.value.aspectRatio;
    if (scale < 1) scale = 1 / scale;

    return Scaffold(
      backgroundColor: Colors.black,
      body: SafeArea(
        child: Stack(
          children: [
            // Full Screen Camera Preview
            Transform.scale(
              scale: scale,
              child: Center(
                child: AspectRatio(
                  aspectRatio: _controller!.value.aspectRatio,
                  child: CameraPreview(_controller!),
                ),
              ),
            ),

            // Instruction Text
            Positioned(
              top: 40,
              left: 0,
              right: 0,
              child: Container(
                padding:
                    const EdgeInsets.symmetric(vertical: 12, horizontal: 20),
                margin: const EdgeInsets.symmetric(horizontal: 40),
                decoration: BoxDecoration(
                  color: Colors.black54,
                  borderRadius: BorderRadius.circular(12),
                ),
                child: Text(
                  widget.instructionText,
                  textAlign: TextAlign.center,
                  style: const TextStyle(
                    color: Colors.white,
                    fontSize: 16,
                    fontWeight: FontWeight.w500,
                  ),
                ),
              ),
            ),

            // Controls
            Positioned(
              bottom: 40,
              left: 0,
              right: 0,
              child: Row(
                mainAxisAlignment: MainAxisAlignment.spaceEvenly,
                children: [
                  // Cancel / Back Button
                  IconButton(
                    icon: const Icon(Icons.arrow_back,
                        color: Colors.white, size: 30),
                    onPressed:
                        widget.onCancel ?? () => Navigator.of(context).pop(),
                  ),

                  // Capture Button
                  GestureDetector(
                    onTap: _takePicture,
                    child: Container(
                      height: 80,
                      width: 80,
                      decoration: BoxDecoration(
                        shape: BoxShape.circle,
                        border: Border.all(color: Colors.white, width: 4),
                      ),
                      child: Center(
                        child: Container(
                          height: 64,
                          width: 64,
                          decoration: const BoxDecoration(
                            color: Colors.white,
                            shape: BoxShape.circle,
                          ),
                        ),
                      ),
                    ),
                  ),

                  // Switch Camera Button
                  _cameras.length > 1
                      ? IconButton(
                          icon: const Icon(Icons.flip_camera_ios,
                              color: Colors.white, size: 30),
                          onPressed: _switchCamera,
                        )
                      : const SizedBox(
                          width: 48), // Empty placeholder to balance row
                ],
              ),
            ),
          ],
        ),
      ),
    );
  }
}
