import 'package:dartz/dartz.dart';
import 'package:dio/dio.dart';
import 'package:hris_app/core/error/failures.dart';
import 'package:hris_app/core/network/api_client.dart';
import 'package:hris_app/features/reimbursement/domain/entities/reimbursement.dart';
import 'package:hris_app/features/reimbursement/domain/repositories/reimbursement_repository.dart';

import 'dart:typed_data';

class ReimbursementRepositoryImpl implements ReimbursementRepository {
  final ApiClient apiClient;

  ReimbursementRepositoryImpl({required this.apiClient});

  @override
  Future<Either<Failure, void>> submitReimbursement({
    required String title,
    String? description,
    required double amount,
    Uint8List? attachmentBytes,
    String? attachmentName,
  }) async {
    try {
      final Map<String, dynamic> data = {
        'title': title,
        'description': description ?? '',
        'amount': amount,
      };

      Map<String, dynamic> formDataMap = {...data};
      if (attachmentBytes != null && attachmentName != null) {
        formDataMap['attachment'] = MultipartFile.fromBytes(
          attachmentBytes,
          filename: attachmentName,
        );
      }
      final body = FormData.fromMap(formDataMap);

      final response =
          await apiClient.client.post('reimbursements', data: body);

      if (response.statusCode == 201) {
        return const Right(null);
      } else {
        return Left(ServerFailure(response.data is Map
            ? (response.data['message'] ?? 'Failed to submit reimbursement')
            : 'Failed to submit reimbursement'));
      }
    } on DioException catch (e) {
      if (e.response != null && e.response?.data is Map) {
        return Left(ServerFailure(
            e.response?.data['message'] ?? 'Failed to submit reimbursement'));
      }
      return Left(ServerFailure(e.message ?? 'Unknown error occurred'));
    } catch (e) {
      return Left(ServerFailure(e.toString()));
    }
  }

  @override
  Future<Either<Failure, List<Reimbursement>>> getMyHistory(
      {int page = 1, int limit = 10}) async {
    try {
      final response =
          await apiClient.client.get('reimbursements/my', queryParameters: {
        'page': page,
        'limit': limit,
      });

      if (response.statusCode == 200) {
        final List<dynamic> data = response.data['data'] ?? [];
        return Right(data.map((e) => Reimbursement.fromJson(e)).toList());
      } else {
        return Left(ServerFailure(response.data is Map
            ? (response.data['message'] ?? 'Failed to load history')
            : 'Failed to load history'));
      }
    } on DioException catch (e) {
      if (e.response != null && e.response?.data is Map) {
        return Left(ServerFailure(
            e.response?.data['message'] ?? 'Failed to load history'));
      }
      return Left(ServerFailure(e.message ?? 'Unknown error occurred'));
    } catch (e) {
      return Left(ServerFailure(e.toString()));
    }
  }

  @override
  Future<Either<Failure, List<Reimbursement>>> getAllPending(
      {int page = 1, int limit = 20}) async {
    try {
      final response =
          await apiClient.client.get('reimbursements/manage', queryParameters: {
        'status': 'PENDING',
        'page': page,
        'limit': limit,
      });

      if (response.statusCode == 200) {
        final List<dynamic> data = response.data['data'] ?? [];
        return Right(data.map((e) => Reimbursement.fromJson(e)).toList());
      } else {
        return const Left(ServerFailure('Gagal mengambil data reimbursement.'));
      }
    } catch (e) {
      return Left(ServerFailure(e.toString()));
    }
  }

  @override
  Future<Either<Failure, void>> approveReimbursement(
      String id, String status) async {
    try {
      final response = await apiClient.client
          .put('reimbursements/manage/$id/approve', data: {
        'status': status,
      });

      if (response.statusCode == 200) {
        return const Right(null);
      } else {
        return const Left(
            ServerFailure('Gagal memperbarui status reimbursement.'));
      }
    } catch (e) {
      return Left(ServerFailure(e.toString()));
    }
  }

  @override
  Future<Either<Failure, void>> updateReimbursement({
    required String id,
    required String title,
    String? description,
    required double amount,
    Uint8List? attachmentBytes,
    String? attachmentName,
  }) async {
    try {
      final Map<String, dynamic> data = {
        'title': title,
        'description': description ?? '',
        'amount': amount,
      };

      Map<String, dynamic> formDataMap = {...data};
      if (attachmentBytes != null && attachmentName != null) {
        formDataMap['attachment'] = MultipartFile.fromBytes(
          attachmentBytes,
          filename: attachmentName,
        );
      }
      final body = FormData.fromMap(formDataMap);

      final response =
          await apiClient.client.put('reimbursements/$id', data: body);

      if (response.statusCode == 200) {
        return const Right(null);
      } else {
        return Left(ServerFailure(response.data is Map
            ? (response.data['message'] ?? 'Failed to update reimbursement')
            : 'Failed to update reimbursement'));
      }
    } on DioException catch (e) {
      if (e.response != null && e.response?.data is Map) {
        return Left(ServerFailure(
            e.response?.data['message'] ?? 'Failed to update reimbursement'));
      }
      return Left(ServerFailure(e.message ?? 'Unknown error occurred'));
    } catch (e) {
      return Left(ServerFailure(e.toString()));
    }
  }

  @override
  Future<Either<Failure, void>> deleteReimbursement(String id) async {
    try {
      final response = await apiClient.client.delete('reimbursements/$id');

      if (response.statusCode == 200) {
        return const Right(null);
      } else {
        return Left(ServerFailure(response.data is Map
            ? (response.data['message'] ?? 'Failed to delete reimbursement')
            : 'Failed to delete reimbursement'));
      }
    } on DioException catch (e) {
      if (e.response != null && e.response?.data is Map) {
        return Left(ServerFailure(
            e.response?.data['message'] ?? 'Failed to delete reimbursement'));
      }
      return Left(ServerFailure(e.message ?? 'Unknown error occurred'));
    } catch (e) {
      return Left(ServerFailure(e.toString()));
    }
  }
}
