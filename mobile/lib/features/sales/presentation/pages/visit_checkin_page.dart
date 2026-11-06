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
import 'package:hris_app/features/sales/data/models/store_model.dart';
import 'package:hris_app/features/sales/presentation/pages/visit_transaction_page.dart';
import 'package:hris_app/core/database/database.dart';
import 'package:hris_app/features/sync/presentation/bloc/sync_bloc.dart';
import 'package:hris_app/injection.dart' as di;
import 'package:drift/drift.dart' hide Column;
import 'package:path_provider/path_provider.dart';
import 'package:geocoding/geocoding.dart';
import 'package:dio/dio.dart';
import 'package:hris_app/core/utils/image_compress_helper.dart';

class VisitCheckinPage extends StatefulWidget {
  final StoreModel store;
  const VisitCheckinPage({super.key, required this.store});

  @override
  State<VisitCheckinPage> createState() => _VisitCheckinPageState();
}

class _VisitCheckinPageState extends State<VisitCheckinPage> {
  CameraController? _controller;
  List<CameraDescription> _cameras = [];
  bool _isInitialized = false;
  bool _isTaking = false;
  String? _capturedPath;
  Uint8List? _capturedBytes;
  String _address = 'Mengambil lokasi...';
  String _timestamp = '';
  double _lat = 0.0, _lng = 0.0;
  double _distanceInMeters = -1.0;
  bool _isLocationLoading = true;
  bool _isSubmitting = false;

  @override
  void initState() {
    super.initState();
    _initCamera();
    _getLocation();
    _timestamp = DateFormat('dd/MM/yyyy HH:mm:ss').format(DateTime.now());
  }

  int _selectedCameraIdx = -1;
  double _minZoom = 1.0;
  double _maxZoom = 1.0;
  double _currentZoom = 1.0;
  double _baseZoom = 1.0;
  bool _showFramingGuide = true;
  bool _guideFaceOnLeft = true;

  Future<void> _initCamera({int? cameraIndex}) async {
    _cameras = await availableCameras();
    if (_cameras.isEmpty) return;

    if (cameraIndex != null) {
      _selectedCameraIdx = cameraIndex;
    } else {
      _selectedCameraIdx = _cameras.indexWhere(
        (c) => c.lensDirection == CameraLensDirection.front,
      );
      if (_selectedCameraIdx == -1) _selectedCameraIdx = 0;
    }

    _controller = CameraController(
      _cameras[_selectedCameraIdx],
      ResolutionPreset.high,
      enableAudio: false,
    );
    await _controller!.initialize();
    try {
      await _controller!.setFlashMode(FlashMode.off);
    } catch (_) {}

    // Initialize zoom levels (detect ultra-wide sensor if available)
    try {
      _minZoom = await _controller!.getMinZoomLevel();
      _maxZoom = await _controller!.getMaxZoomLevel();
      // Auto-set to wide angle if supported (< 1.0), otherwise default to 1.0
      _currentZoom = _minZoom < 1.0 ? _minZoom : 1.0;
      await _controller!.setZoomLevel(_currentZoom);
    } catch (e) {
      debugPrint('Zoom level setup error: $e');
      _minZoom = 1.0;
      _maxZoom = 1.0;
      _currentZoom = 1.0;
    }

    if (mounted) setState(() => _isInitialized = true);
  }

  Future<void> _setZoom(double zoom) async {
    if (_controller == null || !_isInitialized) return;
    final clamped = zoom.clamp(_minZoom, _maxZoom);
    try {
      await _controller!.setZoomLevel(clamped);
      if (mounted) setState(() => _currentZoom = clamped);
    } catch (_) {}
  }

  Future<void> _switchCamera() async {
    if (_cameras.length < 2) return;

    // Memberi jeda sejenak untuk menghindari bug MouseTracker di Flutter Web
    // saat menghapus HtmlElementView (CameraPreview) secara sinkron yang
    // membuat tampilan freeze dengan pesan error Assertion failed: mouse_tracker.dart
    await Future.delayed(const Duration(milliseconds: 100));

    setState(() => _isInitialized = false);
    await _controller?.dispose();

    _selectedCameraIdx = (_selectedCameraIdx + 1) % _cameras.length;
    await _initCamera(cameraIndex: _selectedCameraIdx);
  }

  Future<void> _getLocation() async {
    if (!mounted) return;
    setState(() => _isLocationLoading = true);
    try {
      bool serviceEnabled = await Geolocator.isLocationServiceEnabled();
      if (!serviceEnabled) {
        if (mounted) {
          if (widget.store.latitude != 0.0 && widget.store.longitude != 0.0) {
            setState(() {
              _lat = widget.store.latitude;
              _lng = widget.store.longitude;
              _distanceInMeters = 0.0;
              _address = widget.store.address;
              _isLocationLoading = false;
            });
          } else {
            setState(() {
              _address = 'GPS nonaktif. Harap aktifkan GPS.';
              _isLocationLoading = false;
            });
          }
        }
        return;
      }

      LocationPermission perm = await Geolocator.checkPermission();
      if (perm == LocationPermission.denied) {
        perm = await Geolocator.requestPermission();
      }
      if (perm == LocationPermission.denied ||
          perm == LocationPermission.deniedForever) {
        if (mounted) {
          if (widget.store.latitude != 0.0 && widget.store.longitude != 0.0) {
            setState(() {
              _lat = widget.store.latitude;
              _lng = widget.store.longitude;
              _distanceInMeters = 0.0;
              _address = widget.store.address;
              _isLocationLoading = false;
            });
          } else {
            setState(() {
              _address = 'Izin lokasi ditolak';
              _isLocationLoading = false;
            });
          }
        }
        return;
      }

      Position? pos;
      try {
        pos = await Geolocator.getCurrentPosition(
          desiredAccuracy: LocationAccuracy.high,
          timeLimit: const Duration(seconds: 8),
        );
      } catch (e) {
        debugPrint(
          'VisitCheckin: getCurrentPosition failed ($e), trying lastKnownPosition...',
        );
        try {
          pos = await Geolocator.getLastKnownPosition();
        } catch (_) {}
      }

      // Validasi Anti-Fake GPS / Mock Location
      if (pos != null && pos.isMocked) {
        if (mounted) {
          setState(() {
            _address = 'Terdeteksi Fake GPS / Mock Location!';
            _isLocationLoading = false;
          });
          ScaffoldMessenger.of(context).showSnackBar(
            const SnackBar(
              content: Text(
                'Peringatan: Terdeteksi Fake GPS / Mock Location! Matikan aplikasi pemalsu lokasi.',
              ),
              backgroundColor: Colors.red,
              duration: Duration(seconds: 5),
            ),
          );
        }
        return;
      }

      // If still null and store coordinates are available, fallback to store coordinates
      if (pos == null) {
        if (widget.store.latitude != 0.0 && widget.store.longitude != 0.0) {
          debugPrint('VisitCheckin: Using store coordinates as fallback');
          pos = Position(
            latitude: widget.store.latitude,
            longitude: widget.store.longitude,
            timestamp: DateTime.now(),
            accuracy: 10.0,
            altitude: 0.0,
            altitudeAccuracy: 0.0,
            heading: 0.0,
            headingAccuracy: 0.0,
            speed: 0.0,
            speedAccuracy: 0.0,
          );
        }
      }

      if (pos == null) {
        if (mounted) {
          setState(() {
            _address =
                'Gagal mengunci GPS. Ketuk tombol reload untuk mencoba lagi.';
            _isLocationLoading = false;
          });
        }
        return;
      }

      double calculatedDistance = -1.0;
      if (widget.store.latitude != 0.0 && widget.store.longitude != 0.0) {
        calculatedDistance = Geolocator.distanceBetween(
          pos.latitude,
          pos.longitude,
          widget.store.latitude,
          widget.store.longitude,
        );
      }

      // Unlock GPS immediately so salesman can take photo without waiting for geocoding when offline
      if (mounted) {
        setState(() {
          _lat = pos!.latitude;
          _lng = pos.longitude;
          _distanceInMeters = calculatedDistance;
          _address = widget.store.address.isNotEmpty
              ? widget.store.address
              : '';
          _isLocationLoading = false;
        });
      }

      String readableAddress =
          'GPS: ${pos.latitude.toStringAsFixed(4)}, ${pos.longitude.toStringAsFixed(4)}';
      try {
        try {
          await setLocaleIdentifier("id_ID");
        } catch (_) {}

        List<Placemark>? placemarks;
        try {
          placemarks = await placemarkFromCoordinates(
            pos.latitude,
            pos.longitude,
          ).timeout(const Duration(seconds: 2));
        } catch (e) {
          debugPrint('Geocoding error: $e');
        }

        if (placemarks != null && placemarks.isNotEmpty) {
          Placemark place = placemarks[0];
          List<String> components = [
            place.street ?? '',
            place.subLocality ?? '',
            place.locality ?? '',
            place.subAdministrativeArea ?? '',
          ].where((s) => s.isNotEmpty && !s.contains('null')).toList();

          if (components.isNotEmpty) {
            readableAddress =
                '${components.join(', ')} (GPS: ${pos.latitude.toStringAsFixed(5)}, ${pos.longitude.toStringAsFixed(5)})';
          }
        } else {
          // SYSTEM GEOCODING FAILED -> FALLBACK TO NOMINATIM (OSM) with short 2s timeout
          try {
            final dio = GetIt.I<ApiClient>().client;
            final response = await dio
                .get(
                  'https://nominatim.openstreetmap.org/reverse',
                  queryParameters: {
                    'lat': pos.latitude,
                    'lon': pos.longitude,
                    'format': 'json',
                    'addressdetails': 1,
                  },
                  options: Options(
                    headers: {'User-Agent': 'KarsaHRMobile/1.0'},
                  ),
                )
                .timeout(const Duration(seconds: 2));

            if (response.statusCode == 200 && response.data != null) {
              final addr = response.data['address'];
              if (addr != null) {
                final List<String> osmComponents =
                    [
                          addr['house_number'] ??
                              addr['house_name'] ??
                              addr['building'] ??
                              '',
                          addr['road'] ??
                              addr['amenity'] ??
                              addr['industrial'] ??
                              '',
                          addr['neighbourhood'] ??
                              addr['suburb'] ??
                              addr['city_block'] ??
                              addr['quarter'] ??
                              '',
                          addr['village'] ??
                              addr['hamlet'] ??
                              addr['town'] ??
                              addr['city_district'] ??
                              addr['city'] ??
                              addr['municipality'] ??
                              '',
                          addr['county'] ?? '',
                          addr['postcode'] ?? '',
                        ]
                        .where((s) => s != null && s.toString().isNotEmpty)
                        .map((s) => s.toString())
                        .toList();

                if (osmComponents.isNotEmpty) {
                  readableAddress =
                      '${osmComponents.join(', ')} (GPS: ${pos.latitude.toStringAsFixed(5)}, ${pos.longitude.toStringAsFixed(5)})';
                } else if (response.data['display_name'] != null) {
                  readableAddress =
                      '${response.data['display_name']} (GPS: ${pos.latitude.toStringAsFixed(5)}, ${pos.longitude.toStringAsFixed(5)})';
                }
              }
            }
          } catch (_) {}
        }
      } catch (e) {
        debugPrint('Reverse geocoding ultimate error: $e');
      }

      if (mounted &&
          !readableAddress.startsWith('GPS:') &&
          readableAddress != 'Mengambil lokasi...') {
        setState(() {
          _address = readableAddress;
        });
      }
    } catch (_) {
      if (mounted) {
        if (widget.store.latitude != 0.0 && widget.store.longitude != 0.0) {
          setState(() {
            _lat = widget.store.latitude;
            _lng = widget.store.longitude;
            _distanceInMeters = 0.0;
            _address = widget.store.address;
            _isLocationLoading = false;
          });
        } else {
          setState(() {
            _address = '';
            _isLocationLoading = false;
          });
        }
      }
    }
  }

  Future<void> _showManualAddressDialog() async {
    final TextEditingController controller = TextEditingController(
      text: _address,
    );
    final newAddress = await showDialog<String>(
      context: context,
      builder: (context) {
        return AlertDialog(
          title: const Text('Input Alamat Manual'),
          content: TextField(
            controller: controller,
            decoration: const InputDecoration(
              hintText: 'Masukkan alamat toko / lokasi di sini...',
              border: OutlineInputBorder(),
            ),
            maxLines: 3,
          ),
          actions: [
            TextButton(
              onPressed: () => Navigator.pop(context),
              child: const Text('Batal'),
            ),
            ElevatedButton(
              onPressed: () => Navigator.pop(context, controller.text),
              child: const Text('Simpan'),
            ),
          ],
        );
      },
    );

    if (newAddress != null && newAddress.isNotEmpty) {
      setState(() {
        _address = newAddress.trim();
      });
    }
  }

  Future<void> _takePhoto() async {
    if (_controller == null || !_isInitialized || _isTaking) return;

    if (_isLocationLoading) {
      ScaffoldMessenger.of(context).showSnackBar(
        const SnackBar(
          content: Text(
            'Sedang mengunci koordinat GPS, harap tunggu sebentar...',
          ),
          backgroundColor: Colors.orange,
        ),
      );
      return;
    }

    if (_distanceInMeters > 1000) {
      showDialog(
        context: context,
        builder: (ctx) => AlertDialog(
          title: Text(
            'Di Luar Radius Kunjungan',
            style: GoogleFonts.outfit(
              fontWeight: FontWeight.bold,
              color: Colors.red,
            ),
          ),
          content: Text(
            'Tidak bisa selfie kunjungan karena Anda terdeteksi berada di luar radius toko sejauh ${(_distanceInMeters / 1000).toStringAsFixed(1)} KM.\n\n'
            'Apabila Anda sudah berada di titik yang benar tapi aplikasinya masih menolak, silakan edit titik lokasi tokonya terlebih dahulu di halaman Edit Toko.',
            style: GoogleFonts.outfit(fontSize: 14),
          ),
          actions: [
            ElevatedButton(
              onPressed: () => Navigator.pop(ctx),
              style: ElevatedButton.styleFrom(backgroundColor: Colors.red),
              child: Text(
                'Tutup',
                style: GoogleFonts.outfit(
                  color: Colors.white,
                  fontWeight: FontWeight.bold,
                ),
              ),
            ),
          ],
        ),
      );
      return;
    }

    setState(() => _isTaking = true);

    try {
      final xfile = await _controller!.takePicture();
      final bytes = await xfile.readAsBytes();

      // Burn watermark onto the captured image using canvas
      final resultBytes = await _burnWatermark(bytes);

      String pathOrBase64 = '';
      if (kIsWeb) {
        // On Web, convert to Base64 for storage in local DB
        pathOrBase64 = 'data:image/jpeg;base64,${base64Encode(resultBytes)}';
      } else {
        final dir = await getApplicationDocumentsDirectory();
        final outPath =
            '${dir.path}/checkin_${DateTime.now().millisecondsSinceEpoch}.jpg';
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
    final canvas = Canvas(
      recorder,
      Rect.fromLTWH(0, 0, targetWidth, targetHeight),
    );

    // Draw original image scaled down
    canvas.drawImageRect(
      srcImg,
      Rect.fromLTWH(0, 0, srcImg.width.toDouble(), srcImg.height.toDouble()),
      Rect.fromLTWH(0, 0, targetWidth, targetHeight),
      Paint(),
    );

    final cardMargin = targetWidth * 0.03;
    final paddingH = targetWidth * 0.03;
    final paddingV = targetWidth * 0.022;
    final cardWidth = targetWidth - (cardMargin * 2);

    final gpsCoordText = (_lat != 0.0 && _lng != 0.0)
        ? "GPS: ${_lat.toStringAsFixed(6)}, ${_lng.toStringAsFixed(6)}"
        : "GPS: Menunggu Kunci";

    final spans = <TextSpan>[
      TextSpan(
        text: "${widget.store.name.toUpperCase()}\n",
        style: TextStyle(
          fontSize: targetWidth * 0.034,
          fontWeight: FontWeight.bold,
          color: Colors.white,
          letterSpacing: 0.5,
        ),
      ),
      TextSpan(
        text: "$_timestamp WIB  •  $gpsCoordText\n",
        style: TextStyle(
          fontSize: targetWidth * 0.024,
          fontWeight: FontWeight.w600,
          color: Colors.yellowAccent,
        ),
      ),
      if (_address.isNotEmpty)
        TextSpan(
          text: _address,
          style: TextStyle(
            fontSize: targetWidth * 0.021,
            color: Colors.white.withOpacity(0.9),
          ),
        ),
    ];

    final tp = TextPainter(
      text: TextSpan(children: spans),
      textDirection: ui.TextDirection.ltr,
      textAlign: TextAlign.left,
      maxLines: 4,
      ellipsis: '...',
    );
    tp.layout(maxWidth: cardWidth - (paddingH * 2) - (targetWidth * 0.015));

    final textHeight = tp.size.height;
    final cardHeight = textHeight + (paddingV * 2);
    final cardY = targetHeight - cardHeight - cardMargin;

    // Draw compact rounded card with dark translucent background
    final cardPaint = Paint()..color = Colors.black.withOpacity(0.72);
    final cardRect = Rect.fromLTWH(cardMargin, cardY, cardWidth, cardHeight);
    final cardRRect = RRect.fromRectAndRadius(
      cardRect,
      Radius.circular(targetWidth * 0.02),
    );
    canvas.drawRRect(cardRRect, cardPaint);

    // Subtle green accent stripe on the left edge of the card
    final accentPaint = Paint()..color = const Color(0xFF00E676);
    final accentRRect = RRect.fromRectAndRadius(
      Rect.fromLTWH(cardMargin, cardY, targetWidth * 0.012, cardHeight),
      Radius.circular(targetWidth * 0.02),
    );
    canvas.drawRRect(accentRRect, accentPaint);

    // Draw text inside card with offset past the accent stripe
    tp.paint(
      canvas,
      Offset(cardMargin + paddingH + (targetWidth * 0.012), cardY + paddingV),
    );

    final picture = recorder.endRecording();
    final int outWidth = targetWidth.toInt();
    final int outHeight = targetHeight.toInt();
    final resultImg = await picture.toImage(outWidth, outHeight);
    final byteData = await resultImg.toByteData(
      format: ui.ImageByteFormat.rawRgba,
    );

    // Dispose resources to prevent context/memory loss
    srcImg.dispose();
    resultImg.dispose();
    picture.dispose();

    final rawBytes = byteData!.buffer.asUint8List();
    // Directly encode raw RGBA to JPEG with 75% quality without PNG roundtrip
    return await ImageCompressHelper.encodeRawRgbaToJpg(
      rawBytes,
      width: outWidth,
      height: outHeight,
      quality: 75,
    );
  }

  @override
  void dispose() {
    _controller?.dispose();
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
    final bool isFrontCamera =
        _selectedCameraIdx >= 0 &&
        _selectedCameraIdx < _cameras.length &&
        _cameras[_selectedCameraIdx].lensDirection == CameraLensDirection.front;

    return Stack(
      children: [
        if (_isInitialized && _controller != null)
          GestureDetector(
            onScaleStart: (details) {
              _baseZoom = _currentZoom;
            },
            onScaleUpdate: (details) {
              if (_controller == null || !_isInitialized) return;
              final newZoom = (_baseZoom * details.scale).clamp(
                _minZoom,
                _maxZoom,
              );
              if ((newZoom - _currentZoom).abs() > 0.03) {
                _setZoom(newZoom);
              }
            },
            child: Container(
              color: Colors.black,
              width: double.infinity,
              height: double.infinity,
              child: Center(
                child: AspectRatio(
                  aspectRatio: 1 / _controller!.value.aspectRatio,
                  child: CameraPreview(_controller!),
                ),
              ),
            ),
          )
        else
          const Center(child: CircularProgressIndicator(color: Colors.white)),

        // Framing Guide Overlay (Rule of Thirds: Wajah & Toko)
        if (_isInitialized && _showFramingGuide)
          _buildFramingGuide(isFrontCamera),

        // Top bar
        Positioned(
          top: 0,
          left: 0,
          right: 0,
          child: Container(
            padding: const EdgeInsets.fromLTRB(16, 50, 16, 16),
            decoration: BoxDecoration(
              gradient: LinearGradient(
                begin: Alignment.topCenter,
                end: Alignment.bottomCenter,
                colors: [Colors.black.withOpacity(0.8), Colors.transparent],
              ),
            ),
            child: Row(
              crossAxisAlignment: CrossAxisAlignment.start,
              children: [
                IconButton(
                  icon: const Icon(
                    Icons.arrow_back_ios_rounded,
                    color: Colors.white,
                  ),
                  onPressed: () => Navigator.pop(context),
                ),
                Expanded(
                  child: Column(
                    children: [
                      Text(
                        'FOTO KUNJUNGAN TOKO',
                        style: GoogleFonts.outfit(
                          color: Colors.greenAccent,
                          fontSize: 11,
                          fontWeight: FontWeight.w900,
                          letterSpacing: 1.2,
                        ),
                      ),
                      const SizedBox(height: 2),
                      Text(
                        widget.store.name,
                        style: GoogleFonts.outfit(
                          color: Colors.white,
                          fontSize: 17,
                          fontWeight: FontWeight.w900,
                        ),
                      ),
                      Text(
                        isFrontCamera
                            ? 'Selfie sudut lebar (arahkan plang toko terlihat)'
                            : 'Foto tampak depan toko bersama plang nama',
                        style: GoogleFonts.outfit(
                          color: Colors.white70,
                          fontSize: 11,
                        ),
                        maxLines: 1,
                        overflow: TextOverflow.ellipsis,
                      ),
                    ],
                  ),
                ),
                // Toggle framing guide button
                IconButton(
                  icon: Icon(
                    _showFramingGuide
                        ? Icons.view_quilt_rounded
                        : Icons.view_quilt_outlined,
                    color: _showFramingGuide
                        ? Colors.greenAccent
                        : Colors.white70,
                  ),
                  tooltip: 'Panduan Posisi Foto',
                  onPressed: () =>
                      setState(() => _showFramingGuide = !_showFramingGuide),
                ),
              ],
            ),
          ),
        ),

        // Zoom control buttons (Ultra-wide 0.5x / Standard 1.0x)
        if (_isInitialized && (_minZoom < 0.95 || _maxZoom > 1.05))
          Positioned(
            bottom: 235,
            left: 0,
            right: 0,
            child: Center(
              child: Container(
                padding: const EdgeInsets.symmetric(horizontal: 6, vertical: 3),
                decoration: BoxDecoration(
                  color: Colors.black.withOpacity(0.65),
                  borderRadius: BorderRadius.circular(24),
                  border: Border.all(color: Colors.white24, width: 1),
                ),
                child: Row(
                  mainAxisSize: MainAxisSize.min,
                  children: [
                    if (_minZoom < 0.95) ...[
                      GestureDetector(
                        onTap: () => _setZoom(_minZoom),
                        child: Container(
                          padding: const EdgeInsets.symmetric(
                            horizontal: 10,
                            vertical: 6,
                          ),
                          decoration: BoxDecoration(
                            color: (_currentZoom <= _minZoom + 0.05)
                                ? Colors.greenAccent
                                : Colors.transparent,
                            borderRadius: BorderRadius.circular(16),
                          ),
                          child: Text(
                            '0.5x WIDE',
                            style: GoogleFonts.outfit(
                              fontSize: 11,
                              fontWeight: FontWeight.w900,
                              color: (_currentZoom <= _minZoom + 0.05)
                                  ? Colors.black
                                  : Colors.white,
                            ),
                          ),
                        ),
                      ),
                      const SizedBox(width: 4),
                    ],
                    GestureDetector(
                      onTap: () => _setZoom(1.0),
                      child: Container(
                        padding: const EdgeInsets.symmetric(
                          horizontal: 10,
                          vertical: 6,
                        ),
                        decoration: BoxDecoration(
                          color: ((_currentZoom - 1.0).abs() <= 0.05)
                              ? Colors.greenAccent
                              : Colors.transparent,
                          borderRadius: BorderRadius.circular(16),
                        ),
                        child: Text(
                          '1.0x',
                          style: GoogleFonts.outfit(
                            fontSize: 11,
                            fontWeight: FontWeight.w900,
                            color: ((_currentZoom - 1.0).abs() <= 0.05)
                                ? Colors.black
                                : Colors.white,
                          ),
                        ),
                      ),
                    ),
                    if (_maxZoom >= 2.0) ...[
                      const SizedBox(width: 4),
                      GestureDetector(
                        onTap: () => _setZoom(2.0),
                        child: Container(
                          padding: const EdgeInsets.symmetric(
                            horizontal: 10,
                            vertical: 6,
                          ),
                          decoration: BoxDecoration(
                            color: ((_currentZoom - 2.0).abs() <= 0.1)
                                ? Colors.greenAccent
                                : Colors.transparent,
                            borderRadius: BorderRadius.circular(16),
                          ),
                          child: Text(
                            '2.0x',
                            style: GoogleFonts.outfit(
                              fontSize: 11,
                              fontWeight: FontWeight.w900,
                              color: ((_currentZoom - 2.0).abs() <= 0.1)
                                  ? Colors.black
                                  : Colors.white,
                            ),
                          ),
                        ),
                      ),
                    ],
                  ],
                ),
              ),
            ),
          ),

        // Watermark preview at bottom
        Positioned(
          bottom: 125,
          left: 0,
          right: 0,
          child: Container(
            margin: const EdgeInsets.symmetric(horizontal: 14),
            padding: const EdgeInsets.symmetric(horizontal: 14, vertical: 10),
            decoration: BoxDecoration(
              color: Colors.black.withOpacity(0.68),
              borderRadius: BorderRadius.circular(14),
              border: Border.all(color: Colors.white12, width: 1),
            ),
            child: Column(
              children: [
                Text(
                  widget.store.name.toUpperCase(),
                  style: GoogleFonts.outfit(
                    color: Colors.white,
                    fontWeight: FontWeight.w900,
                    fontSize: 13,
                  ),
                ),
                GestureDetector(
                  onTap: _showManualAddressDialog,
                  child: Row(
                    mainAxisAlignment: MainAxisAlignment.center,
                    children: [
                      Expanded(
                        child: Text(
                          _address.isEmpty
                              ? 'Ketuk untuk input alamat manual'
                              : _address,
                          style: GoogleFonts.outfit(
                            color: _address.isEmpty
                                ? Colors.orangeAccent
                                : Colors.white70,
                            fontSize: 11,
                          ),
                          maxLines: 2,
                          textAlign: TextAlign.center,
                          overflow: TextOverflow.ellipsis,
                        ),
                      ),
                      const SizedBox(width: 4),
                      Icon(
                        Icons.edit_location_alt,
                        size: 14,
                        color: _address.isEmpty
                            ? Colors.orangeAccent
                            : Colors.white70,
                      ),
                    ],
                  ),
                ),
                const SizedBox(height: 4),
                Row(
                  mainAxisAlignment: MainAxisAlignment.center,
                  children: [
                    Container(
                      padding: const EdgeInsets.symmetric(
                        horizontal: 8,
                        vertical: 2,
                      ),
                      decoration: BoxDecoration(
                        color: (_lat != 0.0 && _lng != 0.0)
                            ? Colors.green.withOpacity(0.3)
                            : Colors.red.withOpacity(0.3),
                        borderRadius: BorderRadius.circular(8),
                        border: Border.all(
                          color: (_lat != 0.0 && _lng != 0.0)
                              ? Colors.greenAccent
                              : Colors.redAccent,
                          width: 1,
                        ),
                      ),
                      child: Row(
                        mainAxisSize: MainAxisSize.min,
                        children: [
                          Icon(
                            Icons.my_location,
                            size: 11,
                            color: (_lat != 0.0 && _lng != 0.0)
                                ? Colors.greenAccent
                                : Colors.redAccent,
                          ),
                          const SizedBox(width: 4),
                          Text(
                            (_lat != 0.0 && _lng != 0.0)
                                ? 'GPS: ${_lat.toStringAsFixed(5)}, ${_lng.toStringAsFixed(5)}'
                                : 'GPS Belum Terkunci',
                            style: GoogleFonts.outfit(
                              fontSize: 10,
                              fontWeight: FontWeight.bold,
                              color: (_lat != 0.0 && _lng != 0.0)
                                  ? Colors.greenAccent
                                  : Colors.redAccent,
                            ),
                          ),
                        ],
                      ),
                    ),
                    const SizedBox(width: 6),
                    InkWell(
                      onTap: _getLocation,
                      child: Container(
                        padding: const EdgeInsets.all(4),
                        decoration: BoxDecoration(
                          color: Colors.white24,
                          borderRadius: BorderRadius.circular(6),
                        ),
                        child: _isLocationLoading
                            ? const SizedBox(
                                width: 12,
                                height: 12,
                                child: CircularProgressIndicator(
                                  strokeWidth: 2,
                                  color: Colors.white,
                                ),
                              )
                            : const Icon(
                                Icons.refresh,
                                size: 14,
                                color: Colors.white,
                              ),
                      ),
                    ),
                  ],
                ),
                if (_distanceInMeters != -1.0)
                  Padding(
                    padding: const EdgeInsets.only(top: 4.0),
                    child: Text(
                      'Jarak: ${_distanceInMeters.toStringAsFixed(0)} meter',
                      style: GoogleFonts.outfit(
                        color: _distanceInMeters > 1000
                            ? Colors.redAccent
                            : Colors.lightGreenAccent,
                        fontWeight: FontWeight.w600,
                        fontSize: 12,
                      ),
                    ),
                  ),
                Text(
                  _timestamp,
                  style: GoogleFonts.outfit(
                    color: Colors.yellowAccent,
                    fontSize: 11,
                  ),
                ),
              ],
            ),
          ),
        ),

        // Bottom controls
        Positioned(
          bottom: 35,
          left: 0,
          right: 0,
          child: Row(
            mainAxisAlignment: MainAxisAlignment.spaceEvenly,
            children: [
              // Empty space to balance
              const SizedBox(width: 48),

              // Capture button
              GestureDetector(
                onTap: _takePhoto,
                child: Container(
                  width: 80,
                  height: 80,
                  decoration: BoxDecoration(
                    shape: BoxShape.circle,
                    color: Colors.white,
                    border: Border.all(
                      color: Colors.white.withOpacity(0.5),
                      width: 4,
                    ),
                    boxShadow: [
                      BoxShadow(
                        color: Colors.white.withOpacity(0.3),
                        blurRadius: 20,
                      ),
                    ],
                  ),
                  child: _isTaking
                      ? const Padding(
                          padding: EdgeInsets.all(24),
                          child: CircularProgressIndicator(
                            color: Colors.black,
                            strokeWidth: 3,
                          ),
                        )
                      : const Icon(
                          Icons.camera_alt_rounded,
                          size: 36,
                          color: Colors.black,
                        ),
                ),
              ),

              // Switch camera button
              if (_cameras.length > 1)
                Container(
                  decoration: BoxDecoration(
                    color: Colors.black.withOpacity(0.55),
                    shape: BoxShape.circle,
                    border: Border.all(color: Colors.white24, width: 1.5),
                  ),
                  child: IconButton(
                    icon: const Icon(
                      Icons.flip_camera_ios_rounded,
                      color: Colors.white,
                      size: 28,
                    ),
                    tooltip: 'Ganti Kamera Depan/Belakang',
                    onPressed: _switchCamera,
                  ),
                )
              else
                const SizedBox(width: 48),
            ],
          ),
        ),
      ],
    );
  }

  Widget _buildFramingGuide(bool isFrontCamera) {
    return Positioned(
      top: 135,
      bottom: 275,
      left: 16,
      right: 16,
      child: IgnorePointer(
        ignoring: false,
        child: Column(
          children: [
            // Guide mini toolbar
            Row(
              mainAxisAlignment: MainAxisAlignment.center,
              children: [
                GestureDetector(
                  onTap: () =>
                      setState(() => _guideFaceOnLeft = !_guideFaceOnLeft),
                  child: Container(
                    padding: const EdgeInsets.symmetric(
                      horizontal: 10,
                      vertical: 4,
                    ),
                    decoration: BoxDecoration(
                      color: Colors.black.withOpacity(0.7),
                      borderRadius: BorderRadius.circular(20),
                      border: Border.all(color: Colors.white30, width: 1),
                    ),
                    child: Row(
                      mainAxisSize: MainAxisSize.min,
                      children: [
                        const Icon(
                          Icons.swap_horiz_rounded,
                          size: 14,
                          color: Colors.greenAccent,
                        ),
                        const SizedBox(width: 4),
                        Text(
                          'Tukar Sisi Wajah',
                          style: GoogleFonts.outfit(
                            fontSize: 10,
                            fontWeight: FontWeight.bold,
                            color: Colors.white,
                          ),
                        ),
                      ],
                    ),
                  ),
                ),
                const SizedBox(width: 8),
                GestureDetector(
                  onTap: () => setState(() => _showFramingGuide = false),
                  child: Container(
                    padding: const EdgeInsets.symmetric(
                      horizontal: 8,
                      vertical: 4,
                    ),
                    decoration: BoxDecoration(
                      color: Colors.black.withOpacity(0.7),
                      borderRadius: BorderRadius.circular(20),
                      border: Border.all(color: Colors.white30, width: 1),
                    ),
                    child: Row(
                      mainAxisSize: MainAxisSize.min,
                      children: [
                        const Icon(
                          Icons.close_rounded,
                          size: 12,
                          color: Colors.white70,
                        ),
                        const SizedBox(width: 2),
                        Text(
                          'Tutup',
                          style: GoogleFonts.outfit(
                            fontSize: 10,
                            color: Colors.white70,
                          ),
                        ),
                      ],
                    ),
                  ),
                ),
              ],
            ),
            const SizedBox(height: 8),

            // Visual guide frame boxes
            Expanded(
              child: Row(
                children: _guideFaceOnLeft
                    ? [
                        _buildFaceGuideBox(),
                        const SizedBox(width: 8),
                        _buildStoreGuideBox(),
                      ]
                    : [
                        _buildStoreGuideBox(),
                        const SizedBox(width: 8),
                        _buildFaceGuideBox(),
                      ],
              ),
            ),
          ],
        ),
      ),
    );
  }

  Widget _buildFaceGuideBox() {
    return Expanded(
      flex: 4,
      child: Container(
        decoration: BoxDecoration(
          color: Colors.greenAccent.withOpacity(0.08),
          borderRadius: BorderRadius.circular(24),
          border: Border.all(
            color: Colors.greenAccent.withOpacity(0.75),
            width: 1.5,
          ),
        ),
        child: Column(
          mainAxisAlignment: MainAxisAlignment.center,
          children: [
            Container(
              padding: const EdgeInsets.all(10),
              decoration: BoxDecoration(
                shape: BoxShape.circle,
                color: Colors.greenAccent.withOpacity(0.2),
              ),
              child: const Icon(
                Icons.face_retouching_natural_rounded,
                color: Colors.greenAccent,
                size: 26,
              ),
            ),
            const SizedBox(height: 6),
            Text(
              'Wajah Sales',
              style: GoogleFonts.outfit(
                color: Colors.greenAccent,
                fontSize: 11,
                fontWeight: FontWeight.w800,
              ),
            ),
            Text(
              'Posisikan di sini',
              style: GoogleFonts.outfit(color: Colors.white70, fontSize: 9),
            ),
          ],
        ),
      ),
    );
  }

  Widget _buildStoreGuideBox() {
    return Expanded(
      flex: 6,
      child: Container(
        decoration: BoxDecoration(
          color: Colors.amberAccent.withOpacity(0.06),
          borderRadius: BorderRadius.circular(24),
          border: Border.all(
            color: Colors.amberAccent.withOpacity(0.75),
            width: 1.5,
          ),
        ),
        child: Column(
          mainAxisAlignment: MainAxisAlignment.center,
          children: [
            Container(
              padding: const EdgeInsets.all(10),
              decoration: BoxDecoration(
                shape: BoxShape.circle,
                color: Colors.amberAccent.withOpacity(0.2),
              ),
              child: const Icon(
                Icons.storefront_rounded,
                color: Colors.amberAccent,
                size: 28,
              ),
            ),
            const SizedBox(height: 6),
            Text(
              'Tampak Toko & Plang',
              style: GoogleFonts.outfit(
                color: Colors.amberAccent,
                fontSize: 11,
                fontWeight: FontWeight.w800,
              ),
            ),
            Text(
              'Biarkan area toko terlihat luas',
              style: GoogleFonts.outfit(color: Colors.white70, fontSize: 9),
            ),
          ],
        ),
      ),
    );
  }

  Widget _buildPreview() {
    return Stack(
      children: [
        Container(
          color: Colors.black,
          width: double.infinity,
          height: double.infinity,
          child: Center(
            child: InteractiveViewer(
              minScale: 1.0,
              maxScale: 3.0,
              child: Image.memory(_capturedBytes!, fit: BoxFit.contain),
            ),
          ),
        ),

        // Top bar
        Positioned(
          top: 0,
          left: 0,
          right: 0,
          child: Container(
            padding: const EdgeInsets.fromLTRB(16, 50, 16, 16),
            decoration: BoxDecoration(
              gradient: LinearGradient(
                begin: Alignment.topCenter,
                end: Alignment.bottomCenter,
                colors: [Colors.black.withOpacity(0.8), Colors.transparent],
              ),
            ),
            child: Row(
              crossAxisAlignment: CrossAxisAlignment.start,
              children: [
                IconButton(
                  icon: const Icon(
                    Icons.arrow_back_ios_rounded,
                    color: Colors.white,
                  ),
                  onPressed: () => setState(() {
                    _capturedPath = null;
                    _capturedBytes = null;
                  }),
                ),
                Expanded(
                  child: Column(
                    children: [
                      Text(
                        'PRATINJAU FOTO KUNJUNGAN',
                        style: GoogleFonts.outfit(
                          color: Colors.yellowAccent,
                          fontSize: 12,
                          fontWeight: FontWeight.w900,
                          letterSpacing: 1,
                        ),
                      ),
                      Text(
                        'Apakah toko & wajah sudah jelas?',
                        style: GoogleFonts.outfit(
                          color: Colors.white,
                          fontSize: 15,
                          fontWeight: FontWeight.w700,
                        ),
                      ),
                    ],
                  ),
                ),
                const SizedBox(width: 48), // To balance the center alignment
              ],
            ),
          ),
        ),

        // Action buttons
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
                    _capturedBytes = null;
                  }),
                  icon: const Icon(Icons.replay_rounded, color: Colors.white),
                  label: Text(
                    'FOTO ULANG',
                    style: GoogleFonts.outfit(
                      color: Colors.white,
                      fontWeight: FontWeight.w800,
                    ),
                  ),
                  style: OutlinedButton.styleFrom(
                    side: const BorderSide(color: Colors.white, width: 2),
                    shape: RoundedRectangleBorder(
                      borderRadius: BorderRadius.circular(16),
                    ),
                    padding: const EdgeInsets.symmetric(vertical: 16),
                  ),
                ),
              ),
              const SizedBox(width: 16),
              Expanded(
                child: ElevatedButton.icon(
                  onPressed: _isSubmitting
                      ? null
                      : () async {
                          if (_isSubmitting) return;
                          setState(() => _isSubmitting = true);

                          if (_lat == 0.0 && _lng == 0.0) {
                            if (widget.store.latitude != 0.0 &&
                                widget.store.longitude != 0.0) {
                              _lat = widget.store.latitude;
                              _lng = widget.store.longitude;
                            } else {
                              setState(() => _isSubmitting = false);
                              ScaffoldMessenger.of(context).showSnackBar(
                                const SnackBar(
                                  content: Text(
                                    'Lokasi GPS belum terkunci. Harap ketuk tombol muat ulang GPS terlebih dahulu.',
                                  ),
                                  backgroundColor: Colors.red,
                                ),
                              );
                              _getLocation();
                              return;
                            }
                          }

                          try {
                            final AppDatabase db = di.sl<AppDatabase>();
                            await db
                                .into(db.localCheckins)
                                .insert(
                                  LocalCheckinsCompanion.insert(
                                    storeId: widget.store.id,
                                    storeName: widget.store.name,
                                    latitude: _lat,
                                    longitude: _lng,
                                    selfiePath: _capturedPath!,
                                    createdAt: DateTime.now(),
                                    syncStatus: const Value('pending'),
                                  ),
                                );

                            // Trigger sync in background
                            di.sl<SyncBloc>().add(
                              const SyncDataRequested(force: true),
                            );

                            if (mounted) {
                              Navigator.push(
                                context,
                                MaterialPageRoute(
                                  builder: (_) => VisitTransactionPage(
                                    store: widget.store,
                                    selfiePath: _capturedPath!,
                                    selfieBytes: _capturedBytes,
                                  ),
                                ),
                              );
                            }
                          } finally {
                            if (mounted) {
                              setState(() => _isSubmitting = false);
                            }
                          }
                        },
                  icon: _isSubmitting
                      ? const SizedBox(
                          width: 20,
                          height: 20,
                          child: CircularProgressIndicator(
                            color: Colors.white,
                            strokeWidth: 2,
                          ),
                        )
                      : const Icon(
                          Icons.check_circle_rounded,
                          color: Colors.white,
                        ),
                  label: Text(
                    'LANJUT',
                    style: GoogleFonts.outfit(
                      color: Colors.white,
                      fontWeight: FontWeight.w900,
                      fontSize: 16,
                    ),
                  ),
                  style: ElevatedButton.styleFrom(
                    backgroundColor: Colors.green,
                    shape: RoundedRectangleBorder(
                      borderRadius: BorderRadius.circular(16),
                    ),
                    padding: const EdgeInsets.symmetric(vertical: 16),
                    elevation: 8,
                  ),
                ),
              ),
            ],
          ),
        ),
      ],
    );
  }
}
