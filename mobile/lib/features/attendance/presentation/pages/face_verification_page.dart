import 'dart:io';
import 'dart:async';
import 'dart:ui' as ui;
import 'package:path/path.dart' as p;
import 'package:path_provider/path_provider.dart';
import 'package:flutter/material.dart';
import 'package:flutter/foundation.dart';
import 'package:flutter_bloc/flutter_bloc.dart';
import 'package:google_fonts/google_fonts.dart';
import 'package:intl/intl.dart';
import 'package:camera/camera.dart';
import 'package:geolocator/geolocator.dart';
import 'package:geocoding/geocoding.dart';
// Face & InputImage types diambil dari face_detector_service.dart (conditional export)
// Import langsung MLKit dihapus agar tidak konflik dengan stub di Web
import 'package:image/image.dart' as imglib;
import 'package:hris_app/core/theme/app_colors.dart';
import 'package:hris_app/injection.dart' as di;
import 'package:hris_app/core/services/biometric_service.dart';
import 'package:hris_app/core/utils/dialog_utils.dart';
import 'package:hris_app/core/utils/face_detector_service.dart';
import 'package:hris_app/core/utils/constants.dart';
import 'package:hris_app/core/utils/ml_service.dart';
import 'package:hris_app/core/utils/image_compress_helper.dart';
import 'package:hris_app/features/auth/domain/repositories/auth_repository.dart';
import 'package:hris_app/features/attendance/presentation/bloc/attendance_bloc.dart';
import 'package:hris_app/features/attendance/presentation/bloc/attendance_event.dart';
import 'package:hris_app/features/attendance/presentation/bloc/attendance_state.dart';
import 'package:hris_app/core/utils/snackbar_utils.dart';
import 'package:hris_app/features/attendance/presentation/pages/clock_in_success_page.dart';
import 'package:hris_app/features/attendance/presentation/pages/clock_in_failed_page.dart';
import 'package:hris_app/features/attendance/presentation/pages/clock_out_success_page.dart';
import 'package:hris_app/features/attendance/presentation/pages/clock_out_failed_page.dart';

class FaceVerificationResult {
  final List<double> embedding;
  final String imagePath;

  FaceVerificationResult({required this.embedding, required this.imagePath});
}

class _FaceProcessingParams {
  final String path;
  final Rect boundingBox;
  final bool mirror;
  final String timeText;
  final String locationText;
  final String addressText;

  _FaceProcessingParams(
    this.path,
    this.boundingBox,
    this.timeText,
    this.locationText,
    this.addressText, {
    this.mirror = false,
  });
}

class ProcessedImageResult {
  final imglib.Image faceCrop;
  final List<int> watermarkedBytes;

  ProcessedImageResult(this.faceCrop, this.watermarkedBytes);
}

Future<ProcessedImageResult?> _processImageAndWatermarkBackground(
    _FaceProcessingParams params) async {
  try {
    final bytes = await File(params.path).readAsBytes();
    imglib.Image? capturedImage = imglib.decodeImage(bytes);
    if (capturedImage == null) return null;

    // Bake orientation logic
    capturedImage = imglib.bakeOrientation(capturedImage);

    // If it's from the front camera, we usually want to un-mirror it
    if (params.mirror) {
      capturedImage = imglib.flip(capturedImage,
          direction: imglib.FlipDirection.horizontal);
    }

    // 1. Face Crop for ML
    final faceCrop = imglib.copyCrop(
      capturedImage,
      x: params.boundingBox.left.toInt(),
      y: params.boundingBox.top.toInt(),
      width: params.boundingBox.width.toInt(),
      height: params.boundingBox.height.toInt(),
    );

    // 2. Base Image for Watermark
    int targetWidth = 720;
    
    // NEW: Crop to 1:1 (Square) aspect ratio, centered around the face
    int cropSize = capturedImage.width; // Use width as the size for a square in portrait
    int faceCenterY = params.boundingBox.center.dy.toInt();
    
    int cropX = 0;
    int cropY = faceCenterY - (cropSize ~/ 2);
    
    // Boundary checks to ensure we don't crop outside the image
    if (cropY < 0) cropY = 0;
    if (cropY + cropSize > capturedImage.height) {
      cropY = capturedImage.height - cropSize;
    }

    imglib.Image squareImage = imglib.copyCrop(
      capturedImage, 
      x: cropX, 
      y: cropY, 
      width: cropSize, 
      height: cropSize,
    );

    // Resize the square image
    imglib.Image resizedImage = squareImage;
    if (squareImage.width > targetWidth) {
      resizedImage = imglib.copyResize(squareImage, width: targetWidth);
    }
    
    // Compress quality to 70 to save storage
    final processedBytes = imglib.encodeJpg(resizedImage, quality: 70);

    return ProcessedImageResult(faceCrop, processedBytes);
  } catch (e) {
    debugPrint("Background processing error: $e");
    return null;
  }
}

class FaceVerificationPage extends StatefulWidget {
  final bool isClockIn;
  final bool isRegistration;
  final bool isOutOfTown;
  final String destination;

  const FaceVerificationPage({
    super.key,
    this.isClockIn = true,
    this.isRegistration = false,
    this.isOutOfTown = false,
    this.destination = '',
  });

  @override
  State<FaceVerificationPage> createState() => _FaceVerificationPageState();
}

class _FaceVerificationPageState extends State<FaceVerificationPage> {
  CameraController? _cameraController;
  final FaceDetectorService _faceDetectorService = FaceDetectorService();
  final MLService _mlService = MLService();

  String? _capturedImagePath;
  bool _isVerifying = false;
  Position? _currentPosition;
  String? _currentAddress;
  bool _isLocationLoading = true;
  String? _locationError;
  String? _cameraError;
  late Timer _nowTimer;
  DateTime _now = DateTime.now();
  bool _isBiometricEnabled = false;

  // Shift Info
  String _shiftName = "SHIFT PAGI";
  String _shiftStartTime = "07:30";
  String _shiftEndTime = "16:30";

  // Geofencing & Branch Info
  Map<String, dynamic>? _assignedBranch;
  double _distanceToBranch = -1;
  bool _isWithinGeofence = false;
  String? _targetBranchAddress;
  bool _canAbsenAnywhere = false;

  // Face Recognition Info
  bool _isFaceDetected = false;
  bool _isFaceMatched = false;
  String _statusText = "Menginisialisasi...";
  bool _isProcessing = false;
  List<double>? _currentEmbedding;
  List<double>? _baselineEmbedding;

  late bool _isOutOfTown;
  late String _destination;

  @override
  void initState() {
    super.initState();
    _isOutOfTown = widget.isOutOfTown;
    _destination = widget.destination;
    _initializeServices();
    _checkBiometricStatus();
    _nowTimer = Timer.periodic(const Duration(seconds: 1), (timer) {
      if (mounted) {
        setState(() {
          _now = DateTime.now();
        });
      }
    });
  }

  Future<void> _initializeServices() async {
    // Start camera immediately so the user doesn't see a hang
    _startCamera();

    if (!kIsWeb) {
      _faceDetectorService.initialize();
      _mlService.initialize().catchError((e) {
        debugPrint("MLService initialization error: $e");
      });
    }

    // PENTING: Tunggu data cabang tersedia dulu sebelum menghitung geofence
    // agar tidak terjadi race condition (_assignedBranch masih null saat geofence dihitung)
    await _fetchBranchAndBaseline();
    _getCurrentLocation();
  }

  Future<void> _fetchBranchAndBaseline() async {
    try {
      // Skip authenticated calls during registration
      if (widget.isRegistration) {
        final branchesResult = await di.sl<AuthRepository>().getBranches();
        branchesResult.fold((l) => debugPrint("Error branches: ${l.message}"),
            (branches) {
          if (mounted) {
            // During registration we don't have an assigned branch yet
          }
        });
        return;
      }

      final profileResult = await di.sl<AuthRepository>().getProfile();
      final branchesResult = await di.sl<AuthRepository>().getBranches();
      final embeddingResult =
          await di.sl<AuthRepository>().getStoredFaceEmbedding();

      profileResult.fold((l) => debugPrint("Error profile: ${l.message}"),
          (profile) {
        // Compatibility layer for branch_id
        final branchId = profile['branch_id'] ?? profile['BranchID'];

        if (branchId == null) {
          debugPrint(
              "Warning: No branch_id found in profile. Is a branch assigned to this employee?");
        }
        
        final userObj = profile['user'] ?? profile['User'];
        if (userObj != null && userObj is Map) {
          if (mounted) {
            setState(() {
              _canAbsenAnywhere = userObj['can_absen_anywhere'] == true || userObj['CanAbsenAnywhere'] == true;
            });
          }
        }

        // Fallback embedding extraction from profile
        final remoteEmbedding =
            profile['face_embedding'] ?? profile['FaceEmbedding'];
        if (remoteEmbedding != null &&
            remoteEmbedding is List &&
            remoteEmbedding.isNotEmpty) {
          try {
            final List<double> mappedEmbedding =
                remoteEmbedding.map((e) => (e as num).toDouble()).toList();
            if (mappedEmbedding.length == 192) {
              if (mounted) {
                setState(() {
                  if (_baselineEmbedding == null ||
                      _baselineEmbedding!.isEmpty) {
                    _baselineEmbedding = mappedEmbedding;
                  }
                });
              }
            } else {
              debugPrint(
                  "Warning: Remote embedding has invalid length: ${mappedEmbedding.length}");
            }
          } catch (_) {}
        }

        // Extract applicable shift details dynamically from the preloaded GORM relations
        Map<String, dynamic>? applicableShift;

        final empShifts =
            profile['employee_shifts'] ?? profile['EmployeeShifts'];
        if (empShifts != null && empShifts is List && empShifts.isNotEmpty) {
          final firstShift = empShifts[0];
          if (firstShift != null) {
            final shiftObj = firstShift['shift'] ?? firstShift['Shift'];
            if (shiftObj != null && shiftObj is Map) {
              applicableShift = Map<String, dynamic>.from(shiftObj);
            }
          }
        }

        if (applicableShift == null) {
          final defShift = profile['default_shift'] ?? profile['DefaultShift'];
          if (defShift != null && defShift is Map) {
            applicableShift = Map<String, dynamic>.from(defShift);
          }
        }

        if (applicableShift == null) {
          final branchObj = profile['branch'] ?? profile['Branch'];
          if (branchObj != null && branchObj is Map) {
            final branchDefShift =
                branchObj['default_shift'] ?? branchObj['DefaultShift'];
            if (branchDefShift != null && branchDefShift is Map) {
              applicableShift = Map<String, dynamic>.from(branchDefShift);
            }
          }
        }

        if (applicableShift != null) {
          final name = applicableShift['name'] ?? applicableShift['Name'];
          final startTime =
              applicableShift['start_time'] ?? applicableShift['StartTime'];
          final endTime =
              applicableShift['end_time'] ?? applicableShift['EndTime'];

          String formatTime(String? timeStr) {
            if (timeStr == null || timeStr.isEmpty) return "";
            final parts = timeStr.split(':');
            if (parts.length >= 2) {
              return "${parts[0]}:${parts[1]}";
            }
            return timeStr;
          }

          if (mounted) {
            setState(() {
              if (name != null) _shiftName = name.toString();
              if (startTime != null)
                _shiftStartTime = formatTime(startTime.toString());
              if (endTime != null)
                _shiftEndTime = formatTime(endTime.toString());
            });
          }
        }

        branchesResult.fold((l) => debugPrint("Error branches: ${l.message}"),
            (branches) {
          // Compatibility layer for branch ID lookup
          final branch = branches.firstWhere(
              (b) => (b['id'] ?? b['ID'] ?? b['id_branch']) == branchId,
              orElse: () => {});

          if (mounted) {
            setState(() {
              _assignedBranch = branch.isNotEmpty ? branch : null;
            });

            if (_assignedBranch != null && !kIsWeb) {
              // Compatibility layer for latitude/longitude
              final lat =
                  _assignedBranch!['latitude'] ?? _assignedBranch!['Latitude'];
              final lng = _assignedBranch!['longitude'] ??
                  _assignedBranch!['Longitude'];
              if (lat != null && lng != null) {
                _getBranchAddress(lat, lng);
              }
              // Re-kalkulasi geofence jika posisi user sudah tersedia
              // (menangani kasus di mana location selesai sebelum branch data)
              _recalculateGeofence();
            }
          }
        });
      });

      embeddingResult.fold((l) => debugPrint("Error embedding: ${l.message}"),
          (embedding) {
        if (mounted) {
          if (embedding != null && embedding.isNotEmpty) {
            if (embedding.length == 192) {
              setState(() {
                _baselineEmbedding = embedding;
              });
            } else if (!widget.isRegistration) {
              setState(() {
                _statusText =
                    "Data wajah tidak kompatibel. Silakan registrasi ulang.";
              });
            }
          } else if (!widget.isRegistration) {
            setState(() {
              _statusText = "Data wajah tidak ditemukan.";
            });
          }
        }
      });
    } catch (e) {
      debugPrint("Background fetch error: $e");
    } finally {
      if (mounted && _statusText == "Menginisialisasi...") {
        setState(() {
          _statusText = "Mencari wajah...";
        });
      }
    }
  }

  Future<void> _getBranchAddress(double lat, double lng) async {
    try {
      List<Placemark> placemarks = await placemarkFromCoordinates(lat, lng);
      if (placemarks.isNotEmpty && mounted) {
        Placemark place = placemarks[0];
        setState(() {
          _targetBranchAddress =
              "${place.street}, ${place.subLocality}, ${place.locality}";
        });
      }
    } catch (e) {
      debugPrint("Branch geocoding error: $e");
    }
  }

  /// Re-kalkulasi geofence menggunakan data yang sudah tersedia.
  /// Dipanggil setelah branch data atau posisi user berhasil dimuat,
  /// untuk menangani kasus di mana salah satu tersedia lebih dulu.
  void _recalculateGeofence() {
    if (_currentPosition == null || _assignedBranch == null) return;

    final branchLat =
        _assignedBranch!['latitude'] ?? _assignedBranch!['Latitude'];
    final branchLng =
        _assignedBranch!['longitude'] ?? _assignedBranch!['Longitude'];
    if (branchLat == null || branchLng == null) return;

    final radius = ((_assignedBranch!['radius_meter'] ??
            _assignedBranch!['RadiusMeter'] ??
            100) as num)
        .toDouble();

    final distance = Geolocator.distanceBetween(
      _currentPosition!.latitude,
      _currentPosition!.longitude,
      (branchLat as num).toDouble(),
      (branchLng as num).toDouble(),
    );

    // Tambah toleransi dari akurasi GPS perangkat (untuk WFO / sinyal lemah).
    // GPS di area perkantoran/gedung bisa drifting 20-50m, toleransi ini
    // memastikan user tidak salah ditolak karena noise sensor GPS.
    final gpsAccuracyBuffer = _currentPosition!.accuracy.clamp(0.0, 100.0);
    final effectiveRadius = radius + gpsAccuracyBuffer;

    if (mounted) {
      setState(() {
        _distanceToBranch = distance;
        _isWithinGeofence = distance <= effectiveRadius;
      });
    }
    debugPrint(
        'Geofence recalculated: distance=${distance.toStringAsFixed(1)}m, '
        'radius=${radius.toStringAsFixed(0)}m, '
        'gpsBuffer=${gpsAccuracyBuffer.toStringAsFixed(1)}m, '
        'effectiveRadius=${effectiveRadius.toStringAsFixed(1)}m, '
        'within=$_isWithinGeofence');
  }

  Future<void> _startCamera() async {
    if (mounted) {
      setState(() {
        _cameraError = null;
      });
    }

    try {
      final cameras = await availableCameras();
      if (cameras.isEmpty) {
        if (mounted) {
          setState(() {
            _cameraError = "Kamera tidak ditemukan pada perangkat ini.";
            _statusText = "Kamera tidak tersedia.";
          });
        }
        return;
      }

      final frontCamera = cameras.firstWhere(
        (c) => c.lensDirection == CameraLensDirection.front,
        orElse: () => cameras.first,
      );

      // Attempt 1: Preferred format
      try {
        _cameraController = CameraController(
          frontCamera,
          ResolutionPreset.medium,
          enableAudio: false,
          imageFormatGroup:
              !kIsWeb && defaultTargetPlatform == TargetPlatform.android
                  ? ImageFormatGroup.nv21
                  : ImageFormatGroup.bgra8888,
        );
        await _cameraController!.initialize();
      } catch (camInitError) {
        debugPrint(
            "Initial Camera init failed with nv21/preset, attempting fallback: $camInitError");
        // Fallback Attempt: Default format without forcing nv21
        _cameraController = CameraController(
          frontCamera,
          ResolutionPreset.medium,
          enableAudio: false,
        );
        await _cameraController!.initialize();
      }

      if (!kIsWeb) {
        _cameraController!.startImageStream((CameraImage image) {
          if (!_isProcessing && _capturedImagePath == null) {
            _processCameraImage(image, frontCamera);
          }
        });
      } else {
        // On Web, we can't do real-time stream easily, so we allow manual capture
        setState(() {
          _statusText = "Silakan ambil foto wajah anda.";
          _isFaceDetected = true; // Allow clicking "Ambil Foto"
        });
      }

      if (mounted) {
        setState(() {
          _cameraError = null;
        });
      }
    } catch (e) {
      debugPrint("Camera error: $e");
      if (mounted) {
        setState(() {
          final errStr = e.toString().toLowerCase();
          if (errStr.contains('permission') ||
              errStr.contains('denied') ||
              errStr.contains('cameraaccessdenied')) {
            _cameraError =
                "Izin akses kamera belum diizinkan. Silakan aktifkan izin kamera di Pengaturan HP.";
          } else {
            _cameraError =
                "Gagal mengakses kamera. Pastikan kamera tidak sedang digunakan aplikasi lain.";
          }
          _statusText = "Kamera bermasalah.";
        });
      }
    }
  }

  @override
  void dispose() {
    _nowTimer.cancel();
    _cameraController?.dispose();
    if (!kIsWeb) {
      _faceDetectorService.dispose();
      _mlService.dispose();
    }
    super.dispose();
  }

  Future<void> _checkBiometricStatus() async {
    if (kIsWeb) return;
    try {
      final status = await di.sl<BiometricService>().isBiometricAvailable();
      if (mounted) {
        setState(() {
          _isBiometricEnabled = status;
        });
      }
    } catch (e) {
      debugPrint("Biometric check failed: $e");
    }
  }

  Future<void> _getCurrentLocation() async {
    if (mounted) {
      setState(() {
        _isLocationLoading = true;
        _locationError = null;
      });
    }

    try {
      bool serviceEnabled = await Geolocator.isLocationServiceEnabled();
      if (!serviceEnabled) {
        if (mounted) {
          setState(() {
            _locationError =
                "Layanan GPS dinonaktifkan. Silakan aktifkan GPS Anda.";
            _isLocationLoading = false;
          });
          SnackBarUtils.showError(context,
              "GPS tidak aktif. Silakan aktifkan layanan lokasi di HP Anda.");
        }
        return;
      }

      LocationPermission permission = await Geolocator.checkPermission();
      if (permission == LocationPermission.denied) {
        permission = await Geolocator.requestPermission();
        if (permission == LocationPermission.denied) {
          if (mounted) {
            setState(() {
              _locationError =
                  "Izin lokasi ditolak. Aplikasi butuh izin lokasi untuk absen.";
              _isLocationLoading = false;
            });
            SnackBarUtils.showError(
                context, "Izin lokasi ditolak. Mohon izinkan akses lokasi.");
          }
          return;
        }
      }

      if (permission == LocationPermission.deniedForever) {
        if (mounted) {
          setState(() {
            _locationError =
                "Izin lokasi ditolak permanen. Silakan aktifkan lewat pengaturan.";
            _isLocationLoading = false;
          });
          SnackBarUtils.showError(context,
              "Izin lokasi ditolak permanen. Buka pengaturan untuk mengaktifkan.");
        }
        return;
      }

      Position? position;
      try {
        position = await Geolocator.getCurrentPosition(
          desiredAccuracy: LocationAccuracy.high,
          timeLimit: const Duration(seconds: 10),
        );
      } catch (e) {
        debugPrint("Current position timeout/error: $e. trying last known...");
        try {
          final lastKnown = await Geolocator.getLastKnownPosition();
          if (lastKnown != null) {
            final age = DateTime.now().difference(lastKnown.timestamp);
            // Diperlonggar: gunakan posisi terakhir jika tidak lebih dari 30 menit.
            // Ini membantu user WFO yang berada di dalam gedung dengan sinyal GPS lemah.
            if (age.inMinutes <= 30) {
              position = lastKnown;
              debugPrint(
                  'Using last known position (age: ${age.inMinutes} min)');
            } else {
              debugPrint(
                  'Last known position too old (${age.inMinutes} min), skipping.');
            }
          }
        } catch (_) {}
      }

      if (position == null) {
        if (mounted) {
          setState(() {
            _locationError =
                "Gagal mendeteksi lokasi GPS. Coba lagi di tempat terbuka.";
            _isLocationLoading = false;
          });
          SnackBarUtils.showError(context,
              "Gagal mendeteksi lokasi. Pastikan sinyal GPS Anda baik.");
        }
        return;
      }

      // Reverse Geocoding
      try {
        if (!kIsWeb) {
          List<Placemark> placemarks = await placemarkFromCoordinates(
              position.latitude, position.longitude);
          if (placemarks.isNotEmpty) {
            Placemark place = placemarks[0];
            _currentAddress =
                "${place.street}, ${place.subLocality}, ${place.locality}";
          }
        } else {
          _currentAddress =
              "Web Location (${position.latitude.toStringAsFixed(4)}, ${position.longitude.toStringAsFixed(4)})";
        }
      } catch (e) {
        debugPrint("Geocoding error: $e");
      }

      // Geofencing calculation
      double distance = -1;
      bool withinGeofence = false;

      final branchLat =
          _assignedBranch?['latitude'] ?? _assignedBranch?['Latitude'];
      final branchLng =
          _assignedBranch?['longitude'] ?? _assignedBranch?['Longitude'];
      // Type-safe cast: nilai dari JSON bisa int atau double
      final radius = ((_assignedBranch?['radius_meter'] ??
              _assignedBranch?['RadiusMeter'] ??
              100) as num)
          .toDouble();

      if (!widget.isRegistration &&
          _assignedBranch != null &&
          branchLat != null) {
        distance = Geolocator.distanceBetween(
          position.latitude,
          position.longitude,
          (branchLat as num).toDouble(),
          (branchLng as num).toDouble(),
        );
        // Toleransi dari akurasi GPS perangkat (WFO: sinyal dalam gedung bisa drifting).
        // Nilai accuracy dari Geolocator adalah radius ketidakpastian dalam meter.
        // Dibatasi maks 100m agar tidak terlalu longgar pada GPS yang sangat buruk.
        final gpsAccuracyBuffer = position.accuracy.clamp(0.0, 100.0);
        final effectiveRadius = radius + gpsAccuracyBuffer;
        withinGeofence = distance <= effectiveRadius;
        debugPrint(
            'Geofence check: distance=${distance.toStringAsFixed(1)}m, '
            'radius=${radius.toStringAsFixed(0)}m, '
            'gpsBuffer=${gpsAccuracyBuffer.toStringAsFixed(1)}m, '
            'effectiveRadius=${effectiveRadius.toStringAsFixed(1)}m, '
            'within=$withinGeofence');
      }

      if (mounted) {
        setState(() {
          _currentPosition = position;
          _distanceToBranch = distance;
          _isWithinGeofence = withinGeofence;
          _locationError = null;
          _isLocationLoading = false;
        });
      }
    } catch (e) {
      debugPrint('Error getting location: $e');
      if (mounted) {
        setState(() {
          _locationError = "Terjadi kesalahan saat mencari lokasi.";
          _isLocationLoading = false;
        });
      }
    }
  }

  Future<void> _processCameraImage(
      CameraImage image, CameraDescription camera) async {
    _isProcessing = true;
    try {
      final inputImage =
          _faceDetectorService.createInputImageFromCameraImage(image, camera);
      if (inputImage == null) {
        _isProcessing = false;
        return;
      }

      final faces = await _faceDetectorService.processImage(inputImage);

      if (faces.isEmpty) {
        if (mounted && _isFaceDetected) {
          setState(() {
            _isFaceDetected = false;
            _statusText = "Mencari wajah...";
          });
        }
      } else if (faces.length > 1) {
        if (mounted) {
          setState(() {
            _isFaceDetected = false;
            _statusText = "Terdeteksi lebih dari satu wajah.";
          });
        }
      } else {
        final face = faces.first;
        if (face.headEulerAngleY! > 20 ||
            face.headEulerAngleY! < -20 ||
            face.headEulerAngleZ! > 20 ||
            face.headEulerAngleZ! < -20) {
          if (mounted) {
            setState(() {
              _isFaceDetected = false;
              _statusText = "Tatap lurus ke kamera.";
            });
          }
        } else {
          // Face is good
          if (mounted && !_isFaceDetected) {
            setState(() {
              _isFaceDetected = true;
              _statusText = "Wajah terdeteksi.";
            });
          }

          if (widget.isRegistration) {
            // For registration, we just need to detect
          } else if (_baselineEmbedding != null && !_isFaceMatched) {
            await _extractAndVerifyFace(image, face);
          }
        }
      }
    } catch (e) {
      // Suppress format errors in logs as we have a manual fallback
      if (e.toString().contains("InputImageConverterError")) {
        if (mounted && _statusText == "Mencari wajah...") {
          setState(() => _statusText = "Posisikan wajah & ambil foto.");
        }
      } else {
        debugPrint("Error processing image: $e");
      }
    }
    _isProcessing = false;
  }

  Future<void> _extractAndVerifyFace(CameraImage image, Face face) async {
    try {
      imglib.Image convertedImage = _convertCameraImage(image);
      if (!_mlService.isInitialized) {
        debugPrint("MLService not initialized, skipping verification.");
        return;
      }
      final currentEmbedding = _mlService.predict(convertedImage);

      if (_baselineEmbedding != null) {
        final distance = _mlService.calculateEuclideanDistance(
            currentEmbedding, _baselineEmbedding!);
        // Adjusted threshold to AppConstants.faceVerificationThreshold (currently 1.7) for a looser match as requested
        final isMatched = distance < AppConstants.faceVerificationThreshold;

        if (mounted) {
          setState(() {
            _isFaceMatched = isMatched;
            _statusText =
                isMatched ? "Wajah diverifikasi." : "Wajah tidak sesuai.";
          });

          // AUTO-CAPTURE: If matched, trigger capture automatically
          if (isMatched && !widget.isRegistration && !_isVerifying) {
            _takePicture();
          }
        }
      }
    } catch (e) {
      debugPrint("Error verifying face: $e");
    }
  }

  Future<Uint8List> _burnWatermark({
    required Uint8List bytes,
    required String timeText,
    required String addressText,
    required String branchName,
  }) async {
    final codec = await ui.instantiateImageCodec(bytes);
    final frame = await codec.getNextFrame();
    final srcImg = frame.image;

    final targetWidth = srcImg.width.toDouble();
    final targetHeight = srcImg.height.toDouble();

    final recorder = ui.PictureRecorder();
    final canvas = Canvas(recorder, Rect.fromLTWH(0, 0, targetWidth, targetHeight));

    // Draw Image
    canvas.drawImage(srcImg, Offset.zero, Paint());

    final padding = 16.0;
    
    // We will measure texts first to determine the dynamic box height
    final textSpans = <TextSpan>[];
    void addIconText(IconData icon, String text, {bool isAddress = false}) {
      textSpans.add(TextSpan(
        children: [
          TextSpan(
            text: String.fromCharCode(icon.codePoint),
            style: TextStyle(
              fontFamily: icon.fontFamily,
              package: icon.fontPackage,
              fontSize: 16,
              color: Colors.white,
            ),
          ),
          TextSpan(
            text: "  $text\n", // add newline
            style: GoogleFonts.plusJakartaSans(
              fontSize: 14,
              color: Colors.white,
            ),
          ),
        ],
      ));
    }

    addIconText(Icons.access_time, timeText);
    addIconText(Icons.storefront, branchName);
    addIconText(Icons.location_on, addressText, isAddress: true);

    final fullSpan = TextSpan(children: textSpans);
    final tp = TextPainter(
      text: fullSpan,
      textDirection: ui.TextDirection.ltr,
    );
    tp.layout(maxWidth: targetWidth - (padding * 2));

    final textHeight = tp.size.height;
    final boxHeight = textHeight + (padding * 2);
    final boxY = targetHeight - boxHeight - 20.0; // 20 padding from bottom

    // Draw semi-transparent background
    final bgPaint = Paint()..color = Colors.black.withOpacity(0.5);
    canvas.drawRect(Rect.fromLTWH(0, boxY, targetWidth, boxHeight), bgPaint);

    // Paint Text
    tp.paint(canvas, Offset(padding, boxY + padding));

    final picture = recorder.endRecording();
    final int outWidth = targetWidth.toInt();
    final int outHeight = targetHeight.toInt();
    final resultImg = await picture.toImage(outWidth, outHeight);
    final byteData =
        await resultImg.toByteData(format: ui.ImageByteFormat.rawRgba);

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

  Future<void> _takePicture() async {
    if (_cameraController == null || !_cameraController!.value.isInitialized)
      return;

    try {
      setState(() => _isVerifying = true);

      // Stop stream if it's running
      try {
        await _cameraController!.stopImageStream();
      } catch (_) {}

      final photo = await _cameraController!.takePicture();

      // Face detection from file (fast because it uses native JNI)
      final faces = await _faceDetectorService.processImageFromFile(photo.path);

      if (faces.isEmpty) {
        if (mounted) {
          SnackBarUtils.showError(
              context, "Wajah tidak ditemukan di foto. Coba lagi.");
          setState(() {
            _isVerifying = false;
            _startCamera();
          });
        }
        return;
      }

      final face = faces.first;

      // OFF-LOAD HEAVY DECODING AND CROPPING TO ISOLATE
      final String timeText = DateFormat('EEEE, dd MMMM yyyy HH:mm:ss').format(DateTime.now());
      final String locationText = _currentPosition != null ? '${_currentPosition!.latitude}, ${_currentPosition!.longitude}' : 'Lokasi tidak diketahui';
      final String addressText = _currentAddress ?? 'Alamat tidak diketahui';

      final ProcessedImageResult? processedResult = await compute(
          _processImageAndWatermarkBackground,
          _FaceProcessingParams(
            photo.path,
            face.boundingBox,
            timeText,
            locationText,
            addressText,
            mirror: true,
          ));

      if (processedResult == null) {
        throw Exception("Gagal memproses file foto.");
      }

      final imglib.Image faceCrop = processedResult.faceCrop;

      if (!_mlService.isInitialized) {
        // Coba init ulang sekali sebelum menyerah
        debugPrint('MLService belum siap, mencoba init ulang...');
        await _mlService.initialize();
        if (!_mlService.isInitialized) {
          throw Exception(
              "Sistem AI belum siap. Mohon tutup dan buka kembali aplikasi.");
        }
      }

      final embedding = _mlService.predict(faceCrop);

      // Burn the watermark dynamically using Canvas
      final branchName = _assignedBranch?['name'] ?? 'Cabang Tidak Diketahui';
      final finalWatermarkedBytes = await _burnWatermark(
        bytes: Uint8List.fromList(processedResult.watermarkedBytes),
        timeText: timeText,
        addressText: addressText,
        branchName: branchName,
      );

      // Save the watermarked full image for upload/registration
      final tempDir = await getTemporaryDirectory();
      final facePath = p.join(tempDir.path,
          'verify_face_${DateTime.now().millisecondsSinceEpoch}.jpg');
      final faceFile = File(facePath);
      await faceFile.writeAsBytes(finalWatermarkedBytes);

      if (widget.isRegistration) {
        if (mounted) {
          setState(() {
            _currentEmbedding = embedding;
            _capturedImagePath = faceFile.path;
            _isVerifying = false;
          });
        }
      } else {
        if (_baselineEmbedding == null) {
          // Handle missing baseline - IMPORTANT: prevents hang
          if (mounted) {
            SnackBarUtils.showError(context,
                "Data wajah Anda tidak ditemukan. Silakan registrasi ulang.");
            setState(() {
              _isVerifying = false;
              _startCamera();
            });
          }
          return;
        }

        // Verify against baseline
        final double distance = _mlService.calculateEuclideanDistance(
            _baselineEmbedding!, embedding);
        // Adjusted threshold to AppConstants.faceVerificationThreshold (currently 1.7) for a looser match as requested
        final bool isMatched =
            distance < AppConstants.faceVerificationThreshold;

        if (mounted) {
          setState(() {
            _currentEmbedding = embedding;
            _handleCapturedImage(faceFile.path, isMatched);
          });
        }
      }
    } catch (e) {
      debugPrint("Take picture logic error: $e");
      if (mounted) {
        // Tampilkan pesan ramah ke user — tidak tampilkan raw exception
        final String userMessage;
        final String errStr = e.toString();
        if (errStr.contains('libtensorflow') ||
            errStr.contains('dynamic library') ||
            errStr.contains('AI belum siap')) {
          userMessage =
              "Sistem AI sedang dimuat, silakan coba lagi dalam beberapa detik.";
        } else if (errStr.contains('Gagal memproses')) {
          userMessage =
              "Pastikan wajah Anda terlihat jelas dan pencahayaan cukup, lalu coba lagi.";
        } else {
          userMessage = "Gagal memproses wajah. Silakan coba lagi.";
        }
        DialogUtils.showError(
            context: context, title: "Gagal Absen", message: userMessage);
        setState(() => _isVerifying = false);
        _startCamera();
      }
    }
  }

  void _handleCapturedImage(String path, bool isMatched) {
    _capturedImagePath = path;
    _isFaceMatched = isMatched;
    _isVerifying = false;
    _statusText = isMatched ? "Wajah diverifikasi." : "Wajah tidak sesuai.";

    if (!isMatched) {
      DialogUtils.showError(
          context: context,
          title: "Verifikasi Gagal",
          message:
              "Wajah tidak dikenali atau tidak sesuai. Pastikan pencahayaan cukup dan wajah terlihat jelas.");
      _startCamera();
    }
  }

  imglib.Image _convertCameraImage(CameraImage image) {
    if (image.format.group == ImageFormatGroup.nv21) {
      return _convertNV21(image);
    } else if (image.format.group == ImageFormatGroup.bgra8888) {
      return _convertBGRA8888(image);
    } else {
      throw Exception('Unsupported image format');
    }
  }

  imglib.Image _convertNV21(CameraImage image) {
    final int width = image.width;
    final int height = image.height;
    final imglib.Image img = imglib.Image(width: width, height: height);
    final plane0 = image.planes[0].bytes;
    final plane1 = image.planes[1].bytes;
    final plane2 = image.planes[2].bytes;

    for (int y = 0; y < height; y++) {
      for (int x = 0; x < width; x++) {
        final int uvIndex = (y >> 1) * image.planes[1].bytesPerRow + (x & ~1);
        final int index = y * image.planes[0].bytesPerRow + x;
        final yp = plane0[index];
        final up = plane1[uvIndex];
        final vp = plane2[uvIndex];
        int r = (yp + vp * 1436 / 1024 - 179).round().clamp(0, 255);
        int g = (yp - up * 46549 / 131072 + 44 - vp * 93604 / 131072 + 91)
            .round()
            .clamp(0, 255);
        int b = (yp + up * 1814 / 1024 - 227).round().clamp(0, 255);
        img.setPixelRgb(x, y, r, g, b);
      }
    }
    return imglib.copyRotate(img, angle: -90);
  }

  imglib.Image _convertBGRA8888(CameraImage image) {
    return imglib.Image.fromBytes(
      width: image.width,
      height: image.height,
      bytes: image.planes[0].bytes.buffer,
      order: imglib.ChannelOrder.bgra,
    );
  }

  void _showSwitchModeDialog() {
    final TextEditingController destCtrl =
        TextEditingController(text: _destination);
    bool tempIsOutOfTown = _isOutOfTown;

    showModalBottomSheet(
      context: context,
      isScrollControlled: true,
      backgroundColor: Colors.transparent,
      builder: (ctx) => StatefulBuilder(
        builder: (ctx, setModalState) => Container(
          padding: EdgeInsets.only(
            left: 24,
            right: 24,
            top: 24,
            bottom: MediaQuery.of(ctx).viewInsets.bottom + 24,
          ),
          decoration: const BoxDecoration(
            color: Colors.white,
            borderRadius: BorderRadius.vertical(top: Radius.circular(28)),
          ),
          child: Column(
            mainAxisSize: MainAxisSize.min,
            crossAxisAlignment: CrossAxisAlignment.start,
            children: [
              Center(
                child: Container(
                  width: 40,
                  height: 4,
                  decoration: BoxDecoration(
                    color: Colors.grey.shade300,
                    borderRadius: BorderRadius.circular(2),
                  ),
                ),
              ),
              const SizedBox(height: 20),
              Text(
                'Pilih Mode Absensi',
                style: GoogleFonts.plusJakartaSans(
                  fontSize: 18,
                  fontWeight: FontWeight.w800,
                  color: AppColors.textPrimary,
                ),
              ),
              const SizedBox(height: 6),
              Text(
                'Tentukan jenis presensi yang akan Anda submit saat ini.',
                style: GoogleFonts.plusJakartaSans(
                  fontSize: 13,
                  color: AppColors.textSecondary,
                ),
              ),
              const SizedBox(height: 20),
              // Option 1: Reguler
              InkWell(
                onTap: () {
                  setModalState(() {
                    tempIsOutOfTown = false;
                  });
                },
                borderRadius: BorderRadius.circular(16),
                child: Container(
                  padding: const EdgeInsets.all(16),
                  decoration: BoxDecoration(
                    color: !tempIsOutOfTown
                        ? AppColors.primaryRed.withOpacity(0.06)
                        : Colors.grey.shade50,
                    borderRadius: BorderRadius.circular(16),
                    border: Border.all(
                      color: !tempIsOutOfTown
                          ? AppColors.primaryRed
                          : Colors.grey.shade200,
                      width: !tempIsOutOfTown ? 2 : 1,
                    ),
                  ),
                  child: Row(
                    children: [
                      Container(
                        padding: const EdgeInsets.all(10),
                        decoration: BoxDecoration(
                          color: !tempIsOutOfTown
                              ? AppColors.primaryRed.withOpacity(0.15)
                              : Colors.grey.shade200,
                          shape: BoxShape.circle,
                        ),
                        child: Icon(
                          Icons.business_rounded,
                          color: !tempIsOutOfTown
                              ? AppColors.primaryRed
                              : AppColors.textTertiary,
                          size: 22,
                        ),
                      ),
                      const SizedBox(width: 14),
                      Expanded(
                        child: Column(
                          crossAxisAlignment: CrossAxisAlignment.start,
                          children: [
                            Text(
                              'Kantor / Reguler (WFO)',
                              style: GoogleFonts.plusJakartaSans(
                                fontSize: 14,
                                fontWeight: FontWeight.w800,
                                color: !tempIsOutOfTown
                                    ? AppColors.primaryRed
                                    : AppColors.textPrimary,
                              ),
                            ),
                            const SizedBox(height: 2),
                            Text(
                              'Presensi reguler di lokasi kantor cabang.',
                              style: GoogleFonts.plusJakartaSans(
                                fontSize: 11,
                                color: AppColors.textSecondary,
                              ),
                            ),
                          ],
                        ),
                      ),
                      if (!tempIsOutOfTown)
                        const Icon(Icons.check_circle_rounded,
                            color: AppColors.primaryRed, size: 22),
                    ],
                  ),
                ),
              ),
              const SizedBox(height: 12),
              // Option 2: Dinas Luar Kota
              InkWell(
                onTap: () {
                  setModalState(() {
                    tempIsOutOfTown = true;
                  });
                },
                borderRadius: BorderRadius.circular(16),
                child: Container(
                  padding: const EdgeInsets.all(16),
                  decoration: BoxDecoration(
                    color: tempIsOutOfTown
                        ? const Color(0xFF4F46E5).withOpacity(0.06)
                        : Colors.grey.shade50,
                    borderRadius: BorderRadius.circular(16),
                    border: Border.all(
                      color: tempIsOutOfTown
                          ? const Color(0xFF4F46E5)
                          : Colors.grey.shade200,
                      width: tempIsOutOfTown ? 2 : 1,
                    ),
                  ),
                  child: Row(
                    children: [
                      Container(
                        padding: const EdgeInsets.all(10),
                        decoration: BoxDecoration(
                          color: tempIsOutOfTown
                              ? const Color(0xFF4F46E5).withOpacity(0.15)
                              : Colors.grey.shade200,
                          shape: BoxShape.circle,
                        ),
                        child: Icon(
                          Icons.flight_takeoff_rounded,
                          color: tempIsOutOfTown
                              ? const Color(0xFF4F46E5)
                              : AppColors.textTertiary,
                          size: 22,
                        ),
                      ),
                      const SizedBox(width: 14),
                      Expanded(
                        child: Column(
                          crossAxisAlignment: CrossAxisAlignment.start,
                          children: [
                            Text(
                              'Dinas Luar Kota',
                              style: GoogleFonts.plusJakartaSans(
                                fontSize: 14,
                                fontWeight: FontWeight.w800,
                                color: tempIsOutOfTown
                                    ? const Color(0xFF4F46E5)
                                    : AppColors.textPrimary,
                              ),
                            ),
                            const SizedBox(height: 2),
                            Text(
                              'Bebas radius lokasi kantor untuk tugas luar.',
                              style: GoogleFonts.plusJakartaSans(
                                fontSize: 11,
                                color: AppColors.textSecondary,
                              ),
                            ),
                          ],
                        ),
                      ),
                      if (tempIsOutOfTown)
                        const Icon(Icons.check_circle_rounded,
                            color: Color(0xFF4F46E5), size: 22),
                    ],
                  ),
                ),
              ),
              if (tempIsOutOfTown) ...[
                const SizedBox(height: 16),
                Text(
                  'Kota / Tempat Tujuan Dinas *',
                  style: GoogleFonts.plusJakartaSans(
                    fontSize: 12,
                    fontWeight: FontWeight.w800,
                    color: const Color(0xFF4F46E5),
                  ),
                ),
                const SizedBox(height: 8),
                TextField(
                  controller: destCtrl,
                  decoration: InputDecoration(
                    hintText: 'Contoh: Surabaya - Kunjungan Proyek',
                    hintStyle: GoogleFonts.plusJakartaSans(
                      fontSize: 13,
                      color: AppColors.textTertiary,
                    ),
                    isDense: true,
                    contentPadding: const EdgeInsets.symmetric(
                        horizontal: 14, vertical: 12),
                    filled: true,
                    fillColor: const Color(0xFF4F46E5).withOpacity(0.05),
                    border: OutlineInputBorder(
                      borderRadius: BorderRadius.circular(12),
                      borderSide: BorderSide.none,
                    ),
                  ),
                  style: GoogleFonts.plusJakartaSans(
                    fontSize: 13,
                    fontWeight: FontWeight.w700,
                    color: AppColors.textPrimary,
                  ),
                ),
              ],
              const SizedBox(height: 24),
              SizedBox(
                width: double.infinity,
                height: 48,
                child: ElevatedButton(
                  onPressed: () {
                    final dest = destCtrl.text.trim();
                    if (tempIsOutOfTown && dest.isEmpty) {
                      SnackBarUtils.showError(ctx,
                          'Harap masukkan kota atau tempat tujuan dinas luar kota.');
                      return;
                    }
                    Navigator.of(ctx).pop();
                    setState(() {
                      _isOutOfTown = tempIsOutOfTown;
                      _destination = tempIsOutOfTown ? dest : '';
                    });
                  },
                  style: ElevatedButton.styleFrom(
                    backgroundColor: tempIsOutOfTown
                        ? const Color(0xFF4F46E5)
                        : AppColors.primaryRed,
                    shape: RoundedRectangleBorder(
                      borderRadius: BorderRadius.circular(14),
                    ),
                  ),
                  child: Text(
                    'Terapkan Mode',
                    style: GoogleFonts.plusJakartaSans(
                      fontSize: 14,
                      fontWeight: FontWeight.w800,
                      color: Colors.white,
                    ),
                  ),
                ),
              ),
            ],
          ),
        ),
      ),
    );
  }

  Future<void> _verifyBiometric() async {
    // Bypass geofencing jika profil user mengizinkan absen dari mana saja atau dalam mode dinas luar kota
    if (!_canAbsenAnywhere && !_isOutOfTown) {
      if (_assignedBranch == null ||
          !_isWithinGeofence ||
          _currentPosition == null) {
        SnackBarUtils.showError(
          context,
          'Lokasi Anda berada di luar radius kantor atau GPS belum mendapatkan koordinat presisi. Harap tunggu atau mendekat ke cabang.',
        );
        return;
      }
    }

    // Pastikan foto sudah terambil sebelum verifikasi sidik jari, atau ambil foto otomatis
    if (_capturedImagePath == null) {
      if (_cameraController != null && _cameraController!.value.isInitialized) {
        await _takePicture();
      }
      if (!mounted) return;
      if (_capturedImagePath == null) {
        SnackBarUtils.showError(
          context,
          'Posisikan wajah ke kamera dan ambil foto bukti presensi terlebih dahulu.',
        );
        return;
      }
    }

    try {
      final authenticated = await di.sl<BiometricService>().authenticate(
            reason: 'Silakan verifikasi sidik jari untuk absensi',
          );
      if (!mounted) return;
      if (authenticated) {
        _submitClockIn();
      } else {
        SnackBarUtils.showError(
            context, 'Verifikasi sidik jari gagal atau dibatalkan.');
      }
    } on BiometricException catch (e) {
      if (!mounted) return;
      SnackBarUtils.showError(context, e.message);
    } catch (e) {
      if (!mounted) return;
      SnackBarUtils.showError(
          context, 'Terjadi kesalahan saat autentikasi biometrik: $e');
    }
  }

  void _submitClockIn() {
    if (widget.isRegistration) {
      if (_capturedImagePath == null) return;
      // Navigator should return the result if it's for registration
      Navigator.of(context).pop(FaceVerificationResult(
        embedding: _currentEmbedding ?? [],
        imagePath: _capturedImagePath!,
      ));
    } else {
      if (_capturedImagePath == null) {
        SnackBarUtils.showError(
            context, 'Foto bukti presensi wajib diambil.');
        return;
      }
      if (_isOutOfTown && _destination.isEmpty) {
        SnackBarUtils.showError(
            context, 'Kota / tujuan dinas luar kota wajib diisi.');
        _showSwitchModeDialog();
        return;
      }
      context.read<AttendanceBloc>().add(
            ClockInRequested(
              isClockIn: widget.isClockIn,
              imagePath: _capturedImagePath,
              latitude: _currentPosition?.latitude,
              longitude: _currentPosition?.longitude,
              faceEmbedding: _currentEmbedding,
              isOutOfTown: _isOutOfTown,
              destination: _destination,
            ),
          );
    }
  }

  @override
  Widget build(BuildContext context) {
    String timeStr = DateFormat('HH:mm').format(_now);
    String amPmStr = DateFormat('a').format(_now);

    return BlocConsumer<AttendanceBloc, AttendanceState>(
      listener: (context, state) {
        if (state is AttendanceSuccess) {
          Navigator.of(context).pushReplacement(
            MaterialPageRoute(
              builder: (_) => widget.isClockIn
                  ? ClockInSuccessPage(
                      attendance: state.attendance,
                      branchName: _assignedBranch?['name'] ??
                          _assignedBranch?['Name'] ??
                          'Lokasi Terdaftar',
                      shiftName: _shiftName,
                      shiftStartTime: _shiftStartTime,
                      shiftEndTime: _shiftEndTime,
                    )
                  : ClockOutSuccessPage(
                      attendance: state.attendance,
                      branchName: _assignedBranch?['name'] ??
                          _assignedBranch?['Name'] ??
                          'Lokasi Terdaftar',
                    ),
            ),
          );
        } else if (state is AttendanceFailure) {
          Navigator.of(context).pushReplacement(
            MaterialPageRoute(
              builder: (_) => widget.isClockIn
                  ? ClockInFailedPage(
                      errorMessage: state.message,
                      isClockIn: widget.isClockIn,
                    )
                  : ClockOutFailedPage(
                      errorMessage: state.message,
                    ),
            ),
          );
        } else if (state is FaceRegistrationSuccess) {
          DialogUtils.showSuccess(
            context: context,
            title: 'Wajah Terdaftar',
            message:
                'Data wajah anda telah berhasil disimpan untuk identitas absensi.',
            onConfirm: () => Navigator.of(context).pop(),
          );
        }
      },
      builder: (context, state) {
        bool isLoading =
            state is AttendanceLoading || state is FaceRegistrationLoading;

        return Scaffold(
          backgroundColor: AppColors.backgroundAlt,
          appBar: AppBar(
            title: Text(
              widget.isRegistration
                  ? 'VERIFIKASI WAJAH'
                  : (widget.isClockIn ? 'ABSEN MASUK' : 'ABSEN KELUAR'),
              style: const TextStyle(
                  fontSize: 16,
                  fontWeight: FontWeight.w800,
                  letterSpacing: 1.5,
                  color: AppColors.textPrimary),
            ),
            centerTitle: true,
            backgroundColor: Colors.white,
            elevation: 0,
            leading: IconButton(
              icon: const Icon(Icons.arrow_back_ios_new_rounded,
                  color: AppColors.textPrimary, size: 20),
              onPressed: () => Navigator.of(context).pop(),
            ),
          ),
          body: SingleChildScrollView(
            child: Column(
              children: [
                // Top Info Section
                Container(
                  width: double.infinity,
                  color: Colors.white,
                  padding: const EdgeInsets.symmetric(vertical: 32),
                  child: Column(
                    children: [
                      if (!widget.isRegistration) ...[
                        Container(
                          padding: const EdgeInsets.symmetric(
                              horizontal: 16, vertical: 8),
                          decoration: BoxDecoration(
                            color: AppColors.primaryRed.withOpacity(0.08),
                            borderRadius: BorderRadius.circular(20),
                          ),
                          child: Text(
                            '${_shiftName.toUpperCase()} ($_shiftStartTime${_shiftEndTime.isNotEmpty ? ' - $_shiftEndTime' : ''})',
                            style: const TextStyle(
                                fontSize: 11,
                                fontWeight: FontWeight.w800,
                                letterSpacing: 1,
                                color: AppColors.primaryRed),
                          ),
                        ),
                        const SizedBox(height: 16),
                      ],
                      Row(
                        mainAxisAlignment: MainAxisAlignment.center,
                        crossAxisAlignment: CrossAxisAlignment.end,
                        children: [
                          Text(
                            timeStr,
                            style: const TextStyle(
                                fontSize: 48,
                                fontWeight: FontWeight.w900,
                                color: AppColors.textPrimary,
                                letterSpacing: -1),
                          ),
                          Padding(
                            padding:
                                const EdgeInsets.only(bottom: 8.0, left: 4.0),
                            child: Text(
                              amPmStr.toUpperCase(),
                              style: const TextStyle(
                                  fontSize: 18,
                                  fontWeight: FontWeight.w800,
                                  color: AppColors.textTertiary),
                            ),
                          ),
                        ],
                      ),
                    ],
                  ),
                ),

                if (!widget.isRegistration) ...[
                  const SizedBox(height: 16),
                  Padding(
                    padding: const EdgeInsets.symmetric(horizontal: 24),
                    child: Container(
                      padding: const EdgeInsets.all(14),
                      decoration: BoxDecoration(
                        color: _isOutOfTown
                            ? const Color(0xFF4F46E5).withOpacity(0.08)
                            : AppColors.primaryRed.withOpacity(0.05),
                        borderRadius: BorderRadius.circular(18),
                        border: Border.all(
                          color: _isOutOfTown
                              ? const Color(0xFF4F46E5).withOpacity(0.3)
                              : AppColors.primaryRed.withOpacity(0.1),
                        ),
                      ),
                      child: Row(
                        children: [
                          Icon(
                            _isOutOfTown
                                ? Icons.flight_takeoff_rounded
                                : Icons.business_rounded,
                            color: _isOutOfTown
                                ? const Color(0xFF4F46E5)
                                : AppColors.primaryRed,
                            size: 24,
                          ),
                          const SizedBox(width: 12),
                          Expanded(
                            child: Column(
                              crossAxisAlignment: CrossAxisAlignment.start,
                              children: [
                                Text(
                                  _isOutOfTown
                                      ? 'MODE DINAS LUAR KOTA'
                                      : 'MODE KANTOR / REGULER',
                                  style: TextStyle(
                                    fontSize: 11,
                                    fontWeight: FontWeight.w900,
                                    color: _isOutOfTown
                                        ? const Color(0xFF4F46E5)
                                        : AppColors.primaryRed,
                                    letterSpacing: 0.5,
                                  ),
                                ),
                                const SizedBox(height: 2),
                                Text(
                                  _isOutOfTown
                                      ? (_destination.isNotEmpty
                                          ? 'Tujuan: $_destination'
                                          : 'Bebas radius lokasi kantor')
                                      : 'Wajib berada dalam radius kantor cabang ${_assignedBranch?['name'] ?? ''}.',
                                  style: const TextStyle(
                                    fontSize: 12,
                                    fontWeight: FontWeight.w600,
                                    color: AppColors.textPrimary,
                                    height: 1.3,
                                  ),
                                ),
                              ],
                            ),
                          ),
                          const SizedBox(width: 8),
                          TextButton(
                            onPressed: _showSwitchModeDialog,
                            style: TextButton.styleFrom(
                              padding: const EdgeInsets.symmetric(
                                  horizontal: 10, vertical: 6),
                              backgroundColor: _isOutOfTown
                                  ? const Color(0xFF4F46E5)
                                  : AppColors.primaryRed,
                              foregroundColor: Colors.white,
                              shape: RoundedRectangleBorder(
                                borderRadius: BorderRadius.circular(10),
                              ),
                              minimumSize: Size.zero,
                              tapTargetSize: MaterialTapTargetSize.shrinkWrap,
                            ),
                            child: Text(
                              _isOutOfTown ? 'Ubah' : 'Ganti Dinas',
                              style: const TextStyle(
                                fontSize: 11,
                                fontWeight: FontWeight.w800,
                              ),
                            ),
                          ),
                        ],
                      ),
                    ),
                  ),
                  const SizedBox(height: 8),
                  // Location & Geofencing Details
                  Container(
                    margin: const EdgeInsets.symmetric(horizontal: 24),
                    padding: const EdgeInsets.all(20),
                    decoration: BoxDecoration(
                      color: Colors.white,
                      borderRadius: BorderRadius.circular(24),
                      boxShadow: [
                        BoxShadow(
                            color: Colors.black.withOpacity(0.04),
                            blurRadius: 20,
                            offset: const Offset(0, 10))
                      ],
                    ),
                    child: Column(
                      children: [
                        Row(
                          crossAxisAlignment: CrossAxisAlignment.center,
                          children: [
                            Container(
                              padding: const EdgeInsets.all(12),
                              decoration: BoxDecoration(
                                  color: (_locationError != null)
                                      ? AppColors.primaryRed.withOpacity(0.08)
                                      : (_isWithinGeofence
                                              ? const Color(0xFF10B981)
                                              : AppColors.primaryRed)
                                          .withOpacity(0.08),
                                  borderRadius: BorderRadius.circular(16)),
                              child: Icon(Icons.person_pin_circle_rounded,
                                  color: (_locationError != null)
                                      ? AppColors.primaryRed
                                      : (_isWithinGeofence
                                          ? const Color(0xFF10B981)
                                          : AppColors.primaryRed),
                                  size: 24),
                            ),
                            const SizedBox(width: 16),
                            Expanded(
                              child: Column(
                                crossAxisAlignment: CrossAxisAlignment.start,
                                children: [
                                  Row(
                                    children: [
                                      const Text('LOKASI SAYA SAAT INI',
                                          style: TextStyle(
                                              fontSize: 10,
                                              letterSpacing: 1,
                                              fontWeight: FontWeight.w800,
                                              color: AppColors.textTertiary)),
                                      if (_isLocationLoading) ...[
                                        const SizedBox(width: 8),
                                        const SizedBox(
                                          width: 10,
                                          height: 10,
                                          child: CircularProgressIndicator(
                                              strokeWidth: 1.5,
                                              color: AppColors.primaryRed),
                                        ),
                                      ],
                                    ],
                                  ),
                                  const SizedBox(height: 4),
                                  if (_locationError != null) ...[
                                    Text(
                                      _locationError!,
                                      style: const TextStyle(
                                          fontSize: 12,
                                          fontWeight: FontWeight.w700,
                                          color: AppColors.primaryRed),
                                    ),
                                  ] else ...[
                                    Text(
                                      _currentAddress ??
                                          (_currentPosition != null
                                              ? '${_currentPosition!.latitude.toStringAsFixed(6)}, ${_currentPosition!.longitude.toStringAsFixed(6)}'
                                              : 'Mencari lokasi...'),
                                      style: const TextStyle(
                                          fontSize: 13,
                                          fontWeight: FontWeight.w700,
                                          color: AppColors.textPrimary),
                                      maxLines: 2,
                                      overflow: TextOverflow.ellipsis,
                                    ),
                                  ],
                                ],
                              ),
                            ),
                          ],
                        ),
                        if (_locationError != null ||
                            _currentPosition == null) ...[
                          const SizedBox(height: 12),
                          Row(
                            mainAxisAlignment: MainAxisAlignment.end,
                            children: [
                              if (_locationError != null &&
                                  _locationError!.contains("GPS"))
                                TextButton.icon(
                                  onPressed: () async {
                                    await Geolocator.openLocationSettings();
                                  },
                                  icon: const Icon(
                                      Icons.settings_input_antenna_rounded,
                                      size: 16,
                                      color: AppColors.primaryRed),
                                  label: const Text(
                                    'Aktifkan GPS',
                                    style: TextStyle(
                                        fontSize: 11,
                                        fontWeight: FontWeight.w800,
                                        color: AppColors.primaryRed),
                                  ),
                                  style: TextButton.styleFrom(
                                    padding: const EdgeInsets.symmetric(
                                        horizontal: 8, vertical: 4),
                                    minimumSize: Size.zero,
                                    tapTargetSize:
                                        MaterialTapTargetSize.shrinkWrap,
                                  ),
                                )
                              else if (_locationError != null &&
                                  _locationError!.contains("Izin"))
                                TextButton.icon(
                                  onPressed: () async {
                                    await Geolocator.openAppSettings();
                                  },
                                  icon: const Icon(Icons.settings_rounded,
                                      size: 16, color: AppColors.primaryRed),
                                  label: const Text(
                                    'Buka Pengaturan',
                                    style: TextStyle(
                                        fontSize: 11,
                                        fontWeight: FontWeight.w800,
                                        color: AppColors.primaryRed),
                                  ),
                                  style: TextButton.styleFrom(
                                    padding: const EdgeInsets.symmetric(
                                        horizontal: 8, vertical: 4),
                                    minimumSize: Size.zero,
                                    tapTargetSize:
                                        MaterialTapTargetSize.shrinkWrap,
                                  ),
                                ),
                              const SizedBox(width: 8),
                              ElevatedButton.icon(
                                onPressed: _isLocationLoading
                                    ? null
                                    : _getCurrentLocation,
                                icon: _isLocationLoading
                                    ? const SizedBox(
                                        width: 12,
                                        height: 12,
                                        child: CircularProgressIndicator(
                                            strokeWidth: 2,
                                            color: Colors.white),
                                      )
                                    : const Icon(Icons.refresh_rounded,
                                        size: 16, color: Colors.white),
                                label: Text(
                                  _isLocationLoading
                                      ? 'Mencari...'
                                      : 'Coba Lagi',
                                  style: const TextStyle(
                                      fontSize: 11,
                                      fontWeight: FontWeight.w800,
                                      color: Colors.white),
                                ),
                                style: ElevatedButton.styleFrom(
                                  backgroundColor: AppColors.primaryRed,
                                  padding: const EdgeInsets.symmetric(
                                      horizontal: 12, vertical: 8),
                                  minimumSize: Size.zero,
                                  tapTargetSize:
                                      MaterialTapTargetSize.shrinkWrap,
                                  shape: RoundedRectangleBorder(
                                      borderRadius: BorderRadius.circular(8)),
                                ),
                              ),
                            ],
                          ),
                        ],
                        const SizedBox(height: 16),
                        const Divider(height: 1),
                        const SizedBox(height: 16),
                        Row(
                          children: [
                            Container(
                              padding: const EdgeInsets.all(12),
                              decoration: BoxDecoration(
                                  color:
                                      AppColors.textTertiary.withOpacity(0.08),
                                  borderRadius: BorderRadius.circular(16)),
                              child: const Icon(Icons.business_rounded,
                                  color: AppColors.textTertiary, size: 24),
                            ),
                            const SizedBox(width: 16),
                            Expanded(
                              child: Column(
                                crossAxisAlignment: CrossAxisAlignment.start,
                                children: [
                                  const Text('TARGET LOKASI CABANG',
                                      style: TextStyle(
                                          fontSize: 10,
                                          letterSpacing: 1,
                                          fontWeight: FontWeight.w800,
                                          color: AppColors.textTertiary)),
                                  const SizedBox(height: 4),
                                  Text(
                                    _assignedBranch != null
                                        ? '${_assignedBranch!['name'] ?? _assignedBranch!['Name']}\n${_targetBranchAddress ?? '(${_assignedBranch?['latitude'] ?? _assignedBranch?['Latitude']}, ${_assignedBranch?['longitude'] ?? _assignedBranch?['Longitude']})'}'
                                        : 'Cabang belum ditugaskan atau gagal dimuat',
                                    style: TextStyle(
                                        fontSize: 13,
                                        fontWeight: FontWeight.w700,
                                        color: _assignedBranch != null
                                            ? AppColors.textPrimary
                                            : AppColors.primaryRed),
                                  ),
                                ],
                              ),
                            ),
                            if (_assignedBranch == null &&
                                !widget.isRegistration) ...[
                              IconButton(
                                icon: const Icon(Icons.sync_problem_rounded,
                                    color: AppColors.primaryRed),
                                onPressed: () {
                                  _fetchBranchAndBaseline();
                                  _getCurrentLocation();
                                },
                                tooltip: 'Segarkan data cabang',
                              ),
                            ],
                          ],
                        ),
                        if (_distanceToBranch >= 0 || _isOutOfTown) ...[
                          const SizedBox(height: 16),
                          Container(
                            padding: const EdgeInsets.symmetric(
                                horizontal: 12, vertical: 8),
                            decoration: BoxDecoration(
                              color: (_isWithinGeofence || _canAbsenAnywhere || _isOutOfTown
                                      ? const Color(0xFF10B981)
                                      : AppColors.primaryRed)
                                  .withOpacity(0.1),
                              borderRadius: BorderRadius.circular(12),
                            ),
                            child: Row(
                              children: [
                                Icon(
                                  _isWithinGeofence || _canAbsenAnywhere || _isOutOfTown
                                      ? Icons.check_circle_rounded
                                      : Icons.warning_rounded,
                                  size: 16,
                                  color: _isWithinGeofence || _canAbsenAnywhere || _isOutOfTown
                                      ? const Color(0xFF10B981)
                                      : AppColors.primaryRed,
                                ),
                                const SizedBox(width: 8),
                                Expanded(
                                  child: Text(
                                    _isOutOfTown
                                        ? 'Mode Dinas Luar Kota: ${_destination.isNotEmpty ? _destination : "Lokasi Tugas"}'
                                        : (_isWithinGeofence
                                            ? 'Berada dalam radius (Jarak: ${_distanceToBranch.toStringAsFixed(0)}m)'
                                            : (_canAbsenAnywhere
                                                ? 'Bypass Radius Aktif (Jarak: ${_distanceToBranch.toStringAsFixed(0)}m)'
                                                : 'Di luar radius (Jarak: ${_distanceToBranch.toStringAsFixed(0)}m)')),
                                    style: TextStyle(
                                      fontSize: 12,
                                      fontWeight: FontWeight.w700,
                                      color: _isWithinGeofence || _canAbsenAnywhere || _isOutOfTown
                                          ? const Color(0xFF10B981)
                                          : AppColors.primaryRed,
                                    ),
                                  ),
                                ),
                              ],
                            ),
                          ),
                        ],
                      ],
                    ),
                  ),
                ],

                const SizedBox(height: 32),

                // Circular Camera Section
                if (_capturedImagePath == null) ...[
                  SizedBox(
                    height: 380,
                    child: Center(
                      child: LayoutBuilder(
                        builder: (context, constraints) {
                          final circleSize = constraints.maxWidth * 0.75;

                          return Container(
                            width: circleSize,
                            height: circleSize,
                            decoration: BoxDecoration(
                              shape: BoxShape.circle,
                              border: Border.all(
                                  color: _isFaceDetected
                                      ? (widget.isRegistration || _isFaceMatched
                                          ? const Color(0xFF10B981)
                                          : Colors.orange)
                                      : AppColors.grayBorder,
                                  width: 4),
                              boxShadow: [
                                BoxShadow(
                                    color: (_isFaceDetected
                                            ? Colors.blueAccent
                                            : Colors.black)
                                        .withOpacity(0.1),
                                    blurRadius: 20,
                                    spreadRadius: 5),
                              ],
                            ),
                            clipBehavior: Clip.antiAlias,
                            child: _cameraController != null &&
                                    _cameraController!.value.isInitialized
                                ? FittedBox(
                                    fit: BoxFit.cover,
                                    child: SizedBox(
                                      width: constraints.maxWidth,
                                      height: constraints.maxWidth *
                                          _cameraController!.value.aspectRatio,
                                      child: CameraPreview(_cameraController!),
                                    ),
                                  )
                                : _cameraError != null
                                    ? Center(
                                        child: Padding(
                                          padding: const EdgeInsets.all(16.0),
                                          child: Column(
                                            mainAxisAlignment:
                                                MainAxisAlignment.center,
                                            children: [
                                              const Icon(
                                                Icons.videocam_off_rounded,
                                                color: AppColors.primaryRed,
                                                size: 36,
                                              ),
                                              const SizedBox(height: 8),
                                              Text(
                                                _cameraError!,
                                                textAlign: TextAlign.center,
                                                style: GoogleFonts.plusJakartaSans(
                                                  fontSize: 11,
                                                  fontWeight: FontWeight.w700,
                                                  color: AppColors.textPrimary,
                                                ),
                                              ),
                                              const SizedBox(height: 12),
                                              ElevatedButton.icon(
                                                onPressed: _startCamera,
                                                icon: const Icon(
                                                    Icons.refresh_rounded,
                                                    size: 14,
                                                    color: Colors.white),
                                                label: Text(
                                                  'Coba Lagi',
                                                  style: GoogleFonts
                                                      .plusJakartaSans(
                                                    fontSize: 11,
                                                    fontWeight: FontWeight.w700,
                                                    color: Colors.white,
                                                  ),
                                                ),
                                                style: ElevatedButton.styleFrom(
                                                  backgroundColor:
                                                      AppColors.primaryRed,
                                                  padding:
                                                      const EdgeInsets.symmetric(
                                                          horizontal: 14,
                                                          vertical: 6),
                                                  shape: RoundedRectangleBorder(
                                                      borderRadius:
                                                          BorderRadius.circular(
                                                              10)),
                                                ),
                                              )
                                            ],
                                          ),
                                        ),
                                      )
                                    : const Center(
                                        child: CircularProgressIndicator()),
                          );
                        },
                      ),
                    ),
                  ),
                  Container(
                    width: double.infinity,
                    margin: const EdgeInsets.symmetric(horizontal: 48),
                    child: Column(
                      children: [
                        Text(
                          _statusText,
                          textAlign: TextAlign.center,
                          style: TextStyle(
                              color: _isFaceMatched
                                  ? const Color(0xFF10B981)
                                  : AppColors.textPrimary,
                              fontWeight: FontWeight.w800,
                              fontSize: 15),
                        ),
                        const SizedBox(height: 24),
                        SizedBox(
                          width: double.infinity,
                          child: ElevatedButton.icon(
                            onPressed: (!_isVerifying &&
                                    _cameraController != null &&
                                    _cameraController!.value.isInitialized)
                                ? _takePicture
                                : null,
                            icon: const Icon(Icons.camera_alt_rounded,
                                color: Colors.white, size: 20),
                            label: Text(
                                _isVerifying ? 'MEMPROSES...' : 'AMBIL FOTO'),
                            style: ElevatedButton.styleFrom(
                              backgroundColor: AppColors.primaryRed,
                              disabledBackgroundColor: Colors.grey.shade300,
                              shape: RoundedRectangleBorder(
                                  borderRadius: BorderRadius.circular(16)),
                              padding: const EdgeInsets.symmetric(vertical: 16),
                            ),
                          ),
                        ),
                      ],
                    ),
                  ),
                  if (_isBiometricEnabled && !widget.isRegistration) ...[
                    const SizedBox(height: 16),
                    Padding(
                      padding: const EdgeInsets.symmetric(horizontal: 48),
                      child: OutlinedButton.icon(
                        onPressed: (isLoading ||
                                _isVerifying ||
                                (!_isOutOfTown &&
                                    (_assignedBranch == null ||
                                        (!_isWithinGeofence && !_canAbsenAnywhere))))
                            ? null
                            : _verifyBiometric,
                        icon: const Icon(Icons.fingerprint_rounded, size: 24),
                        label: const Text('VERIFIKASI SIDIK JARI'),
                        style: OutlinedButton.styleFrom(
                          padding: const EdgeInsets.symmetric(vertical: 14),
                          side: BorderSide(
                              color: (!_isOutOfTown &&
                                      (_assignedBranch == null ||
                                          (!_isWithinGeofence && !_canAbsenAnywhere)))
                                  ? Colors.grey
                                  : const Color(0xFF10B981),
                              width: 2),
                          foregroundColor:
                              (!_isOutOfTown &&
                                      (_assignedBranch == null ||
                                          (!_isWithinGeofence && !_canAbsenAnywhere)))
                                  ? Colors.grey
                                  : const Color(0xFF10B981),
                          shape: RoundedRectangleBorder(
                              borderRadius: BorderRadius.circular(16)),
                        ),
                      ),
                    ),
                  ],
                ] else ...[
                  // Captured Image Preview
                  Center(
                    child: Container(
                      height: 280,
                      width: 280,
                      decoration: BoxDecoration(
                        shape: BoxShape.circle,
                        border: Border.all(
                            color: _isFaceMatched || widget.isRegistration
                                ? const Color(0xFF10B981)
                                : AppColors.primaryRed,
                            width: 4),
                        image: DecorationImage(
                          image: kIsWeb
                              ? NetworkImage(_capturedImagePath!)
                                  as ImageProvider
                              : FileImage(File(_capturedImagePath!)),
                          fit: BoxFit.cover,
                        ),
                        boxShadow: [
                          BoxShadow(
                              color: (_isFaceMatched || widget.isRegistration
                                      ? const Color(0xFF10B981)
                                      : AppColors.primaryRed)
                                  .withOpacity(0.2),
                              blurRadius: 30,
                              offset: const Offset(0, 15))
                        ],
                      ),
                    ),
                  ),
                  const SizedBox(height: 24),
                  TextButton.icon(
                    onPressed: () {
                      setState(() {
                        _capturedImagePath = null;
                        _isFaceMatched = false;
                        _isFaceDetected = false;
                        _statusText = "Mencari wajah...";
                      });
                      _startCamera();
                    },
                    icon: const Icon(Icons.refresh_rounded,
                        color: AppColors.primaryRed),
                    label: const Text('Ambil Ulang Foto',
                        style: TextStyle(
                            color: AppColors.primaryRed,
                            fontWeight: FontWeight.w800)),
                  ),
                  const SizedBox(height: 32),

                  Container(
                    width: double.infinity,
                    margin: const EdgeInsets.symmetric(horizontal: 32),
                    height: 56,
                    decoration: BoxDecoration(
                      borderRadius: BorderRadius.circular(16),
                      gradient: const LinearGradient(
                        colors: AppColors.primaryGradient,
                        begin: Alignment.topLeft,
                        end: Alignment.bottomRight,
                      ),
                    ),
                    child: ElevatedButton(
                      onPressed: (isLoading ||
                              _isVerifying ||
                              (!widget.isRegistration &&
                                  !_isOutOfTown &&
                                  (_assignedBranch == null ||
                                      (!_isWithinGeofence && !_canAbsenAnywhere))))
                          ? null
                          : _submitClockIn,
                      style: ElevatedButton.styleFrom(
                        backgroundColor: Colors.transparent,
                        shadowColor: Colors.transparent,
                        disabledBackgroundColor: Colors.transparent,
                      ),
                      child: (isLoading || _isVerifying)
                          ? const SizedBox(
                              height: 24,
                              width: 24,
                              child: CircularProgressIndicator(
                                  color: Colors.white, strokeWidth: 3))
                          : Text(
                              (!widget.isRegistration &&
                                      !_isOutOfTown &&
                                      _assignedBranch == null)
                                  ? 'GAGAL MEMUAT DATA CABANG'
                                  : (!widget.isRegistration &&
                                          !_isOutOfTown &&
                                          !_isWithinGeofence &&
                                          !_canAbsenAnywhere)
                                      ? 'DI LUAR RADIUS'
                                      : (widget.isRegistration
                                          ? 'SIMPAN REGISTRASI'
                                          : (widget.isClockIn
                                              ? (_isOutOfTown
                                                  ? 'SUBMIT DINAS MASUK'
                                                  : 'SUBMIT ABSEN MASUK')
                                              : (_isOutOfTown
                                                  ? 'SUBMIT DINAS KELUAR'
                                                  : 'SUBMIT ABSEN KELUAR'))),
                              style: const TextStyle(
                                  fontSize: 16,
                                  fontWeight: FontWeight.w800,
                                  letterSpacing: 1,
                                  color: Colors.white),
                            ),
                    ),
                  ),
                  const SizedBox(height: 16),
                  TextButton(
                    onPressed: () {
                      setState(() {
                        _capturedImagePath = null;
                        _isFaceMatched = false;
                        _isFaceDetected = false;
                        _statusText = "Mencari wajah...";
                      });
                      _startCamera();
                    },
                    child: const Text('Ambil Ulang Foto',
                        style: TextStyle(
                            color: AppColors.textTertiary,
                            fontWeight: FontWeight.w700)),
                  ),
                ],
                const SizedBox(height: 40),
              ],
            ),
          ),
        );
      },
    );
  }
}
