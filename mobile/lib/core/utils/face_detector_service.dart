// face_detector_service.dart — Entry point untuk FaceDetectorService.
//
// Platform routing:
//   Web (Chrome)          → stub  (MLKit / dart:io tidak tersedia di web)
//   Android / iOS fisik   → impl  (MLKit penuh)
//   iOS Simulator         → impl  (dart:io tersedia, kUseMlKit=false men-guard calls-nya)
//
// Untuk iOS Simulator, jalankan: ./run_simulator.sh
// Untuk Android / iPhone fisik:  ./run_device.sh  atau  flutter run
export 'face_detector_service_stub.dart'
    if (dart.library.io) 'face_detector_service_impl.dart';
