import 'package:hris_app/core/network/api_client.dart';
import 'package:get_it/get_it.dart';
import 'dart:io';
import 'dart:convert';
import 'package:flutter/foundation.dart';
import 'dart:ui' as ui;
import 'package:flutter/material.dart';
import 'package:google_fonts/google_fonts.dart';
import 'package:camera/camera.dart';
import 'package:geolocator/geolocator.dart';
import 'package:intl/intl.dart';
import 'package:geocoding/geocoding.dart';
import 'package:dio/dio.dart';
import 'package:path_provider/path_provider.dart';
import 'package:hris_app/core/theme/app_colors.dart';
import 'package:hris_app/features/sales/data/services/sales_api_service.dart';
import 'package:hris_app/core/utils/image_compress_helper.dart';
import 'package:hris_app/features/sync/presentation/bloc/sync_bloc.dart';

class EndJourneyPage extends StatefulWidget {
  const EndJourneyPage({super.key});

  @override
  State<EndJourneyPage> createState() => _EndJourneyPageState();
}

class _EndJourneyPageState extends State<EndJourneyPage> {
  CameraController? _controller;
  List<CameraDescription> _cameras = [];
  bool _isInitialized = false;
  bool _isTaking = false;
  bool _isSubmitting = false;
  
  String? _capturedPath;
  Uint8List? _capturedBytes;
  
  String _address = 'Mengambil lokasi...';
  String _timestamp = '';
  double _lat = 0.0, _lng = 0.0;
  bool _isLocationLoading = true;
  
  final TextEditingController _kmController = TextEditingController();
  final SalesApiService _apiService = SalesApiService(apiClient: GetIt.I<ApiClient>());

  @override
  void initState() {
    super.initState();
    _initCamera();
    _getLocation();
    _timestamp = DateFormat('dd/MM/yyyy HH:mm:ss').format(DateTime.now());
  }

  int _selectedCameraIdx = -1;

  Future<void> _initCamera({int? cameraIndex}) async {
    _cameras = await availableCameras();
    if (_cameras.isEmpty) return;

    if (cameraIndex != null) {
      _selectedCameraIdx = cameraIndex;
    } else {
      final backIdx = _cameras.indexWhere((c) => c.lensDirection == CameraLensDirection.back);
      _selectedCameraIdx = (backIdx != -1) ? backIdx : 0;
    }

    _controller = CameraController(
        _cameras[_selectedCameraIdx], ResolutionPreset.high,
        enableAudio: false);
    await _controller!.initialize();
    try {
      await _controller!.setFlashMode(FlashMode.off);
    } catch (_) {}
    if (mounted) setState(() => _isInitialized = true);
  }

  Future<void> _switchCamera() async {
    if (_cameras.length < 2) return;
    await Future.delayed(const Duration(milliseconds: 100));
    setState(() => _isInitialized = false);
    await _controller?.dispose();
    _selectedCameraIdx = (_selectedCameraIdx + 1) % _cameras.length;
    await _initCamera(cameraIndex: _selectedCameraIdx);
  }

  Future<void> _getLocation() async {
    setState(() => _isLocationLoading = true);
    try {
      bool serviceEnabled = await Geolocator.isLocationServiceEnabled();
      if (!serviceEnabled) {
        if (mounted) setState(() { _address = 'GPS nonaktif'; _isLocationLoading = false; });
        return;
      }
      
      LocationPermission perm = await Geolocator.checkPermission();
      if (perm == LocationPermission.denied) perm = await Geolocator.requestPermission();
      if (perm == LocationPermission.denied || perm == LocationPermission.deniedForever) {
        if (mounted) setState(() { _address = 'Izin lokasi ditolak'; _isLocationLoading = false; });
        return;
      }
      
      final pos = await Geolocator.getCurrentPosition(desiredAccuracy: LocationAccuracy.high);
      String readableAddress = 'GPS: ${pos.latitude.toStringAsFixed(4)}, ${pos.longitude.toStringAsFixed(4)}';
      if (mounted) {
        setState(() {
          _lat = pos.latitude;
          _lng = pos.longitude;
          _address = readableAddress;
          _isLocationLoading = false;
        });
      }
      
      try {
        List<Placemark> placemarks = await placemarkFromCoordinates(pos.latitude, pos.longitude).timeout(const Duration(seconds: 2));
        if (placemarks.isNotEmpty) {
          Placemark place = placemarks[0];
          List<String> components = [
            place.street ?? '',
            place.subLocality ?? '',
            place.locality ?? '',
            place.subAdministrativeArea ?? '',
          ].where((s) => s.isNotEmpty && !s.contains('null')).toList();
          
          if (components.isNotEmpty && mounted) {
            setState(() {
              _address = '${components.join(', ')} (GPS: ${pos.latitude.toStringAsFixed(5)}, ${pos.longitude.toStringAsFixed(5)})';
            });
          }
        }
      } catch (e) {
        debugPrint('Geocoding error: $e');
      }
    } catch (_) {
      if (mounted) setState(() { _address = ''; _isLocationLoading = false; });
    }
  }

  Future<void> _takePhoto() async {
    if (_controller == null || !_isInitialized || _isTaking) return;
    if (_isLocationLoading) {
      ScaffoldMessenger.of(context).showSnackBar(
        const SnackBar(content: Text('Sedang mengunci koordinat GPS, harap tunggu sebentar...'), backgroundColor: Colors.orange),
      );
      return;
    }

    setState(() => _isTaking = true);
    try {
      final xfile = await _controller!.takePicture();
      final bytes = await xfile.readAsBytes();
      final resultBytes = await _burnWatermark(bytes);

      String pathOrBase64 = '';
      if (kIsWeb) {
        pathOrBase64 = 'data:image/jpeg;base64,${base64Encode(resultBytes)}';
      } else {
        final dir = await getApplicationDocumentsDirectory();
        final outPath = '${dir.path}/journey_end_${DateTime.now().millisecondsSinceEpoch}.jpg';
        await File(outPath).writeAsBytes(resultBytes);
        pathOrBase64 = outPath;
      }

      setState(() {
        _capturedBytes = resultBytes;
        _capturedPath = pathOrBase64;
      });
    } catch (e) {
      debugPrint('Camera error: $e');
    } finally {
      if (mounted) setState(() => _isTaking = false);
    }
  }

  Future<Uint8List> _burnWatermark(Uint8List bytes) async {
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

    final padding = 16.0;
    final spans = <TextSpan>[
      TextSpan(
        text: "KM PULANG\n",
        style: TextStyle(fontSize: targetWidth * 0.04, fontWeight: FontWeight.bold, color: Colors.white),
      ),
      TextSpan(
        text: "$_address\n",
        style: TextStyle(fontSize: targetWidth * 0.03, color: Colors.white70),
      ),
      TextSpan(
        text: _timestamp,
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

    final barPaint = Paint()..color = Colors.black.withOpacity(0.6);
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

  Future<void> _submitJourney() async {
    if (_kmController.text.isEmpty) {
      ScaffoldMessenger.of(context).showSnackBar(
        const SnackBar(content: Text('Masukkan angka KM pada speedometer'), backgroundColor: Colors.red),
      );
      return;
    }

    final double? km = double.tryParse(_kmController.text);
    if (km == null) {
      ScaffoldMessenger.of(context).showSnackBar(
        const SnackBar(content: Text('KM harus berupa angka'), backgroundColor: Colors.red),
      );
      return;
    }

    setState(() => _isSubmitting = true);
    try {
      String uploadedUrl = '';
      if (_capturedBytes != null && _capturedBytes!.isNotEmpty) {
        try {
          final formData = FormData.fromMap({
            'file': MultipartFile.fromBytes(
              _capturedBytes!,
              filename: 'end_journey_${DateTime.now().millisecondsSinceEpoch}.png',
            ),
          });
          final uploadRes = await GetIt.I<ApiClient>().client.post(
            'sales/upload',
            data: formData,
            options: Options(
              contentType: 'multipart/form-data',
              sendTimeout: const Duration(seconds: 6),
              receiveTimeout: const Duration(seconds: 6),
            ),
          );
          if (uploadRes.data != null && uploadRes.data['url'] != null) {
            uploadedUrl = uploadRes.data['url'];
          }
        } catch (uploadErr) {
          debugPrint('Failed to upload end photo to server: $uploadErr');
        }
      }

      final Map<String, dynamic> data = {
        'end_km': km,
        'end_location': _address,
        'end_photo_url': uploadedUrl.isNotEmpty ? uploadedUrl : (_capturedPath ?? ''),
        'end_latitude': _lat,
        'end_longitude': _lng,
      };

      await _apiService.endJourney(data);
      try {
        GetIt.I<SyncBloc>().add(const SyncDataRequested(force: true));
      } catch (_) {}
      
      if (mounted) {
        ScaffoldMessenger.of(context).showSnackBar(
          const SnackBar(
            content: Text('KM Pulang berhasil dicatat! Pekerjaan hari ini selesai.'),
            backgroundColor: Colors.green,
          ),
        );
        Navigator.pop(context, true);
      }
    } catch (e) {
      if (mounted) {
        final errorMsg = e.toString().replaceAll('Exception: ', '');
        ScaffoldMessenger.of(context).showSnackBar(
          SnackBar(
            content: Text(errorMsg),
            backgroundColor: Colors.red,
          ),
        );
      }
    } finally {
      if (mounted) setState(() => _isSubmitting = false);
    }
  }

  @override
  void dispose() {
    _controller?.dispose();
    _kmController.dispose();
    super.dispose();
  }

  @override
  Widget build(BuildContext context) {
    return Scaffold(
      backgroundColor: Colors.black,
      body: _capturedBytes != null ? _buildPreview() : _buildCamera(),
    );
  }

  Widget _buildCamera() {
    return Stack(
      children: [
        if (_isInitialized && _controller != null)
          SizedBox.expand(child: CameraPreview(_controller!))
        else
          const Center(child: CircularProgressIndicator(color: Colors.white)),
          
        Positioned(
          top: 0, left: 0, right: 0,
          child: Container(
            padding: const EdgeInsets.fromLTRB(16, 60, 16, 20),
            decoration: BoxDecoration(gradient: LinearGradient(begin: Alignment.topCenter, end: Alignment.bottomCenter, colors: [Colors.black.withOpacity(0.7), Colors.transparent])),
            child: Row(
              crossAxisAlignment: CrossAxisAlignment.start,
              children: [
                IconButton(
                  icon: const Icon(Icons.arrow_back_ios_rounded, color: Colors.white),
                  onPressed: () => Navigator.pop(context),
                ),
                Expanded(
                  child: Column(
                    children: [
                      Text('CATAT KM PULANG', style: GoogleFonts.outfit(color: AppColors.blueSales, fontSize: 16, fontWeight: FontWeight.w900)),
                      Text('Ambil foto speedometer / odometer kendaraan', style: GoogleFonts.outfit(color: Colors.white70, fontSize: 12)),
                    ],
                  ),
                ),
                const SizedBox(width: 48),
              ],
            ),
          ),
        ),

        Positioned(
          bottom: 130, left: 0, right: 0,
          child: Container(
            margin: const EdgeInsets.symmetric(horizontal: 16),
            padding: const EdgeInsets.symmetric(horizontal: 16, vertical: 12),
            decoration: BoxDecoration(color: Colors.black.withOpacity(0.6), borderRadius: BorderRadius.circular(16)),
            child: Column(
              children: [
                Text('KM PULANG', style: GoogleFonts.outfit(color: Colors.white, fontWeight: FontWeight.w900, fontSize: 13)),
                Text(_address, style: GoogleFonts.outfit(color: Colors.white70, fontSize: 11), textAlign: TextAlign.center, maxLines: 2, overflow: TextOverflow.ellipsis),
                Text(_timestamp, style: GoogleFonts.outfit(color: Colors.yellowAccent, fontSize: 11)),
              ],
            ),
          ),
        ),

        Positioned(
          bottom: 40, left: 0, right: 0,
          child: Row(
            mainAxisAlignment: MainAxisAlignment.spaceEvenly,
            children: [
              const SizedBox(width: 48),
              GestureDetector(
                onTap: _takePhoto,
                child: Container(
                  width: 80, height: 80,
                  decoration: BoxDecoration(
                    shape: BoxShape.circle, color: Colors.white,
                    border: Border.all(color: Colors.white.withOpacity(0.5), width: 4),
                    boxShadow: [BoxShadow(color: Colors.white.withOpacity(0.3), blurRadius: 20)],
                  ),
                  child: _isTaking
                      ? const Padding(padding: EdgeInsets.all(24), child: CircularProgressIndicator(color: Colors.black, strokeWidth: 3))
                      : const Icon(Icons.camera_alt_rounded, size: 36, color: Colors.black),
                ),
              ),
              if (_cameras.length > 1)
                Container(
                  decoration: BoxDecoration(color: Colors.black.withOpacity(0.5), shape: BoxShape.circle),
                  child: IconButton(icon: const Icon(Icons.flip_camera_ios_rounded, color: Colors.white, size: 28), onPressed: _switchCamera),
                )
              else
                const SizedBox(width: 48),
            ],
          ),
        ),
      ],
    );
  }

  Widget _buildPreview() {
    return Scaffold(
      backgroundColor: Colors.white,
      appBar: AppBar(
        title: Text('Catat KM Pulang & Selesai', style: GoogleFonts.outfit(color: Colors.white)),
        backgroundColor: AppColors.blueSales,
        iconTheme: const IconThemeData(color: Colors.white),
        leading: IconButton(
          icon: const Icon(Icons.arrow_back),
          onPressed: () => setState(() {
            _capturedPath = null;
            _capturedBytes = null;
          }),
        ),
      ),
      body: SingleChildScrollView(
        padding: const EdgeInsets.all(16),
        child: Column(
          crossAxisAlignment: CrossAxisAlignment.stretch,
          children: [
            ClipRRect(
              borderRadius: BorderRadius.circular(16),
              child: Image.memory(_capturedBytes!, height: 300, fit: BoxFit.cover),
            ),
            const SizedBox(height: 24),
            Text('Angka KM Odometer Pulang', style: GoogleFonts.outfit(fontWeight: FontWeight.bold, fontSize: 16)),
            const SizedBox(height: 8),
            TextField(
              controller: _kmController,
              keyboardType: const TextInputType.numberWithOptions(decimal: true),
              decoration: InputDecoration(
                hintText: 'Masukkan angka KM akhir (misal: 15450)',
                border: OutlineInputBorder(borderRadius: BorderRadius.circular(12)),
                prefixIcon: const Icon(Icons.speed),
                filled: true,
                fillColor: Colors.grey[100],
              ),
            ),
            const SizedBox(height: 40),
            ElevatedButton(
              onPressed: _isSubmitting ? null : _submitJourney,
              style: ElevatedButton.styleFrom(
                backgroundColor: AppColors.blueSales,
                padding: const EdgeInsets.symmetric(vertical: 16),
                shape: RoundedRectangleBorder(borderRadius: BorderRadius.circular(12)),
              ),
              child: _isSubmitting 
                ? const SizedBox(width: 24, height: 24, child: CircularProgressIndicator(color: Colors.white))
                : Text('SIMPAN KM PULANG', style: GoogleFonts.outfit(color: Colors.white, fontWeight: FontWeight.bold, fontSize: 16)),
            ),
          ],
        ),
      ),
    );
  }
}
