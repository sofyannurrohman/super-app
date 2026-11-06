import 'package:flutter/foundation.dart';
import 'package:flutter/services.dart';
import 'package:local_auth/local_auth.dart';

class BiometricException implements Exception {
  final String message;
  BiometricException(this.message);

  @override
  String toString() => message;
}

class BiometricService {
  final LocalAuthentication _auth = LocalAuthentication();

  Future<bool> isBiometricAvailable() async {
    final bool canAuthenticateWithBiometrics = await _auth.canCheckBiometrics;
    return canAuthenticateWithBiometrics;
  }

  Future<List<BiometricType>> getAvailableBiometrics() async {
    try {
      return await _auth.getAvailableBiometrics();
    } on PlatformException {
      return <BiometricType>[];
    }
  }

  Future<bool> authenticate({
    required String reason,
    bool localizedReason = true,
  }) async {
    try {
      return await _auth.authenticate(
        localizedReason: reason,
        options: const AuthenticationOptions(
          stickyAuth: true,
          biometricOnly: true,
        ),
      );
    } on PlatformException catch (e) {
      debugPrint(
          "Biometric authenticate PlatformException: ${e.code} - ${e.message}");
      switch (e.code) {
        case 'NotAvailable':
          throw BiometricException(
              'Sensor biometrik tidak tersedia atau tidak didukung pada perangkat ini.');
        case 'NotEnrolled':
          throw BiometricException(
              'Sidik jari atau wajah belum didaftarkan di HP Anda. Silakan daftarkan terlebih dahulu di pengaturan keamanan HP.');
        case 'LockedOut':
          throw BiometricException(
              'Autentikasi biometrik terkunci sementara karena terlalu banyak percobaan salah. Silakan coba lagi beberapa saat lagi.');
        case 'PermanentlyLockedOut':
          throw BiometricException(
              'Autentikasi biometrik terkunci permanen. Harap buka kunci perangkat Anda menggunakan PIN/Password terlebih dahulu.');
        case 'passcodeNotSet':
          throw BiometricException(
              'Kunci layar (PIN/Pola/Password) belum diaktifkan di HP Anda.');
        default:
          throw BiometricException(
              'Terjadi kesalahan sistem pada biometrik: ${e.message ?? e.code}');
      }
    } catch (e) {
      debugPrint("Biometric authenticate generic exception: $e");
      throw BiometricException('Gagal memverifikasi biometrik: $e');
    }
  }
}
