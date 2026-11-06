import 'package:flutter/material.dart';
import 'package:google_fonts/google_fonts.dart';
import 'package:hris_app/core/network/api_client.dart';
import 'package:hris_app/core/theme/app_colors.dart';
import 'package:hris_app/injection.dart' as di;

class MemoRequestPage extends StatefulWidget {
  const MemoRequestPage({super.key});

  @override
  State<MemoRequestPage> createState() => _MemoRequestPageState();
}

class _MemoRequestPageState extends State<MemoRequestPage> {
  final _formKey = GlobalKey<FormState>();
  final _apiClient = di.sl<ApiClient>();

  final _destinationController = TextEditingController();
  final _subjectController = TextEditingController();
  final _requesterNameController = TextEditingController();

  String? _selectedType;
  String? _selectedBranchId;

  bool _isLoadingBranches = true;
  bool _isSubmitting = false;

  List<dynamic> _branches = [];
  final List<String> _memoTypes = ['INTERNAL', 'EXTERNAL'];

  @override
  void initState() {
    super.initState();
    _fetchBranches();
  }

  Future<void> _fetchBranches() async {
    try {
      final res = await _apiClient.client.get('/branches');
      if (res.statusCode == 200) {
        setState(() {
          _branches = res.data['data'] ?? [];
          _isLoadingBranches = false;
        });
      }
    } catch (e) {
      setState(() => _isLoadingBranches = false);
      // Optional: show error toast
    }
  }

  Future<void> _submitRequest() async {
    if (!_formKey.currentState!.validate()) return;

    setState(() => _isSubmitting = true);
    try {
      final payload = {
        'type': _selectedType,
        'branch_id': _selectedBranchId,
        'destination': _destinationController.text,
        'subject': _subjectController.text,
        'requester_name': _requesterNameController.text,
      };

      final res = await _apiClient.client.post('/memo-requests', data: payload);
      if (res.statusCode == 201) {
        ScaffoldMessenger.of(context).showSnackBar(
          const SnackBar(content: Text('Request memo berhasil dikirim!')),
        );
        Navigator.pop(context);
      }
    } catch (e) {
      ScaffoldMessenger.of(context).showSnackBar(
        const SnackBar(content: Text('Gagal mengirim request')),
      );
    } finally {
      if (mounted) setState(() => _isSubmitting = false);
    }
  }

  @override
  void dispose() {
    _destinationController.dispose();
    _subjectController.dispose();
    _requesterNameController.dispose();
    super.dispose();
  }

  @override
  Widget build(BuildContext context) {
    return Scaffold(
      backgroundColor: AppColors.backgroundAlt,
      appBar: AppBar(
        title: Text(
          'Request Nomor Memo',
          style: GoogleFonts.plusJakartaSans(
            fontWeight: FontWeight.w900,
            fontSize: 16,
            color: Colors.white,
          ),
        ),
        backgroundColor: AppColors.primaryGreen,
        iconTheme: const IconThemeData(color: Colors.white),
        elevation: 0,
      ),
      body: SingleChildScrollView(
        padding: const EdgeInsets.all(24),
        child: Form(
          key: _formKey,
          child: Column(
            crossAxisAlignment: CrossAxisAlignment.start,
            children: [
              Text(
                'Data Request Surat/Memo',
                style: GoogleFonts.plusJakartaSans(
                  fontWeight: FontWeight.w900,
                  fontSize: 16,
                  color: AppColors.textPrimary,
                ),
              ),
              const SizedBox(height: 4),
              Text(
                'Isi form untuk mengajukan request penomoran surat secara otomatis.',
                style: GoogleFonts.plusJakartaSans(
                  fontSize: 12,
                  color: AppColors.textSecondary,
                  height: 1.4,
                ),
              ),
              const SizedBox(height: 24),
              _buildCardContainer([
                _buildDropdownField(
                  label: 'Tipe Memo',
                  value: _selectedType,
                  items: _memoTypes
                      .map((e) => DropdownMenuItem(value: e, child: Text(e)))
                      .toList(),
                  onChanged: (val) => setState(() => _selectedType = val),
                  validator: (val) => val == null ? 'Tipe wajib dipilih' : null,
                ),
                const SizedBox(height: 16),
                if (_isLoadingBranches)
                  const Center(child: CircularProgressIndicator())
                else
                  _buildDropdownField(
                    label: 'Cabang Asal',
                    value: _selectedBranchId,
                    items: [
                      const DropdownMenuItem(
                          value: null,
                          child: Text('Pusat / Umum (Tanpa Cabang)')),
                      ..._branches.map((b) => DropdownMenuItem(
                            value: b['id'],
                            child: Text(b['name'] ?? ''),
                          )),
                    ],
                    onChanged: (val) => setState(() => _selectedBranchId = val),
                    validator: (val) => null, // Optional
                  ),
                const SizedBox(height: 16),
                _buildTextField(
                  controller: _destinationController,
                  label: 'Tujuan Surat',
                  validator: (val) =>
                      val == null || val.isEmpty ? 'Tujuan wajib diisi' : null,
                ),
                const SizedBox(height: 16),
                _buildTextField(
                  controller: _subjectController,
                  label: 'Perihal Surat',
                  validator: (val) =>
                      val == null || val.isEmpty ? 'Perihal wajib diisi' : null,
                ),
                const SizedBox(height: 16),
                _buildTextField(
                  controller: _requesterNameController,
                  label: 'Nama Pemohon (Karyawan)',
                  validator: (val) => val == null || val.isEmpty
                      ? 'Nama pemohon wajib diisi'
                      : null,
                ),
              ]),
              const SizedBox(height: 32),
              SizedBox(
                width: double.infinity,
                height: 52,
                child: ElevatedButton(
                  onPressed: _isSubmitting ? null : _submitRequest,
                  style: ElevatedButton.styleFrom(
                    backgroundColor: AppColors.primaryGreen,
                    shape: RoundedRectangleBorder(
                      borderRadius: BorderRadius.circular(16),
                    ),
                    elevation: 2,
                  ),
                  child: _isSubmitting
                      ? const CircularProgressIndicator(color: Colors.white)
                      : Text(
                          'KIRIM REQUEST',
                          style: GoogleFonts.plusJakartaSans(
                            fontWeight: FontWeight.w800,
                            color: Colors.white,
                            fontSize: 13,
                            letterSpacing: 0.5,
                          ),
                        ),
                ),
              ),
            ],
          ),
        ),
      ),
    );
  }

  Widget _buildCardContainer(List<Widget> children) {
    return Container(
      padding: const EdgeInsets.all(20),
      decoration: BoxDecoration(
        color: Colors.white,
        borderRadius: BorderRadius.circular(20),
        boxShadow: [
          BoxShadow(
            color: Colors.black.withOpacity(0.02),
            blurRadius: 10,
            offset: const Offset(0, 5),
          )
        ],
      ),
      child: Column(
        crossAxisAlignment: CrossAxisAlignment.start,
        children: children,
      ),
    );
  }

  Widget _buildTextField({
    required TextEditingController controller,
    required String label,
    String? Function(String?)? validator,
  }) {
    return TextFormField(
      controller: controller,
      validator: validator,
      style: GoogleFonts.plusJakartaSans(
        fontWeight: FontWeight.bold,
        fontSize: 14,
        color: AppColors.textPrimary,
      ),
      decoration: InputDecoration(
        labelText: label,
        labelStyle: GoogleFonts.plusJakartaSans(
          color: AppColors.textTertiary,
          fontWeight: FontWeight.w600,
          fontSize: 12,
        ),
        filled: true,
        fillColor: Colors.grey.shade50,
        contentPadding:
            const EdgeInsets.symmetric(horizontal: 16, vertical: 14),
        border: OutlineInputBorder(
          borderRadius: BorderRadius.circular(12),
          borderSide: BorderSide.none,
        ),
        enabledBorder: OutlineInputBorder(
          borderRadius: BorderRadius.circular(12),
          borderSide: BorderSide.none,
        ),
        focusedBorder: OutlineInputBorder(
          borderRadius: BorderRadius.circular(12),
          borderSide:
              const BorderSide(color: AppColors.primaryGreen, width: 1.5),
        ),
      ),
    );
  }

  Widget _buildDropdownField({
    required String label,
    required dynamic value,
    required List<DropdownMenuItem<dynamic>> items,
    required Function(dynamic) onChanged,
    String? Function(dynamic)? validator,
  }) {
    return DropdownButtonFormField<dynamic>(
      value: value,
      items: items,
      onChanged: onChanged,
      validator: validator,
      style: GoogleFonts.plusJakartaSans(
        fontWeight: FontWeight.bold,
        fontSize: 14,
        color: AppColors.textPrimary,
      ),
      decoration: InputDecoration(
        labelText: label,
        labelStyle: GoogleFonts.plusJakartaSans(
          color: AppColors.textTertiary,
          fontWeight: FontWeight.w600,
          fontSize: 12,
        ),
        filled: true,
        fillColor: Colors.grey.shade50,
        contentPadding: const EdgeInsets.symmetric(horizontal: 16),
        border: OutlineInputBorder(
          borderRadius: BorderRadius.circular(12),
          borderSide: BorderSide.none,
        ),
        enabledBorder: OutlineInputBorder(
          borderRadius: BorderRadius.circular(12),
          borderSide: BorderSide.none,
        ),
      ),
    );
  }
}
