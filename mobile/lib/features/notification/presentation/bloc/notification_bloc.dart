import 'package:equatable/equatable.dart';
import 'package:flutter_bloc/flutter_bloc.dart';
import 'package:dartz/dartz.dart';
import 'package:hris_app/core/error/failures.dart';
import 'package:hris_app/features/announcement/data/repositories/announcement_repository.dart';
import 'package:hris_app/features/announcement/data/models/announcement_model.dart';
import 'package:hris_app/features/leave/domain/repositories/leave_repository.dart';
import 'package:hris_app/features/leave/domain/entities/leave.dart';
import 'package:hris_app/features/overtime/data/repositories/overtime_repository.dart';
import 'package:hris_app/features/overtime/domain/entities/overtime.dart';
import 'package:hris_app/features/reimbursement/domain/repositories/reimbursement_repository.dart';
import 'package:hris_app/features/reimbursement/domain/entities/reimbursement.dart';
import 'package:hris_app/features/notification/domain/entities/app_notification.dart';

// Events
abstract class NotificationEvent extends Equatable {
  const NotificationEvent();
  @override
  List<Object?> get props => [];
}

class FetchNotificationsRequested extends NotificationEvent {}

// States
abstract class NotificationState extends Equatable {
  const NotificationState();
  @override
  List<Object?> get props => [];
}

class NotificationInitial extends NotificationState {}

class NotificationLoading extends NotificationState {}

class NotificationLoaded extends NotificationState {
  final List<AppNotification> notifications;
  const NotificationLoaded(this.notifications);
  @override
  List<Object?> get props => [notifications];
}

class NotificationFailure extends NotificationState {
  final String message;
  const NotificationFailure(this.message);
  @override
  List<Object?> get props => [message];
}

// BLoC
class NotificationBloc extends Bloc<NotificationEvent, NotificationState> {
  final AnnouncementRepository announcementRepository;
  final LeaveRepository leaveRepository;
  final OvertimeRepository overtimeRepository;
  final ReimbursementRepository reimbursementRepository;

  NotificationBloc({
    required this.announcementRepository,
    required this.leaveRepository,
    required this.overtimeRepository,
    required this.reimbursementRepository,
  }) : super(NotificationInitial()) {
    on<FetchNotificationsRequested>((event, emit) async {
      emit(NotificationLoading());
      try {
        final results = await Future.wait([
          announcementRepository.getAnnouncements(),
          leaveRepository.getMyLeaves(1, 20),
          overtimeRepository.fetchMyOvertimes(),
          reimbursementRepository.getMyHistory(page: 1, limit: 20),
        ]);

        List<AppNotification> notifications = [];

        // Announcements
        final announcementsResult =
            results[0] as Either<Failure, List<Announcement>>;
        announcementsResult.fold((_) => null, (list) {
          for (var item in list) {
            notifications.add(AppNotification(
              id: item.id,
              title: item.title,
              content: item.content,
              timestamp: item.createdAt,
              type: AppNotificationType.announcement,
              originalData: item,
            ));
          }
        });

        // Leaves
        final leavesResult = results[1] as Either<Failure, List<Leave>>;
        leavesResult.fold((_) => null, (list) {
          for (var item in list) {
            notifications.add(AppNotification(
              id: item.id,
              title: 'Update Pengajuan Cuti',
              content:
                  'Pengajuan cuti Anda pada ${item.startDate.toString().split(' ')[0]} statusnya adalah ${item.status}',
              timestamp: item.createdAt ?? DateTime.now(),
              type: AppNotificationType.leave,
              status: item.status,
              originalData: item,
            ));
          }
        });

        // Overtimes (Returns List<Overtime> directly or throws)
        try {
          final overtimesList = results[2] as List<Overtime>;
          for (var item in overtimesList) {
            notifications.add(AppNotification(
              id: item.id,
              title: 'Update Pengajuan Lembur',
              content:
                  'Pengajuan lembur Anda pada ${item.date.toString().split(' ')[0]} statusnya adalah ${item.status}',
              timestamp: item.date,
              type: AppNotificationType.overtime,
              status: item.status,
              originalData: item,
            ));
          }
        } catch (_) {}

        // Reimbursements
        final reimbursementResult =
            results[3] as Either<Failure, List<Reimbursement>>;
        reimbursementResult.fold((_) => null, (list) {
          for (var item in list) {
            notifications.add(AppNotification(
              id: item.id,
              title: 'Update Reimbursement',
              content:
                  'Pengajuan reimbursement "${item.title}" statusnya adalah ${item.status}',
              timestamp: item.createdAt ?? DateTime.now(),
              type: AppNotificationType.reimbursement,
              status: item.status,
              originalData: item,
            ));
          }
        });

        // Sort by timestamp newest first
        notifications.sort((a, b) => b.timestamp.compareTo(a.timestamp));

        emit(NotificationLoaded(notifications));
      } catch (e) {
        emit(NotificationFailure(e.toString()));
      }
    });
  }
}
