import 'package:dio/dio.dart';
import 'package:hris_app/core/network/api_client.dart';
import 'package:hris_app/features/epkwt/domain/entities/epkwt_request.dart';

class EpkwtRepository {
  final ApiClient apiClient;

  EpkwtRepository({required this.apiClient});

  Future<List<EpkwtRequest>> fetchMyEpkwts() async {
    try {
      final response = await apiClient.client.get('epkwt');
      final List<dynamic> data = response.data['data'] ?? [];
      return data.map((json) => EpkwtRequest.fromJson(json)).toList();
    } on DioException catch (e) {
      final errorMsg = (e.response?.data is Map)
          ? (e.response?.data['message'] ?? 'Gagal mengambil data EPKWT')
          : 'Gagal mengambil data EPKWT';
      throw Exception(errorMsg);
    } catch (e) {
      throw Exception(e.toString());
    }
  }

  Future<void> submitEpkwtRequest({
    required String identityNumber,
    required String firstName,
    required String? lastName,
    required String? phoneNumber,
    required String? gender,
    required DateTime? birthDate,
    required String? birthPlace,
    required String? addressKtp,
    required String? addressResidential,
    required String? maritalStatus,
    String? signatureData,
  }) async {
    try {
      final reqData = {
        "identity_number": identityNumber,
        "first_name": firstName,
        "last_name": lastName,
        "phone_number": phoneNumber,
        "gender": gender,
        "birth_date": birthDate?.toUtc().toIso8601String(),
        "birth_place": birthPlace,
        "address_ktp": addressKtp,
        "address_residential": addressResidential,
        "marital_status": maritalStatus,
        if (signatureData != null) "signature_data": signatureData,
      };

      await apiClient.client.post(
        'epkwt',
        data: reqData,
      );
    } on DioException catch (e) {
      final errorMsg = (e.response?.data is Map)
          ? (e.response?.data['message'] ?? 'Gagal mengirim pengajuan EPKWT')
          : 'Gagal mengirim pengajuan EPKWT';
      throw Exception(errorMsg);
    } catch (e) {
      throw Exception(e.toString());
    }
  }

  Future<void> signEpkwt({
    required String id,
    required String base64Signature,
  }) async {
    try {
      final reqData = {
        "signature_data": base64Signature,
      };

      await apiClient.client.put(
        'epkwt/$id/sign',
        data: reqData,
      );
    } on DioException catch (e) {
      final errorMsg = (e.response?.data is Map)
          ? (e.response?.data['message'] ?? 'Gagal menandatangani EPKWT')
          : 'Gagal menandatangani EPKWT';
      throw Exception(errorMsg);
    } catch (e) {
      throw Exception(e.toString());
    }
  }
}
