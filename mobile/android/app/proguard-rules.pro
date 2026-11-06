# ============================================================
# TensorFlow Lite ProGuard Rules
# WAJIB agar libtensorflowlite_jni.so tidak di-strip oleh R8/ProGuard
# ============================================================

# Core TFLite classes
-keep class org.tensorflow.** { *; }
-keep class org.tensorflow.lite.** { *; }
-keep class org.tensorflow.lite.gpu.** { *; }
-keep class org.tensorflow.lite.nnapi.** { *; }
-keep class org.tensorflow.lite.delegates.** { *; }

# Keep native methods (JNI)
-keepclasseswithmembernames class * {
    native <methods>;
}

# Suppress warnings
-dontwarn org.tensorflow.lite.**
-dontwarn org.tensorflow.lite.gpu.**
-dontwarn org.tensorflow.lite.nnapi.**

# ============================================================
# Google ML Kit (face detection)
# ============================================================
-keep class com.google.mlkit.** { *; }
-keep class com.google.android.gms.internal.mlkit_vision_face.** { *; }
-dontwarn com.google.mlkit.**
