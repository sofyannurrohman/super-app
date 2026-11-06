import 'package:equatable/equatable.dart';

abstract class AuthState extends Equatable {
  final bool isBiometricSupported;
  final bool isBiometricEnabled;
  final bool isRememberMeEnabled;

  const AuthState({
    this.isBiometricSupported = false,
    this.isBiometricEnabled = false,
    this.isRememberMeEnabled = false,
  });

  @override
  List<Object?> get props =>
      [isBiometricSupported, isBiometricEnabled, isRememberMeEnabled];
}

class AuthInitial extends AuthState {}

class AuthLoading extends AuthState {}

class Authenticated extends AuthState {
  final Map<String, dynamic>? userProfile;

  const Authenticated({
    this.userProfile,
    super.isBiometricSupported = false,
    super.isBiometricEnabled = false,
    super.isRememberMeEnabled = false,
  });

  @override
  List<Object?> get props => [...super.props, userProfile ?? {}];
}

class Unauthenticated extends AuthState {
  final Map<String, String>? rememberedCredentials;
  final String? sessionMessage;

  const Unauthenticated({
    this.rememberedCredentials,
    this.sessionMessage,
    super.isBiometricSupported = false,
    super.isBiometricEnabled = false,
    super.isRememberMeEnabled = false,
  });

  @override
  List<Object?> get props =>
      [...super.props, rememberedCredentials, sessionMessage];
}

class AuthError extends AuthState {
  final String message;
  final bool isAuthenticated;

  const AuthError(
    this.message, {
    this.isAuthenticated = false,
    super.isBiometricSupported = false,
    super.isBiometricEnabled = false,
    super.isRememberMeEnabled = false,
  });

  @override
  List<Object?> get props => [...super.props, message, isAuthenticated];
}

class FaceRegistrationSuccess extends AuthState {}

class RegisterSuccess extends AuthState {}

class ChangePasswordSuccess extends AuthState {}

class ForgotPasswordSuccess extends AuthState {}
