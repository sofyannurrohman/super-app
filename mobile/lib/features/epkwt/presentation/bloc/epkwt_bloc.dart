import 'package:flutter_bloc/flutter_bloc.dart';
import 'package:hris_app/features/epkwt/data/repositories/epkwt_repository.dart';
import 'package:hris_app/features/epkwt/presentation/bloc/epkwt_event.dart';
import 'package:hris_app/features/epkwt/presentation/bloc/epkwt_state.dart';

class EpkwtBloc extends Bloc<EpkwtEvent, EpkwtState> {
  final EpkwtRepository repository;

  EpkwtBloc({required this.repository}) : super(EpkwtInitial()) {
    on<FetchMyEpkwtsRequested>((event, emit) async {
      emit(EpkwtLoading());
      try {
        final epkwts = await repository.fetchMyEpkwts();
        emit(EpkwtListLoaded(epkwts));
      } catch (e) {
        emit(EpkwtFailure(e.toString().replaceAll('Exception: ', '')));
      }
    });

    on<SubmitEpkwtRequested>((event, emit) async {
      emit(EpkwtLoading());
      try {
        await repository.submitEpkwtRequest(
          identityNumber: event.identityNumber,
          firstName: event.firstName,
          lastName: event.lastName,
          phoneNumber: event.phoneNumber,
          gender: event.gender,
          birthDate: event.birthDate,
          birthPlace: event.birthPlace,
          addressKtp: event.addressKtp,
          addressResidential: event.addressResidential,
          maritalStatus: event.maritalStatus,
          signatureData: event.signatureData,
        );
        emit(const EpkwtSuccess('Pengajuan EPKWT berhasil dikirim!'));
      } catch (e) {
        emit(EpkwtFailure(e.toString().replaceAll('Exception: ', '')));
      }
    });

    on<SignEpkwtRequested>((event, emit) async {
      emit(EpkwtLoading());
      try {
        await repository.signEpkwt(
          id: event.id,
          base64Signature: event.base64Signature,
        );
        emit(const EpkwtSuccess('Dokumen EPKWT berhasil ditandatangani!'));
      } catch (e) {
        emit(EpkwtFailure(e.toString().replaceAll('Exception: ', '')));
      }
    });
  }
}
