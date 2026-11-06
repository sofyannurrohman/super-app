import 'package:equatable/equatable.dart';

abstract class ProfileEvent extends Equatable {
  const ProfileEvent();

  @override
  List<Object?> get props => [];
}

class LoadProfileRequested extends ProfileEvent {}

class UpdateProfileRequested extends ProfileEvent {
  final String firstName;
  final String lastName;
  final String birthDate;
  final String birthPlace;
  final String gender;
  final String maritalStatus;
  final String phoneNumber;
  final String addressKtp;
  final String addressResidential;
  final String emergencyContact;
  final String identityNumber;
  final String npwpNumber;
  final String bankName;
  final String bankAccountNumber;
  final String accountHolderName;

  const UpdateProfileRequested({
    required this.firstName,
    required this.lastName,
    required this.birthDate,
    required this.birthPlace,
    required this.gender,
    required this.maritalStatus,
    required this.phoneNumber,
    required this.addressKtp,
    required this.addressResidential,
    required this.emergencyContact,
    required this.identityNumber,
    required this.npwpNumber,
    required this.bankName,
    required this.bankAccountNumber,
    required this.accountHolderName,
  });

  @override
  List<Object?> get props => [
        firstName,
        lastName,
        birthDate,
        birthPlace,
        gender,
        maritalStatus,
        phoneNumber,
        addressKtp,
        addressResidential,
        emergencyContact,
        identityNumber,
        npwpNumber,
        bankName,
        bankAccountNumber,
        accountHolderName,
      ];
}

class UploadAvatarRequested extends ProfileEvent {
  final String imagePath;

  const UploadAvatarRequested(this.imagePath);

  @override
  List<Object?> get props => [imagePath];
}

