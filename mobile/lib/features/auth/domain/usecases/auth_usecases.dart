import 'package:dartz/dartz.dart';
import 'package:hris_app/core/error/failures.dart';
import 'package:hris_app/features/auth/domain/repositories/auth_repository.dart';

class LoginUseCase {
  final AuthRepository repository;

  LoginUseCase(this.repository);

  Future<Either<Failure, void>> call(String email, String password) {
    return repository.login(email, password);
  }
}

class CheckAuthStatusUseCase {
  final AuthRepository repository;

  CheckAuthStatusUseCase(this.repository);

  Future<Either<Failure, bool>> call() {
    return repository.checkAuthStatus();
  }
}

class LogoutUseCase {
  final AuthRepository repository;

  LogoutUseCase(this.repository);

  Future<Either<Failure, void>> call() {
    return repository.logout();
  }
}

class RegisterUseCase {
  final AuthRepository repository;

  RegisterUseCase(this.repository);

  Future<Either<Failure, void>> call(
      String name,
      String email,
      String employeeId,
      String password,
      String jobPositionId,
      String branchId,
      String? shiftId,
      {List<double>? embedding,
      String? selfiePath}) {
    return repository.register(
        name, email, employeeId, password, jobPositionId, branchId, shiftId,
        embedding: embedding, selfiePath: selfiePath);
  }
}

class RegisterFaceUseCase {
  final AuthRepository repository;

  RegisterFaceUseCase(this.repository);

  Future<Either<Failure, void>> call(
      List<double> embedding, String selfiePath) {
    return repository.registerFace(embedding, selfiePath);
  }
}

class GetProfileUseCase {
  final AuthRepository repository;

  GetProfileUseCase(this.repository);

  Future<Either<Failure, Map<String, dynamic>>> call() {
    return repository.getProfile();
  }
}

class ChangePasswordUseCase {
  final AuthRepository repository;

  ChangePasswordUseCase(this.repository);

  Future<Either<Failure, void>> call(String newPassword) {
    return repository.changePassword(newPassword);
  }
}

class ForgotPasswordUseCase {
  final AuthRepository repository;

  ForgotPasswordUseCase(this.repository);

  Future<Either<Failure, void>> call(String email) {
    return repository.forgotPassword(email);
  }
}
