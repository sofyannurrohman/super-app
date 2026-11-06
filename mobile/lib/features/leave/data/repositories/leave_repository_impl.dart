import 'package:dartz/dartz.dart';
import 'package:dio/dio.dart';
import 'package:hris_app/core/error/failures.dart';
import 'package:hris_app/core/network/api_client.dart';
import 'package:hris_app/features/leave/domain/entities/leave.dart';
import 'package:hris_app/features/leave/domain/entities/leave_balance.dart';
import 'package:hris_app/features/leave/domain/repositories/leave_repository.dart';

import 'dart:typed_data';

class LeaveRepositoryImpl implements LeaveRepository {
  final ApiClient apiClient;

  LeaveRepositoryImpl({required this.apiClient});

  @override
  Future<Either<Failure, void>> submitLeave(
      String leaveTypeId, String startDate, String endDate, String reason,
      {Uint8List? attachmentBytes, String? attachmentName}) async {
    try {
      final Map<String, dynamic> data = {
        'leave_type_id': leaveTypeId,
        'start_date': startDate,
        'end_date': endDate,
        'reason': reason,
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
          await apiClient.client.post('time-off/request', data: body);

      if (response.statusCode == 201) {
        return const Right(null);
      } else {
        return Left(ServerFailure(response.data is Map
            ? (response.data['message'] ?? 'Failed to submit leave')
            : 'Failed to submit leave'));
      }
    } on DioException catch (e) {
      if (e.response != null && e.response?.data is Map) {
        return Left(ServerFailure(
            e.response?.data['message'] ?? 'Failed to submit leave'));
      }
      return Left(ServerFailure(e.message ?? 'Unknown error occurred'));
    } catch (e) {
      return Left(ServerFailure(e.toString()));
    }
  }

  @override
  Future<Either<Failure, List<Leave>>> getMyLeaves(int page, int limit) async {
    try {
      final response =
          await apiClient.client.get('time-off/history', queryParameters: {
        'page': page,
        'limit': limit,
      });

      if (response.statusCode == 200) {
        final List<dynamic> data = response.data['data'] ?? [];
        final List<Leave> leaves = data.map((e) => Leave.fromJson(e)).toList();
        return Right(leaves);
      } else {
        return Left(ServerFailure(response.data is Map
            ? (response.data['message'] ?? 'Failed to load leaves')
            : 'Failed to load leaves'));
      }
    } on DioException catch (e) {
      if (e.response != null && e.response?.data is Map) {
        return Left(ServerFailure(
            e.response?.data['message'] ?? 'Failed to load leaves'));
      }
      return Left(ServerFailure(e.message ?? 'Unknown error occurred'));
    } catch (e) {
      return Left(ServerFailure(e.toString()));
    }
  }

  @override
  Future<Either<Failure, List<Leave>>> getAllLeaves(
      String? status, int page, int limit) async {
    try {
      final Map<String, dynamic> queryParams = {
        'page': page,
        'limit': limit,
      };
      if (status != null && status.isNotEmpty) {
        queryParams['status'] = status;
      }

      final response = await apiClient.client
          .get('time-off/manage', queryParameters: queryParams);

      if (response.statusCode == 200) {
        final List<dynamic> data = response.data['data'] ?? [];
        final List<Leave> leaves = data.map((e) => Leave.fromJson(e)).toList();
        return Right(leaves);
      } else {
        return Left(ServerFailure(response.data is Map
            ? (response.data['message'] ?? 'Failed to load all leaves')
            : 'Failed to load all leaves'));
      }
    } on DioException catch (e) {
      if (e.response != null && e.response?.data is Map) {
        return Left(ServerFailure(
            e.response?.data['message'] ?? 'Failed to load all leaves'));
      }
      return Left(ServerFailure(e.message ?? 'Unknown error occurred'));
    } catch (e) {
      return Left(ServerFailure(e.toString()));
    }
  }

  @override
  Future<Either<Failure, void>> approveLeave(
      String leaveId, String status, {String? rejectReason}) async {
    try {
      final Map<String, dynamic> data = {
        'status': status.toUpperCase(),
      };
      if (rejectReason != null && rejectReason.isNotEmpty) {
        data['reject_reason'] = rejectReason;
      }

      final response =
          await apiClient.client.put('time-off/manage/$leaveId/approve', data: data);

      if (response.statusCode == 200) {
        return const Right(null);
      } else {
        return Left(ServerFailure(response.data is Map
            ? (response.data['message'] ?? 'Failed to update leave status')
            : 'Failed to update leave status'));
      }
    } on DioException catch (e) {
      if (e.response != null && e.response?.data is Map) {
        return Left(ServerFailure(
            e.response?.data['message'] ?? 'Failed to update leave status'));
      }
      return Left(ServerFailure(e.message ?? 'Unknown error occurred'));
    } catch (e) {
      return Left(ServerFailure(e.toString()));
    }
  }

  @override
  Future<Either<Failure, List<LeaveBalance>>> getMyBalances() async {
    try {
      final response = await apiClient.client.get('leave-types');
      if (response.statusCode == 200) {
        final List<dynamic> data = response.data['data'] ?? [];
        return Right(data.map((e) => LeaveBalance.fromJson(e)).toList());
      } else {
        return Left(ServerFailure(response.data is Map
            ? (response.data['message'] ?? 'Failed to load balances')
            : 'Failed to load balances'));
      }
    } on DioException catch (e) {
      if (e.response != null && e.response?.data is Map) {
        return Left(ServerFailure(
            e.response?.data['message'] ?? 'Failed to load balances'));
      }
      return Left(ServerFailure(e.message ?? 'Unknown error occurred'));
    } catch (e) {
      return Left(ServerFailure(e.toString()));
    }
  }

  @override
  Future<Either<Failure, void>> updateLeave(String leaveId, String leaveTypeId,
      String startDate, String endDate, String reason,
      {Uint8List? attachmentBytes, String? attachmentName}) async {
    try {
      final Map<String, dynamic> data = {
        'leave_type_id': leaveTypeId,
        'start_date': startDate,
        'end_date': endDate,
        'reason': reason,
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
          await apiClient.client.put('time-off/$leaveId', data: body);

      if (response.statusCode == 200) {
        return const Right(null);
      } else {
        return Left(ServerFailure(response.data is Map
            ? (response.data['message'] ?? 'Failed to update leave')
            : 'Failed to update leave'));
      }
    } on DioException catch (e) {
      if (e.response != null && e.response?.data is Map) {
        return Left(ServerFailure(
            e.response?.data['message'] ?? 'Failed to update leave'));
      }
      return Left(ServerFailure(e.message ?? 'Unknown error occurred'));
    } catch (e) {
      return Left(ServerFailure(e.toString()));
    }
  }

  @override
  Future<Either<Failure, void>> deleteLeave(String leaveId) async {
    try {
      final response = await apiClient.client.delete('time-off/$leaveId');

      if (response.statusCode == 200) {
        return const Right(null);
      } else {
        return Left(ServerFailure(response.data is Map
            ? (response.data['message'] ?? 'Failed to cancel/delete leave')
            : 'Failed to cancel/delete leave'));
      }
    } on DioException catch (e) {
      if (e.response != null && e.response?.data is Map) {
        return Left(ServerFailure(
            e.response?.data['message'] ?? 'Failed to cancel/delete leave'));
      }
      return Left(ServerFailure(e.message ?? 'Unknown error occurred'));
    } catch (e) {
      return Left(ServerFailure(e.toString()));
    }
  }
}
