import 'dart:convert';
import 'package:flutter/material.dart';
import 'package:flutter_bloc/flutter_bloc.dart';
import 'package:google_fonts/google_fonts.dart';
import 'package:intl/intl.dart';
import 'package:signature/signature.dart';
import 'package:hris_app/core/theme/app_colors.dart';
import 'package:hris_app/features/auth/presentation/bloc/auth_bloc.dart';
import 'package:hris_app/features/auth/presentation/bloc/auth_state.dart';
import 'package:hris_app/features/epkwt/presentation/bloc/epkwt_bloc.dart';
import 'package:hris_app/features/epkwt/presentation/bloc/epkwt_event.dart';
import 'package:hris_app/features/epkwt/presentation/bloc/epkwt_state.dart';

class EpkwtRequestPage extends StatefulWidget {
  const EpkwtRequestPage({super.key});

  @override
  State<EpkwtRequestPage> createState() => _EpkwtRequestPageState();
}

class _EpkwtRequestPageState extends State<EpkwtRequestPage> {
  final _formKey = GlobalKey<FormState>();

  final _identityNumberController = TextEditingController();
  final _firstNameController = TextEditingController();
  final _lastNameController = TextEditingController();
  final _phoneNumberController = TextEditingController();
  final _birthPlaceController = TextEditingController();
  final _ktpStreetController = TextEditingController();
  final _ktpRtRwController = TextEditingController();
  final _ktpVillageController = TextEditingController();
  final _ktpDistrictController = TextEditingController();
  final _ktpCityController = TextEditingController();
  final _ktpProvinceController = TextEditingController();

  final _resStreetController = TextEditingController();
  final _resRtRwController = TextEditingController();
  final _resVillageController = TextEditingController();
  final _resDistrictController = TextEditingController();
  final _resCityController = TextEditingController();
  final _resProvinceController = TextEditingController();

  final SignatureController _signatureController = SignatureController(
    penStrokeWidth: 3,
    penColor: Colors.black,
    exportBackgroundColor: Colors.transparent,
  );

  bool _isResidentialSameAsKtp = true;

  String? _selectedGender;
  String? _selectedMaritalStatus;
  DateTime? _selectedBirthDate;

  final List<String> _genders = ['Laki-laki', 'Perempuan'];
  final List<String> _maritalStatuses = ['Lajang', 'Menikah', 'Duda/Janda'];

  @override
  void initState() {
    super.initState();
    _prepopulateFromProfile();
  }

  void _prepopulateFromProfile() {
    final authState = context.read<AuthBloc>().state;
    if (authState is Authenticated) {
      final profile = authState.userProfile;
      if (profile != null) {
        _identityNumberController.text = profile['identity_number'] ?? '';
        _firstNameController.text = profile['first_name'] ?? '';
        _lastNameController.text = profile['last_name'] ?? '';
        _phoneNumberController.text = profile['phone_number'] ?? '';
        _birthPlaceController.text = profile['birth_place'] ?? '';
        _ktpStreetController.text = profile['address_ktp'] ?? '';
        _resStreetController.text = profile['address_residential'] ?? '';

        final genderStr = profile['gender']?.toString();
        if (genderStr != null && _genders.contains(genderStr)) {
          _selectedGender = genderStr;
        }

        final maritalStr = profile['marital_status']?.toString();
        if (maritalStr != null && _maritalStatuses.contains(maritalStr)) {
          _selectedMaritalStatus = maritalStr;
        }

        final bdateStr = profile['birth_date']?.toString();
        if (bdateStr != null) {
          try {
            _selectedBirthDate = DateTime.parse(bdateStr);
          } catch (_) {}
        }
      }
    }
  }

  @override
  void dispose() {
    _identityNumberController.dispose();
    _firstNameController.dispose();
    _lastNameController.dispose();
    _phoneNumberController.dispose();
    _birthPlaceController.dispose();
    _ktpStreetController.dispose();
    _ktpRtRwController.dispose();
    _ktpVillageController.dispose();
    _ktpDistrictController.dispose();
    _ktpCityController.dispose();
    _ktpProvinceController.dispose();
    _resStreetController.dispose();
    _resRtRwController.dispose();
    _resVillageController.dispose();
    _resDistrictController.dispose();
    _resCityController.dispose();
    _resProvinceController.dispose();
    _signatureController.dispose();
    super.dispose();
  }

  Future<void> _selectBirthDate() async {
    final picked = await showDatePicker(
      context: context,
      initialDate: _selectedBirthDate ?? DateTime(2000, 1, 1),
      firstDate: DateTime(1950),
      lastDate: DateTime.now(),
      builder: (context, child) {
        return Theme(
          data: Theme.of(context).copyWith(
            colorScheme: const ColorScheme.light(
              primary: AppColors.primaryGreen,
              onPrimary: Colors.white,
              onSurface: AppColors.textPrimary,
            ),
          ),
          child: child!,
        );
      },
    );

    if (picked != null) {
      setState(() {
        _selectedBirthDate = picked;
      });
    }
  }

  String _combineAddress(
    TextEditingController street,
    TextEditingController rtrw,
    TextEditingController village,
    TextEditingController district,
    TextEditingController city,
    TextEditingController province,
  ) {
    final parts = <String>[];
    if (street.text.isNotEmpty) parts.add(street.text);
    if (rtrw.text.isNotEmpty) parts.add('RT/RW ${rtrw.text}');
    if (village.text.isNotEmpty) parts.add('Kel. ${village.text}');
    if (district.text.isNotEmpty) parts.add('Kec. ${district.text}');
    if (city.text.isNotEmpty) parts.add(city.text);
    if (province.text.isNotEmpty) parts.add('Prov. ${province.text}');
    return parts.join(', ');
  }

  Future<void> _submitForm() async {
    if (_signatureController.isEmpty) {
      ScaffoldMessenger.of(context).showSnackBar(
        const SnackBar(content: Text('Tanda tangan wajib diisi')),
      );
      return;
    }

    if (_formKey.currentState!.validate()) {
      final sigBytes = await _signatureController.toPngBytes();
      final String? signatureBase64 =
          sigBytes != null ? base64Encode(sigBytes) : null;

      final ktpAddr = _combineAddress(
        _ktpStreetController,
        _ktpRtRwController,
        _ktpVillageController,
        _ktpDistrictController,
        _ktpCityController,
        _ktpProvinceController,
      );

      final resAddr = _isResidentialSameAsKtp
          ? ktpAddr
          : _combineAddress(
              _resStreetController,
              _resRtRwController,
              _resVillageController,
              _resDistrictController,
              _resCityController,
              _resProvinceController,
            );

      context.read<EpkwtBloc>().add(
            SubmitEpkwtRequested(
              identityNumber: _identityNumberController.text,
              firstName: _firstNameController.text,
              lastName: _lastNameController.text.isNotEmpty
                  ? _lastNameController.text
                  : null,
              phoneNumber: _phoneNumberController.text.isNotEmpty
                  ? _phoneNumberController.text
                  : null,
              gender: _selectedGender,
              birthDate: _selectedBirthDate,
              birthPlace: _birthPlaceController.text.isNotEmpty
                  ? _birthPlaceController.text
                  : null,
              addressKtp: ktpAddr.isNotEmpty ? ktpAddr : null,
              addressResidential: resAddr.isNotEmpty ? resAddr : null,
              maritalStatus: _selectedMaritalStatus,
              signatureData: signatureBase64,
            ),
          );
    }
  }

  @override
  Widget build(BuildContext context) {
    return Scaffold(
      backgroundColor: AppColors.backgroundAlt,
      appBar: AppBar(
        title: Text(
          'Ajukan Pembuatan PKWT',
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
      body: BlocListener<EpkwtBloc, EpkwtState>(
        listener: (context, state) {
          if (state is EpkwtSuccess) {
            ScaffoldMessenger.of(context).showSnackBar(
              SnackBar(content: Text(state.message)),
            );
            Navigator.pop(context);
          }
          if (state is EpkwtFailure) {
            ScaffoldMessenger.of(context).showSnackBar(
              SnackBar(content: Text(state.message)),
            );
          }
        },
        child: SingleChildScrollView(
          padding: const EdgeInsets.all(24),
          child: Form(
            key: _formKey,
            child: Column(
              crossAxisAlignment: CrossAxisAlignment.start,
              children: [
                Text(
                  'Data Personal Kontrak',
                  style: GoogleFonts.plusJakartaSans(
                    fontWeight: FontWeight.w900,
                    fontSize: 16,
                    color: AppColors.textPrimary,
                  ),
                ),
                const SizedBox(height: 4),
                Text(
                  'Lengkapi data Anda secara mandiri untuk proses cetak draf PKWT.',
                  style: GoogleFonts.plusJakartaSans(
                    fontSize: 12,
                    color: AppColors.textSecondary,
                    height: 1.4,
                  ),
                ),
                const SizedBox(height: 24),
                _buildCardContainer([
                  _buildTextField(
                    controller: _identityNumberController,
                    label: 'No. KTP / NIK',
                    validator: (val) =>
                        val == null || val.isEmpty ? 'NIK wajib diisi' : null,
                    keyboardType: TextInputType.number,
                  ),
                  const SizedBox(height: 16),
                  _buildTextField(
                    controller: _firstNameController,
                    label: 'Nama Depan',
                    validator: (val) => val == null || val.isEmpty
                        ? 'Nama depan wajib diisi'
                        : null,
                  ),
                  const SizedBox(height: 16),
                  _buildTextField(
                    controller: _lastNameController,
                    label: 'Nama Belakang',
                  ),
                ]),
                const SizedBox(height: 20),
                _buildCardContainer([
                  _buildDropdownField(
                    label: 'Jenis Kelamin',
                    value: _selectedGender,
                    items: _genders,
                    onChanged: (val) => setState(() => _selectedGender = val),
                    validator: (val) =>
                        val == null ? 'Jenis kelamin wajib dipilih' : null,
                  ),
                  const SizedBox(height: 16),
                  _buildDatePickerField(),
                  const SizedBox(height: 16),
                  _buildTextField(
                    controller: _birthPlaceController,
                    label: 'Tempat Lahir',
                  ),
                  const SizedBox(height: 16),
                  _buildDropdownField(
                    label: 'Status Pernikaran',
                    value: _selectedMaritalStatus,
                    items: _maritalStatuses,
                    onChanged: (val) =>
                        setState(() => _selectedMaritalStatus = val),
                  ),
                ]),
                const SizedBox(height: 20),
                _buildCardContainer([
                  _buildTextField(
                    controller: _phoneNumberController,
                    label: 'No. Telepon / WhatsApp',
                    keyboardType: TextInputType.phone,
                  ),
                  const SizedBox(height: 16),
                  Text(
                    'Alamat (Sesuai KTP)',
                    style: GoogleFonts.plusJakartaSans(
                      fontWeight: FontWeight.bold,
                      fontSize: 14,
                      color: AppColors.textPrimary,
                    ),
                  ),
                  const SizedBox(height: 12),
                  _buildTextField(
                    controller: _ktpStreetController,
                    label: 'Nama/Nomor Jalan',
                    validator: (val) => val == null || val.isEmpty ? 'Jalan wajib diisi' : null,
                  ),
                  const SizedBox(height: 12),
                  Row(
                    children: [
                      Expanded(
                        child: _buildTextField(
                          controller: _ktpRtRwController,
                          label: 'RT/RW',
                        ),
                      ),
                      const SizedBox(width: 12),
                      Expanded(
                        child: _buildTextField(
                          controller: _ktpVillageController,
                          label: 'Kelurahan',
                        ),
                      ),
                    ],
                  ),
                  const SizedBox(height: 12),
                  Row(
                    children: [
                      Expanded(
                        child: _buildTextField(
                          controller: _ktpDistrictController,
                          label: 'Kecamatan',
                        ),
                      ),
                      const SizedBox(width: 12),
                      Expanded(
                        child: _buildTextField(
                          controller: _ktpCityController,
                          label: 'Kabupaten/Kota',
                        ),
                      ),
                    ],
                  ),
                  const SizedBox(height: 12),
                  _buildTextField(
                    controller: _ktpProvinceController,
                    label: 'Provinsi',
                  ),
                  const SizedBox(height: 16),
                  Row(
                    children: [
                      Checkbox(
                        value: _isResidentialSameAsKtp,
                        activeColor: AppColors.primaryGreen,
                        onChanged: (val) {
                          setState(() {
                            _isResidentialSameAsKtp = val ?? true;
                          });
                        },
                      ),
                      Expanded(
                        child: Text(
                          'Alamat Domisili sama dengan KTP',
                          style: GoogleFonts.plusJakartaSans(
                            fontSize: 13,
                            fontWeight: FontWeight.w600,
                            color: AppColors.textSecondary,
                          ),
                        ),
                      ),
                    ],
                  ),
                  if (!_isResidentialSameAsKtp) ...[
                    const SizedBox(height: 12),
                    Text(
                      'Alamat Domisili',
                      style: GoogleFonts.plusJakartaSans(
                        fontWeight: FontWeight.bold,
                        fontSize: 14,
                        color: AppColors.textPrimary,
                      ),
                    ),
                    const SizedBox(height: 12),
                    _buildTextField(
                      controller: _resStreetController,
                      label: 'Nama/Nomor Jalan',
                    ),
                    const SizedBox(height: 12),
                    Row(
                      children: [
                        Expanded(
                          child: _buildTextField(
                            controller: _resRtRwController,
                            label: 'RT/RW',
                          ),
                        ),
                        const SizedBox(width: 12),
                        Expanded(
                          child: _buildTextField(
                            controller: _resVillageController,
                            label: 'Kelurahan',
                          ),
                        ),
                      ],
                    ),
                    const SizedBox(height: 12),
                    Row(
                      children: [
                        Expanded(
                          child: _buildTextField(
                            controller: _resDistrictController,
                            label: 'Kecamatan',
                          ),
                        ),
                        const SizedBox(width: 12),
                        Expanded(
                          child: _buildTextField(
                            controller: _resCityController,
                            label: 'Kabupaten/Kota',
                          ),
                        ),
                      ],
                    ),
                    const SizedBox(height: 12),
                    _buildTextField(
                      controller: _resProvinceController,
                      label: 'Provinsi',
                    ),
                  ],
                ]),
                const SizedBox(height: 20),
                _buildCardContainer([
                  Text(
                    'Tanda Tangan',
                    style: GoogleFonts.plusJakartaSans(
                      fontWeight: FontWeight.bold,
                      fontSize: 14,
                      color: AppColors.textPrimary,
                    ),
                  ),
                  const SizedBox(height: 12),
                  Container(
                    decoration: BoxDecoration(
                      border: Border.all(color: Colors.grey.shade300),
                      borderRadius: BorderRadius.circular(12),
                      color: Colors.grey.shade50,
                    ),
                    clipBehavior: Clip.hardEdge,
                    child: Signature(
                      controller: _signatureController,
                      height: 150,
                      backgroundColor: Colors.transparent,
                    ),
                  ),
                  const SizedBox(height: 8),
                  Align(
                    alignment: Alignment.centerRight,
                    child: TextButton.icon(
                      onPressed: () => _signatureController.clear(),
                      icon: const Icon(Icons.clear, size: 16, color: Colors.red),
                      label: Text(
                        'Hapus',
                        style: GoogleFonts.plusJakartaSans(color: Colors.red),
                      ),
                    ),
                  ),
                ]),
                const SizedBox(height: 32),
                SizedBox(
                  width: double.infinity,
                  height: 52,
                  child: BlocBuilder<EpkwtBloc, EpkwtState>(
                    builder: (context, state) {
                      final isLoading = state is EpkwtLoading;
                      return ElevatedButton(
                        onPressed: isLoading ? null : _submitForm,
                        style: ElevatedButton.styleFrom(
                          backgroundColor: AppColors.primaryGreen,
                          shape: RoundedRectangleBorder(
                            borderRadius: BorderRadius.circular(16),
                          ),
                          elevation: 2,
                        ),
                        child: isLoading
                            ? const CircularProgressIndicator(
                                color: Colors.white)
                            : Text(
                                'KIRIM PENGAJUAN',
                                style: GoogleFonts.plusJakartaSans(
                                  fontWeight: FontWeight.w800,
                                  color: Colors.white,
                                  fontSize: 13,
                                  letterSpacing: 0.5,
                                ),
                              ),
                      );
                    },
                  ),
                ),
              ],
            ),
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
    TextInputType? keyboardType,
    int maxLines = 1,
  }) {
    return TextFormField(
      controller: controller,
      validator: validator,
      keyboardType: keyboardType,
      maxLines: maxLines,
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
    required String? value,
    required List<String> items,
    required Function(String?) onChanged,
    String? Function(String?)? validator,
  }) {
    return DropdownButtonFormField<String>(
      initialValue: value,
      items: items
          .map((e) => DropdownMenuItem(
                value: e,
                child:
                    Text(e, style: GoogleFonts.plusJakartaSans(fontSize: 14)),
              ))
          .toList(),
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

  Widget _buildDatePickerField() {
    final displayStr = _selectedBirthDate == null
        ? 'Pilih Tanggal Lahir'
        : DateFormat('dd MMMM yyyy', 'id_ID').format(_selectedBirthDate!);

    return InkWell(
      onTap: _selectBirthDate,
      borderRadius: BorderRadius.circular(12),
      child: Container(
        padding: const EdgeInsets.symmetric(horizontal: 16, vertical: 14),
        decoration: BoxDecoration(
          color: Colors.grey.shade50,
          borderRadius: BorderRadius.circular(12),
        ),
        child: Row(
          mainAxisAlignment: MainAxisAlignment.spaceBetween,
          children: [
            Column(
              crossAxisAlignment: CrossAxisAlignment.start,
              children: [
                Text(
                  'Tanggal Lahir',
                  style: GoogleFonts.plusJakartaSans(
                    color: AppColors.textTertiary,
                    fontWeight: FontWeight.w600,
                    fontSize: 11,
                  ),
                ),
                const SizedBox(height: 4),
                Text(
                  displayStr,
                  style: GoogleFonts.plusJakartaSans(
                    fontWeight: FontWeight.bold,
                    fontSize: 14,
                    color: _selectedBirthDate == null
                        ? AppColors.textTertiary
                        : AppColors.textPrimary,
                  ),
                ),
              ],
            ),
            const Icon(Icons.calendar_month_rounded,
                color: AppColors.textTertiary),
          ],
        ),
      ),
    );
  }
}
