import 'package:dartz/dartz.dart';
import 'package:hris_app/core/error/failures.dart';
import 'package:hris_app/features/auth/domain/repositories/auth_repository.dart';

class BiometricLoginUseCase {
  final AuthRepository repository;

  BiometricLoginUseCase(this.repository);

  Future<Either<Failure, void>> call() async {
    return await repository.loginWithBiometric();
  }
}

class GetBiometricStatusUseCase {
  final AuthRepository repository;

  GetBiometricStatusUseCase(this.repository);

  Future<Map<String, bool>> call() async {
    final isSupported = await repository.isBiometricSupported();
    final isEnabled = await repository.getBiometricEnabled();

    bool supported = false;
    bool enabled = false;

    isSupported.fold((_) => supported = false, (v) => supported = v);
    isEnabled.fold((_) => enabled = false, (v) => enabled = v);

    return {'supported': supported, 'enabled': enabled};
  }
}

class SetBiometricEnabledUseCase {
  final AuthRepository repository;

  SetBiometricEnabledUseCase(this.repository);

  Future<Either<Failure, void>> call(bool enabled) async {
    return await repository.setBiometricEnabled(enabled);
  }
}
