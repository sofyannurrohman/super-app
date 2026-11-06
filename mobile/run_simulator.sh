#!/bin/bash
# =============================================================================
# run_simulator.sh — Build & jalankan di iOS Simulator (tanpa MLKit/TFLite)
# =============================================================================
# Cara pakai: ./run_simulator.sh
# 
# Script ini:
#   1. Backup pubspec.yaml → pubspec.yaml.device
#   2. Buat pubspec.yaml tanpa google_mlkit_face_detection & tflite_flutter  
#   3. Jalankan flutter pub get + pod install
#   4. Jalankan flutter run di iOS Simulator
#   5. Saat keluar (Ctrl+C), kembalikan pubspec.yaml ke semula
# =============================================================================

set -e

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
PUBSPEC="$SCRIPT_DIR/pubspec.yaml"
PUBSPEC_BACKUP="$SCRIPT_DIR/pubspec.yaml.device"
ML_SERVICE="$SCRIPT_DIR/lib/core/utils/ml_service.dart"
FACE_SERVICE="$SCRIPT_DIR/lib/core/utils/face_detector_service.dart"

# ─── Cleanup function — dipanggil saat script keluar ──────────────────────
cleanup() {
  echo ""
  echo "🔄 Mengembalikan pubspec.yaml ke mode device (MLKit ON)..."
  
  if [ -f "$PUBSPEC_BACKUP" ]; then
    cp "$PUBSPEC_BACKUP" "$PUBSPEC"
    rm -f "$PUBSPEC_BACKUP"
  fi

  # Kembalikan ml_service dan face_detector_service ke mode device
  echo "export 'ml_service_impl.dart';" > "$ML_SERVICE"
  echo "export 'face_detector_service_impl.dart';" > "$FACE_SERVICE"

  echo "✅ pubspec.yaml dikembalikan. Jalankan 'flutter pub get' untuk sync kembali."
}

# Pastikan cleanup selalu berjalan saat script selesai
trap cleanup EXIT INT TERM

echo "📱 Mode: iOS Simulator (MLKit DISABLED)"
echo "   Fitur face recognition dinonaktifkan untuk simulator testing"
echo ""

# ─── 1. Backup pubspec.yaml ────────────────────────────────────────────────
echo "💾 Backup pubspec.yaml..."
cp "$PUBSPEC" "$PUBSPEC_BACKUP"

# ─── 2. Hapus MLKit & TFLite dari pubspec.yaml ────────────────────────────
echo "✂️  Menonaktifkan MLKit dari pubspec.yaml..."
sed -i '' \
  -e 's/^  google_mlkit_face_detection:.*$/#  google_mlkit_face_detection: (disabled for simulator)/' \
  -e 's/^  tflite_flutter:.*$/#  tflite_flutter: (disabled for simulator)/' \
  "$PUBSPEC"

# ─── 3. Arahkan ml_service ke stub ────────────────────────────────────────
echo "🔀 Mengalihkan MLService ke stub implementation..."
cat > "$ML_SERVICE" << 'EOF'
// Simulator mode: gunakan stub (tanpa TFLite)
export 'ml_service_stub.dart';
EOF

# ─── 4. Arahkan face_detector_service ke stub ─────────────────────────────
cat > "$FACE_SERVICE" << 'EOF'
// Simulator mode: gunakan stub (tanpa MLKit)
export 'face_detector_service_stub.dart';
EOF

# ─── 5. flutter pub get + pod install ─────────────────────────────────────
echo "🔄 Sinkronisasi dependencies..."
flutter pub get

echo "🔄 Pod install (tanpa MLKit pods)..."
cd "$SCRIPT_DIR/ios"
pod install --silent
cd "$SCRIPT_DIR"

# ─── 6. Jalankan di Simulator ─────────────────────────────────────────────
echo ""
echo "🚀 Menjalankan di iOS Simulator (tekan Ctrl+C untuk keluar)..."
echo ""
flutter run \
  --dart-define=USE_MLKIT=false \
  "$@"

# cleanup() dipanggil otomatis oleh trap saat script selesai
