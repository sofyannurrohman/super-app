import 'dart:async';
import 'dart:convert';
import 'package:flutter_bloc/flutter_bloc.dart';
import 'package:shared_preferences/shared_preferences.dart';
import 'package:hris_app/features/auth/domain/usecases/auth_usecases.dart';
import 'package:hris_app/features/auth/domain/usecases/biometric_usecases.dart';
import 'package:hris_app/features/auth/domain/usecases/remember_me_usecases.dart';
import 'package:hris_app/injection.dart' as di;
import 'package:hris_app/core/services/notification_service.dart';
import 'package:hris_app/core/services/heartbeat_service.dart';
import 'auth_event.dart';
import 'auth_state.dart';

class AuthBloc extends Bloc<AuthEvent, AuthState> {
  final LoginUseCase loginUseCase;
  final RegisterUseCase registerUseCase;
  final LogoutUseCase logoutUseCase;
  final CheckAuthStatusUseCase checkAuthStatusUseCase;
  final RegisterFaceUseCase registerFaceUseCase;
  final BiometricLoginUseCase biometricLoginUseCase;
  final GetBiometricStatusUseCase getBiometricStatusUseCase;
  final SetBiometricEnabledUseCase setBiometricEnabledUseCase;
  final GetProfileUseCase getProfileUseCase;
  final ChangePasswordUseCase changePasswordUseCase;
  final ForgotPasswordUseCase forgotPasswordUseCase;
  final GetRememberMeStatusUseCase getRememberMeStatusUseCase;
  final SetRememberMeEnabledUseCase setRememberMeEnabledUseCase;
  final GetRememberedCredentialsUseCase getRememberedCredentialsUseCase;
  final ClearRememberedCredentialsUseCase clearRememberedCredentialsUseCase;

  AuthBloc({
    required this.loginUseCase,
    required this.registerUseCase,
    required this.logoutUseCase,
    required this.checkAuthStatusUseCase,
    required this.registerFaceUseCase,
    required this.biometricLoginUseCase,
    required this.getBiometricStatusUseCase,
    required this.setBiometricEnabledUseCase,
    required this.getProfileUseCase,
    required this.changePasswordUseCase,
    required this.forgotPasswordUseCase,
    required this.getRememberMeStatusUseCase,
    required this.setRememberMeEnabledUseCase,
    required this.getRememberedCredentialsUseCase,
    required this.clearRememberedCredentialsUseCase,
  }) : super(AuthInitial()) {
    on<LoginRequested>(_onLoginRequested);
    on<RegisterRequested>(_onRegisterRequested);
    on<LogoutRequested>(_onLogoutRequested);
    on<CheckAuthStatusRequested>(_onCheckAuthStatusRequested);
    on<RegisterFaceRequested>(_onRegisterFaceRequested);
    on<SessionExpired>(_onSessionExpired);
    on<BiometricLoginRequested>(_onBiometricLoginRequested);
    on<CheckBiometricSupportRequested>(_onCheckBiometricSupportRequested);
    on<ToggleBiometricRequested>(_onToggleBiometricRequested);
    on<ChangePasswordRequested>(_onChangePasswordRequested);
    on<ForgotPasswordRequested>(_onForgotPasswordRequested);
    on<LoadRememberMeStatusRequested>(_onLoadRememberMeStatusRequested);
    on<ToggleRememberMeRequested>(_onToggleRememberMeRequested);
    on<LoadRememberedCredentialsRequested>(
        _onLoadRememberedCredentialsRequested);
  }

  Future<Map<String, bool>> _getAuthParams() async {
    final bio = await getBiometricStatusUseCase();
    final rememberMe = await getRememberMeStatusUseCase();
    return {
      'bio_supported': bio['supported'] ?? false,
      'bio_enabled': bio['enabled'] ?? false,
      'remember_me': rememberMe,
    };
  }

  Future<void> _onLoginRequested(
      LoginRequested event, Emitter<AuthState> emit) async {
    emit(AuthLoading());
    final params = await _getAuthParams();
    final result = await loginUseCase(event.email, event.password);

    await result.fold(
      (failure) async {
        emit(AuthError(
          failure.message,
          isBiometricSupported: params['bio_supported']!,
          isBiometricEnabled: params['bio_enabled']!,
          isRememberMeEnabled: params['remember_me']!,
        ));
      },
      (_) async {
        // Handle Remember Me
        await setRememberMeEnabledUseCase(event.rememberMe);
        if (!event.rememberMe) {
          await clearRememberedCredentialsUseCase();
        }

        final profileResult = await getProfileUseCase();
        Map<String, dynamic>? profile;
        profileResult.fold((_) => null, (p) => profile = p);
        emit(Authenticated(
          userProfile: profile,
          isBiometricSupported: params['bio_supported']!,
          isBiometricEnabled: params['bio_enabled']!,
          isRememberMeEnabled: event.rememberMe,
        ));
        // Sync FCM Token with backend
        di.sl<NotificationService>().syncTokenWithBackend();
        
        // Force heartbeat after login
        di.sl<HeartbeatService>().forceHeartbeat();
      },
    );
  }

  Future<void> _onRegisterRequested(
      RegisterRequested event, Emitter<AuthState> emit) async {
    emit(AuthLoading());
    final result = await registerUseCase(
      event.name,
      event.email,
      event.employeeId,
      event.password,
      event.jobPositionId,
      event.branchId,
      event.shiftId,
      embedding: event.embedding,
      selfiePath: event.selfiePath,
    );
    result.fold(
      (failure) => emit(AuthError(failure.message)),
      (_) {
        emit(RegisterSuccess());
        // Restore unauthenticated state
        add(CheckAuthStatusRequested());
      },
    );
  }

  Future<void> _onLogoutRequested(
      LogoutRequested event, Emitter<AuthState> emit) async {
    emit(AuthLoading());
    final params = await _getAuthParams();
    final result = await logoutUseCase();
    result.fold(
      (failure) => emit(AuthError(
        failure.message,
        isAuthenticated: true,
        isBiometricSupported: params['bio_supported']!,
        isBiometricEnabled: params['bio_enabled']!,
        isRememberMeEnabled: params['remember_me']!,
      )),
      (_) => emit(Unauthenticated(
        isBiometricSupported: params['bio_supported']!,
        isBiometricEnabled: params['bio_enabled']!,
        isRememberMeEnabled: params['remember_me']!,
      )),
    );
  }

  Future<void> _onCheckAuthStatusRequested(
      CheckAuthStatusRequested event, Emitter<AuthState> emit) async {
    final startTime = DateTime.now();
    final params = await _getAuthParams();
    final result = await checkAuthStatusUseCase();
    await result.fold(
      (failure) async => emit(Unauthenticated(
        isBiometricSupported: params['bio_supported']!,
        isBiometricEnabled: params['bio_enabled']!,
        isRememberMeEnabled: params['remember_me']!,
      )),
      (isAuthenticated) async {
        final elapsed = DateTime.now().difference(startTime);
        if (elapsed < const Duration(seconds: 3)) {
          await Future.delayed(const Duration(seconds: 3) - elapsed);
        }

        if (isAuthenticated) {
          Map<String, dynamic>? cachedProfile;
          try {
            final prefs = di.sl<SharedPreferences>();
            final raw = prefs.getString('cached_employee_profile');
            if (raw != null && raw.isNotEmpty) {
              cachedProfile = jsonDecode(raw) as Map<String, dynamic>;
            }
          } catch (_) {}

          emit(Authenticated(
            userProfile: cachedProfile, // Emit cached profile immediately so offline cold start works without waiting for network timeout
            isBiometricSupported: params['bio_supported']!,
            isBiometricEnabled: params['bio_enabled']!,
            isRememberMeEnabled: params['remember_me']!,
          ));
          // Sync FCM Token with backend
          di.sl<NotificationService>().syncTokenWithBackend();

          // Fetch profile in background
          final profileResult = await getProfileUseCase();
          profileResult.fold((_) => null, (p) {
            if (!emit.isDone) {
              emit(Authenticated(
                userProfile: p,
                isBiometricSupported: params['bio_supported']!,
                isBiometricEnabled: params['bio_enabled']!,
                isRememberMeEnabled: params['remember_me']!,
              ));
            }
          });
        } else {
          emit(Unauthenticated(
            isBiometricSupported: params['bio_supported']!,
            isBiometricEnabled: params['bio_enabled']!,
            isRememberMeEnabled: params['remember_me']!,
          ));
        }
      },
    );
  }

  Future<void> _onRegisterFaceRequested(
      RegisterFaceRequested event, Emitter<AuthState> emit) async {
    final currentState = state;
    final bool currentlyAuthenticated = currentState is Authenticated;
    emit(AuthLoading());
    try {
      // Menambahkan hard timeout 45 detik untuk fallback jika ISP nge-hang
      final result = await registerFaceUseCase(event.embedding, event.selfiePath)
          .timeout(const Duration(seconds: 45));
      result.fold(
        (failure) {
          emit(AuthError(failure.message,
              isAuthenticated: currentlyAuthenticated));
          add(CheckAuthStatusRequested());
        },
        (_) {
          emit(FaceRegistrationSuccess());
          // Restore auth status (could be authenticated or unauthenticated depending on context)
          add(CheckAuthStatusRequested());
        },
      );
    } on TimeoutException catch (_) {
      emit(AuthError("Koneksi internet bermasalah atau server tidak merespons (Timeout). Silakan coba lagi.",
          isAuthenticated: currentlyAuthenticated));
      add(CheckAuthStatusRequested());
    } catch (e) {
      emit(AuthError("Terjadi kesalahan yang tidak terduga: $e",
          isAuthenticated: currentlyAuthenticated));
      add(CheckAuthStatusRequested());
    }
  }

  Future<void> _onSessionExpired(
      SessionExpired event, Emitter<AuthState> emit) async {
    final params = await _getAuthParams();
    emit(Unauthenticated(
      sessionMessage: event.message,
      isBiometricSupported: params['bio_supported']!,
      isBiometricEnabled: params['bio_enabled']!,
      isRememberMeEnabled: params['remember_me']!,
    ));
  }

  Future<void> _onBiometricLoginRequested(
      BiometricLoginRequested event, Emitter<AuthState> emit) async {
    emit(AuthLoading());
    final params = await _getAuthParams();
    final result = await biometricLoginUseCase();

    await result.fold(
      (failure) async => emit(AuthError(
        failure.message,
        isBiometricSupported: params['bio_supported']!,
        isBiometricEnabled: params['bio_enabled']!,
        isRememberMeEnabled: params['remember_me']!,
      )),
      (_) async {
        final profileResult = await getProfileUseCase();
        Map<String, dynamic>? profile;
        profileResult.fold((_) => null, (p) => profile = p);
        emit(Authenticated(
          userProfile: profile,
          isBiometricSupported: params['bio_supported']!,
          isBiometricEnabled: params['bio_enabled']!,
          isRememberMeEnabled: params['remember_me']!,
        ));
        // Sync FCM Token with backend
        di.sl<NotificationService>().syncTokenWithBackend();
        
        // Force heartbeat after login
        di.sl<HeartbeatService>().forceHeartbeat();
      },
    );
  }

  Future<void> _onCheckBiometricSupportRequested(
      CheckBiometricSupportRequested event, Emitter<AuthState> emit) async {
    final params = await _getAuthParams();
    if (state is Authenticated) {
      emit(Authenticated(
        userProfile: (state as Authenticated).userProfile,
        isBiometricSupported: params['bio_supported']!,
        isBiometricEnabled: params['bio_enabled']!,
        isRememberMeEnabled: params['remember_me']!,
      ));
    } else {
      emit(Unauthenticated(
        isBiometricSupported: params['bio_supported']!,
        isBiometricEnabled: params['bio_enabled']!,
        isRememberMeEnabled: params['remember_me']!,
      ));
    }
  }

  Future<void> _onToggleBiometricRequested(
      ToggleBiometricRequested event, Emitter<AuthState> emit) async {
    final currentState = state;
    Map<String, dynamic>? currentProfile;
    final bool currentlyAuthenticated = currentState is Authenticated;
    if (currentState is Authenticated) {
      currentProfile = currentState.userProfile;
    }

    final result = await setBiometricEnabledUseCase(event.enabled);
    final params = await _getAuthParams();

    await result.fold(
      (failure) async {
        emit(AuthError(
          failure.message,
          isAuthenticated: currentlyAuthenticated,
          isBiometricSupported: params['bio_supported']!,
          isBiometricEnabled: params['bio_enabled']!,
          isRememberMeEnabled: params['remember_me']!,
        ));

        if (currentlyAuthenticated) {
          emit(Authenticated(
            userProfile: currentProfile,
            isBiometricSupported: params['bio_supported']!,
            isBiometricEnabled: params['bio_enabled']!,
            isRememberMeEnabled: params['remember_me']!,
          ));
        }
      },
      (_) async {
        add(CheckBiometricSupportRequested());
      },
    );
  }

  Future<void> _onChangePasswordRequested(
      ChangePasswordRequested event, Emitter<AuthState> emit) async {
    final currentState = state;
    Map<String, dynamic>? currentProfile;
    if (currentState is Authenticated) {
      currentProfile = currentState.userProfile;
    }

    emit(AuthLoading());
    final params = await _getAuthParams();
    final result = await changePasswordUseCase(event.newPassword);

    result.fold(
      (failure) {
        emit(AuthError(
          failure.message,
          isAuthenticated: true,
          isBiometricSupported: params['bio_supported']!,
          isBiometricEnabled: params['bio_enabled']!,
          isRememberMeEnabled: params['remember_me']!,
        ));
        emit(Authenticated(
          userProfile: currentProfile,
          isBiometricSupported: params['bio_supported']!,
          isBiometricEnabled: params['bio_enabled']!,
          isRememberMeEnabled: params['remember_me']!,
        ));
      },
      (_) {
        emit(ChangePasswordSuccess());
        // Restore authenticated state after success so AuthWrapper doesn't get stuck in loading
        emit(Authenticated(
          userProfile: currentProfile,
          isBiometricSupported: params['bio_supported']!,
          isBiometricEnabled: params['bio_enabled']!,
          isRememberMeEnabled: params['remember_me']!,
        ));
      },
    );
  }

  Future<void> _onForgotPasswordRequested(
      ForgotPasswordRequested event, Emitter<AuthState> emit) async {
    emit(AuthLoading());
    final params = await _getAuthParams();
    final result = await forgotPasswordUseCase(event.email);

    result.fold(
      (failure) => emit(AuthError(
        failure.message,
        isBiometricSupported: params['bio_supported']!,
        isBiometricEnabled: params['bio_enabled']!,
        isRememberMeEnabled: params['remember_me']!,
      )),
      (_) {
        emit(ForgotPasswordSuccess());
        // Restore unauthenticated state so AuthWrapper doesn't get stuck
        emit(Unauthenticated(
          isBiometricSupported: params['bio_supported']!,
          isBiometricEnabled: params['bio_enabled']!,
          isRememberMeEnabled: params['remember_me']!,
        ));
      },
    );
  }

  Future<void> _onLoadRememberMeStatusRequested(
      LoadRememberMeStatusRequested event, Emitter<AuthState> emit) async {
    final params = await _getAuthParams();
    if (state is Unauthenticated) {
      emit(Unauthenticated(
        isBiometricSupported: params['bio_supported']!,
        isBiometricEnabled: params['bio_enabled']!,
        isRememberMeEnabled: params['remember_me']!,
        rememberedCredentials: (state as Unauthenticated).rememberedCredentials,
      ));
    }
  }

  Future<void> _onToggleRememberMeRequested(
      ToggleRememberMeRequested event, Emitter<AuthState> emit) async {
    await setRememberMeEnabledUseCase(event.enabled);
    if (!event.enabled) {
      await clearRememberedCredentialsUseCase();
    }
    add(LoadRememberMeStatusRequested());
  }

  Future<void> _onLoadRememberedCredentialsRequested(
      LoadRememberedCredentialsRequested event, Emitter<AuthState> emit) async {
    final params = await _getAuthParams();
    if (params['remember_me']!) {
      final credResult = await getRememberedCredentialsUseCase();
      credResult.fold(
        (_) => null,
        (credentials) {
          if (credentials != null) {
            emit(Unauthenticated(
              rememberedCredentials: credentials,
              isBiometricSupported: params['bio_supported']!,
              isBiometricEnabled: params['bio_enabled']!,
              isRememberMeEnabled: params['remember_me']!,
            ));
          }
        },
      );
    }
  }
}
