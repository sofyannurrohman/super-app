import 'dart:typed_data';
import 'package:flutter/material.dart';
import 'package:google_fonts/google_fonts.dart';
import 'package:hris_app/core/network/api_client.dart';
import 'package:hris_app/features/sales/data/models/store_model.dart';
import 'package:hris_app/injection.dart' as di;
import 'package:hris_app/core/database/database.dart';
import './order_entry_page.dart';

class SelectCompanyPage extends StatefulWidget {
  final StoreModel store;
  final String selfiePath;
  final Uint8List? selfieBytes;
  final String jobPositionTitle;

  const SelectCompanyPage({
    super.key,
    required this.store,
    required this.selfiePath,
    this.selfieBytes,
    required this.jobPositionTitle,
  });

  @override
  State<SelectCompanyPage> createState() => _SelectCompanyPageState();
}

class _SelectCompanyPageState extends State<SelectCompanyPage> {
  final ApiClient apiClient = di.sl<ApiClient>();
  List<dynamic> _companies = [];
  bool _isLoading = false;

  @override
  void initState() {
    super.initState();
    _fetchCompanies();
  }

  Future<void> _fetchCompanies() async {
    setState(() => _isLoading = true);
    try {
      final db = di.sl<AppDatabase>();
      final companies = await db.select(db.companies).get();
      setState(() {
        _companies = companies
            .map((c) => {
                  'id': c.id,
                  'name': c.name,
                  'code': c.code,
                })
            .toList();
      });
    } catch (e) {
      if (mounted) {
        ScaffoldMessenger.of(context).showSnackBar(
            SnackBar(content: Text('Gagal memuat entitas lokal: $e')));
      }
    } finally {
      if (mounted) setState(() => _isLoading = false);
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
            Text('LANGKAH 3 DARI 5',
                style: GoogleFonts.outfit(
                    fontSize: 10,
                    color: Colors.blueAccent,
                    fontWeight: FontWeight.w800,
                    letterSpacing: 1)),
            Text('Pilih Entitas Transaksi',
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
                  width: double.infinity,
                  padding: const EdgeInsets.all(24),
                  color: Colors.white,
                  child: Text(
                    'Untuk perusahaan mana transaksi ini dibuat?',
                    style: GoogleFonts.outfit(
                        fontSize: 18,
                        fontWeight: FontWeight.w900,
                        color: const Color(0xFF1E293B)),
                  ),
                ),
                Expanded(
                  child: ListView.builder(
                    padding: const EdgeInsets.all(20),
                    itemCount: _companies.length,
                    itemBuilder: (context, index) {
                      final comp = _companies[index];
                      return Container(
                        margin: const EdgeInsets.only(bottom: 16),
                        decoration: BoxDecoration(
                          color: Colors.white,
                          borderRadius: BorderRadius.circular(20),
                          boxShadow: [
                            BoxShadow(
                                color: Colors.black.withOpacity(0.02),
                                blurRadius: 10,
                                offset: const Offset(0, 4))
                          ],
                        ),
                        child: ListTile(
                          contentPadding: const EdgeInsets.all(20),
                          leading: Container(
                            padding: const EdgeInsets.all(12),
                            decoration: BoxDecoration(
                                color: Colors.blueAccent.withOpacity(0.1),
                                shape: BoxShape.circle),
                            child: const Icon(Icons.business_rounded,
                                color: Colors.blueAccent),
                          ),
                          title: Text(comp['name'],
                              style: GoogleFonts.outfit(
                                  fontWeight: FontWeight.w900, fontSize: 16)),
                          subtitle: Text(
                              'ID Perusahaan: ${comp['code'] ?? comp['id'].toString().substring(0, 8)}',
                              style: GoogleFonts.outfit(
                                  fontSize: 12, color: Colors.grey)),
                          trailing: const Icon(Icons.arrow_forward_ios_rounded,
                              size: 16, color: Colors.grey),
                          onTap: () => Navigator.push(
                            context,
                            MaterialPageRoute(
                              builder: (_) => OrderEntryPage(
                                store: widget.store,
                                selfiePath: widget.selfiePath,
                                selfieBytes: widget.selfieBytes,
                                companyId: comp['id'],
                                companyName: comp['name'],
                                jobPositionTitle: widget.jobPositionTitle,
                              ),
                            ),
                          ),
                        ),
                      );
                    },
                  ),
                ),
              ],
            ),
    );
  }
}
