import 'dart:convert';
import 'dart:io';
import 'package:flutter/foundation.dart' show kIsWeb;
import 'package:shared_preferences/shared_preferences.dart';
import 'package:flutter/material.dart';
import 'package:google_fonts/google_fonts.dart';
import 'package:hris_app/features/sales/data/models/store_model.dart';
import 'package:hris_app/features/sales/data/services/store_api_service.dart';
import 'package:hris_app/features/sales/presentation/pages/add_store_page.dart';
import 'package:hris_app/features/sales/presentation/pages/store_order_history_page.dart';
import 'package:hris_app/injection.dart' as di;
import 'package:hris_app/core/network/api_client.dart';
import 'package:hris_app/core/utils/constants.dart';
import 'package:flutter_bloc/flutter_bloc.dart';
import 'package:hris_app/features/profile/presentation/bloc/profile_bloc.dart';
import 'package:hris_app/features/profile/presentation/bloc/profile_state.dart';
import 'package:cached_network_image/cached_network_image.dart';
import 'package:hris_app/core/theme/app_colors.dart';

class StoreListPage extends StatefulWidget {
  const StoreListPage({super.key});

  @override
  State<StoreListPage> createState() => _StoreListPageState();
}

class _StoreListPageState extends State<StoreListPage> {
  final _apiService = StoreApiService(apiClient: di.sl<ApiClient>(), prefs: di.sl<SharedPreferences>());
  bool _isLoading = true;
  List<StoreModel> _stores = [];
  String _searchQuery = '';
  final _searchController = TextEditingController();

  @override
  void initState() {
    super.initState();
    _fetchStores();
  }

  @override
  void dispose() {
    _searchController.dispose();
    super.dispose();
  }

  Future<void> _fetchStores({bool forceRefresh = false}) async {
    setState(() => _isLoading = true);
    try {
      String? employeeId;
      final profileState = context.read<ProfileBloc>().state;
      if (profileState is ProfileLoaded) {
        employeeId = profileState.profile['id']?.toString();
      }
      
      final stores = await _apiService.getStores(
        assignedEmployeeId: employeeId,
        forceRefresh: forceRefresh,
      );
      if (mounted) {
        setState(() {
          _stores = stores;
        });
      }
    } catch (e) {
      debugPrint('Error fetching stores: $e');
    } finally {
      if (mounted) setState(() => _isLoading = false);
    }
  }

  Future<void> _deleteStore(StoreModel store) async {
    final confirm = await showDialog<bool>(
      context: context,
      builder: (ctx) => AlertDialog(
        title: Text('Hapus Toko',
            style: GoogleFonts.outfit(fontWeight: FontWeight.bold)),
        content: Text('Apakah Anda yakin ingin menghapus toko ${store.name}?',
            style: GoogleFonts.outfit()),
        actions: [
          TextButton(
            onPressed: () => Navigator.pop(ctx, false),
            child: Text('Batal',
                style: GoogleFonts.outfit(color: Colors.blueGrey)),
          ),
          ElevatedButton(
            onPressed: () => Navigator.pop(ctx, true),
            style: ElevatedButton.styleFrom(backgroundColor: Colors.redAccent),
            child:
                Text('Hapus', style: GoogleFonts.outfit(color: Colors.white)),
          ),
        ],
      ),
    );

    if (confirm == true) {
      setState(() => _isLoading = true);
      try {
        await _apiService.deleteStore(store.id);
        if (mounted) {
          ScaffoldMessenger.of(context).showSnackBar(
            SnackBar(
              content: Text('Toko berhasil dihapus',
                  style: GoogleFonts.outfit(color: Colors.white)),
              backgroundColor: Colors.green,
            ),
          );
        }
        _fetchStores();
      } catch (e) {
        final errorMessage = e.toString().toLowerCase();
        if (errorMessage.contains('tidak bisa dihapus') || errorMessage.contains('nonaktifkan')) {
          if (mounted) {
            setState(() => _isLoading = false);
            final confirmDeactivate = await showDialog<bool>(
              context: context,
              builder: (ctx) => AlertDialog(
                title: Text('Nonaktifkan Toko?', style: GoogleFonts.outfit(fontWeight: FontWeight.bold)),
                content: Text('Toko tidak bisa dihapus karena sudah memiliki data transaksi. Apakah Anda ingin menonaktifkannya agar tidak muncul di daftar?', style: GoogleFonts.outfit()),
                actions: [
                  TextButton(
                    onPressed: () => Navigator.pop(ctx, false),
                    child: Text('Batal', style: GoogleFonts.outfit(color: Colors.blueGrey)),
                  ),
                  ElevatedButton(
                    onPressed: () => Navigator.pop(ctx, true),
                    style: ElevatedButton.styleFrom(backgroundColor: Colors.orange),
                    child: Text('Nonaktifkan', style: GoogleFonts.outfit(color: Colors.white)),
                  ),
                ],
              ),
            );

            if (confirmDeactivate == true) {
              setState(() => _isLoading = true);
              try {
                final payload = store.toJson();
                payload['is_active'] = false;
                await _apiService.updateStore(store.id, payload);
                if (mounted) {
                  ScaffoldMessenger.of(context).showSnackBar(
                    SnackBar(
                      content: Text('Toko berhasil dinonaktifkan',
                          style: GoogleFonts.outfit(color: Colors.white)),
                      backgroundColor: Colors.green,
                    ),
                  );
                }
                _fetchStores();
              } catch (updateErr) {
                if (mounted) {
                  ScaffoldMessenger.of(context).showSnackBar(
                    SnackBar(
                      content: Text('Gagal menonaktifkan toko: $updateErr',
                          style: GoogleFonts.outfit(color: Colors.white)),
                      backgroundColor: Colors.red,
                    ),
                  );
                  setState(() => _isLoading = false);
                }
              }
            }
          }
          return;
        }

        if (mounted) {
          ScaffoldMessenger.of(context).showSnackBar(
            SnackBar(
              content: Text('Gagal menghapus toko: $e',
                  style: GoogleFonts.outfit(color: Colors.white)),
              backgroundColor: Colors.red,
            ),
          );
          setState(() => _isLoading = false);
        }
      }
    }
  }

  @override
  Widget build(BuildContext context) {
    return Scaffold(
      backgroundColor: const Color(0xFFF1F5F9),
      appBar: AppBar(
        backgroundColor: Colors.white,
        elevation: 0,
        leading: IconButton(
          icon: const Icon(Icons.arrow_back_ios_rounded,
              color: Color(0xFF334155)),
          onPressed: () => Navigator.pop(context),
        ),
        title: Column(
          crossAxisAlignment: CrossAxisAlignment.start,
          children: [
            Text('DAFTAR TOKO',
                style: GoogleFonts.outfit(
                    fontSize: 10,
                    color: AppColors.primaryGreen,
                    fontWeight: FontWeight.w800,
                    letterSpacing: 1)),
            Text('Manajemen Toko',
                style: GoogleFonts.outfit(
                    fontSize: 17,
                    fontWeight: FontWeight.w800,
                    color: const Color(0xFF1E293B))),
          ],
        ),
        actions: [
          IconButton(
            icon: const Icon(Icons.refresh_rounded, color: AppColors.primaryGreen),
            onPressed: () => _fetchStores(forceRefresh: true),
            tooltip: 'Muat Ulang',
          ),
        ],
      ),
      body: _isLoading
          ? const Center(child: CircularProgressIndicator())
          : Column(
              children: [
                Padding(
                  padding: const EdgeInsets.all(20),
                  child: TextField(
                    controller: _searchController,
                    onChanged: (value) => setState(() => _searchQuery = value),
                    decoration: InputDecoration(
                      hintText: 'Cari nama toko, owner, atau alamat...',
                      hintStyle: GoogleFonts.outfit(color: Colors.blueGrey),
                      prefixIcon: const Icon(Icons.search, color: Colors.blueGrey),
                      filled: true,
                      fillColor: Colors.white,
                      border: OutlineInputBorder(
                        borderRadius: BorderRadius.circular(16),
                        borderSide: BorderSide.none,
                      ),
                      contentPadding: const EdgeInsets.symmetric(horizontal: 20, vertical: 16),
                    ),
                  ),
                ),
                Expanded(
                  child: RefreshIndicator(
                    onRefresh: () => _fetchStores(forceRefresh: true),
                    child: Builder(builder: (context) {
                      final filteredStores = _stores.where((s) {
                        final q = _searchQuery.toLowerCase();
                        return s.name.toLowerCase().contains(q) ||
                               s.ownerName.toLowerCase().contains(q) ||
                               s.address.toLowerCase().contains(q);
                      }).toList();
                      
                      if (filteredStores.isEmpty) {
                        return _buildEmptyState();
                      }
                      return ListView.builder(
                        physics: const AlwaysScrollableScrollPhysics(),
                        padding: const EdgeInsets.symmetric(horizontal: 20),
                        itemCount: filteredStores.length,
                        itemBuilder: (ctx, i) => _buildStoreCard(filteredStores[i]),
                      );
                    }),
                  ),
                ),
              ],
            ),
      floatingActionButton: FloatingActionButton.extended(
        onPressed: () async {
          final result = await Navigator.push(
              context, MaterialPageRoute(builder: (_) => const AddStorePage()));
          if (result != null) {
            if (result is StoreModel) {
              setState(() {
                _stores.removeWhere((s) => s.id == result.id);
                _stores.insert(0, result);
              });
            }
            _fetchStores(forceRefresh: true);
          }
        },
        backgroundColor: AppColors.primaryGreen,
        icon: const Icon(Icons.add_rounded, color: Colors.white),
        label: Text('Tambah Toko Baru',
            style: GoogleFonts.outfit(
                color: Colors.white, fontWeight: FontWeight.w800)),
      ),
    );
  }

  Widget _buildStorePhotoWidget(String photoRef,
      {BoxFit fit = BoxFit.cover, Widget? fallback}) {
    final defaultFallback = fallback ??
        const Icon(Icons.storefront_rounded, size: 28, color: Colors.grey);

    if (photoRef.isEmpty) return defaultFallback;

    if (photoRef.startsWith('data:image/')) {
      try {
        final commaIdx = photoRef.indexOf(',');
        final b64 = commaIdx != -1 ? photoRef.substring(commaIdx + 1) : photoRef;
        return Image.memory(
          base64Decode(b64),
          fit: fit,
          errorBuilder: (_, __, ___) => defaultFallback,
        );
      } catch (_) {
        return defaultFallback;
      }
    }

    final isServerPath = photoRef.startsWith('http') ||
        photoRef.startsWith('/uploads/') ||
        photoRef.startsWith('uploads/');

    if (!kIsWeb && !isServerPath) {
      final file = File(photoRef);
      if (file.existsSync()) {
        return Image.file(
          file,
          fit: fit,
          errorBuilder: (_, __, ___) => defaultFallback,
        );
      }
    }

    return CachedNetworkImage(
      imageUrl: AppConstants.buildImageUrl(photoRef),
      fit: fit,
      placeholder: (context, url) => defaultFallback,
      errorWidget: (context, url, error) => defaultFallback,
    );
  }

  Widget _buildStoreCard(StoreModel store) {
    final photoRef = store.effectivePhoto;
    return Container(
      margin: const EdgeInsets.only(bottom: 16),
      decoration: BoxDecoration(
        color: Colors.white,
        borderRadius: BorderRadius.circular(24),
        boxShadow: [
          BoxShadow(
              color: Colors.black.withOpacity(0.04),
              blurRadius: 20,
              offset: const Offset(0, 8))
        ],
      ),
      child: ListTile(
        contentPadding:
            const EdgeInsets.symmetric(horizontal: 20, vertical: 12),
        leading: InkWell(
          onTap: () {
            if (photoRef.isNotEmpty) {
              showDialog(
                context: context,
                builder: (_) => Dialog(
                  backgroundColor: Colors.transparent,
                  insetPadding: const EdgeInsets.all(16),
                  child: Stack(
                    alignment: Alignment.center,
                    children: [
                      InteractiveViewer(
                        child: ClipRRect(
                          borderRadius: BorderRadius.circular(16),
                          child: _buildStorePhotoWidget(
                            photoRef,
                            fit: BoxFit.contain,
                            fallback: const Icon(Icons.error,
                                color: Colors.white, size: 40),
                          ),
                        ),
                      ),
                      Positioned(
                        top: 0,
                        right: 0,
                        child: IconButton(
                          icon: const Icon(Icons.close,
                              color: Colors.white, size: 30),
                          onPressed: () => Navigator.pop(context),
                        ),
                      ),
                    ],
                  ),
                ),
              );
            }
          },
          child: Container(
            width: 56,
            height: 56,
            decoration: BoxDecoration(
              color: store.isNew
                  ? Colors.orange.withOpacity(0.1)
                  : AppColors.primaryGreen.withOpacity(0.1),
              shape: BoxShape.circle,
            ),
            clipBehavior: Clip.antiAlias,
            child: photoRef.isNotEmpty
                ? _buildStorePhotoWidget(photoRef, fit: BoxFit.cover)
                : Icon(Icons.storefront_rounded,
                    color: store.isNew ? Colors.orange : AppColors.primaryGreen,
                    size: 28),
          ),
        ),
        title: Row(
          children: [
            Flexible(
              child: Text(
                store.name,
                style: GoogleFonts.outfit(
                  fontWeight: FontWeight.w800,
                  fontSize: 16,
                  color: const Color(0xFF1E293B),
                ),
                maxLines: 1,
                overflow: TextOverflow.ellipsis,
              ),
            ),
            const SizedBox(width: 8),
            Container(
              padding: const EdgeInsets.symmetric(horizontal: 8, vertical: 2),
              decoration: BoxDecoration(
                color: store.isNew
                    ? Colors.orange.withOpacity(0.1)
                    : Colors.green.withOpacity(0.1),
                borderRadius: BorderRadius.circular(20),
              ),
              child: Text(
                store.isNew ? 'TOKO BARU' : 'TOKO LAMA',
                style: GoogleFonts.outfit(
                    fontSize: 9,
                    fontWeight: FontWeight.w900,
                    color: store.isNew ? Colors.orange : Colors.green),
              ),
            ),
            if (store.isOffline || store.syncStatus == 'pending') ...[
              const SizedBox(width: 4),
              Container(
                padding: const EdgeInsets.symmetric(horizontal: 8, vertical: 2),
                decoration: BoxDecoration(
                  color: Colors.amber.shade100,
                  borderRadius: BorderRadius.circular(20),
                ),
                child: Text(
                  'OFFLINE',
                  style: GoogleFonts.outfit(
                      fontSize: 9,
                      fontWeight: FontWeight.w900,
                      color: Colors.amber.shade900),
                ),
              ),
            ],
            if (store.assignedEmployeeName != null &&
                store.assignedEmployeeName!.isNotEmpty) ...[
              const SizedBox(width: 4),
              Container(
                padding: const EdgeInsets.symmetric(horizontal: 8, vertical: 2),
                decoration: BoxDecoration(
                  color: Colors.blue.withOpacity(0.1),
                  borderRadius: BorderRadius.circular(20),
                ),
                child: Text(
                  'PIC: ${store.assignedEmployeeName}',
                  style: GoogleFonts.outfit(
                      fontSize: 9,
                      fontWeight: FontWeight.w700,
                      color: Colors.blue.shade800),
                ),
              ),
            ],
          ],
        ),
        subtitle: Column(
          crossAxisAlignment: CrossAxisAlignment.start,
          children: [
            const SizedBox(height: 4),
            Text(store.ownerName,
                style:
                    GoogleFonts.outfit(fontSize: 13, color: Colors.blueGrey)),
            const SizedBox(height: 2),
            if (store.address.isNotEmpty)
              Text(store.address,
                  style: GoogleFonts.outfit(
                      fontSize: 12, color: Colors.blueGrey.withOpacity(0.9)),
                  maxLines: 1,
                  overflow: TextOverflow.ellipsis),
            Builder(
              builder: (context) {
                final regions = [
                  if (store.village.isNotEmpty) store.village,
                  if (store.district.isNotEmpty) store.district,
                  if (store.regency.isNotEmpty) store.regency,
                  if (store.province.isNotEmpty) store.province,
                ];
                
                if (regions.isEmpty) {
                   if (store.address.isEmpty) {
                     return Text(
                       'Belum ada alamat/wilayah',
                       style: GoogleFonts.outfit(
                           fontSize: 12, color: Colors.blueGrey.withOpacity(0.8)),
                       maxLines: 1,
                       overflow: TextOverflow.ellipsis,
                     );
                   }
                   return const SizedBox.shrink();
                }

                return Padding(
                  padding: EdgeInsets.only(top: store.address.isNotEmpty ? 2 : 0),
                  child: Text(
                    regions.join(', '),
                    style: GoogleFonts.outfit(
                        fontSize: 11, color: Colors.blueGrey.withOpacity(0.7)),
                    maxLines: 1,
                    overflow: TextOverflow.ellipsis,
                  ),
                );
              },
            ),
            const SizedBox(height: 6),
            Row(
              children: [
                const Icon(Icons.repeat_rounded,
                    size: 12, color: AppColors.primaryGreen),
                const SizedBox(width: 4),
                Text(
                  '${store.visitFrequency} • ',
                  style: GoogleFonts.outfit(
                      fontSize: 11,
                      fontWeight: FontWeight.w700,
                      color: AppColors.primaryGreen),
                ),
                const Icon(Icons.calendar_month_rounded,
                    size: 12, color: Colors.blueGrey),
                const SizedBox(width: 4),
                Text(
                  store.visitDays.isEmpty
                      ? 'Belum diatur'
                      : store.visitDays
                          .map((d) => [
                                'Sen',
                                'Sel',
                                'Rab',
                                'Kam',
                                'Jum',
                                'Sab',
                                'Min'
                              ][d - 1])
                          .join(', '),
                  style: GoogleFonts.outfit(
                      fontSize: 11,
                      fontWeight: FontWeight.w600,
                      color: Colors.blueGrey),
                ),
              ],
            ),
          ],
        ),
        trailing: PopupMenuButton<String>(
          icon: const Icon(Icons.more_vert_rounded, color: Colors.blueGrey),
          shape: RoundedRectangleBorder(borderRadius: BorderRadius.circular(12)),
          onSelected: (value) async {
            if (value == 'history') {
              Navigator.push(
                context,
                MaterialPageRoute(
                  builder: (_) => StoreOrderHistoryPage(store: store),
                ),
              );
            } else if (value == 'edit') {
              final result = await Navigator.push(
                  context,
                  MaterialPageRoute(
                      builder: (_) => AddStorePage(existingStore: store)));
              if (result != null) {
                _fetchStores(forceRefresh: true);
              }
            } else if (value == 'delete') {
              _deleteStore(store);
            }
          },
          itemBuilder: (context) => [
            PopupMenuItem(
              value: 'history',
              child: Row(
                children: [
                  const Icon(Icons.history_rounded, color: Colors.blueAccent, size: 20),
                  const SizedBox(width: 12),
                  Text('Riwayat Pesanan', style: GoogleFonts.outfit(fontSize: 14)),
                ],
              ),
            ),
            PopupMenuItem(
              value: 'edit',
              child: Row(
                children: [
                  const Icon(Icons.edit_rounded, color: AppColors.primaryGreen, size: 20),
                  const SizedBox(width: 12),
                  Text('Edit Toko', style: GoogleFonts.outfit(fontSize: 14)),
                ],
              ),
            ),
            PopupMenuItem(
              value: 'delete',
              child: Row(
                children: [
                  const Icon(Icons.delete_outline_rounded, color: Colors.redAccent, size: 20),
                  const SizedBox(width: 12),
                  Text('Hapus Toko', style: GoogleFonts.outfit(fontSize: 14, color: Colors.redAccent)),
                ],
              ),
            ),
          ],
        ),
      ),
    );
  }

  Widget _buildEmptyState() {
    return Center(
      child: Column(
        mainAxisAlignment: MainAxisAlignment.center,
        children: [
          Icon(Icons.storefront_outlined,
              size: 80, color: Colors.grey.shade300),
          const SizedBox(height: 24),
          Text('Belum Ada Toko',
              style: GoogleFonts.outfit(
                  fontSize: 20,
                  fontWeight: FontWeight.w800,
                  color: const Color(0xFF334155))),
          const SizedBox(height: 8),
          Text('Tekan "Tambah Toko Baru" untuk\nmenambahkan daftar toko Anda.',
              textAlign: TextAlign.center,
              style: GoogleFonts.outfit(fontSize: 14, color: Colors.blueGrey)),
        ],
      ),
    );
  }
}
