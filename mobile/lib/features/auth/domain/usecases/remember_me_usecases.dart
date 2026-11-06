import 'package:dartz/dartz.dart';
import 'package:hris_app/core/error/failures.dart';
import 'package:hris_app/features/auth/domain/repositories/auth_repository.dart';

class SetRememberMeEnabledUseCase {
  final AuthRepository repository;
  SetRememberMeEnabledUseCase(this.repository);

  Future<Either<Failure, void>> call(bool enabled) async {
    return await repository.setRememberMeEnabled(enabled);
  }
}

class GetRememberMeStatusUseCase {
  final AuthRepository repository;
  GetRememberMeStatusUseCase(this.repository);

  Future<bool> call() async {
    final result = await repository.getRememberMeEnabled();
    return result.fold((_) => false, (enabled) => enabled);
  }
}

class GetRememberedCredentialsUseCase {
  final AuthRepository repository;
  GetRememberedCredentialsUseCase(this.repository);

  Future<Either<Failure, Map<String, String>?>> call() async {
    return await repository.getRememberedCredentials();
  }
}

class ClearRememberedCredentialsUseCase {
  final AuthRepository repository;
  ClearRememberedCredentialsUseCase(this.repository);

  Future<Either<Failure, void>> call() async {
    return await repository.clearRememberedCredentials();
  }
}
