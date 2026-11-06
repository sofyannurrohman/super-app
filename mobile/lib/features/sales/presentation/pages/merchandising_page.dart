import 'dart:convert';
import 'package:shared_preferences/shared_preferences.dart';
import 'package:cached_network_image/cached_network_image.dart';
import 'dart:io';
import 'package:dio/dio.dart';
import 'package:drift/drift.dart' hide Column;
import 'package:flutter/foundation.dart';
import 'package:flutter/material.dart';
import 'package:google_fonts/google_fonts.dart';
import 'package:image_picker/image_picker.dart';
import 'package:path_provider/path_provider.dart';
import 'package:uuid/uuid.dart';
import 'package:hris_app/injection.dart' as di;
import 'package:hris_app/core/database/database.dart';
import 'package:hris_app/core/network/api_client.dart';
import 'package:flutter_bloc/flutter_bloc.dart';
import 'package:hris_app/features/auth/presentation/bloc/auth_bloc.dart';
import 'package:hris_app/features/auth/presentation/bloc/auth_state.dart';
import 'package:hris_app/features/profile/presentation/bloc/profile_bloc.dart';
import 'package:hris_app/features/profile/presentation/bloc/profile_state.dart';
import 'package:hris_app/features/sales/data/models/store_model.dart';
import 'package:hris_app/features/sales/data/services/store_api_service.dart';
import 'package:hris_app/features/sync/presentation/bloc/sync_bloc.dart';

class MerchandisingPage extends StatefulWidget {
  final StoreModel? initialStore;
  const MerchandisingPage({super.key, this.initialStore});

  @override
  State<MerchandisingPage> createState() => _MerchandisingPageState();
}

class _MerchandisingPageState extends State<MerchandisingPage> {
  final ApiClient _apiClient = di.sl<ApiClient>();
  final _formKey = GlobalKey<FormState>();

  bool _isLoading = false;
  List<dynamic> _keyAccounts = [];
  List<StoreModel> _stores = [];
  StoreModel? _selectedStore;

  XFile? _displayPhoto;

  @override
  void initState() {
    super.initState();
    if (widget.initialStore != null) {
      _selectedStore = widget.initialStore;
      _stores = [widget.initialStore!];
    }
    _fetchData();
  }

  Future<void> _fetchData() async {
    final prefs = di.sl<SharedPreferences>();
    final cachedKa = prefs.getString('cached_mt_key_accounts');
    if (cachedKa != null) {
      try {
        _keyAccounts = jsonDecode(cachedKa) as List<dynamic>;
      } catch (_) {}
    }

    if (_stores.isEmpty) {
      setState(() => _isLoading = true);
    }
    try {
      try {
        final kaRes = await _apiClient.client.get('modern-trade/key-accounts');
        if (kaRes.statusCode == 200) {
          _keyAccounts = kaRes.data['data'] ?? [];
          await prefs.setString(
            'cached_mt_key_accounts',
            jsonEncode(_keyAccounts),
          );
        }
      } catch (_) {}

      final storeApiService = StoreApiService(
        apiClient: _apiClient,
        prefs: prefs,
      );
      final stores = await storeApiService.getStores();

      if (mounted) {
        setState(() {
          if (stores.isNotEmpty) {
            _stores = stores;
          }
          if (widget.initialStore != null) {
            try {
              _selectedStore = _stores.firstWhere(
                (s) => s.id == widget.initialStore!.id,
              );
            } catch (_) {
              _selectedStore = widget.initialStore;
            }
          } else if (_stores.isNotEmpty && _selectedStore == null) {
            _selectedStore = _stores[0];
          }
        });
      }
    } catch (e) {
      debugPrint('Error fetch data: $e');
    } finally {
      if (mounted) setState(() => _isLoading = false);
    }
  }

  Future<void> _pickImage() async {
    final picker = ImagePicker();
    final picked = await picker.pickImage(
      maxWidth: 1024,
      maxHeight: 1024,
      source: ImageSource.camera,
      imageQuality: 70,
    );
    if (picked != null) {
      setState(() {
        _displayPhoto = picked;
      });
    }
  }

  String _getMdName() {
    Map<String, dynamic>? profile;
    final profileState = context.read<ProfileBloc>().state;
    if (profileState is ProfileLoaded) {
      profile = profileState.profile;
    } else {
      final authState = context.read<AuthBloc>().state;
      if (authState is Authenticated) {
        profile = authState.userProfile;
      }
    }

    if (profile != null) {
      final firstName = profile['first_name'] ?? profile['FirstName'] ?? '';
      final lastName = profile['last_name'] ?? profile['LastName'] ?? '';
      String displayName = '$firstName $lastName'.trim();
      if (displayName.isEmpty) {
        displayName =
            profile['name'] ??
            profile['Name'] ??
            profile['full_name'] ??
            'MD Karsa HR';
      }
      return displayName;
    }
    return 'MD Karsa HR';
  }

  Future<void> _submitForm() async {
    if (_isLoading) return;
    if (!_formKey.currentState!.validate()) return;
    if (_selectedStore == null) {
      ScaffoldMessenger.of(context).showSnackBar(
        const SnackBar(content: Text('Pilih Mitra terlebih dahulu')),
      );
      return;
    }
    if (_displayPhoto == null) {
      ScaffoldMessenger.of(context).showSnackBar(
        const SnackBar(content: Text('Wajib menyertakan foto produk display')),
      );
      return;
    }

    setState(() => _isLoading = true);

    final mdName = _getMdName();
    String? keyAccountId;
    try {
      final storeLower = _selectedStore!.name.toLowerCase();
      final ownerLower = _selectedStore!.ownerName.toLowerCase();
      final ka = _keyAccounts.firstWhere((element) {
        final kaName = (element['name'] ?? element['NAME'] ?? '')
            .toString()
            .toLowerCase();
        final kaCompName =
            (element['company_name'] ?? element['CompanyName'] ?? '')
                .toString()
                .toLowerCase();
        return (kaName.isNotEmpty &&
                (storeLower.contains(kaName) ||
                    ownerLower.contains(kaName) ||
                    kaName.contains(storeLower))) ||
            (kaCompName.isNotEmpty &&
                (storeLower.contains(kaCompName) ||
                    ownerLower.contains(kaCompName) ||
                    kaCompName.contains(storeLower)));
      });
      keyAccountId = ka['id'];
    } catch (e) {
      if (_keyAccounts.isNotEmpty) {
        keyAccountId = _keyAccounts[0]['id'];
      }
    }

    String persistentPhotoPath = _displayPhoto!.path;
    if (!kIsWeb) {
      try {
        final dir = await getApplicationDocumentsDirectory();
        final dest =
            '${dir.path}/merchandising_${DateTime.now().millisecondsSinceEpoch}.jpg';
        await File(_displayPhoto!.path).copy(dest);
        persistentPhotoPath = dest;
      } catch (_) {}
    }

    final payload = <String, dynamic>{
      if (keyAccountId != null && keyAccountId.isNotEmpty)
        'key_account_id': keyAccountId,
      'store_name': _selectedStore!.name,
      'supermarket_name': _selectedStore!.name,
      'merchandiser_name': mdName,
      'md_name': mdName,
      'visit_time': DateTime.now().toUtc().toIso8601String(),
      'planogram_compliant': true,
      'shelf_share_percent': 0.0,
      'out_of_stock_skus': '',
      'photo_url': '',
      'status': 'PENDING',
    };

    try {
      String? photoUrl;
      MultipartFile filePayload;
      if (kIsWeb) {
        final bytes = await _displayPhoto!.readAsBytes();
        filePayload = MultipartFile.fromBytes(
          bytes,
          filename: 'display_photo.jpg',
        );
      } else {
        filePayload = await MultipartFile.fromFile(
          persistentPhotoPath,
          filename: 'display_photo.jpg',
        );
      }
      final formData = FormData.fromMap({'file': filePayload});

      final uploadRes = await _apiClient.client.post(
        'modern-trade/merchandising/upload-photo',
        data: formData,
        options: Options(
          sendTimeout: const Duration(seconds: 6),
          receiveTimeout: const Duration(seconds: 6),
        ),
      );
      if (uploadRes.statusCode == 200 || uploadRes.statusCode == 201) {
        photoUrl = (uploadRes.data['url'] ?? '').toString().trim();
        payload['photo_url'] = photoUrl;
      }

      final submitRes = await _apiClient.client.post(
        'modern-trade/merchandising',
        data: payload,
      );
      if (submitRes.statusCode == 200 || submitRes.statusCode == 201) {
        if (!mounted) return;
        ScaffoldMessenger.of(context).showSnackBar(
          const SnackBar(
            content: Text('Laporan Merchandising berhasil dikirim'),
            backgroundColor: Colors.green,
          ),
        );
        Navigator.pop(context, true);
        return;
      }
    } catch (e) {
      debugPrint('Merchandising online submit failed, saving offline: $e');
      try {
        final db = di.sl<AppDatabase>();
        await db
            .into(db.offlineTasks)
            .insert(
              OfflineTasksCompanion.insert(
                id: const Uuid().v4(),
                type: 'MERCHANDISING',
                payload: jsonEncode(payload),
                filePath: Value(persistentPhotoPath),
                createdAt: DateTime.now(),
                status: const Value('pending'),
              ),
            );
        try {
          di.sl<SyncBloc>().add(const SyncDataRequested(force: true));
        } catch (_) {}
        if (!mounted) return;
        ScaffoldMessenger.of(context).showSnackBar(
          const SnackBar(
            content: Text(
              'Laporan Merchandising disimpan offline & akan disinkronkan otomatis',
            ),
            backgroundColor: Colors.green,
          ),
        );
        Navigator.pop(context, true);
        return;
      } catch (dbErr) {
        if (!mounted) return;
        ScaffoldMessenger.of(context).showSnackBar(
          SnackBar(content: Text('Gagal menyimpan laporan: $dbErr')),
        );
      }
    } finally {
      if (mounted) setState(() => _isLoading = false);
    }
  }

  @override
  Widget build(BuildContext context) {
    return Scaffold(
      backgroundColor: const Color(0xFFF8FAFC),
      appBar: AppBar(
        title: Text(
          'Laporan Merchandising',
          style: GoogleFonts.outfit(
            fontWeight: FontWeight.w800,
            color: Colors.white,
            fontSize: 16,
          ),
        ),
        backgroundColor: const Color(0xFFD97706),
        centerTitle: true,
        elevation: 0,
        leading: IconButton(
          icon: const Icon(Icons.arrow_back_rounded, color: Colors.white),
          onPressed: () => Navigator.pop(context),
        ),
      ),
      body: _isLoading && _stores.isEmpty
          ? const Center(child: CircularProgressIndicator())
          : SingleChildScrollView(
              padding: const EdgeInsets.all(16),
              child: Form(
                key: _formKey,
                child: Column(
                  crossAxisAlignment: CrossAxisAlignment.start,
                  children: [
                    _buildSectionTitle('Pilih Mitra / Toko Outlet'),
                    Container(
                      padding: const EdgeInsets.symmetric(horizontal: 12),
                      decoration: BoxDecoration(
                        color: Colors.white,
                        borderRadius: BorderRadius.circular(12),
                        border: Border.all(color: Colors.grey.shade300),
                      ),
                      child: DropdownButtonHideUnderline(
                        child: DropdownButton<StoreModel>(
                          isExpanded: true,
                          value: _selectedStore,
                          hint: const Text('Pilih Mitra'),
                          items: _stores.map((store) {
                            return DropdownMenuItem<StoreModel>(
                              value: store,
                              child: Text(
                                '${store.name} (${store.ownerName})',
                                style: GoogleFonts.outfit(fontSize: 14),
                              ),
                            );
                          }).toList(),
                          onChanged: (val) {
                            setState(() => _selectedStore = val);
                          },
                        ),
                      ),
                    ),
                    const SizedBox(height: 16),
                    _buildSectionTitle('Foto Produk Display (Wajib)'),
                    GestureDetector(
                      onTap: _pickImage,
                      child: Container(
                        width: double.infinity,
                        height: 200,
                        decoration: BoxDecoration(
                          color: Colors.white,
                          borderRadius: BorderRadius.circular(16),
                          border: Border.all(
                            color: Colors.grey.shade300,
                            style: BorderStyle.solid,
                          ),
                        ),
                        child: _displayPhoto == null
                            ? Column(
                                mainAxisAlignment: MainAxisAlignment.center,
                                children: [
                                  Icon(
                                    Icons.camera_alt_rounded,
                                    size: 48,
                                    color: Colors.grey.shade400,
                                  ),
                                  const SizedBox(height: 8),
                                  Text(
                                    'Tap untuk ambil foto display',
                                    style: GoogleFonts.outfit(
                                      color: Colors.blueGrey,
                                      fontSize: 14,
                                    ),
                                  ),
                                ],
                              )
                            : ClipRRect(
                                borderRadius: BorderRadius.circular(16),
                                child: kIsWeb
                                    ? CachedNetworkImage(
                                        imageUrl: _displayPhoto!.path,
                                        fit: BoxFit.cover,
                                        placeholder: (context, url) =>
                                            const Center(
                                              child:
                                                  CircularProgressIndicator(),
                                            ),
                                        errorWidget: (context, url, error) =>
                                            const Icon(Icons.error),
                                        memCacheWidth: 800,
                                      )
                                    : Image.file(
                                        File(_displayPhoto!.path),
                                        fit: BoxFit.cover,
                                      ),
                              ),
                      ),
                    ),
                    const SizedBox(height: 32),
                    SizedBox(
                      width: double.infinity,
                      height: 54,
                      child: ElevatedButton(
                        onPressed: _isLoading ? null : _submitForm,
                        style: ElevatedButton.styleFrom(
                          backgroundColor: const Color(0xFFD97706),
                          shape: RoundedRectangleBorder(
                            borderRadius: BorderRadius.circular(16),
                          ),
                          elevation: 0,
                        ),
                        child: _isLoading
                            ? const CircularProgressIndicator(
                                color: Colors.white,
                              )
                            : Text(
                                'CHECKOUT LAPORAN',
                                style: GoogleFonts.outfit(
                                  fontWeight: FontWeight.w900,
                                  color: Colors.white,
                                  fontSize: 14,
                                  letterSpacing: 1,
                                ),
                              ),
                      ),
                    ),
                    const SizedBox(height: 32),
                  ],
                ),
              ),
            ),
    );
  }

  Widget _buildSectionTitle(String title) {
    return Padding(
      padding: const EdgeInsets.only(bottom: 8),
      child: Text(
        title,
        style: GoogleFonts.outfit(
          fontSize: 14,
          fontWeight: FontWeight.w800,
          color: const Color(0xFF1E293B),
        ),
      ),
    );
  }

  Widget _buildTextField({
    required TextEditingController controller,
    required String label,
    String? hint,
    TextInputType? keyboardType,
    String? Function(String?)? validator,
  }) {
    return Column(
      crossAxisAlignment: CrossAxisAlignment.start,
      children: [
        Text(
          label,
          style: GoogleFonts.outfit(
            fontSize: 12,
            fontWeight: FontWeight.w700,
            color: Colors.blueGrey,
          ),
        ),
        const SizedBox(height: 4),
        TextFormField(
          controller: controller,
          keyboardType: keyboardType,
          validator: validator,
          style: GoogleFonts.outfit(fontSize: 14),
          decoration: InputDecoration(
            hintText: hint,
            hintStyle: GoogleFonts.outfit(
              fontSize: 14,
              color: Colors.grey.shade400,
            ),
            filled: true,
            fillColor: Colors.white,
            contentPadding: const EdgeInsets.symmetric(
              horizontal: 16,
              vertical: 12,
            ),
            border: OutlineInputBorder(
              borderRadius: BorderRadius.circular(12),
              borderSide: BorderSide(color: Colors.grey.shade300),
            ),
            enabledBorder: OutlineInputBorder(
              borderRadius: BorderRadius.circular(12),
              borderSide: BorderSide(color: Colors.grey.shade300),
            ),
            focusedBorder: OutlineInputBorder(
              borderRadius: BorderRadius.circular(12),
              borderSide: const BorderSide(
                color: Color(0xFFD97706),
                width: 1.5,
              ),
            ),
            errorBorder: OutlineInputBorder(
              borderRadius: BorderRadius.circular(12),
              borderSide: const BorderSide(color: Colors.red, width: 1.5),
            ),
          ),
        ),
      ],
    );
  }
}
