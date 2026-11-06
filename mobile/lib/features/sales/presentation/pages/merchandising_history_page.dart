import 'package:cached_network_image/cached_network_image.dart';
import 'package:flutter/material.dart';
import 'package:google_fonts/google_fonts.dart';
import 'package:hris_app/core/network/api_client.dart';
import 'package:hris_app/core/utils/constants.dart';
import 'package:hris_app/injection.dart' as di;
import 'package:intl/intl.dart';

class MerchandisingHistoryPage extends StatefulWidget {
  const MerchandisingHistoryPage({Key? key}) : super(key: key);

  @override
  State<MerchandisingHistoryPage> createState() =>
      _MerchandisingHistoryPageState();
}

class _MerchandisingHistoryPageState extends State<MerchandisingHistoryPage> {
  final ApiClient _apiClient = di.sl<ApiClient>();
  bool _isLoading = true;
  List<dynamic> _history = [];

  @override
  void initState() {
    super.initState();
    _fetchHistory();
  }

  Future<void> _fetchHistory() async {
    setState(() => _isLoading = true);
    try {
      final res = await _apiClient.client.get('modern-trade/merchandising');
      if (res.statusCode == 200) {
        if (mounted) {
          setState(() {
            _history = res.data['data'] ?? [];
            // Sort by visit time descending
            _history.sort((a, b) {
              final dateA =
                  DateTime.tryParse(a['visit_time'] ?? '') ?? DateTime.now();
              final dateB =
                  DateTime.tryParse(b['visit_time'] ?? '') ?? DateTime.now();
              return dateB.compareTo(dateA);
            });
          });
        }
      }
    } catch (e) {
      debugPrint('Error fetch merchandising history: $e');
      if (mounted) {
        ScaffoldMessenger.of(context).showSnackBar(
          SnackBar(content: Text('Gagal mengambil data riwayat: $e')),
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
          'Riwayat Merchandising',
          style: GoogleFonts.outfit(
              fontWeight: FontWeight.w800, color: Colors.white, fontSize: 16),
        ),
        backgroundColor: const Color(0xFFD97706),
        centerTitle: true,
        elevation: 0,
        leading: IconButton(
          icon: const Icon(Icons.arrow_back_rounded, color: Colors.white),
          onPressed: () => Navigator.pop(context),
        ),
      ),
      body: _isLoading
          ? const Center(
              child: CircularProgressIndicator(color: Color(0xFFD97706)))
          : _history.isEmpty
              ? _buildEmptyState()
              : RefreshIndicator(
                  onRefresh: _fetchHistory,
                  color: const Color(0xFFD97706),
                  child: ListView.separated(
                    padding: const EdgeInsets.all(16),
                    itemCount: _history.length,
                    separatorBuilder: (_, __) => const SizedBox(height: 12),
                    itemBuilder: (context, index) {
                      final item = _history[index];
                      final storeName = item['store_name'] ??
                          item['supermarket_name'] ??
                          'Outlet';
                      final visitTimeStr = item['visit_time'];
                      final visitDate = visitTimeStr != null
                          ? DateFormat('dd MMM yyyy, HH:mm')
                              .format(DateTime.parse(visitTimeStr).toLocal())
                          : '-';
                      final compliant = item['planogram_compliant'] == true;

                      String? photoUrl = item['photo_url'];
                      if (photoUrl != null && photoUrl.trim().isNotEmpty) {
                        photoUrl = AppConstants.buildImageUrl(photoUrl);
                      }

                      return Container(
                        padding: const EdgeInsets.all(16),
                        decoration: BoxDecoration(
                          color: Colors.white,
                          borderRadius: BorderRadius.circular(16),
                          border: Border.all(color: Colors.grey.shade200),
                          boxShadow: [
                            BoxShadow(
                              color: Colors.black.withOpacity(0.02),
                              blurRadius: 10,
                              offset: const Offset(0, 4),
                            ),
                          ],
                        ),
                        child: Column(
                          crossAxisAlignment: CrossAxisAlignment.start,
                          children: [
                            Row(
                              mainAxisAlignment: MainAxisAlignment.spaceBetween,
                              children: [
                                Expanded(
                                  child: Text(
                                    storeName,
                                    style: GoogleFonts.outfit(
                                      fontWeight: FontWeight.w800,
                                      fontSize: 16,
                                      color: const Color(0xFF1E293B),
                                    ),
                                  ),
                                ),
                                Container(
                                  padding: const EdgeInsets.symmetric(
                                      horizontal: 8, vertical: 4),
                                  decoration: BoxDecoration(
                                    color: compliant
                                        ? const Color(0xFFDCFCE7)
                                        : const Color(0xFFFFE4E6),
                                    borderRadius: BorderRadius.circular(8),
                                  ),
                                  child: Text(
                                    compliant
                                        ? 'Patuh Planogram'
                                        : 'Tidak Patuh',
                                    style: GoogleFonts.outfit(
                                      fontWeight: FontWeight.w700,
                                      fontSize: 10,
                                      color: compliant
                                          ? const Color(0xFF16A34A)
                                          : const Color(0xFFE11D48),
                                    ),
                                  ),
                                ),
                              ],
                            ),
                            const SizedBox(height: 8),
                            Row(
                              children: [
                                const Icon(Icons.calendar_month_rounded,
                                    size: 14, color: Colors.blueGrey),
                                const SizedBox(width: 4),
                                Text(
                                  visitDate,
                                  style: GoogleFonts.outfit(
                                    fontWeight: FontWeight.w500,
                                    fontSize: 12,
                                    color: Colors.blueGrey,
                                  ),
                                ),
                              ],
                            ),
                            const SizedBox(height: 12),
                            const Divider(height: 1, color: Color(0xFFF1F5F9)),
                            const SizedBox(height: 12),
                            Row(
                              mainAxisAlignment: MainAxisAlignment.spaceBetween,
                              children: [
                                Column(
                                  crossAxisAlignment: CrossAxisAlignment.start,
                                  children: [
                                    Text(
                                      'MD / PIC',
                                      style: GoogleFonts.outfit(
                                        fontWeight: FontWeight.w600,
                                        fontSize: 10,
                                        color: Colors.blueGrey.shade400,
                                      ),
                                    ),
                                    Text(
                                      item['md_name'] ??
                                          item['merchandiser_name'] ??
                                          '-',
                                      style: GoogleFonts.outfit(
                                        fontWeight: FontWeight.w700,
                                        fontSize: 12,
                                        color: const Color(0xFF334155),
                                      ),
                                    ),
                                  ],
                                ),
                                if (photoUrl != null &&
                                    photoUrl.toString().isNotEmpty)
                                  InkWell(
                                    onTap: () =>
                                        _showImagePreview(context, photoUrl!),
                                    borderRadius: BorderRadius.circular(8),
                                    child: ClipRRect(
                                      borderRadius: BorderRadius.circular(8),
                                      child: CachedNetworkImage(imageUrl: photoUrl, width: 48,
                                        height: 48,
                                        fit: BoxFit.cover, placeholder: (context, url) => const Center(child: CircularProgressIndicator()), errorWidget: (context, url, error) => const Icon(Icons.error), memCacheWidth: 800,),
                                    ),
                                  )
                              ],
                            ),
                          ],
                        ),
                      );
                    },
                  ),
                ),
    );
  }

  void _showImagePreview(BuildContext context, String imageUrl) {
    showDialog(
      context: context,
      builder: (_) => Dialog(
        backgroundColor: Colors.transparent,
        insetPadding: const EdgeInsets.all(16),
        child: Stack(
          alignment: Alignment.center,
          children: [
            InteractiveViewer(
              panEnabled: true,
              minScale: 0.5,
              maxScale: 4.0,
              child: ClipRRect(
                borderRadius: BorderRadius.circular(16),
                child: CachedNetworkImage(imageUrl: imageUrl, fit: BoxFit.contain, placeholder: (context, url) => const Center(child: CircularProgressIndicator()), errorWidget: (context, url, error) => const Icon(Icons.error), memCacheWidth: 800,),
              ),
            ),
            Positioned(
              top: 0,
              right: 0,
              child: IconButton(
                icon: const Icon(Icons.cancel, color: Colors.white, size: 36),
                onPressed: () => Navigator.pop(context),
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
          Icon(Icons.history_rounded, size: 64, color: Colors.grey.shade300),
          const SizedBox(height: 16),
          Text(
            'Belum Ada Riwayat',
            style: GoogleFonts.outfit(
              fontWeight: FontWeight.w800,
              fontSize: 18,
              color: Colors.blueGrey,
            ),
          ),
          const SizedBox(height: 8),
          Text(
            'Anda belum membuat laporan merchandising.',
            style: GoogleFonts.outfit(
              fontWeight: FontWeight.w500,
              fontSize: 14,
              color: Colors.grey,
            ),
          ),
        ],
      ),
    );
  }
}
