import 'package:shared_preferences/shared_preferences.dart';
import 'package:flutter/material.dart';
import 'package:google_fonts/google_fonts.dart';
import 'package:hris_app/features/sales/data/models/store_model.dart';
import 'package:hris_app/features/sales/data/models/visit_plan_model.dart';
import 'package:hris_app/features/sales/data/services/store_api_service.dart';
import 'package:hris_app/features/sales/presentation/pages/optimal_route_map_page.dart';
import 'package:hris_app/injection.dart' as di;
import 'package:hris_app/core/network/api_client.dart';

class RoutePlanningPage extends StatefulWidget {
  const RoutePlanningPage({super.key});

  @override
  State<RoutePlanningPage> createState() => _RoutePlanningPageState();
}

class _RoutePlanningPageState extends State<RoutePlanningPage> {
  final _apiService = StoreApiService(apiClient: di.sl<ApiClient>(), prefs: di.sl<SharedPreferences>());
  bool _isLoading = true;
  List<StoreModel> _stores = [];
  final Set<String> _selectedStoreIds = {};

  @override
  void initState() {
    super.initState();
    _fetchStores();
  }

  Future<void> _fetchStores() async {
    setState(() => _isLoading = true);
    try {
      final stores = await _apiService.getStores();
      // Only show stores with valid coordinates for routing
      setState(() {
        _stores = stores
            .where((s) => s.latitude != 0.0 && s.longitude != 0.0)
            .toList();
      });
    } catch (e) {
      debugPrint('Error fetching stores: $e');
    } finally {
      if (mounted) setState(() => _isLoading = false);
    }
  }

  void _toggleSelection(String id) {
    setState(() {
      if (_selectedStoreIds.contains(id)) {
        _selectedStoreIds.remove(id);
      } else {
        _selectedStoreIds.add(id);
      }
    });
  }

  void _navigateToMap() {
    if (_selectedStoreIds.length < 2) {
      ScaffoldMessenger.of(context).showSnackBar(
        const SnackBar(
            content: Text('Pilih minimal 2 toko untuk rute optimal'),
            backgroundColor: Colors.orange),
      );
      return;
    }

    final selectedStores =
        _stores.where((s) => _selectedStoreIds.contains(s.id)).toList();
    final plans = selectedStores
        .map((s) => VisitPlanModel(
              storeId: s.id,
              storeName: s.name,
              address: s.address,
              status: 'PLANNED',
              isExtra: false,
              latitude: s.latitude,
              longitude: s.longitude,
            ))
        .toList();
    Navigator.push(context,
        MaterialPageRoute(builder: (_) => OptimalRouteMapPage(plans: plans)));
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
            Text('RUTE KUNJUNGAN',
                style: GoogleFonts.outfit(
                    fontSize: 10,
                    color: Colors.blueAccent,
                    fontWeight: FontWeight.w800,
                    letterSpacing: 1)),
            Text('Pilih Toko Tujuan',
                style: GoogleFonts.outfit(
                    fontSize: 17,
                    fontWeight: FontWeight.w800,
                    color: const Color(0xFF1E293B))),
          ],
        ),
      ),
      body: _isLoading
          ? const Center(child: CircularProgressIndicator())
          : Column(
              children: [
                Container(
                  padding:
                      const EdgeInsets.symmetric(horizontal: 20, vertical: 16),
                  color: Colors.white,
                  child: Row(
                    children: [
                      const Icon(Icons.info_outline_rounded,
                          color: Colors.blueAccent, size: 20),
                      const SizedBox(width: 12),
                      Expanded(
                        child: Text(
                          'Pilih beberapa toko untuk melihat rute urutan kunjungan paling optimal.',
                          style: GoogleFonts.outfit(
                              fontSize: 12,
                              color: Colors.blueGrey,
                              fontWeight: FontWeight.w600),
                        ),
                      ),
                    ],
                  ),
                ),
                Expanded(
                  child: _stores.isEmpty
                      ? _buildEmptyState()
                      : ListView.builder(
                          padding: const EdgeInsets.all(20),
                          itemCount: _stores.length,
                          itemBuilder: (ctx, i) => _buildStoreItem(_stores[i]),
                        ),
                ),
              ],
            ),
      floatingActionButton: _selectedStoreIds.length >= 2
          ? FloatingActionButton.extended(
              onPressed: _navigateToMap,
              backgroundColor: Colors.blueAccent,
              icon: const Icon(Icons.map_rounded, color: Colors.white),
              label: Text('Tampilkan Rute',
                  style: GoogleFonts.outfit(
                      color: Colors.white, fontWeight: FontWeight.w900)),
            )
          : null,
    );
  }

  Widget _buildStoreItem(StoreModel store) {
    final isSelected = _selectedStoreIds.contains(store.id);
    return Container(
      margin: const EdgeInsets.only(bottom: 12),
      decoration: BoxDecoration(
        color: isSelected ? Colors.blueAccent.withOpacity(0.05) : Colors.white,
        borderRadius: BorderRadius.circular(20),
        border: Border.all(
            color: isSelected ? Colors.blueAccent : Colors.transparent,
            width: 2),
        boxShadow: [
          if (!isSelected)
            BoxShadow(
                color: Colors.black.withOpacity(0.03),
                blurRadius: 10,
                offset: const Offset(0, 4))
        ],
      ),
      child: CheckboxListTile(
        value: isSelected,
        onChanged: (val) => _toggleSelection(store.id),
        activeColor: Colors.blueAccent,
        shape: RoundedRectangleBorder(borderRadius: BorderRadius.circular(20)),
        contentPadding: const EdgeInsets.symmetric(horizontal: 16, vertical: 8),
        title: Text(store.name,
            style: GoogleFonts.outfit(
                fontWeight: FontWeight.w800,
                fontSize: 15,
                color: const Color(0xFF1E293B))),
        subtitle: Text(store.address,
            style: GoogleFonts.outfit(fontSize: 12, color: Colors.blueGrey),
            maxLines: 1,
            overflow: TextOverflow.ellipsis),
        secondary: Container(
          width: 48,
          height: 48,
          decoration: BoxDecoration(
              color: Colors.blueAccent.withOpacity(0.1),
              shape: BoxShape.circle),
          child: const Icon(Icons.storefront_rounded, color: Colors.blueAccent),
        ),
      ),
    );
  }

  Widget _buildEmptyState() {
    return Center(
      child: Column(
        mainAxisAlignment: MainAxisAlignment.center,
        children: [
          Icon(Icons.location_off_rounded,
              size: 80, color: Colors.grey.shade300),
          const SizedBox(height: 24),
          Text('Belum Ada Toko Valid',
              style: GoogleFonts.outfit(
                  fontSize: 20,
                  fontWeight: FontWeight.w800,
                  color: const Color(0xFF334155))),
          const SizedBox(height: 8),
          Text(
              'Toko harus memiliki koordinat GPS\nuntuk bisa dihitung rutenya.',
              textAlign: TextAlign.center,
              style: GoogleFonts.outfit(fontSize: 14, color: Colors.blueGrey)),
        ],
      ),
    );
  }
}
