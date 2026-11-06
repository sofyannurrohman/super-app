import 'package:equatable/equatable.dart';
import 'package:flutter_bloc/flutter_bloc.dart';
import 'package:hris_app/features/announcement/data/models/announcement_model.dart';
import 'package:hris_app/features/announcement/data/repositories/announcement_repository.dart';

// Events
abstract class AnnouncementEvent extends Equatable {
  const AnnouncementEvent();
  @override
  List<Object?> get props => [];
}

class FetchAnnouncementsRequested extends AnnouncementEvent {}

class GetPendingPopupsEvent extends AnnouncementEvent {}

class AcknowledgePopupEvent extends AnnouncementEvent {
  final String announcementId;
  const AcknowledgePopupEvent(this.announcementId);
  @override
  List<Object?> get props => [announcementId];
}

class ToggleLikeAnnouncementEvent extends AnnouncementEvent {
  final String announcementId;
  const ToggleLikeAnnouncementEvent(this.announcementId);
  @override
  List<Object?> get props => [announcementId];
}

// Unified State
class AnnouncementState extends Equatable {
  final List<Announcement> announcements;
  final List<Announcement> pendingPopups;
  final bool isLoading;
  final String? errorMessage;
  final String? acknowledgedPopupId;

  const AnnouncementState({
    this.announcements = const [],
    this.pendingPopups = const [],
    this.isLoading = false,
    this.errorMessage,
    this.acknowledgedPopupId,
  });

  AnnouncementState copyWith({
    List<Announcement>? announcements,
    List<Announcement>? pendingPopups,
    bool? isLoading,
    String? errorMessage,
    String? acknowledgedPopupId,
  }) {
    return AnnouncementState(
      announcements: announcements ?? this.announcements,
      pendingPopups: pendingPopups ?? this.pendingPopups,
      isLoading: isLoading ?? this.isLoading,
      errorMessage: errorMessage,
      acknowledgedPopupId: acknowledgedPopupId,
    );
  }

  @override
  List<Object?> get props => [
        announcements,
        pendingPopups,
        isLoading,
        errorMessage,
        acknowledgedPopupId,
      ];
}

// BLoC
class AnnouncementBloc extends Bloc<AnnouncementEvent, AnnouncementState> {
  final AnnouncementRepository repository;

  AnnouncementBloc({required this.repository})
      : super(const AnnouncementState()) {
    on<FetchAnnouncementsRequested>((event, emit) async {
      emit(state.copyWith(isLoading: true, errorMessage: null));
      final result = await repository.getAnnouncements();
      result.fold(
        (failure) => emit(state.copyWith(
          isLoading: false,
          errorMessage: failure.message,
        )),
        (lists) => emit(state.copyWith(
          isLoading: false,
          announcements: lists,
        )),
      );
    });

    on<ToggleLikeAnnouncementEvent>((event, emit) async {
      // Optimistic update
      final previousAnnouncements = List<Announcement>.from(state.announcements);
      final updatedAnnouncements = state.announcements.map((ann) {
        if (ann.id == event.announcementId) {
          final willBeLiked = !ann.isLikedByMe;
          final updatedLikesCount = willBeLiked
              ? ann.likesCount + 1
              : (ann.likesCount > 0 ? ann.likesCount - 1 : 0);
          return ann.copyWith(
            isLikedByMe: willBeLiked,
            likesCount: updatedLikesCount,
          );
        }
        return ann;
      }).toList();

      emit(state.copyWith(announcements: updatedAnnouncements));

      final result = await repository.toggleLike(event.announcementId);
      result.fold(
        (failure) {
          // Revert on error
          emit(state.copyWith(
            announcements: previousAnnouncements,
            errorMessage: failure.message,
          ));
        },
        (data) {
          // Sync with exact server response if available
          final isLiked = data['is_liked'] as bool?;
          final count = data['likes_count'] as int?;
          if (isLiked != null && count != null) {
            final syncedList = state.announcements.map((ann) {
              if (ann.id == event.announcementId) {
                return ann.copyWith(
                  isLikedByMe: isLiked,
                  likesCount: count,
                );
              }
              return ann;
            }).toList();
            emit(state.copyWith(announcements: syncedList));
          }
        },
      );
    });

    on<GetPendingPopupsEvent>((event, emit) async {
      final result = await repository.getPendingPopups();
      result.fold(
        (failure) => emit(state.copyWith(errorMessage: failure.message)),
        (popups) => emit(state.copyWith(
          pendingPopups: popups,
        )),
      );
    });

    on<AcknowledgePopupEvent>((event, emit) async {
      final result =
          await repository.acknowledgeAnnouncement(event.announcementId);
      result.fold(
        (failure) => emit(state.copyWith(errorMessage: failure.message)),
        (_) {
          final updatedPopups = state.pendingPopups
              .where((p) => p.id != event.announcementId)
              .toList();
          emit(state.copyWith(
            pendingPopups: updatedPopups,
            acknowledgedPopupId: event.announcementId,
          ));
        },
      );
    });
  }
}
