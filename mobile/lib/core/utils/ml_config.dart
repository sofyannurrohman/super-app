// ---------------------------------------------------------------------------
// ml_config.dart — Konfigurasi sentral build-time flags untuk ML features
// ---------------------------------------------------------------------------
//
// Nilai di-set via --dart-define saat build:
//
//   USE_MLKIT=true  (default) → MLKit & TFLite AKTIF (Android / iPhone fisik)
//   USE_MLKIT=false           → MLKit & TFLite NONAKTIF (iOS Simulator)
//
// Cara pakai di kode lain:
//   import 'package:hris_app/core/utils/ml_config.dart';
//   if (kUseMlKit) { ... }
// ---------------------------------------------------------------------------

/// Konstanta build-time. Default: true (MLKit aktif).
const bool kUseMlKit = bool.fromEnvironment('USE_MLKIT', defaultValue: true);
