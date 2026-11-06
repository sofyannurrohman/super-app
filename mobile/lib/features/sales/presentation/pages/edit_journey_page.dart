import 'dart:convert';
import 'dart:io';
import 'dart:ui' as ui;
import 'package:camera/camera.dart';
import 'package:dio/dio.dart';
import 'package:flutter/foundation.dart';
import 'package:flutter/material.dart';
import 'package:geocoding/geocoding.dart';
import 'package:geolocator/geolocator.dart';
import 'package:get_it/get_it.dart';
import 'package:google_fonts/google_fonts.dart';
import 'package:intl/intl.dart';
import 'package:path_provider/path_provider.dart';

import 'package:hris_app/core/network/api_client.dart';
import 'package:hris_app/core/theme/app_colors.dart';
import 'package:hris_app/core/utils/constants.dart';
import 'package:hris_app/core/utils/image_compress_helper.dart';
import 'package:hris_app/features/sales/data/services/sales_api_service.dart';

class EditJourneyPage extends StatefulWidget {
  final String journeyId;
  final String editMode; // 'start' | 'end' | 'both'
  final double? initialStartKM;
  final double? initialEndKM;
  final String? currentStartPhotoUrl;
  final String? currentEndPhotoUrl;
  final String? currentStartLocation;
  final String? currentEndLocation;
  final String formattedDate;

  const EditJourneyPage({
    super.key,
    required this.journeyId,
    required this.editMode,
    this.initialStartKM,
    this.initialEndKM,
    this.currentStartPhotoUrl,
    this.currentEndPhotoUrl,
    this.currentStartLocation,
    this.currentEndLocation,
    required this.formattedDate,
  });

  @override
  State<EditJourneyPage> createState() => _EditJourneyPageState();
}

class _EditJourneyPageState extends State<EditJourneyPage> {
  final _apiService = SalesApiService(apiClient: GetIt.I<ApiClient>());

  late final TextEditingController _startKMController;
  late final TextEditingController _endKMController;

  Uint8List? _newStartBytes;
  String? _newStartPath;

  Uint8List? _newEndBytes;
  String? _newEndPath;

  bool _isSubmitting = false;

  // In-page camera state for taking a new photo
  bool _isCameraActive = false;
  String _cameraTarget = 'start'; // 'start' or 'end'
  CameraController? _cameraController;
  List<CameraDescription> _cameras = [];
  bool _isCameraInitialized = false;
  int _selectedCameraIdx = -1;
  bool _isTakingPhoto = false;

  String _currentGpsAddress = 'Mengambil lokasi...';
  String _currentGpsTimestamp = '';

  @override
  void initState() {
    super.initState();
    _startKMController = TextEditingController(
      text: widget.initialStartKM != null ? widget.initialStartKM.toString() : '',
    );
    _endKMController = TextEditingController(
      text: widget.initialEndKM != null && widget.initialEndKM! > 0
          ? widget.initialEndKM.toString()
          : '',
    );
  }

  @override
  void dispose() {
    _startKMController.dispose();
    _endKMController.dispose();
    _cameraController?.dispose();
    super.dispose();
  }

  bool _isLocalFile(String path) {
    if (path.isEmpty) return false;
    if (path.startsWith('http')) return false;
    if (path.startsWith('/uploads/') || path.startsWith('/api/v1/uploads/')) return false;
    if (path.startsWith('/data/') ||
        path.startsWith('/var/') ||
        path.startsWith('/storage/') ||
        path.startsWith('/private/')) {
      return true;
    }
    try {
      return File(path).existsSync();
    } catch (_) {
      return false;
    }
  }

  Widget _buildPhotoThumbnail(String rawPath) {
    if (rawPath.isEmpty) {
      return Container(
        height: 100,
        width: double.infinity,
        decoration: BoxDecoration(
          color: const Color(0xFFF1F5F9),
          borderRadius: BorderRadius.circular(12),
        ),
        child: const Center(
          child: Icon(Icons.broken_image_rounded, color: Colors.grey, size: 30),
        ),
      );
    }

    if (_isLocalFile(rawPath)) {
      final file = File(rawPath);
      if (file.existsSync()) {
        return ClipRRect(
          borderRadius: BorderRadius.circular(12),
          child: Image.file(
            file,
            height: 140,
            width: double.infinity,
            fit: BoxFit.cover,
            errorBuilder: (_, __, ___) => _buildThumbError(),
          ),
        );
      }
    }

    return ClipRRect(
      borderRadius: BorderRadius.circular(12),
      child: Image.network(
        AppConstants.buildImageUrl(rawPath),
        height: 140,
        width: double.infinity,
        fit: BoxFit.cover,
        loadingBuilder: (context, child, progress) {
          if (progress == null) return child;
          return Container(
            height: 140,
            color: const Color(0xFFF1F5F9),
            child: const Center(
              child: SizedBox(
                width: 24,
                height: 24,
                child: CircularProgressIndicator(strokeWidth: 2, color: AppColors.primaryGreen),
              ),
            ),
          );
        },
        errorBuilder: (_, __, ___) => _buildThumbError(),
      ),
    );
  }

  Widget _buildThumbError() {
    return Container(
      height: 100,
      decoration: BoxDecoration(
        color: const Color(0xFFF1F5F9),
        borderRadius: BorderRadius.circular(12),
      ),
      child: Center(
        child: Column(
          mainAxisAlignment: MainAxisAlignment.center,
          children: [
            const Icon(Icons.broken_image_rounded, size: 24, color: Colors.grey),
            const SizedBox(height: 4),
            Text('Foto tidak dapat dimuat', style: GoogleFonts.outfit(fontSize: 11, color: Colors.grey.shade600)),
          ],
        ),
      ),
    );
  }

  Future<void> _startCameraFor(String target) async {
    setState(() {
      _cameraTarget = target;
      _isCameraActive = true;
      _isCameraInitialized = false;
    });

    _currentGpsTimestamp = DateFormat('dd/MM/yyyy HH:mm:ss').format(DateTime.now());
    _getGpsLocation();

    try {
      _cameras = await availableCameras();
      if (_cameras.isEmpty) {
        if (mounted) {
          setState(() => _isCameraActive = false);
          ScaffoldMessenger.of(context).showSnackBar(
            const SnackBar(content: Text('Kamera tidak tersedia pada perangkat ini'), backgroundColor: Colors.red),
          );
        }
        return;
      }

      final backIdx = _cameras.indexWhere((c) => c.lensDirection == CameraLensDirection.back);
      _selectedCameraIdx = (backIdx != -1) ? backIdx : 0;

      _cameraController = CameraController(
        _cameras[_selectedCameraIdx],
        ResolutionPreset.high,
        enableAudio: false,
      );
      await _cameraController!.initialize();
      try {
        await _cameraController!.setFlashMode(FlashMode.off);
      } catch (_) {}
      if (mounted) {
        setState(() => _isCameraInitialized = true);
      }
    } catch (e) {
      if (mounted) {
        setState(() => _isCameraActive = false);
        ScaffoldMessenger.of(context).showSnackBar(
          SnackBar(content: Text('Gagal membuka kamera: $e'), backgroundColor: Colors.red),
        );
      }
    }
  }

  Future<void> _switchCamera() async {
    if (_cameras.length < 2) return;
    setState(() => _isCameraInitialized = false);
    await _cameraController?.dispose();
    _selectedCameraIdx = (_selectedCameraIdx + 1) % _cameras.length;
    _cameraController = CameraController(
      _cameras[_selectedCameraIdx],
      ResolutionPreset.high,
      enableAudio: false,
    );
    await _cameraController!.initialize();
    try {
      await _cameraController!.setFlashMode(FlashMode.off);
    } catch (_) {}
    if (mounted) setState(() => _isCameraInitialized = true);
  }

  Future<void> _getGpsLocation() async {
    try {
      bool serviceEnabled = await Geolocator.isLocationServiceEnabled();
      if (!serviceEnabled) {
        if (mounted) setState(() { _currentGpsAddress = 'GPS nonaktif'; });
        return;
      }

      LocationPermission perm = await Geolocator.checkPermission();
      if (perm == LocationPermission.denied) perm = await Geolocator.requestPermission();
      if (perm == LocationPermission.denied || perm == LocationPermission.deniedForever) {
        if (mounted) setState(() { _currentGpsAddress = 'Izin GPS ditolak'; });
        return;
      }

      final pos = await Geolocator.getCurrentPosition(desiredAccuracy: LocationAccuracy.high);
      String readableAddress = 'GPS: ${pos.latitude.toStringAsFixed(4)}, ${pos.longitude.toStringAsFixed(4)}';

      try {
        List<Placemark> placemarks = await placemarkFromCoordinates(pos.latitude, pos.longitude).timeout(const Duration(seconds: 5));
        if (placemarks.isNotEmpty) {
          Placemark place = placemarks[0];
          List<String> components = [
            place.street ?? '',
            place.subLocality ?? '',
            place.locality ?? '',
            place.subAdministrativeArea ?? '',
          ].where((s) => s.isNotEmpty && !s.contains('null')).toList();

          if (components.isNotEmpty) {
            readableAddress = '${components.join(', ')} (GPS: ${pos.latitude.toStringAsFixed(5)}, ${pos.longitude.toStringAsFixed(5)})';
          }
        }
      } catch (_) {}

      if (mounted) {
        setState(() {
          _currentGpsAddress = readableAddress;
        });
      }
    } catch (_) {
      if (mounted) setState(() { _currentGpsAddress = ''; });
    }
  }

  Future<void> _capturePhoto() async {
    if (_cameraController == null || !_isCameraInitialized || _isTakingPhoto) return;

    setState(() => _isTakingPhoto = true);
    try {
      final xfile = await _cameraController!.takePicture();
      final bytes = await xfile.readAsBytes();
      final resultBytes = await _burnWatermark(bytes, _cameraTarget);

      String pathOrBase64 = '';
      if (kIsWeb) {
        pathOrBase64 = 'data:image/jpeg;base64,${base64Encode(resultBytes)}';
      } else {
        final dir = await getTemporaryDirectory();
        final outPath = '${dir.path}/journey_edit_${_cameraTarget}_${DateTime.now().millisecondsSinceEpoch}.jpg';
        await File(outPath).writeAsBytes(resultBytes);
        pathOrBase64 = outPath;
      }

      await _cameraController?.dispose();
      _cameraController = null;

      if (mounted) {
        setState(() {
          _isCameraActive = false;
          _isTakingPhoto = false;
          if (_cameraTarget == 'start') {
            _newStartBytes = resultBytes;
            _newStartPath = pathOrBase64;
          } else {
            _newEndBytes = resultBytes;
            _newEndPath = pathOrBase64;
          }
        });
      }
    } catch (e) {
      debugPrint('Camera capture error: $e');
      if (mounted) {
        setState(() => _isTakingPhoto = false);
        ScaffoldMessenger.of(context).showSnackBar(
          SnackBar(content: Text('Gagal mengambil foto: $e'), backgroundColor: Colors.red),
        );
      }
    }
  }

  Future<Uint8List> _burnWatermark(Uint8List bytes, String target) async {
    final codec = await ui.instantiateImageCodec(bytes);
    final frame = await codec.getNextFrame();
    final srcImg = frame.image;

    double targetWidth = srcImg.width.toDouble();
    double targetHeight = srcImg.height.toDouble();
    const double maxDimension = 1024.0;
    if (targetWidth > maxDimension || targetHeight > maxDimension) {
      if (targetWidth > targetHeight) {
        targetHeight = (targetHeight * maxDimension) / targetWidth;
        targetWidth = maxDimension;
      } else {
        targetWidth = (targetWidth * maxDimension) / targetHeight;
        targetHeight = maxDimension;
      }
    }

    final recorder = ui.PictureRecorder();
    final canvas = Canvas(recorder, Rect.fromLTWH(0, 0, targetWidth, targetHeight));

    canvas.drawImageRect(
      srcImg,
      Rect.fromLTWH(0, 0, srcImg.width.toDouble(), srcImg.height.toDouble()),
      Rect.fromLTWH(0, 0, targetWidth, targetHeight),
      Paint(),
    );

    final titleLabel = target == 'start' ? 'KM BERANGKAT (REVISI)' : 'KM PULANG (REVISI)';
    const padding = 16.0;
    final spans = <TextSpan>[
      TextSpan(
        text: "$titleLabel\n",
        style: TextStyle(fontSize: targetWidth * 0.04, fontWeight: FontWeight.bold, color: Colors.white),
      ),
      if (_currentGpsAddress.isNotEmpty)
        TextSpan(
          text: "$_currentGpsAddress\n",
          style: TextStyle(fontSize: targetWidth * 0.03, color: Colors.white70),
        ),
      TextSpan(
        text: _currentGpsTimestamp,
        style: TextStyle(fontSize: targetWidth * 0.03, color: Colors.yellowAccent),
      ),
    ];

    final tp = TextPainter(
      text: TextSpan(children: spans),
      textDirection: ui.TextDirection.ltr,
      textAlign: TextAlign.center,
    );
    tp.layout(maxWidth: targetWidth - (padding * 2));

    final textHeight = tp.size.height;
    final boxHeight = textHeight + (padding * 2);
    final boxY = targetHeight - boxHeight;

    final barPaint = Paint()..color = Colors.black.withOpacity(0.65);
    canvas.drawRect(Rect.fromLTWH(0, boxY, targetWidth, boxHeight), barPaint);
    tp.paint(canvas, Offset(padding, boxY + padding));

    final picture = recorder.endRecording();
    final int outWidth = targetWidth.toInt();
    final int outHeight = targetHeight.toInt();
    final resultImg = await picture.toImage(outWidth, outHeight);
    final byteData = await resultImg.toByteData(format: ui.ImageByteFormat.rawRgba);

    srcImg.dispose();
    resultImg.dispose();
    picture.dispose();

    final rawBytes = byteData!.buffer.asUint8List();
    return await ImageCompressHelper.encodeRawRgbaToJpg(
      rawBytes,
      width: outWidth,
      height: outHeight,
      quality: 75,
    );
  }

  Future<String?> _uploadPhoto(Uint8List bytes, String filenamePrefix) async {
    try {
      final formData = FormData.fromMap({
        'file': MultipartFile.fromBytes(
          bytes,
          filename: '${filenamePrefix}_${DateTime.now().millisecondsSinceEpoch}.jpg',
        ),
      });
      final res = await GetIt.I<ApiClient>().client.post(
        'sales/upload',
        data: formData,
        options: Options(contentType: 'multipart/form-data'),
      );
      if (res.data != null && res.data['url'] != null) {
        return res.data['url'].toString();
      }
    } catch (e) {
      debugPrint('Upload error: $e');
    }
    return null;
  }

  Future<void> _submitChanges() async {
    final startText = _startKMController.text.trim();
    final endText = _endKMController.text.trim();

    if (startText.isEmpty) {
      ScaffoldMessenger.of(context).showSnackBar(
        const SnackBar(content: Text('Angka KM Berangkat tidak boleh kosong'), backgroundColor: Colors.red),
      );
      return;
    }

    final double? startKM = double.tryParse(startText);
    if (startKM == null) {
      ScaffoldMessenger.of(context).showSnackBar(
        const SnackBar(content: Text('KM Berangkat harus berupa angka'), backgroundColor: Colors.red),
      );
      return;
    }

    double? endKM;
    if (widget.editMode == 'both' && endText.isNotEmpty) {
      endKM = double.tryParse(endText);
      if (endKM == null) {
        ScaffoldMessenger.of(context).showSnackBar(
          const SnackBar(content: Text('KM Pulang harus berupa angka'), backgroundColor: Colors.red),
        );
        return;
      }
      if (endKM < startKM) {
        ScaffoldMessenger.of(context).showSnackBar(
          const SnackBar(
            content: Text('KM Pulang tidak boleh lebih kecil dari KM Berangkat'),
            backgroundColor: Colors.red,
          ),
        );
        return;
      }
    }

    setState(() => _isSubmitting = true);

    try {
      final payload = <String, dynamic>{};

      // Always pass KM values
      payload['start_km'] = startKM;
      if (endKM != null) {
        payload['end_km'] = endKM;
      }

      // Upload start photo if newly captured
      if (_newStartBytes != null) {
        final uploadedStartUrl = await _uploadPhoto(_newStartBytes!, 'edit_start');
        if (uploadedStartUrl != null && uploadedStartUrl.isNotEmpty) {
          payload['start_photo_url'] = uploadedStartUrl;
        } else if (_newStartPath != null) {
          payload['start_photo_url'] = _newStartPath;
        }
      }

      // Upload end photo if newly captured
      if (_newEndBytes != null) {
        final uploadedEndUrl = await _uploadPhoto(_newEndBytes!, 'edit_end');
        if (uploadedEndUrl != null && uploadedEndUrl.isNotEmpty) {
          payload['end_photo_url'] = uploadedEndUrl;
        } else if (_newEndPath != null) {
          payload['end_photo_url'] = _newEndPath;
        }
      }

      await _apiService.updateJourney(widget.journeyId, payload);

      if (mounted) {
        ScaffoldMessenger.of(context).showSnackBar(
          const SnackBar(
            content: Text('Data perjalanan berhasil diperbarui!'),
            backgroundColor: Colors.green,
          ),
        );
        Navigator.pop(context, true);
      }
    } catch (e) {
      if (mounted) {
        final err = e.toString().replaceAll('Exception: ', '');
        ScaffoldMessenger.of(context).showSnackBar(
          SnackBar(content: Text(err), backgroundColor: Colors.red),
        );
      }
    } finally {
      if (mounted) setState(() => _isSubmitting = false);
    }
  }

  @override
  Widget build(BuildContext context) {
    if (_isCameraActive) {
      return _buildCameraView();
    }

    final isEnded = widget.editMode == 'both';

    return Scaffold(
      backgroundColor: const Color(0xFFF8FAFC),
      appBar: AppBar(
        title: Column(
          crossAxisAlignment: CrossAxisAlignment.start,
          children: [
            Text(
              'Edit Pencatatan KM',
              style: GoogleFonts.outfit(fontWeight: FontWeight.w800, fontSize: 16),
            ),
            Text(
              widget.formattedDate,
              style: GoogleFonts.outfit(fontSize: 11, color: Colors.blueGrey.shade600, fontWeight: FontWeight.normal),
            ),
          ],
        ),
        backgroundColor: Colors.white,
        foregroundColor: const Color(0xFF0F172A),
        elevation: 0,
        leading: IconButton(
          icon: const Icon(Icons.arrow_back_rounded),
          onPressed: () => Navigator.pop(context),
        ),
      ),
      body: SingleChildScrollView(
        padding: const EdgeInsets.all(16),
        child: Column(
          crossAxisAlignment: CrossAxisAlignment.stretch,
          children: [
            // ═══ Section 1: KM Berangkat ═══
            _buildSectionCard(
              title: 'KM BERANGKAT',
              accentColor: const Color(0xFF0284C7),
              bgColor: const Color(0xFFE0F2FE),
              icon: Icons.flight_takeoff_rounded,
              child: Column(
                crossAxisAlignment: CrossAxisAlignment.stretch,
                children: [
                  Text(
                    'Angka KM Speedometer Awal',
                    style: GoogleFonts.outfit(fontSize: 12, fontWeight: FontWeight.w700, color: const Color(0xFF334155)),
                  ),
                  const SizedBox(height: 6),
                  TextField(
                    controller: _startKMController,
                    keyboardType: const TextInputType.numberWithOptions(decimal: true),
                    decoration: InputDecoration(
                      hintText: 'Contoh: 15420',
                      prefixIcon: const Icon(Icons.speed_rounded, color: Color(0xFF0284C7)),
                      filled: true,
                      fillColor: Colors.white,
                      border: OutlineInputBorder(
                        borderRadius: BorderRadius.circular(12),
                        borderSide: const BorderSide(color: Color(0xFFCBD5E1)),
                      ),
                      enabledBorder: OutlineInputBorder(
                        borderRadius: BorderRadius.circular(12),
                        borderSide: const BorderSide(color: Color(0xFFCBD5E1)),
                      ),
                      focusedBorder: OutlineInputBorder(
                        borderRadius: BorderRadius.circular(12),
                        borderSide: const BorderSide(color: Color(0xFF0284C7), width: 2),
                      ),
                    ),
                  ),
                  const SizedBox(height: 14),

                  // Photo Display
                  Text(
                    'Foto Speedometer Berangkat',
                    style: GoogleFonts.outfit(fontSize: 12, fontWeight: FontWeight.w700, color: const Color(0xFF334155)),
                  ),
                  const SizedBox(height: 6),
                  if (_newStartBytes != null)
                    Stack(
                      children: [
                        ClipRRect(
                          borderRadius: BorderRadius.circular(12),
                          child: Image.memory(
                            _newStartBytes!,
                            height: 140,
                            width: double.infinity,
                            fit: BoxFit.cover,
                          ),
                        ),
                        Positioned(
                          top: 8,
                          right: 8,
                          child: Container(
                            padding: const EdgeInsets.symmetric(horizontal: 8, vertical: 4),
                            decoration: BoxDecoration(
                              color: const Color(0xFF16A34A),
                              borderRadius: BorderRadius.circular(20),
                            ),
                            child: Row(
                              mainAxisSize: MainAxisSize.min,
                              children: [
                                const Icon(Icons.check_circle_rounded, size: 12, color: Colors.white),
                                const SizedBox(width: 4),
                                Text(
                                  'Foto Baru Dipilih',
                                  style: GoogleFonts.outfit(fontSize: 10, color: Colors.white, fontWeight: FontWeight.bold),
                                ),
                              ],
                            ),
                          ),
                        ),
                      ],
                    )
                  else if (widget.currentStartPhotoUrl != null && widget.currentStartPhotoUrl!.isNotEmpty)
                    _buildPhotoThumbnail(widget.currentStartPhotoUrl!)
                  else
                    Container(
                      height: 80,
                      decoration: BoxDecoration(
                        color: const Color(0xFFF1F5F9),
                        borderRadius: BorderRadius.circular(12),
                        border: Border.all(color: const Color(0xFFE2E8F0)),
                      ),
                      child: Center(
                        child: Text(
                          'Belum ada foto tersimpan',
                          style: GoogleFonts.outfit(fontSize: 12, color: const Color(0xFF94A3B8)),
                        ),
                      ),
                    ),

                  const SizedBox(height: 10),
                  OutlinedButton.icon(
                    onPressed: () => _startCameraFor('start'),
                    icon: const Icon(Icons.camera_alt_rounded, size: 16, color: Color(0xFF0284C7)),
                    label: Text(
                      _newStartBytes != null ? 'Ambil Ulang Foto Berangkat' : 'Ganti Foto Berangkat',
                      style: GoogleFonts.outfit(fontSize: 12, fontWeight: FontWeight.w700, color: const Color(0xFF0284C7)),
                    ),
                    style: OutlinedButton.styleFrom(
                      padding: const EdgeInsets.symmetric(vertical: 10),
                      side: const BorderSide(color: Color(0xFF0284C7)),
                      shape: RoundedRectangleBorder(borderRadius: BorderRadius.circular(10)),
                    ),
                  ),
                ],
              ),
            ),

            const SizedBox(height: 16),

            // ═══ Section 2: KM Pulang (hanya jika ENDED / status keduanya) ═══
            if (isEnded)
              _buildSectionCard(
                title: 'KM PULANG',
                accentColor: const Color(0xFF16A34A),
                bgColor: const Color(0xFFDCFCE7),
                icon: Icons.flight_land_rounded,
                child: Column(
                  crossAxisAlignment: CrossAxisAlignment.stretch,
                  children: [
                    Text(
                      'Angka KM Speedometer Akhir',
                      style: GoogleFonts.outfit(fontSize: 12, fontWeight: FontWeight.w700, color: const Color(0xFF334155)),
                    ),
                    const SizedBox(height: 6),
                    TextField(
                      controller: _endKMController,
                      keyboardType: const TextInputType.numberWithOptions(decimal: true),
                      decoration: InputDecoration(
                        hintText: 'Contoh: 15480',
                        prefixIcon: const Icon(Icons.speed_rounded, color: Color(0xFF16A34A)),
                        filled: true,
                        fillColor: Colors.white,
                        border: OutlineInputBorder(
                          borderRadius: BorderRadius.circular(12),
                          borderSide: const BorderSide(color: Color(0xFFCBD5E1)),
                        ),
                        enabledBorder: OutlineInputBorder(
                          borderRadius: BorderRadius.circular(12),
                          borderSide: const BorderSide(color: Color(0xFFCBD5E1)),
                        ),
                        focusedBorder: OutlineInputBorder(
                          borderRadius: BorderRadius.circular(12),
                          borderSide: const BorderSide(color: Color(0xFF16A34A), width: 2),
                        ),
                      ),
                    ),
                    const SizedBox(height: 14),

                    // Photo Display
                    Text(
                      'Foto Speedometer Pulang',
                      style: GoogleFonts.outfit(fontSize: 12, fontWeight: FontWeight.w700, color: const Color(0xFF334155)),
                    ),
                    const SizedBox(height: 6),
                    if (_newEndBytes != null)
                      Stack(
                        children: [
                          ClipRRect(
                            borderRadius: BorderRadius.circular(12),
                            child: Image.memory(
                              _newEndBytes!,
                              height: 140,
                              width: double.infinity,
                              fit: BoxFit.cover,
                            ),
                          ),
                          Positioned(
                            top: 8,
                            right: 8,
                            child: Container(
                              padding: const EdgeInsets.symmetric(horizontal: 8, vertical: 4),
                              decoration: BoxDecoration(
                                color: const Color(0xFF16A34A),
                                borderRadius: BorderRadius.circular(20),
                              ),
                              child: Row(
                                mainAxisSize: MainAxisSize.min,
                                children: [
                                  const Icon(Icons.check_circle_rounded, size: 12, color: Colors.white),
                                  const SizedBox(width: 4),
                                  Text(
                                    'Foto Baru Dipilih',
                                    style: GoogleFonts.outfit(fontSize: 10, color: Colors.white, fontWeight: FontWeight.bold),
                                  ),
                                ],
                              ),
                            ),
                          ),
                        ],
                      )
                    else if (widget.currentEndPhotoUrl != null && widget.currentEndPhotoUrl!.isNotEmpty)
                      _buildPhotoThumbnail(widget.currentEndPhotoUrl!)
                    else
                      Container(
                        height: 80,
                        decoration: BoxDecoration(
                          color: const Color(0xFFF1F5F9),
                          borderRadius: BorderRadius.circular(12),
                          border: Border.all(color: const Color(0xFFE2E8F0)),
                        ),
                        child: Center(
                          child: Text(
                            'Belum ada foto tersimpan',
                            style: GoogleFonts.outfit(fontSize: 12, color: const Color(0xFF94A3B8)),
                          ),
                        ),
                      ),

                    const SizedBox(height: 10),
                    OutlinedButton.icon(
                      onPressed: () => _startCameraFor('end'),
                      icon: const Icon(Icons.camera_alt_rounded, size: 16, color: Color(0xFF16A34A)),
                      label: Text(
                        _newEndBytes != null ? 'Ambil Ulang Foto Pulang' : 'Ganti Foto Pulang',
                        style: GoogleFonts.outfit(fontSize: 12, fontWeight: FontWeight.w700, color: const Color(0xFF16A34A)),
                      ),
                      style: OutlinedButton.styleFrom(
                        padding: const EdgeInsets.symmetric(vertical: 10),
                        side: const BorderSide(color: Color(0xFF16A34A)),
                        shape: RoundedRectangleBorder(borderRadius: BorderRadius.circular(10)),
                      ),
                    ),
                  ],
                ),
              ),

            const SizedBox(height: 24),

            // ═══ Save Button ═══
            ElevatedButton(
              onPressed: _isSubmitting ? null : _submitChanges,
              style: ElevatedButton.styleFrom(
                backgroundColor: AppColors.blueSales,
                padding: const EdgeInsets.symmetric(vertical: 16),
                shape: RoundedRectangleBorder(borderRadius: BorderRadius.circular(14)),
                elevation: 0,
              ),
              child: _isSubmitting
                  ? const SizedBox(
                      width: 22,
                      height: 22,
                      child: CircularProgressIndicator(strokeWidth: 2.5, color: Colors.white),
                    )
                  : Text(
                      'SIMPAN PERUBAHAN',
                      style: GoogleFonts.outfit(
                        fontSize: 15,
                        fontWeight: FontWeight.w800,
                        color: Colors.white,
                        letterSpacing: 0.5,
                      ),
                    ),
            ),
          ],
        ),
      ),
    );
  }

  Widget _buildSectionCard({
    required String title,
    required Color accentColor,
    required Color bgColor,
    required IconData icon,
    required Widget child,
  }) {
    return Container(
      padding: const EdgeInsets.all(16),
      decoration: BoxDecoration(
        color: Colors.white,
        borderRadius: BorderRadius.circular(18),
        border: Border.all(color: const Color(0xFFE2E8F0)),
        boxShadow: [
          BoxShadow(
            color: Colors.black.withOpacity(0.02),
            blurRadius: 8,
            offset: const Offset(0, 2),
          ),
        ],
      ),
      child: Column(
        crossAxisAlignment: CrossAxisAlignment.stretch,
        children: [
          Row(
            children: [
              Container(
                padding: const EdgeInsets.all(6),
                decoration: BoxDecoration(
                  color: bgColor,
                  borderRadius: BorderRadius.circular(8),
                ),
                child: Icon(icon, size: 16, color: accentColor),
              ),
              const SizedBox(width: 8),
              Text(
                title,
                style: GoogleFonts.outfit(
                  fontSize: 12,
                  fontWeight: FontWeight.w800,
                  color: accentColor,
                  letterSpacing: 0.5,
                ),
              ),
            ],
          ),
          const Divider(height: 20, color: Color(0xFFF1F5F9)),
          child,
        ],
      ),
    );
  }

  Widget _buildCameraView() {
    final title = _cameraTarget == 'start' ? 'GANTI FOTO BERANGKAT' : 'GANTI FOTO PULANG';

    return Scaffold(
      backgroundColor: Colors.black,
      body: Stack(
        children: [
          if (_isCameraInitialized && _cameraController != null)
            SizedBox.expand(child: CameraPreview(_cameraController!))
          else
            const Center(child: CircularProgressIndicator(color: Colors.white)),

          // Header
          Positioned(
            top: 0,
            left: 0,
            right: 0,
            child: Container(
              padding: const EdgeInsets.fromLTRB(16, 50, 16, 20),
              decoration: BoxDecoration(
                gradient: LinearGradient(
                  begin: Alignment.topCenter,
                  end: Alignment.bottomCenter,
                  colors: [Colors.black.withOpacity(0.7), Colors.transparent],
                ),
              ),
              child: Row(
                crossAxisAlignment: CrossAxisAlignment.start,
                children: [
                  IconButton(
                    icon: const Icon(Icons.close_rounded, color: Colors.white, size: 28),
                    onPressed: () {
                      _cameraController?.dispose();
                      _cameraController = null;
                      setState(() => _isCameraActive = false);
                    },
                  ),
                  Expanded(
                    child: Column(
                      children: [
                        Text(
                          title,
                          style: GoogleFonts.outfit(
                            color: AppColors.blueSales,
                            fontSize: 15,
                            fontWeight: FontWeight.w900,
                          ),
                        ),
                        Text(
                          'Arahkan kamera ke speedometer / odometer',
                          style: GoogleFonts.outfit(color: Colors.white70, fontSize: 11),
                        ),
                      ],
                    ),
                  ),
                  const SizedBox(width: 48),
                ],
              ),
            ),
          ),

          // GPS Info overlay
          Positioned(
            bottom: 130,
            left: 0,
            right: 0,
            child: Container(
              margin: const EdgeInsets.symmetric(horizontal: 16),
              padding: const EdgeInsets.symmetric(horizontal: 16, vertical: 10),
              decoration: BoxDecoration(
                color: Colors.black.withOpacity(0.6),
                borderRadius: BorderRadius.circular(16),
              ),
              child: Column(
                children: [
                  Text(
                    title,
                    style: GoogleFonts.outfit(color: Colors.white, fontWeight: FontWeight.w900, fontSize: 12),
                  ),
                  if (_currentGpsAddress.isNotEmpty)
                    Text(
                      _currentGpsAddress,
                      style: GoogleFonts.outfit(color: Colors.white70, fontSize: 10),
                      textAlign: TextAlign.center,
                      maxLines: 2,
                      overflow: TextOverflow.ellipsis,
                    ),
                  Text(_currentGpsTimestamp, style: GoogleFonts.outfit(color: Colors.yellowAccent, fontSize: 10)),
                ],
              ),
            ),
          ),

          // Capture button
          Positioned(
            bottom: 40,
            left: 0,
            right: 0,
            child: Row(
              mainAxisAlignment: MainAxisAlignment.spaceEvenly,
              children: [
                const SizedBox(width: 48),
                GestureDetector(
                  onTap: _capturePhoto,
                  child: Container(
                    width: 76,
                    height: 76,
                    decoration: BoxDecoration(
                      shape: BoxShape.circle,
                      color: Colors.white,
                      border: Border.all(color: Colors.white.withOpacity(0.5), width: 4),
                      boxShadow: [BoxShadow(color: Colors.white.withOpacity(0.3), blurRadius: 20)],
                    ),
                    child: _isTakingPhoto
                        ? const Padding(
                            padding: EdgeInsets.all(22),
                            child: CircularProgressIndicator(color: Colors.black, strokeWidth: 3),
                          )
                        : const Icon(Icons.camera_alt_rounded, size: 34, color: Colors.black),
                  ),
                ),
                if (_cameras.length > 1)
                  Container(
                    decoration: BoxDecoration(color: Colors.black.withOpacity(0.5), shape: BoxShape.circle),
                    child: IconButton(
                      icon: const Icon(Icons.flip_camera_ios_rounded, color: Colors.white, size: 26),
                      onPressed: _switchCamera,
                    ),
                  )
                else
                  const SizedBox(width: 48),
              ],
            ),
          ),
        ],
      ),
    );
  }
}
