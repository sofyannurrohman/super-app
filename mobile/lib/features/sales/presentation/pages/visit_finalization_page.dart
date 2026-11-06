import 'package:flutter/material.dart';
import 'package:google_fonts/google_fonts.dart';
import 'package:hris_app/core/theme/app_colors.dart';
import 'package:hris_app/features/sales/data/services/sales_api_service.dart';
import 'package:hris_app/injection.dart' as di;
import 'package:hris_app/core/network/api_client.dart';

class VisitFinalizationPage extends StatefulWidget {
  final String storeName;
  final String? receiptPath;
  final String transactionId;
  final String? notes;

  const VisitFinalizationPage({
    super.key,
    required this.storeName,
    required this.transactionId,
    this.receiptPath,
    this.notes,
  });

  @override
  State<VisitFinalizationPage> createState() => _VisitFinalizationPageState();
}

class _VisitFinalizationPageState extends State<VisitFinalizationPage> {
  final _noNotaCtrl = TextEditingController();
  final _totalCtrl = TextEditingController();
  final _catatanCtrl = TextEditingController();
  bool _isSubmitting = false;

  @override
  void initState() {
    super.initState();
    if (widget.notes != null) {
      _catatanCtrl.text = widget.notes!;
    }
    // Set a default receipt number if empty
    _noNotaCtrl.text =
        'DIG-${DateTime.now().millisecondsSinceEpoch.toString().substring(7)}';
  }

  @override
  void dispose() {
    _noNotaCtrl.dispose();
    _totalCtrl.dispose();
    _catatanCtrl.dispose();
    super.dispose();
  }

  Future<void> _submitToSystem() async {
    if (_isSubmitting) return;
    setState(() => _isSubmitting = true);

    final apiService = SalesApiService(apiClient: di.sl<ApiClient>());
    try {
      await apiService.verifyTransaction(widget.transactionId, {
        'receipt_no': _noNotaCtrl.text,
        'total_amount': double.tryParse(
                _totalCtrl.text.replaceAll(RegExp(r'[^0-9]'), '')) ??
            0.0,
        'notes': _catatanCtrl.text,
      });

      if (mounted) {
        Navigator.pop(context, true);
        ScaffoldMessenger.of(context).showSnackBar(SnackBar(
          content: Text('✓ Data ${widget.storeName} berhasil diverifikasi!',
              style: GoogleFonts.outfit(fontWeight: FontWeight.w700)),
          backgroundColor: Colors.green,
          behavior: SnackBarBehavior.floating,
          shape:
              RoundedRectangleBorder(borderRadius: BorderRadius.circular(14)),
        ));
      }
    } catch (e) {
      if (mounted) {
        ScaffoldMessenger.of(context).showSnackBar(SnackBar(
          content:
              Text('Gagal verifikasi data: $e', style: GoogleFonts.outfit()),
          backgroundColor: Colors.red,
        ));
      }
    } finally {
      if (mounted) setState(() => _isSubmitting = false);
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
            onPressed: () => Navigator.pop(context)),
        title: Column(crossAxisAlignment: CrossAxisAlignment.start, children: [
          Text('KONFIRMASI DATA DIGITAL',
              style: GoogleFonts.outfit(
                  fontSize: 10,
                  color: Colors.blueAccent,
                  fontWeight: FontWeight.w800,
                  letterSpacing: 1)),
          Text(widget.storeName,
              style: GoogleFonts.outfit(
                  fontSize: 15,
                  fontWeight: FontWeight.w800,
                  color: const Color(0xFF1E293B))),
        ]),
      ),
      body: SingleChildScrollView(
        padding: const EdgeInsets.all(24),
        child: Column(crossAxisAlignment: CrossAxisAlignment.start, children: [
          Container(
            padding: const EdgeInsets.all(24),
            decoration: BoxDecoration(
              color: Colors.white,
              borderRadius: BorderRadius.circular(28),
              boxShadow: [
                BoxShadow(
                    color: Colors.black.withOpacity(0.04),
                    blurRadius: 20,
                    offset: const Offset(0, 8))
              ],
            ),
            child: Column(
              crossAxisAlignment: CrossAxisAlignment.start,
              children: [
                Row(
                  children: [
                    Container(
                      padding: const EdgeInsets.all(12),
                      decoration: BoxDecoration(
                          color: Colors.green.withOpacity(0.1),
                          shape: BoxShape.circle),
                      child: const Icon(Icons.check_circle_outline_rounded,
                          color: Colors.green, size: 28),
                    ),
                    const SizedBox(width: 16),
                    Expanded(
                      child: Column(
                        crossAxisAlignment: CrossAxisAlignment.start,
                        children: [
                          Text('Order Siap Kirim',
                              style: GoogleFonts.outfit(
                                  fontWeight: FontWeight.w900,
                                  fontSize: 18,
                                  color: Colors.green)),
                          Text('Tinjau ulang data sebelum finalisasi.',
                              style: GoogleFonts.outfit(
                                  fontSize: 13, color: Colors.blueGrey)),
                        ],
                      ),
                    ),
                  ],
                ),
                const SizedBox(height: 24),
                const Divider(),
                const SizedBox(height: 24),
                _field('Nomor Nota Digital', 'Contoh: DIG-123', _noNotaCtrl,
                    Icons.confirmation_number_rounded),
                const SizedBox(height: 16),
                _field('Total Transaksi (Rp)', 'Total yang disepakati',
                    _totalCtrl, Icons.payments_rounded,
                    keyboard: TextInputType.number),
                const SizedBox(height: 16),
                _field('Catatan Final', 'Tambahkan instruksi jika ada',
                    _catatanCtrl, Icons.notes_rounded,
                    maxLines: 3),
              ],
            ),
          ),
          const SizedBox(height: 40),
          SizedBox(
            width: double.infinity,
            height: 64,
            child: ElevatedButton(
              onPressed: _isSubmitting ? null : _submitToSystem,
              style: ElevatedButton.styleFrom(
                backgroundColor: AppColors.primaryRed,
                shape: RoundedRectangleBorder(
                    borderRadius: BorderRadius.circular(20)),
                elevation: 12,
                shadowColor: AppColors.primaryRed.withOpacity(0.4),
              ),
              child: _isSubmitting
                  ? const CircularProgressIndicator(color: Colors.white)
                  : Text('KONFIRMASI & SUBMIT',
                      style: GoogleFonts.outfit(
                          fontWeight: FontWeight.w900,
                          fontSize: 16,
                          color: Colors.white)),
            ),
          ),
        ]),
      ),
    );
  }

  Widget _field(
      String label, String hint, TextEditingController ctrl, IconData icon,
      {TextInputType keyboard = TextInputType.text, int maxLines = 1}) {
    return TextFormField(
      controller: ctrl,
      keyboardType: keyboard,
      maxLines: maxLines,
      style: GoogleFonts.outfit(
          fontSize: 15,
          fontWeight: FontWeight.w600,
          color: const Color(0xFF1E293B)),
      decoration: InputDecoration(
        labelText: label,
        hintText: hint,
        prefixIcon: Icon(icon, color: Colors.blueAccent, size: 20),
        labelStyle: GoogleFonts.outfit(
            fontWeight: FontWeight.w700, color: Colors.blueGrey, fontSize: 13),
        hintStyle:
            GoogleFonts.outfit(color: Colors.grey.shade400, fontSize: 13),
        border: OutlineInputBorder(
            borderRadius: BorderRadius.circular(14),
            borderSide: BorderSide(color: Colors.grey.shade200)),
        enabledBorder: OutlineInputBorder(
            borderRadius: BorderRadius.circular(14),
            borderSide: BorderSide(color: Colors.grey.shade200)),
        focusedBorder: OutlineInputBorder(
            borderRadius: BorderRadius.circular(14),
            borderSide: const BorderSide(color: Colors.blueAccent, width: 2)),
        filled: true,
        fillColor: const Color(0xFFF8FAFC),
      ),
    );
  }
}
