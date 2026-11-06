#!/bin/bash
# =============================================================================
# run_device.sh — Build & jalankan di Android / iPhone fisik (MLKit penuh)
# =============================================================================
# Cara pakai: 
#   ./run_device.sh                    (deteksi device otomatis)
#   ./run_device.sh -d <device_id>    (pilih device tertentu)
# =============================================================================

set -e

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
PUBSPEC="$SCRIPT_DIR/pubspec.yaml"
PUBSPEC_BACKUP="$SCRIPT_DIR/pubspec.yaml.device"
ML_SERVICE="$SCRIPT_DIR/lib/core/utils/ml_service.dart"
FACE_SERVICE="$SCRIPT_DIR/lib/core/utils/face_detector_service.dart"

echo "🤖 Mode: Physical Device (MLKit ENABLED)"
echo "   Fitur face recognition aktif penuh — cocok untuk Android & iPhone fisik"
echo ""

# ─── Pastikan pubspec.yaml dalam kondisi full (restore jika perlu) ─────────
if [ -f "$PUBSPEC_BACKUP" ]; then
  echo "🔄 Mengembalikan pubspec.yaml dari backup..."
  cp "$PUBSPEC_BACKUP" "$PUBSPEC"
  rm -f "$PUBSPEC_BACKUP"
else
  # Pastikan MLKit tidak dalam kondisi dikomentari
  sed -i '' \
    -e 's/^#  google_mlkit_face_detection:.*$/  google_mlkit_face_detection: ^0.13.0/' \
    -e 's/^#  tflite_flutter:.*$/  tflite_flutter: ^0.12.1/' \
    "$PUBSPEC"
fi

# ─── Pastikan ml_service dan face_detector_service ke implementasi asli ────
echo "// Device mode: gunakan MLKit & TFLite implementation (real)" > "$ML_SERVICE"
echo "export 'ml_service_impl.dart';" >> "$ML_SERVICE"

echo "// Device mode: gunakan MLKit implementation (real)" > "$FACE_SERVICE"
echo "export 'face_detector_service_impl.dart';" >> "$FACE_SERVICE"

# ─── Sinkronisasi dependencies ─────────────────────────────────────────────
echo "🔄 Sinkronisasi dependencies..."
flutter pub get

# ─── Pod install hanya jika kita build untuk iOS ──────────────────────────
if [[ "$*" == *"-d"* ]] || flutter devices 2>/dev/null | grep -q "ios"; then
  echo "🔄 Pod install (dengan MLKit pods)..."
  cd "$SCRIPT_DIR/ios"
  pod install --silent 2>/dev/null || true
  cd "$SCRIPT_DIR"
fi

# ─── Jalankan di device ────────────────────────────────────────────────────
echo ""
echo "🚀 Menjalankan di device (tekan Ctrl+C untuk keluar)..."
echo ""
flutter run \
  --dart-define=USE_MLKIT=true \
  "$@"
