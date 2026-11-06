import 'package:flutter_bloc/flutter_bloc.dart';
import 'package:hris_app/features/overtime/data/repositories/overtime_repository.dart';
import 'package:hris_app/features/overtime/presentation/bloc/overtime_event.dart';
import 'package:hris_app/features/overtime/presentation/bloc/overtime_state.dart';

class OvertimeBloc extends Bloc<OvertimeEvent, OvertimeState> {
  final OvertimeRepository repository;

  OvertimeBloc({required this.repository}) : super(OvertimeInitial()) {
    on<SubmitOvertimeRequested>((event, emit) async {
      emit(OvertimeLoading());
      try {
        await repository.submitOvertimeRequest(
          date: event.date,
          startTime: event.startTime,
          endTime: event.endTime,
          durationMinutes: event.durationMinutes,
          type: event.type,
          reason: event.reason,
          attachment: event.attachment,
        );
        emit(const OvertimeSuccess('Overtime submitted successfully!'));
      } catch (e) {
        emit(OvertimeFailure(e.toString().replaceAll('Exception: ', '')));
      }
    });

    on<UpdateOvertimeRequested>((event, emit) async {
      emit(OvertimeLoading());
      try {
        await repository.updateOvertimeRequest(
          id: event.id,
          date: event.date,
          startTime: event.startTime,
          endTime: event.endTime,
          durationMinutes: event.durationMinutes,
          type: event.type,
          reason: event.reason,
          attachment: event.attachment,
        );
        emit(const OvertimeSuccess('Overtime updated successfully!'));
      } catch (e) {
        emit(OvertimeFailure(e.toString().replaceAll('Exception: ', '')));
      }
    });

    on<DeleteOvertimeRequested>((event, emit) async {
      emit(OvertimeLoading());
      try {
        await repository.deleteOvertimeRequest(event.id);
        emit(const OvertimeSuccess('Overtime deleted successfully!'));
      } catch (e) {
        emit(OvertimeFailure(e.toString().replaceAll('Exception: ', '')));
      }
    });

    on<FetchMyOvertimesRequested>((event, emit) async {
      emit(OvertimeLoading());
      try {
        final overtimes = await repository.fetchMyOvertimes();
        emit(MyOvertimesLoaded(overtimes));
      } catch (e) {
        emit(OvertimeFailure(e.toString().replaceAll('Exception: ', '')));
      }
    });
  }
}
