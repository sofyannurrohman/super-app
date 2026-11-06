import 'package:equatable/equatable.dart';

abstract class AuthEvent extends Equatable {
  const AuthEvent();

  @override
  List<Object> get props => [];
}

class LoginRequested extends AuthEvent {
  final String email;
  final String password;
  final bool rememberMe;

  const LoginRequested(this.email, this.password, {this.rememberMe = false});

  @override
  List<Object> get props => [email, password, rememberMe];
}

class LogoutRequested extends AuthEvent {}

class RegisterRequested extends AuthEvent {
  final String name;
  final String email;
  final String employeeId;
  final String password;
  final String jobPositionId;
  final String branchId;
  final String? shiftId;
  final List<double>? embedding;
  final String? selfiePath;

  const RegisterRequested(
    this.name,
    this.email,
    this.employeeId,
    this.password,
    this.jobPositionId,
    this.branchId,
    this.shiftId, {
    this.embedding,
    this.selfiePath,
  });

  @override
  List<Object> get props => [
        name,
        email,
        employeeId,
        password,
        jobPositionId,
        branchId,
        shiftId ?? '',
        embedding ?? [],
        selfiePath ?? ''
      ];
}

class CheckAuthStatusRequested extends AuthEvent {}

class RegisterFaceRequested extends AuthEvent {
  final List<double> embedding;
  final String selfiePath;

  const RegisterFaceRequested(this.embedding, this.selfiePath);

  @override
  List<Object> get props => [embedding, selfiePath];
}

class SessionExpired extends AuthEvent {
  final String? message;
  const SessionExpired({this.message});

  @override
  List<Object> get props => [message ?? ''];
}

class BiometricLoginRequested extends AuthEvent {}

class CheckBiometricSupportRequested extends AuthEvent {}

class ToggleBiometricRequested extends AuthEvent {
  final bool enabled;
  const ToggleBiometricRequested(this.enabled);
  @override
  List<Object> get props => [enabled];
}

class ChangePasswordRequested extends AuthEvent {
  final String newPassword;

  const ChangePasswordRequested(this.newPassword);

  @override
  List<Object> get props => [newPassword];
}

class ForgotPasswordRequested extends AuthEvent {
  final String email;

  const ForgotPasswordRequested(this.email);

  @override
  List<Object> get props => [email];
}

class LoadRememberMeStatusRequested extends AuthEvent {}

class ToggleRememberMeRequested extends AuthEvent {
  final bool enabled;
  const ToggleRememberMeRequested(this.enabled);
  @override
  List<Object> get props => [enabled];
}

class LoadRememberedCredentialsRequested extends AuthEvent {}
