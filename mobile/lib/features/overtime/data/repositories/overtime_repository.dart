import 'package:dio/dio.dart';
import 'package:image_picker/image_picker.dart';
import 'package:hris_app/core/network/api_client.dart';
import 'package:hris_app/features/overtime/domain/entities/overtime.dart';

class OvertimeRepository {
  final ApiClient apiClient;

  OvertimeRepository({required this.apiClient});

  Future<void> submitOvertimeRequest({
    required DateTime date,
    required DateTime startTime,
    required DateTime endTime,
    required int durationMinutes,
    required String type,
    required String reason,
    XFile? attachment,
  }) async {
    try {
      final reqData = {
        "date": date.toIso8601String().split('T')[0],
        "start_time": startTime.toIso8601String(),
        "end_time": endTime.toIso8601String(),
        "duration_minutes": durationMinutes,
        "type": type,
        "reason": reason,
      };

      dynamic dataToSend = reqData;
      if (attachment != null) {
        final fileName = attachment.name;
        final bytes = await attachment.readAsBytes();
        final multipartFile =
            MultipartFile.fromBytes(bytes, filename: fileName);
        dataToSend = FormData.fromMap({
          ...reqData,
          "attachment": multipartFile,
        });
      }

      await apiClient.client.post(
        'overtimes/',
        data: dataToSend,
      );
    } on DioException catch (e) {
      final errorMsg = (e.response?.data is Map)
          ? (e.response?.data['message'] ?? 'Failed to submit overtime request')
          : 'Failed to submit overtime request';
      throw Exception(errorMsg);
    } catch (e) {
      throw Exception(e.toString());
    }
  }

  Future<void> updateOvertimeRequest({
    required String id,
    required DateTime date,
    required DateTime startTime,
    required DateTime endTime,
    required int durationMinutes,
    required String type,
    required String reason,
    XFile? attachment,
  }) async {
    try {
      final reqData = {
        "date": date.toIso8601String().split('T')[0],
        "start_time": startTime.toIso8601String(),
        "end_time": endTime.toIso8601String(),
        "duration_minutes": durationMinutes,
        "type": type,
        "reason": reason,
      };

      dynamic dataToSend = reqData;
      if (attachment != null) {
        final fileName = attachment.name;
        final bytes = await attachment.readAsBytes();
        final multipartFile =
            MultipartFile.fromBytes(bytes, filename: fileName);
        dataToSend = FormData.fromMap({
          ...reqData,
          "attachment": multipartFile,
        });
      }

      await apiClient.client.put(
        'overtimes/$id',
        data: dataToSend,
      );
    } on DioException catch (e) {
      final errorMsg = (e.response?.data is Map)
          ? (e.response?.data['message'] ?? 'Failed to update overtime request')
          : 'Failed to update overtime request';
      throw Exception(errorMsg);
    } catch (e) {
      throw Exception(e.toString());
    }
  }

  Future<void> deleteOvertimeRequest(String id) async {
    try {
      await apiClient.client.delete('overtimes/$id');
    } on DioException catch (e) {
      final errorMsg = (e.response?.data is Map)
          ? (e.response?.data['message'] ?? 'Failed to delete overtime request')
          : 'Failed to delete overtime request';
      throw Exception(errorMsg);
    } catch (e) {
      throw Exception(e.toString());
    }
  }

  Future<List<Overtime>> fetchMyOvertimes() async {
    try {
      final response = await apiClient.client.get('overtimes/me');
      final List<dynamic> data = response.data['data'] ?? [];
      return data.map((json) => Overtime.fromJson(json)).toList();
    } on DioException catch (e) {
      final errorMsg = (e.response?.data is Map)
          ? (e.response?.data['message'] ?? 'Failed to fetch overtime history')
          : 'Failed to fetch overtime history';
      throw Exception(errorMsg);
    } catch (e) {
      throw Exception(e.toString());
    }
  }
}
