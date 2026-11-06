import 'package:equatable/equatable.dart';

abstract class EpkwtEvent extends Equatable {
  const EpkwtEvent();

  @override
  List<Object?> get props => [];
}

class FetchMyEpkwtsRequested extends EpkwtEvent {}

class SubmitEpkwtRequested extends EpkwtEvent {
  final String identityNumber;
  final String firstName;
  final String? lastName;
  final String? phoneNumber;
  final String? gender;
  final DateTime? birthDate;
  final String? birthPlace;
  final String? addressKtp;
  final String? addressResidential;
  final String? maritalStatus;
  final String? signatureData;

  const SubmitEpkwtRequested({
    required this.identityNumber,
    required this.firstName,
    this.lastName,
    this.phoneNumber,
    this.gender,
    this.birthDate,
    this.birthPlace,
    this.addressKtp,
    this.addressResidential,
    this.maritalStatus,
    this.signatureData,
  });

  @override
  List<Object?> get props => [
        identityNumber,
        firstName,
        lastName,
        phoneNumber,
        gender,
        birthDate,
        birthPlace,
        addressKtp,
        addressResidential,
        maritalStatus,
      ];
}

class SignEpkwtRequested extends EpkwtEvent {
  final String id;
  final String base64Signature;

  const SignEpkwtRequested({
    required this.id,
    required this.base64Signature,
  });

  @override
  List<Object?> get props => [id, base64Signature];
}
