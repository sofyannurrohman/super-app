import 'package:flutter/material.dart';
import 'package:google_fonts/google_fonts.dart';
import 'package:hris_app/core/theme/app_colors.dart';
import 'package:hris_app/core/network/api_client.dart';
import 'package:hris_app/injection.dart' as di;
import 'package:url_launcher/url_launcher.dart';
import 'package:intl/intl.dart';
import 'dart:io';
import 'package:image_picker/image_picker.dart';
import 'package:dio/dio.dart';
import 'delivery_checkout_page.dart';

class MtPoListPage extends StatefulWidget {
  const MtPoListPage({super.key});

  @override
  State<MtPoListPage> createState() => _MtPoListPageState();
}

class _MtPoListPageState extends State<MtPoListPage> {
  final ApiClient apiClient = di.sl<ApiClient>();
  bool _isLoading = true;
  List<dynamic> _poList = [];
  final ImagePicker _picker = ImagePicker();

  @override
  void initState() {
    super.initState();
    _fetchPoList();
  }

  Future<void> _fetchPoList() async {
    setState(() => _isLoading = true);
    try {
      final response = await apiClient.client.get('sales/modern-trade/orders');
      setState(() {
        _poList = response.data['data'] ?? [];
      });
    } catch (e) {
      if (mounted) {
        ScaffoldMessenger.of(context).showSnackBar(
          SnackBar(
              content: Text('Gagal memuat data PO: $e'),
              backgroundColor: Colors.red),
        );
      }
      setState(() {
        _poList = [];
      });
    } finally {
      setState(() => _isLoading = false);
    }
  }

  Future<void> _uploadSettlementProof(String poId, double amount) async {
    try {
      final XFile? image = await _picker.pickImage(
          source: ImageSource.gallery, imageQuality: 70);
      if (image == null) return;

      setState(() => _isLoading = true);

      final bytes = await image.readAsBytes();
      final formData = FormData.fromMap({
        'file': MultipartFile.fromBytes(bytes, filename: image.name),
        'type': 'settlement',
      });
      final uploadRes =
          await apiClient.client.post('sales/upload', data: formData);
      final photoUrl = uploadRes.data['url'] ?? '';

      // 2. Submit settlement
      await apiClient.client
          .post('sales/modern-trade/orders/$poId/submit-settlement', data: {
        'so_id': poId,
        'transfer_proof_url': photoUrl,
        'amount': amount,
        'payment_method': 'CASH', // default or you can ask user
      });

      if (mounted) {
        ScaffoldMessenger.of(context).showSnackBar(
          const SnackBar(
              content: Text('Bukti tagihan berhasil dikirim!'),
              backgroundColor: Colors.green),
        );
        _fetchPoList();
      }
    } catch (e) {
      if (mounted) {
        ScaffoldMessenger.of(context).showSnackBar(
          SnackBar(content: Text('Gagal: $e'), backgroundColor: Colors.red),
        );
      }
    } finally {
      if (mounted) setState(() => _isLoading = false);
    }
  }

  Future<void> _updateStatus(String soId, String newStatus) async {
    setState(() => _isLoading = true);
    try {
      await apiClient.client.put('sales/modern-trade/orders/$soId/status',
          data: {'status': newStatus});
      if (mounted) {
        ScaffoldMessenger.of(context).showSnackBar(
          const SnackBar(
              content: Text('Status berhasil diperbarui!'),
              backgroundColor: Colors.green),
        );
        _fetchPoList();
      }
    } catch (e) {
      if (mounted) {
        ScaffoldMessenger.of(context).showSnackBar(
          SnackBar(
              content: Text('Gagal mengubah status: $e'),
              backgroundColor: Colors.red),
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
          'Daftar PO Modern Trade',
          style: GoogleFonts.outfit(
              fontWeight: FontWeight.w800, color: Colors.white),
        ),
        backgroundColor: const Color(0xFF059669), // Emerald Aksen MT
        centerTitle: true,
        elevation: 0,
      ),
      body: _isLoading
          ? const Center(
              child: CircularProgressIndicator(color: Color(0xFF059669)))
          : _poList.isEmpty
              ? _buildEmptyState()
              : RefreshIndicator(
                  onRefresh: _fetchPoList,
                  child: ListView.builder(
                    padding: const EdgeInsets.all(20),
                    itemCount: _poList.length,
                    itemBuilder: (context, index) {
                      final po = _poList[index];
                      return _buildPoCard(po);
                    },
                  ),
                ),
    );
  }

  Widget _buildPoCard(dynamic po) {
    final status = po['status'] ?? 'DRAFT';
    final storeName = po['store']?['name'] ?? 'Outlet MT';
    final soNo = po['so_number'] ?? 'SO-MT-XXXX';
    final amount = (po['total_amount'] ?? 0.0).toDouble();
    final dateStr = po['created_at'] ?? po['order_date'] ?? '';
    final formattedDate = dateStr.isNotEmpty
        ? DateFormat('dd MMM yyyy, HH:mm')
            .format(DateTime.parse(dateStr).toLocal())
        : '-';
    final currencyFormat =
        NumberFormat.currency(locale: 'id', symbol: 'Rp ', decimalDigits: 0);

    Color statusColor = const Color(0xFF64748B); // slate-500
    String statusLabel = status;
    if (status == 'CONFIRMED' || status == 'APPROVED') {
      statusColor = const Color(0xFF059669);
      statusLabel = 'DISETUJUI';
    } // emerald-600
    if (status == 'WAREHOUSE_RECEIVED') {
      statusColor = const Color(0xFF7C3AED);
      statusLabel = 'SJ MASUK GUDANG';
    } // violet-600
    if (status == 'LOADING') {
      statusColor = const Color(0xFF2563EB);
      statusLabel = 'PROSES MUAT';
    } // blue-600
    if (status == 'DELIVERY' || status == 'IN_DELIVERY') {
      statusColor = const Color(0xFFD97706);
      statusLabel = 'SEDANG DIKIRIM';
    } // amber-600
    if (status == 'DELIVERED_AND_UPLOADED' || status == 'DELIVERED') {
      statusColor = const Color(0xFF0D9488);
      statusLabel = 'SELESAI KIRIM & TF';
    } // teal-600
    if (status == 'PAID' || status == 'SETTLED') {
      statusColor = const Color(0xFF15803D);
      statusLabel = 'SETTLEMENT SELESAI';
    } // green-700
    if (status == 'WAITING_TAX_DOCS') {
      statusColor = const Color(0xFF10B981);
      statusLabel = 'MENUNGGU PAJAK';
    } // emerald-500
    if (status == 'DRAFT' || status == 'PENDING') {
      statusColor = const Color(0xFF94A3B8);
      statusLabel = 'PENDING';
    } // slate-400

    return Container(
      margin: const EdgeInsets.only(bottom: 16),
      decoration: BoxDecoration(
        color: Colors.white,
        borderRadius: BorderRadius.circular(16),
        boxShadow: [
          BoxShadow(
              color: const Color(0xFF0F172A).withOpacity(0.04),
              blurRadius: 12,
              offset: const Offset(0, 4)),
        ],
        border:
            Border.all(color: const Color(0xFFE2E8F0), width: 1), // slate-200
      ),
      child: ClipRRect(
        borderRadius: BorderRadius.circular(16),
        child: Material(
          color: Colors.transparent,
          child: InkWell(
            onTap: () => _showPoDetailBottomSheet(po),
            child: Padding(
              padding: const EdgeInsets.all(16),
              child: Column(
                crossAxisAlignment: CrossAxisAlignment.start,
                children: [
                  Row(
                    mainAxisAlignment: MainAxisAlignment.spaceBetween,
                    children: [
                      Row(
                        children: [
                          Container(
                            padding: const EdgeInsets.all(8),
                            decoration: BoxDecoration(
                              color: const Color(0xFFECFDF5), // emerald-50
                              borderRadius: BorderRadius.circular(10),
                            ),
                            child: const Icon(Icons.receipt_long_rounded,
                                size: 16,
                                color: Color(0xFF10B981)), // emerald-500
                          ),
                          const SizedBox(width: 12),
                          Column(
                            crossAxisAlignment: CrossAxisAlignment.start,
                            children: [
                              Text(
                                soNo,
                                style: GoogleFonts.outfit(
                                    fontWeight: FontWeight.w800,
                                    color: const Color(0xFF0F172A),
                                    fontSize: 14),
                              ),
                              const SizedBox(height: 2),
                              Text(
                                formattedDate,
                                style: GoogleFonts.outfit(
                                    fontSize: 11,
                                    color: const Color(0xFF64748B),
                                    fontWeight: FontWeight.w500),
                              ),
                            ],
                          ),
                        ],
                      ),
                      Container(
                        padding: const EdgeInsets.symmetric(
                            horizontal: 10, vertical: 6),
                        decoration: BoxDecoration(
                          color: statusColor.withOpacity(0.1),
                          borderRadius: BorderRadius.circular(8),
                        ),
                        child: Text(
                          statusLabel,
                          style: GoogleFonts.outfit(
                              fontSize: 10,
                              fontWeight: FontWeight.w800,
                              color: statusColor,
                              letterSpacing: 0.5),
                        ),
                      ),
                    ],
                  ),
                  const SizedBox(height: 16),
                  Container(
                    padding: const EdgeInsets.all(12),
                    decoration: BoxDecoration(
                      color: const Color(0xFFF8FAFC), // slate-50
                      borderRadius: BorderRadius.circular(12),
                      border: Border.all(
                          color: const Color(0xFFF1F5F9)), // slate-100
                    ),
                    child: Row(
                      children: [
                        const Icon(Icons.storefront_rounded,
                            size: 16, color: Color(0xFF64748B)),
                        const SizedBox(width: 8),
                        Expanded(
                          child: Text(
                            storeName,
                            style: GoogleFonts.outfit(
                                fontWeight: FontWeight.w700,
                                fontSize: 13,
                                color: const Color(0xFF334155)),
                            maxLines: 1,
                            overflow: TextOverflow.ellipsis,
                          ),
                        ),
                        Text(
                          currencyFormat.format(amount),
                          style: GoogleFonts.outfit(
                              fontSize: 14,
                              fontWeight: FontWeight.w900,
                              color: const Color(0xFF059669)), // emerald-600
                        ),
                      ],
                    ),
                  ),
                  if (status == 'WAREHOUSE_RECEIVED' ||
                      status == 'LOADING' ||
                      status == 'APPROVED') ...[
                    const SizedBox(height: 16),
                    SizedBox(
                      width: double.infinity,
                      child: ElevatedButton.icon(
                        style: ElevatedButton.styleFrom(
                          backgroundColor: const Color(0xFF2563EB), // blue-600
                          foregroundColor: Colors.white,
                          elevation: 0,
                          shape: RoundedRectangleBorder(
                              borderRadius: BorderRadius.circular(10)),
                          padding: const EdgeInsets.symmetric(vertical: 12),
                        ),
                        onPressed: () => _updateStatus(po['id'], 'IN_DELIVERY'),
                        icon: const Icon(Icons.check_circle_outline_rounded,
                            size: 16),
                        label: Text(
                          'KONFIRMASI MUATAN & MULAI PENGIRIMAN',
                          style: GoogleFonts.outfit(
                              fontWeight: FontWeight.w800,
                              fontSize: 12,
                              letterSpacing: 0.5),
                        ),
                      ),
                    ),
                  ],
                  if (status == 'IN_DELIVERY' || status == 'SHIPPED') ...[
                    const SizedBox(height: 16),
                    SizedBox(
                      width: double.infinity,
                      child: ElevatedButton.icon(
                        style: ElevatedButton.styleFrom(
                          backgroundColor:
                              const Color(0xFF10B981), // emerald-500
                          foregroundColor: Colors.white,
                          elevation: 0,
                          shape: RoundedRectangleBorder(
                              borderRadius: BorderRadius.circular(10)),
                          padding: const EdgeInsets.symmetric(vertical: 12),
                        ),
                        onPressed: () async {
                          await _uploadSettlementProof(po['id'], amount);
                          _updateStatus(po['id'], 'DELIVERED_AND_UPLOADED');
                        },
                        icon: const Icon(Icons.upload_file_rounded, size: 16),
                        label: Text(
                          'SELESAI KIRIM & UPLOAD BUKTI TF',
                          style: GoogleFonts.outfit(
                              fontWeight: FontWeight.w800,
                              fontSize: 12,
                              letterSpacing: 0.5),
                        ),
                      ),
                    ),
                  ],
                ],
              ),
            ),
          ),
        ),
      ),
    );
  }

  String _getStatusDescription(String status) {
    switch (status) {
      case 'PENDING':
      case 'DRAFT':
        return 'Pesanan baru dibuat dan belum diajukan atau sedang diproses awal.';
      case 'WAITING_TAX_DOCS':
        return 'Menunggu kelengkapan dokumen Faktur Pajak dari Admin Pajak Pusat.';
      case 'APPROVED':
      case 'CONFIRMED':
        return 'Pesanan telah disetujui Pajak. Saat ini menunggu tim Gudang memproses pengiriman.';
      case 'WAREHOUSE_RECEIVED':
        return 'Surat Jalan telah masuk dan diterima oleh pihak Gudang.';
      case 'LOADING':
        return 'Barang sedang diproses muat oleh Admin Gudang ke armada pengiriman.';
      case 'IN_DELIVERY':
      case 'SHIPPED':
        return 'Pesanan sedang dikirim ke ritel. Klik tombol Selesai Kirim setelah barang sampai.';
      case 'DELIVERED_AND_UPLOADED':
      case 'DELIVERED':
        return 'Pesanan telah sampai dan bukti tagihan sudah diunggah, menunggu Settlement admin.';
      case 'PAID':
      case 'SETTLED':
        return 'Settlement / Penagihan telah selesai dan diverifikasi oleh Admin Pusat.';
      default:
        return 'Pesanan sedang diproses.';
    }
  }

  void _showPoDetailBottomSheet(dynamic po) {
    final status = po['status'] ?? 'DRAFT';
    final storeName = po['store']?['name'] ?? 'Outlet MT';
    final address = po['store']?['address'] ?? 'Alamat tidak tersedia';
    final soNo = po['so_number'] ?? 'SO-MT-XXXX';
    final amount = (po['total_amount'] ?? 0.0).toDouble();
    final items = po['items'] as List? ?? [];
    final currencyFormat =
        NumberFormat.currency(locale: 'id', symbol: 'Rp ', decimalDigits: 0);

    final billingUrl = po['billing_note_url'] as String?;
    final taxUrl = po['tax_invoice_url'] as String?;
    final deliveryUrl = po['delivery_note_url'] as String?;
    final bool hasDocs =
        billingUrl != null || taxUrl != null || deliveryUrl != null;

    showModalBottomSheet(
      context: context,
      isScrollControlled: true,
      shape: const RoundedRectangleBorder(
        borderRadius: BorderRadius.vertical(top: Radius.circular(32)),
      ),
      builder: (context) => DraggableScrollableSheet(
        initialChildSize: 0.7,
        maxChildSize: 0.95,
        minChildSize: 0.5,
        expand: false,
        builder: (context, scrollController) => ListView(
          controller: scrollController,
          padding: const EdgeInsets.all(24),
          children: [
            Center(
              child: Container(
                width: 40,
                height: 4,
                decoration: BoxDecoration(
                    color: Colors.grey[300],
                    borderRadius: BorderRadius.circular(10)),
              ),
            ),
            const SizedBox(height: 24),
            Text(
              soNo,
              style: GoogleFonts.outfit(
                  fontWeight: FontWeight.w900,
                  fontSize: 22,
                  color: const Color(0xFF0F172A)),
            ),
            const SizedBox(height: 4),
            Text(
              storeName,
              style: GoogleFonts.outfit(
                  fontWeight: FontWeight.w800,
                  fontSize: 15,
                  color: const Color(0xFF475569)),
            ),
            Text(
              address,
              style: GoogleFonts.outfit(fontSize: 12, color: Colors.blueGrey),
            ),
            const SizedBox(height: 16),
            Container(
              padding: const EdgeInsets.all(12),
              decoration: BoxDecoration(
                color: Colors.blue.shade50,
                borderRadius: BorderRadius.circular(12),
                border: Border.all(color: Colors.blue.shade100),
              ),
              child: Row(
                crossAxisAlignment: CrossAxisAlignment.start,
                children: [
                  Icon(Icons.info_outline_rounded,
                      size: 20, color: Colors.blue.shade700),
                  const SizedBox(width: 12),
                  Expanded(
                    child: Text(
                      _getStatusDescription(status),
                      style: GoogleFonts.outfit(
                          fontSize: 12,
                          color: Colors.blue.shade900,
                          height: 1.4),
                    ),
                  ),
                ],
              ),
            ),
            const SizedBox(height: 24),
            const Divider(),
            const SizedBox(height: 16),

            // Dokumen Pajak Section
            if (hasDocs) ...[
              Text(
                'DOKUMEN DARI ADMIN PAJAK',
                style: GoogleFonts.outfit(
                    fontSize: 11,
                    fontWeight: FontWeight.w900,
                    color: const Color(0xFF059669),
                    letterSpacing: 1),
              ),
              const SizedBox(height: 12),
              Row(
                children: [
                  if (billingUrl != null)
                    Expanded(child: _buildDocBtn('Nota Penagihan', billingUrl)),
                  if (billingUrl != null && taxUrl != null)
                    const SizedBox(width: 8),
                  if (taxUrl != null)
                    Expanded(child: _buildDocBtn('Faktur Pajak', taxUrl)),
                  if (taxUrl != null && deliveryUrl != null)
                    const SizedBox(width: 8),
                  if (deliveryUrl != null)
                    Expanded(child: _buildDocBtn('Surat Jalan', deliveryUrl)),
                ],
              ),
              const SizedBox(height: 24),
              const Divider(),
              const SizedBox(height: 16),
            ],

            // Rincian Barang
            Text(
              'RINCIAN BARANG',
              style: GoogleFonts.outfit(
                  fontSize: 11,
                  fontWeight: FontWeight.w900,
                  color: Colors.blueGrey,
                  letterSpacing: 1),
            ),
            const SizedBox(height: 12),
            ...items.map((item) {
              final prodName = item['product']?['name'] ?? 'Produk';
              final qty = item['ordered_quantity'] ?? 0;
              final unit = item['unit'] ?? 'PCS';
              final sub = (item['subtotal'] ?? 0.0).toDouble();

              return Container(
                margin: const EdgeInsets.only(bottom: 8),
                padding: const EdgeInsets.all(12),
                decoration: BoxDecoration(
                  color: const Color(0xFFF8FAFC),
                  borderRadius: BorderRadius.circular(16),
                ),
                child: Row(
                  mainAxisAlignment: MainAxisAlignment.spaceBetween,
                  children: [
                    Expanded(
                      child: Column(
                        crossAxisAlignment: CrossAxisAlignment.start,
                        children: [
                          Text(prodName,
                              style: GoogleFonts.outfit(
                                  fontWeight: FontWeight.w700,
                                  fontSize: 13,
                                  color: const Color(0xFF1E293B))),
                          Text('$qty $unit',
                              style: GoogleFonts.outfit(
                                  fontSize: 11,
                                  color: Colors.blueGrey,
                                  fontWeight: FontWeight.bold)),
                        ],
                      ),
                    ),
                    Text(currencyFormat.format(sub),
                        style: GoogleFonts.outfit(
                            fontWeight: FontWeight.w900,
                            fontSize: 13,
                            color: const Color(0xFF0F172A))),
                  ],
                ),
              );
            }).toList(),

            const SizedBox(height: 24),
            Container(
              padding: const EdgeInsets.all(16),
              decoration: BoxDecoration(
                color: const Color(0xFFF1F5F9),
                borderRadius: BorderRadius.circular(20),
              ),
              child: Row(
                mainAxisAlignment: MainAxisAlignment.spaceBetween,
                children: [
                  Text('TOTAL VALUE PO',
                      style: GoogleFonts.outfit(
                          fontWeight: FontWeight.w900,
                          color: Colors.blueGrey,
                          fontSize: 12)),
                  Text(currencyFormat.format(amount),
                      style: GoogleFonts.outfit(
                          fontWeight: FontWeight.w900,
                          color: const Color(0xFF0F172A),
                          fontSize: 16)),
                ],
              ),
            ),

            // Tombol Tandai Sedang Dikirim jika status PO = LOADING atau APPROVED atau WAREHOUSE_RECEIVED
            if (status == 'WAREHOUSE_RECEIVED' ||
                status == 'LOADING' ||
                status == 'APPROVED') ...[
              const SizedBox(height: 24),
              ElevatedButton.icon(
                style: ElevatedButton.styleFrom(
                  backgroundColor: const Color(0xFF2563EB), // blue-600
                  foregroundColor: Colors.white,
                  minimumSize: const Size(double.infinity, 54),
                  shape: RoundedRectangleBorder(
                      borderRadius: BorderRadius.circular(16)),
                  elevation: 2,
                  shadowColor: const Color(0xFF2563EB).withOpacity(0.3),
                ),
                onPressed: () {
                  Navigator.pop(context); // Tutup bottom sheet
                  _updateStatus(po['id'], 'IN_DELIVERY');
                },
                icon: const Icon(Icons.check_circle_outline_rounded),
                label: Text(
                  'KONFIRMASI MUATAN & MULAI PENGIRIMAN',
                  style: GoogleFonts.outfit(
                      fontWeight: FontWeight.w900, fontSize: 14),
                ),
              ),
            ],

            // Tombol Selesai Kirim & Upload TF jika status IN_DELIVERY atau SHIPPED
            if (status == 'IN_DELIVERY' || status == 'SHIPPED') ...[
              const SizedBox(height: 24),
              ElevatedButton.icon(
                style: ElevatedButton.styleFrom(
                  backgroundColor: const Color(0xFF10B981), // emerald-500
                  foregroundColor: Colors.white,
                  minimumSize: const Size(double.infinity, 54),
                  shape: RoundedRectangleBorder(
                      borderRadius: BorderRadius.circular(16)),
                  elevation: 2,
                ),
                onPressed: () async {
                  Navigator.pop(context); // Tutup bottom sheet
                  await _uploadSettlementProof(po['id'], amount);
                  _updateStatus(po['id'], 'DELIVERED_AND_UPLOADED');
                },
                icon: const Icon(Icons.upload_file_rounded),
                label: Text(
                  'SELESAI KIRIM & UPLOAD BUKTI TF',
                  style: GoogleFonts.outfit(
                      fontWeight: FontWeight.w900, fontSize: 14),
                ),
              ),
            ],
            const SizedBox(height: 24),
          ],
        ),
      ),
    );
  }

  Widget _buildDocBtn(String label, String url) {
    return ElevatedButton.icon(
      style: ElevatedButton.styleFrom(
        backgroundColor: const Color(0xFFECFDF5),
        foregroundColor: const Color(0xFF059669),
        padding: const EdgeInsets.symmetric(vertical: 12, horizontal: 4),
        shape: RoundedRectangleBorder(borderRadius: BorderRadius.circular(12)),
        elevation: 0,
      ),
      onPressed: () async {
        try {
          String finalUrl = url;
          if (!finalUrl.startsWith('http')) {
            String baseUrl = apiClient.client.options.baseUrl;
            if (baseUrl.endsWith('/'))
              baseUrl = baseUrl.substring(0, baseUrl.length - 1);
            if (!finalUrl.startsWith('/')) finalUrl = '/$finalUrl';
            finalUrl = '$baseUrl$finalUrl';
          }
          final uri = Uri.parse(finalUrl);
          await launchUrl(uri, mode: LaunchMode.externalApplication);
        } catch (e) {
          debugPrint('Could not launch doc: $e');
        }
      },
      icon: const Icon(Icons.file_download_rounded, size: 16),
      label: Text(label,
          style: GoogleFonts.outfit(fontSize: 9, fontWeight: FontWeight.w800)),
    );
  }

  Widget _buildEmptyState() {
    return Center(
      child: Column(
        mainAxisAlignment: MainAxisAlignment.center,
        children: [
          const Icon(Icons.receipt_long_rounded,
              size: 80, color: Color(0xFFE2E8F0)),
          const SizedBox(height: 16),
          Text(
            'Tidak ada PO Modern Trade',
            style: GoogleFonts.outfit(
                fontSize: 18,
                fontWeight: FontWeight.bold,
                color: const Color(0xFF94A3B8)),
          ),
        ],
      ),
    );
  }
}
