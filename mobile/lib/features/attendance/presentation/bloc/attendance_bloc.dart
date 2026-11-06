import 'package:flutter_bloc/flutter_bloc.dart';
import 'package:hris_app/features/attendance/presentation/bloc/attendance_event.dart';
import 'package:hris_app/features/attendance/presentation/bloc/attendance_state.dart';
import 'package:hris_app/features/attendance/domain/entities/attendance.dart';
import 'package:hris_app/features/attendance/data/repositories/attendance_repository.dart';
import 'package:hris_app/core/services/live_location_tracking_service.dart';
import 'package:hris_app/features/auth/domain/repositories/auth_repository.dart';

class AttendanceBloc extends Bloc<AttendanceEvent, AttendanceState> {
  final AttendanceRepository repository;
  final AuthRepository authRepository;

  AttendanceBloc({required this.repository, required this.authRepository})
      : super(AttendanceInitial()) {
    on<FetchHomeDataRequested>((event, emit) async {
      emit(AttendanceLoading());
      try {
        final profileFuture = authRepository.getProfile();
        final statsFuture = repository.fetchStats();
        final historyFuture = repository.getHistory();
        final queueCountFuture = repository.getQueueCount();

        final profileResult = await profileFuture;
        final statsResult = await statsFuture;
        final historyResult = await historyFuture;
        final queueCount = await queueCountFuture;

        profileResult.fold(
          (failure) => emit(HomeDataFailure(failure.message)),
          (profile) {
            final stats = statsResult.getOrElse(() => AttendanceStats(
                  totalHours: 0,
                  overtimeHours: 0,
                  attendanceDays: 0,
                  leaveDays: 0,
                ));

            final history = historyResult.getOrElse(() => <Attendance>[]);

            final mockStats = {
              'PresentCount': stats.attendanceDays,
              'LeaveCount': stats.leaveDays,
              'AlphaCount': 0,
            };

            emit(HomeDataLoaded(
              profile: profile,
              stats: stats,
              history: history,
              queueCount: queueCount,
              statistics: mockStats,
            ));

            // Restore background tracking service if employee is currently clocked in
            if (history.isNotEmpty && history.first.checkOut == null) {
              LiveLocationTrackingService.start();
            }
          },
        );
      } catch (e) {
        emit(HomeDataFailure(e.toString()));
      }
    });

    on<ClockInRequested>((event, emit) async {
      emit(AttendanceLoading());
      final lat = event.latitude ?? 0.0;
      final lng = event.longitude ?? 0.0;

      if (event.isClockIn) {
        final result = await repository.checkIn(
          lat,
          lng,
          event.imagePath,
          event.faceEmbedding,
          isOutOfTown: event.isOutOfTown,
          destination: event.destination,
        );
        result.fold(
          (failure) => emit(AttendanceFailure(failure.message)),
          (attendance) => emit(AttendanceSuccess(
              message: event.isOutOfTown
                  ? 'Absen Masuk Dinas Luar Kota Berhasil'
                  : 'Absen Masuk Berhasil',
              attendance: attendance)),
        );
        if (result.isRight()) {
          await LiveLocationTrackingService.start(
            initialLat: lat != 0.0 ? lat : null,
            initialLng: lng != 0.0 ? lng : null,
          );
        }
      } else {
        final result = await repository.checkOut(
          lat,
          lng,
          event.imagePath,
          event.faceEmbedding,
          isOutOfTown: event.isOutOfTown,
          destination: event.destination,
        );
        result.fold(
          (failure) => emit(AttendanceFailure(failure.message)),
          (attendance) => emit(AttendanceSuccess(
              message: event.isOutOfTown
                  ? 'Absen Keluar Dinas Luar Kota Berhasil'
                  : 'Absen Keluar Berhasil',
              attendance: attendance)),
        );
        if (result.isRight()) {
          await LiveLocationTrackingService.stop();
        }
      }
    });

    on<RegisterFaceRequested>((event, emit) async {
      emit(FaceRegistrationLoading());
      final result = await authRepository.registerFace(
          event.faceEmbedding ?? [], event.imagePath);
      result.fold(
        (failure) => emit(AttendanceFailure(failure.message)),
        (_) => emit(FaceRegistrationSuccess()),
      );
    });

    on<FetchHistoryRequested>((event, emit) async {
      final currentProfile = (state is HomeDataLoaded)
          ? (state as HomeDataLoaded).profile
          : (state is AttendanceHistoryLoaded
              ? (state as AttendanceHistoryLoaded).profile
              : null);
      final currentStats = (state is HomeDataLoaded)
          ? (state as HomeDataLoaded).stats
          : (state is AttendanceHistoryLoaded
              ? (state as AttendanceHistoryLoaded).stats
              : null);
      final currentStatistics = (state is HomeDataLoaded)
          ? (state as HomeDataLoaded).statistics
          : (state is AttendanceHistoryLoaded
              ? (state as AttendanceHistoryLoaded).statistics
              : null);

      emit(AttendanceLoading());
      final result = await repository.getHistory(
          startDate: event.startDate, endDate: event.endDate);
      result.fold(
        (failure) => emit(AttendanceHistoryFailure(failure.message)),
        (history) => emit(AttendanceHistoryLoaded(
          history,
          profile: currentProfile,
          stats: currentStats,
          statistics: currentStatistics,
        )),
      );
    });

    on<SaveOfflineAttendanceEvent>((event, emit) async {
      emit(AttendanceLoading());
      try {
        await repository.saveOffline(
          event.type,
          event.photoPath,
          event.latitude,
          event.longitude,
          [], // Still empty for now as offline queue needs embedding support refactor
        );
        final mockAttendance = Attendance(
          id: 'offline_${DateTime.now().millisecondsSinceEpoch}',
          userId: 'local_user',
          checkIn: DateTime.now(),
          status: 'offline',
          selfiePath: event.photoPath,
        );
        emit(AttendanceSuccess(
            message: 'Absensi disimpan secara offline.',
            attendance: mockAttendance));
        add(GetQueueCountRequested());
      } catch (e) {
        emit(AttendanceFailure(e.toString()));
      }
    });

    on<SyncAttendanceQueueRequested>((event, emit) async {
      emit(AttendanceLoading());
      try {
        final result = await repository.syncQueue();
        emit(SyncSuccess(result['success'] ?? 0, result['fail'] ?? 0));
        add(FetchHomeDataRequested());
      } catch (e) {
        emit(AttendanceFailure(e.toString()));
      }
    });

    on<GetQueueCountRequested>((event, emit) async {
      final count = await repository.getQueueCount();
      emit(QueueCountLoaded(count));
    });

    on<SubmitLateReasonRequested>((event, emit) async {
      emit(LateReasonSubmitting());
      final result = await repository.updateAttendanceNotes(
          event.attendanceId, event.reason);
      result.fold(
        (failure) => emit(LateReasonFailure(failure.message)),
        (_) {
          emit(LateReasonSuccess(event.attendanceId, event.reason));
          add(FetchHomeDataRequested());
        },
      );
    });
  }
}
