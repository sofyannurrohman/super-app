import 'package:dartz/dartz.dart';
import 'package:hris_app/core/error/failures.dart';
import 'package:hris_app/core/network/api_client.dart';
import 'package:hris_app/features/announcement/data/models/announcement_model.dart';

abstract class AnnouncementRepository {
  Future<Either<Failure, List<Announcement>>> getAnnouncements();
  Future<Either<Failure, Announcement>> getAnnouncementDetail(String announcementId);
  Future<Either<Failure, List<Announcement>>> getPendingPopups();
  Future<Either<Failure, void>> acknowledgeAnnouncement(String announcementId);
  Future<Either<Failure, Map<String, dynamic>>> toggleLike(String announcementId);
  Future<Either<Failure, List<AnnouncementComment>>> getComments(String announcementId, {int page = 1, int limit = 20});
  Future<Either<Failure, AnnouncementComment>> addComment(String announcementId, String content);
  Future<Either<Failure, void>> deleteComment(String announcementId, String commentId);
  Future<Either<Failure, void>> recordActivity(String announcementId, String actionType, {String? channel});
}

class AnnouncementRepositoryImpl implements AnnouncementRepository {
  final ApiClient apiClient;

  AnnouncementRepositoryImpl({required this.apiClient});

  @override
  Future<Either<Failure, List<Announcement>>> getAnnouncements() async {
    try {
      final response = await apiClient.client.get('announcements');
      if (response.statusCode == 200) {
        final List<dynamic> data = response.data['data'] ?? [];
        return Right(data.map((e) => Announcement.fromJson(e)).toList());
      }
      return const Left(ServerFailure('Gagal mengambil pengumuman.'));
    } catch (e) {
      return Left(ServerFailure(e.toString()));
    }
  }

  @override
  Future<Either<Failure, Announcement>> getAnnouncementDetail(String announcementId) async {
    try {
      final response = await apiClient.client.get('announcements/$announcementId');
      if (response.statusCode == 200 && response.data['data'] != null) {
        return Right(Announcement.fromJson(response.data['data']));
      }
      return const Left(ServerFailure('Gagal memuat detail pengumuman.'));
    } catch (e) {
      return Left(ServerFailure(e.toString()));
    }
  }

  @override
  Future<Either<Failure, List<Announcement>>> getPendingPopups() async {
    try {
      final response = await apiClient.client.get('announcements/pending-popup');
      if (response.statusCode == 200) {
        final List<dynamic> data = response.data['data'] ?? [];
        return Right(data.map((e) => Announcement.fromJson(e)).toList());
      }
      return const Left(ServerFailure('Gagal mengambil pending popups.'));
    } catch (e) {
      return Left(ServerFailure(e.toString()));
    }
  }

  @override
  Future<Either<Failure, void>> acknowledgeAnnouncement(String announcementId) async {
    try {
      final response = await apiClient.client.post('announcements/$announcementId/acknowledge');
      if (response.statusCode == 200) {
        return const Right(null);
      }
      return const Left(ServerFailure('Gagal melakukan acknowledge.'));
    } catch (e) {
      return Left(ServerFailure(e.toString()));
    }
  }

  @override
  Future<Either<Failure, Map<String, dynamic>>> toggleLike(String announcementId) async {
    try {
      final response = await apiClient.client.post('announcements/$announcementId/like');
      if (response.statusCode == 200) {
        final Map<String, dynamic> data = response.data['data'] ?? {};
        return Right(data);
      }
      return const Left(ServerFailure('Gagal menyukai postingan.'));
    } catch (e) {
      return Left(ServerFailure(e.toString()));
    }
  }

  @override
  Future<Either<Failure, List<AnnouncementComment>>> getComments(
      String announcementId,
      {int page = 1,
      int limit = 20}) async {
    try {
      final response = await apiClient.client.get(
        'announcements/$announcementId/comments',
        queryParameters: {'page': page, 'limit': limit},
      );
      if (response.statusCode == 200) {
        final List<dynamic> list = response.data['data']?['comments'] ?? [];
        return Right(list.map((e) => AnnouncementComment.fromJson(e)).toList());
      }
      return const Left(ServerFailure('Gagal memuat komentar.'));
    } catch (e) {
      return Left(ServerFailure(e.toString()));
    }
  }

  @override
  Future<Either<Failure, AnnouncementComment>> addComment(
      String announcementId, String content) async {
    try {
      final response = await apiClient.client.post(
        'announcements/$announcementId/comments',
        data: {'content': content},
      );
      if (response.statusCode == 201 || response.statusCode == 200) {
        final data = response.data['data'];
        return Right(AnnouncementComment.fromJson(data));
      }
      return const Left(ServerFailure('Gagal mengirim komentar.'));
    } catch (e) {
      return Left(ServerFailure(e.toString()));
    }
  }

  @override
  Future<Either<Failure, void>> deleteComment(
      String announcementId, String commentId) async {
    try {
      final response = await apiClient.client.delete(
        'announcements/$announcementId/comments/$commentId',
      );
      if (response.statusCode == 200) {
        return const Right(null);
      }
      return const Left(ServerFailure('Gagal menghapus komentar.'));
    } catch (e) {
      return Left(ServerFailure(e.toString()));
    }
  }

  @override
  Future<Either<Failure, void>> recordActivity(
      String announcementId, String actionType, {String? channel}) async {
    try {
      final response = await apiClient.client.post(
        'announcements/$announcementId/activity',
        data: {
          'action_type': actionType,
          if (channel != null) 'channel': channel,
        },
      );
      if (response.statusCode == 200) {
        return const Right(null);
      }
      return const Left(ServerFailure('Gagal merekam aktivitas.'));
    } catch (e) {
      return Left(ServerFailure(e.toString()));
    }
  }
}

