// ml_service.dart — Entry point untuk MLService.
//
// Platform routing:
//   Web (Chrome)          → stub  (dart:ffi / dart:io tidak tersedia di web)
//   Android / iOS fisik   → impl  (MLKit + TFLite penuh)
//   iOS Simulator         → impl  (dart:io tersedia, tapi kUseMlKit=false men-guard calls-nya)
//
// Untuk iOS Simulator, jalankan: ./run_simulator.sh
// Untuk Android / iPhone fisik:  ./run_device.sh  atau  flutter run
export 'ml_service_stub.dart' if (dart.library.io) 'ml_service_impl.dart';
