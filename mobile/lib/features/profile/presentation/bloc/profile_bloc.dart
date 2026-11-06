import 'package:flutter_bloc/flutter_bloc.dart';
import 'package:hris_app/features/auth/domain/repositories/auth_repository.dart';
import 'profile_event.dart';
import 'profile_state.dart';

class ProfileBloc extends Bloc<ProfileEvent, ProfileState> {
  final AuthRepository authRepository;

  ProfileBloc({required this.authRepository}) : super(ProfileInitial()) {
    on<LoadProfileRequested>(_onLoadProfileRequested);
    on<UpdateProfileRequested>(_onUpdateProfileRequested);
    on<UploadAvatarRequested>(_onUploadAvatarRequested);
  }

  Future<void> _onLoadProfileRequested(
    LoadProfileRequested event,
    Emitter<ProfileState> emit,
  ) async {
    emit(ProfileLoading());
    final result = await authRepository.getProfile();
    result.fold(
      (failure) => emit(ProfileFailure(failure.message)),
      (profile) => emit(ProfileLoaded(profile)),
    );
  }

  Future<void> _onUpdateProfileRequested(
    UpdateProfileRequested event,
    Emitter<ProfileState> emit,
  ) async {
    emit(ProfileUpdating());
    final result = await authRepository.updateProfile({
      'firstName': event.firstName,
      'lastName': event.lastName,
      'birthDate': event.birthDate,
      'birthPlace': event.birthPlace,
      'gender': event.gender,
      'maritalStatus': event.maritalStatus,
      'phoneNumber': event.phoneNumber,
      'addressKTP': event.addressKtp,
      'addressResidential': event.addressResidential,
      'emergencyContact': event.emergencyContact,
      'identityNumber': event.identityNumber,
      'npwpNumber': event.npwpNumber,
      'bankName': event.bankName,
      'bankAccountNumber': event.bankAccountNumber,
      'accountHolderName': event.accountHolderName,
    });
    result.fold(
      (failure) => emit(ProfileFailure(failure.message)),
      (_) => emit(ProfileUpdateSuccess()),
    );
  }

  Future<void> _onUploadAvatarRequested(
    UploadAvatarRequested event,
    Emitter<ProfileState> emit,
  ) async {
    emit(ProfileAvatarUpdating());
    final result = await authRepository.uploadAvatar(event.imagePath);
    result.fold(
      (failure) => emit(ProfileFailure(failure.message)),
      (avatarUrl) => emit(ProfileAvatarSuccess(avatarUrl)),
    );
  }
}

