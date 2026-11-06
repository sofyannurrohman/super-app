import 'package:dio/dio.dart';

class RegionApiService {
  static const String baseUrl = 'https://www.emsifa.com/api-wilayah-indonesia/api';
  final Dio _dio = Dio();

  Future<List<dynamic>> getProvinces() async {
    try {
      final response = await _dio.get('$baseUrl/provinces.json');
      if (response.statusCode == 200) {
        return response.data;
      }
      return [];
    } catch (e) {
      return [];
    }
  }

  Future<List<dynamic>> getRegencies(String provinceId) async {
    try {
      final response = await _dio.get('$baseUrl/regencies/$provinceId.json');
      if (response.statusCode == 200) {
        return response.data;
      }
      return [];
    } catch (e) {
      return [];
    }
  }

  Future<List<dynamic>> getDistricts(String regencyId) async {
    try {
      final response = await _dio.get('$baseUrl/districts/$regencyId.json');
      if (response.statusCode == 200) {
        return response.data;
      }
      return [];
    } catch (e) {
      return [];
    }
  }

  Future<List<dynamic>> getVillages(String districtId) async {
    try {
      final response = await _dio.get('$baseUrl/villages/$districtId.json');
      if (response.statusCode == 200) {
        return response.data;
      }
      return [];
    } catch (e) {
      return [];
    }
  }
}
