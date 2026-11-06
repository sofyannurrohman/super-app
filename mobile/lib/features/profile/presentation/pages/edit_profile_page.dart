import 'package:flutter/material.dart';
import 'package:flutter_bloc/flutter_bloc.dart';
import 'package:google_fonts/google_fonts.dart';
import 'package:image_picker/image_picker.dart';
import 'package:intl/intl.dart';
import 'package:hris_app/features/profile/presentation/bloc/profile_bloc.dart';
import 'package:hris_app/features/profile/presentation/bloc/profile_event.dart';
import 'package:hris_app/features/profile/presentation/bloc/profile_state.dart';
import 'package:hris_app/core/utils/constants.dart';
import 'package:hris_app/core/utils/snackbar_utils.dart';
import 'package:hris_app/core/theme/app_colors.dart';

class EditProfilePage extends StatefulWidget {
  const EditProfilePage({super.key});

  @override
  State<EditProfilePage> createState() => _EditProfilePageState();
}

class _EditProfilePageState extends State<EditProfilePage> {
  final _firstNameController = TextEditingController();
  final _lastNameController = TextEditingController();
  final _birthDateController = TextEditingController();
  final _birthPlaceController = TextEditingController();
  final _phoneController = TextEditingController();
  final _addressKtpController = TextEditingController();
  final _addressResidentialController = TextEditingController();
  final _emergencyContactController = TextEditingController();
  final _identityNumberController = TextEditingController();
  final _npwpNumberController = TextEditingController();
  final _bankNameController = TextEditingController();
  final _bankAccountNumberController = TextEditingController();
  final _accountHolderNameController = TextEditingController();
  final ImagePicker _picker = ImagePicker();

  String? _gender;
  String? _maritalStatus;

  @override
  void initState() {
    super.initState();
    context.read<ProfileBloc>().add(LoadProfileRequested());
  }

  Future<void> _pickAndUploadAvatar(ImageSource source) async {
    try {
      final XFile? pickedFile = await _picker.pickImage(
        source: source,
        maxWidth: 1024,
        maxHeight: 1024,
        imageQuality: 85,
      );

      if (pickedFile != null && mounted) {
        context.read<ProfileBloc>().add(UploadAvatarRequested(pickedFile.path));
      }
    } catch (e) {
      if (mounted) {
        SnackBarUtils.showError(context, 'Gagal memilih gambar: $e');
      }
    }
  }

  void _showImageSourceDialog(BuildContext context) {
    showModalBottomSheet(
      context: context,
      backgroundColor: Colors.white,
      shape: const RoundedRectangleBorder(
        borderRadius: BorderRadius.vertical(top: Radius.circular(24)),
      ),
      builder: (ctx) {
        return SafeArea(
          child: Padding(
            padding: const EdgeInsets.symmetric(vertical: 20, horizontal: 24),
            child: Column(
              mainAxisSize: MainAxisSize.min,
              crossAxisAlignment: CrossAxisAlignment.start,
              children: [
                Text(
                  'Ubah Foto Profil',
                  style: GoogleFonts.plusJakartaSans(
                    fontSize: 18,
                    fontWeight: FontWeight.w800,
                    color: AppColors.textPrimary,
                  ),
                ),
                const SizedBox(height: 8),
                Text(
                  'Pilih sumber foto yang ingin Anda gunakan',
                  style: GoogleFonts.plusJakartaSans(
                    fontSize: 13,
                    color: AppColors.textSecondary,
                  ),
                ),
                const SizedBox(height: 20),
                ListTile(
                  leading: Container(
                    padding: const EdgeInsets.all(10),
                    decoration: BoxDecoration(
                      color: AppColors.primaryRed.withOpacity(0.1),
                      borderRadius: BorderRadius.circular(12),
                    ),
                    child: const Icon(Icons.photo_library_rounded,
                        color: AppColors.primaryRed),
                  ),
                  title: Text(
                    'Pilih dari Galeri',
                    style: GoogleFonts.plusJakartaSans(
                      fontWeight: FontWeight.w700,
                      color: AppColors.textPrimary,
                    ),
                  ),
                  subtitle: const Text('Gunakan foto yang sudah ada di HP'),
                  onTap: () {
                    Navigator.pop(ctx);
                    _pickAndUploadAvatar(ImageSource.gallery);
                  },
                ),
                const SizedBox(height: 8),
                ListTile(
                  leading: Container(
                    padding: const EdgeInsets.all(10),
                    decoration: BoxDecoration(
                      color: Colors.blue.withOpacity(0.1),
                      borderRadius: BorderRadius.circular(12),
                    ),
                    child: const Icon(Icons.camera_alt_rounded,
                        color: Colors.blue),
                  ),
                  title: Text(
                    'Ambil Foto Kamera',
                    style: GoogleFonts.plusJakartaSans(
                      fontWeight: FontWeight.w700,
                      color: AppColors.textPrimary,
                    ),
                  ),
                  subtitle: const Text('Ambil foto selfie langsung dengan kamera'),
                  onTap: () {
                    Navigator.pop(ctx);
                    _pickAndUploadAvatar(ImageSource.camera);
                  },
                ),
              ],
            ),
          ),
        );
      },
    );
  }

  @override
  void dispose() {
    _firstNameController.dispose();
    _lastNameController.dispose();
    _birthDateController.dispose();
    _birthPlaceController.dispose();
    _phoneController.dispose();
    _addressKtpController.dispose();
    _addressResidentialController.dispose();
    _emergencyContactController.dispose();
    _identityNumberController.dispose();
    _npwpNumberController.dispose();
    _bankNameController.dispose();
    _bankAccountNumberController.dispose();
    _accountHolderNameController.dispose();
    super.dispose();
  }

  @override
  Widget build(BuildContext context) {
    const textColor = AppColors.textPrimary;
    const subtitleColor = AppColors.textSecondary;
    const primaryRed = AppColors.primaryRed;
    const borderColor = AppColors.grayBorder;
    const disabledBgColor = AppColors.grayLight;

    return BlocListener<ProfileBloc, ProfileState>(
      listener: (context, state) {
        if (state is ProfileLoaded) {
          final profile = state.profile;
          _firstNameController.text = profile['first_name'] ?? '';
          _lastNameController.text = profile['last_name'] ?? '';

          if (profile['birth_date'] != null) {
            final bd = profile['birth_date'].toString();
            _birthDateController.text =
                bd.length >= 10 ? bd.substring(0, 10) : bd;
          }

          _birthPlaceController.text = profile['birth_place'] ?? '';
          _gender = profile['gender'];
          _maritalStatus = profile['marital_status'];
          _phoneController.text = profile['phone_number'] ?? '';
          _addressKtpController.text = profile['address_ktp'] ?? '';
          _addressResidentialController.text =
              profile['address_residential'] ?? '';
          _emergencyContactController.text = profile['emergency_contact'] ?? '';
          _identityNumberController.text = profile['identity_number'] ?? '';
          _npwpNumberController.text = profile['npwp_number'] ?? '';
          _bankNameController.text = profile['bank_name'] ?? '';
          _bankAccountNumberController.text =
              profile['bank_account_number'] ?? '';
          _accountHolderNameController.text =
              profile['account_holder_name'] ?? '';
        }
        if (state is ProfileAvatarSuccess) {
          SnackBarUtils.showSuccess(context, 'Foto profil berhasil diperbarui');
          context.read<ProfileBloc>().add(LoadProfileRequested());
        }
        if (state is ProfileUpdateSuccess) {
          SnackBarUtils.showSuccess(context, 'Profile Updated Successfully');
          context.read<ProfileBloc>().add(LoadProfileRequested());
        }
        if (state is ProfileFailure) {
          SnackBarUtils.showError(context, state.message);
        }
      },
      child: Scaffold(
        backgroundColor: Colors.white,
        appBar: AppBar(
          title: Text('Edit Profile',
              style: GoogleFonts.plusJakartaSans(
                  fontSize: 18, fontWeight: FontWeight.bold, color: textColor)),
          centerTitle: true,
          backgroundColor: Colors.white,
          elevation: 0,
          actions: [
            IconButton(
              icon: const Icon(Icons.help_outline_rounded,
                  color: AppColors.textSecondary),
              onPressed: () {},
            ),
          ],
          leading: IconButton(
            icon: const Icon(Icons.arrow_back_ios_new,
                color: AppColors.textPrimary, size: 20),
            onPressed: () => Navigator.of(context).pop(),
          ),
        ),
        body: SafeArea(
          child: BlocBuilder<ProfileBloc, ProfileState>(
            builder: (context, state) {
              if (state is ProfileLoading && state is! ProfileLoaded) {
                return const Center(child: CircularProgressIndicator());
              }

              final profile = (state is ProfileLoaded) ? state.profile : {};
              final nik = profile['employee_id_number'] ?? '-';
              final department = profile['department']?['name'] ?? '-';
              final isUpdatingAvatar = state is ProfileAvatarUpdating;

              return SingleChildScrollView(
                padding:
                    const EdgeInsets.symmetric(horizontal: 20, vertical: 24)
                        .copyWith(bottom: 110),
                child: Column(
                  crossAxisAlignment: CrossAxisAlignment.start,
                  children: [
                    Center(
                      child: Column(
                        children: [
                          GestureDetector(
                            onTap: isUpdatingAvatar
                                ? null
                                : () => _showImageSourceDialog(context),
                            child: Stack(
                              children: [
                                CircleAvatar(
                                  radius: 50,
                                  backgroundColor: AppColors.grayLight,
                                  backgroundImage:
                                      (profile['face_reference_url'] != null &&
                                              profile['face_reference_url']
                                                  .toString()
                                                  .isNotEmpty)
                                          ? NetworkImage(
                                              AppConstants.buildFaceImageUrl(
                                                  profile['face_reference_url']
                                                      .toString()))
                                          : null,
                                  child: isUpdatingAvatar
                                      ? Container(
                                          decoration: const BoxDecoration(
                                            color: Colors.black45,
                                            shape: BoxShape.circle,
                                          ),
                                          child: const Center(
                                            child: CircularProgressIndicator(
                                              color: Colors.white,
                                              strokeWidth: 3,
                                            ),
                                          ),
                                        )
                                      : (profile['face_reference_url'] ==
                                                  null ||
                                              profile['face_reference_url']
                                                  .toString()
                                                  .isEmpty)
                                          ? Text(
                                              (profile['first_name'] ?? 'U')[0]
                                                  .toUpperCase(),
                                              style: const TextStyle(
                                                  fontSize: 32,
                                                  fontWeight: FontWeight.w900,
                                                  color: AppColors.primaryRed))
                                          : null,
                                ),
                                Positioned(
                                  bottom: 0,
                                  right: 0,
                                  child: Container(
                                    padding: const EdgeInsets.all(8),
                                    decoration: BoxDecoration(
                                      color: primaryRed,
                                      shape: BoxShape.circle,
                                      border: Border.all(
                                          color: Colors.white, width: 2),
                                      boxShadow: [
                                        BoxShadow(
                                            color:
                                                Colors.black.withOpacity(0.1),
                                            blurRadius: 10)
                                      ],
                                    ),
                                    child: isUpdatingAvatar
                                        ? const SizedBox(
                                            width: 14,
                                            height: 14,
                                            child: CircularProgressIndicator(
                                              color: Colors.white,
                                              strokeWidth: 2,
                                            ),
                                          )
                                        : const Icon(
                                            Icons.edit_rounded,
                                            color: Colors.white,
                                            size: 14,
                                          ),
                                  ),
                                ),
                              ],
                            ),
                          ),
                          const SizedBox(height: 12),
                          Text(
                            'Informasi Data Diri',
                            style: GoogleFonts.plusJakartaSans(
                                fontSize: 18,
                                fontWeight: FontWeight.w800,
                                color: textColor),
                          ),
                          Text(
                            'Lengkapi data untuk keperluan administrasi',
                            style: GoogleFonts.plusJakartaSans(
                                fontSize: 12, color: subtitleColor),
                          ),
                        ],
                      ),
                    ),
                    const SizedBox(height: 40),
                    _buildSectionCard(
                      title: 'EMPLOYMENT DETAILS',
                      children: [
                        _buildLabel('NIK Karyawan'),
                        const SizedBox(height: 10),
                        _buildReadOnlyField(nik, Icons.badge_outlined,
                            disabledBgColor, borderColor),
                        const SizedBox(height: 20),
                        _buildLabel('Departemen'),
                        const SizedBox(height: 10),
                        _buildReadOnlyField(
                            department,
                            Icons.account_tree_outlined,
                            disabledBgColor,
                            borderColor),
                        if (profile['salary'] != null) ...[
                          const SizedBox(height: 20),
                          _buildLabel('Gaji Pokok'),
                          const SizedBox(height: 10),
                          _buildReadOnlyField(
                            'Rp ${NumberFormat.currency(locale: 'id_ID', symbol: '', decimalDigits: 0).format(profile['salary'])}',
                            Icons.payments_outlined,
                            disabledBgColor,
                            borderColor,
                          ),
                        ],
                      ],
                    ),
                    const SizedBox(height: 24),
                    _buildSectionCard(
                      title: 'PERSONAL INFORMATION',
                      children: [
                        Row(
                          children: [
                            Expanded(
                              child: Column(
                                crossAxisAlignment: CrossAxisAlignment.start,
                                children: [
                                  _buildLabel('Nama Depan'),
                                  const SizedBox(height: 12),
                                  _buildEditableField(
                                      _firstNameController,
                                      Icons.person_outline_rounded,
                                      borderColor,
                                      primaryRed),
                                ],
                              ),
                            ),
                            const SizedBox(width: 16),
                            Expanded(
                              child: Column(
                                crossAxisAlignment: CrossAxisAlignment.start,
                                children: [
                                  _buildLabel('Nama Belakang'),
                                  const SizedBox(height: 12),
                                  _buildEditableField(
                                      _lastNameController,
                                      Icons.person_outline_rounded,
                                      borderColor,
                                      primaryRed),
                                ],
                              ),
                            ),
                          ],
                        ),
                        const SizedBox(height: 20),
                        Row(
                          children: [
                            Expanded(
                              child: Column(
                                crossAxisAlignment: CrossAxisAlignment.start,
                                children: [
                                  _buildLabel('Tempat Lahir'),
                                  const SizedBox(height: 12),
                                  _buildEditableField(
                                      _birthPlaceController,
                                      Icons.location_city_rounded,
                                      borderColor,
                                      primaryRed),
                                ],
                              ),
                            ),
                            const SizedBox(width: 16),
                            Expanded(
                              child: Column(
                                crossAxisAlignment: CrossAxisAlignment.start,
                                children: [
                                  _buildLabel('Tanggal Lahir'),
                                  const SizedBox(height: 12),
                                  _buildDatePickerField(
                                      _birthDateController,
                                      Icons.calendar_today_rounded,
                                      borderColor,
                                      primaryRed),
                                ],
                              ),
                            ),
                          ],
                        ),
                        const SizedBox(height: 20),
                        _buildLabel('Jenis Kelamin'),
                        const SizedBox(height: 12),
                        _buildDropdownField(
                            _gender,
                            ['Laki-laki', 'Perempuan'],
                            Icons.wc_rounded,
                            borderColor,
                            primaryRed,
                            (v) => setState(() => _gender = v)),
                        const SizedBox(height: 20),
                        _buildLabel('Status Nikah'),
                        const SizedBox(height: 12),
                        _buildDropdownField(
                            _maritalStatus,
                            ['Lajang', 'Menikah', 'Cerai'],
                            Icons.favorite_rounded,
                            borderColor,
                            primaryRed,
                            (v) => setState(() => _maritalStatus = v)),
                        const SizedBox(height: 20),
                        _buildLabel('Nomor Telepon'),
                        const SizedBox(height: 12),
                        _buildEditableField(
                            _phoneController,
                            Icons.phone_android_rounded,
                            borderColor,
                            primaryRed),
                      ],
                    ),
                    const SizedBox(height: 24),
                    _buildSectionCard(
                      title: 'IDENTIFICATION & ADDRESS',
                      children: [
                        _buildLabel('Nomor KTP (NIK)'),
                        const SizedBox(height: 12),
                        _buildEditableField(_identityNumberController,
                            Icons.badge_rounded, borderColor, primaryRed),
                        const SizedBox(height: 20),
                        _buildLabel('Nomor NPWP'),
                        const SizedBox(height: 12),
                        _buildEditableField(
                            _npwpNumberController,
                            Icons.receipt_long_outlined,
                            borderColor,
                            primaryRed),
                        const SizedBox(height: 20),
                        _buildLabel('Alamat Sesuai KTP'),
                        const SizedBox(height: 12),
                        _buildEditableField(_addressKtpController,
                            Icons.map_rounded, borderColor, primaryRed,
                            maxLines: 2),
                        const SizedBox(height: 20),
                        _buildLabel('Alamat Domisili'),
                        const SizedBox(height: 12),
                        _buildEditableField(_addressResidentialController,
                            Icons.home_work_rounded, borderColor, primaryRed,
                            maxLines: 2),
                      ],
                    ),
                    const SizedBox(height: 24),
                    _buildSectionCard(
                      title: 'BANK INFORMATION',
                      children: [
                        _buildLabel('Nama Bank'),
                        const SizedBox(height: 12),
                        _buildEditableField(
                            _bankNameController,
                            Icons.account_balance_rounded,
                            borderColor,
                            primaryRed),
                        const SizedBox(height: 20),
                        _buildLabel('Nomor Rekening'),
                        const SizedBox(height: 12),
                        _buildEditableField(_bankAccountNumberController,
                            Icons.credit_card_rounded, borderColor, primaryRed),
                        const SizedBox(height: 20),
                        _buildLabel('Nama Pemilik Rekening'),
                        const SizedBox(height: 12),
                        _buildEditableField(_accountHolderNameController,
                            Icons.face_rounded, borderColor, primaryRed),
                      ],
                    ),
                  ],
                ),
              );
            },
          ),
        ),
        bottomSheet: Container(
          padding: const EdgeInsets.symmetric(horizontal: 24, vertical: 16),
          decoration: BoxDecoration(
            color: Colors.white,
            boxShadow: [
              BoxShadow(
                color: Colors.black.withOpacity(0.05),
                blurRadius: 10,
                offset: const Offset(0, -4),
              ),
            ],
          ),
          child: SafeArea(
            child: SizedBox(
              width: double.infinity,
              height: 56,
              child: BlocBuilder<ProfileBloc, ProfileState>(
                builder: (context, state) {
                  return ElevatedButton(
                    onPressed: state is ProfileUpdating
                        ? null
                        : () {
                            context.read<ProfileBloc>().add(
                                  UpdateProfileRequested(
                                    firstName: _firstNameController.text,
                                    lastName: _lastNameController.text,
                                    birthDate: _birthDateController.text,
                                    birthPlace: _birthPlaceController.text,
                                    gender: _gender ?? '',
                                    maritalStatus: _maritalStatus ?? '',
                                    phoneNumber: _phoneController.text,
                                    addressKtp: _addressKtpController.text,
                                    addressResidential:
                                        _addressResidentialController.text,
                                    emergencyContact:
                                        _emergencyContactController.text,
                                    identityNumber:
                                        _identityNumberController.text,
                                    npwpNumber: _npwpNumberController.text,
                                    bankName: _bankNameController.text,
                                    bankAccountNumber:
                                        _bankAccountNumberController.text,
                                    accountHolderName:
                                        _accountHolderNameController.text,
                                  ),
                                );
                          },
                    style: ElevatedButton.styleFrom(
                      backgroundColor: AppColors.primaryRed,
                      elevation: 8,
                      shadowColor: AppColors.primaryRed.withOpacity(0.3),
                      shape: RoundedRectangleBorder(
                          borderRadius: BorderRadius.circular(16)),
                    ),
                    child: state is ProfileUpdating
                        ? const SizedBox(
                            width: 20,
                            height: 20,
                            child: CircularProgressIndicator(
                                color: Colors.white, strokeWidth: 2),
                          )
                        : const Text(
                            'SIMPAN PERUBAHAN',
                            style: TextStyle(
                                fontSize: 14,
                                fontWeight: FontWeight.w900,
                                color: Colors.white,
                                letterSpacing: 1),
                          ),
                  );
                },
              ),
            ),
          ),
        ),
      ),
    );
  }

  Widget _buildLabel(String text) {
    return Text(
      text,
      style: GoogleFonts.plusJakartaSans(
          fontSize: 14,
          fontWeight: FontWeight.w700,
          color: AppColors.textPrimary),
    );
  }

  Widget _buildReadOnlyField(
      String initialValue, IconData icon, Color bg, Color border) {
    return TextFormField(
      initialValue: initialValue,
      readOnly: true,
      style: GoogleFonts.plusJakartaSans(
          fontSize: 14,
          color: AppColors.textSecondary,
          fontWeight: FontWeight.w600),
      decoration: InputDecoration(
        filled: true,
        fillColor: bg,
        suffixIcon: Icon(icon, color: AppColors.textTertiary, size: 20),
        contentPadding:
            const EdgeInsets.symmetric(horizontal: 16, vertical: 16),
        border: OutlineInputBorder(
            borderRadius: BorderRadius.circular(12),
            borderSide: BorderSide.none),
      ),
    );
  }

  Widget _buildEditableField(TextEditingController controller, IconData icon,
      Color border, Color focusBorder,
      {int maxLines = 1}) {
    return TextFormField(
      controller: controller,
      maxLines: maxLines,
      style: GoogleFonts.plusJakartaSans(
          fontSize: 14,
          color: AppColors.textPrimary,
          fontWeight: FontWeight.w600),
      decoration: InputDecoration(
        filled: true,
        fillColor: AppColors.grayLight,
        suffixIcon: Icon(icon, color: AppColors.textTertiary, size: 20),
        contentPadding:
            const EdgeInsets.symmetric(horizontal: 16, vertical: 16),
        border: OutlineInputBorder(
            borderRadius: BorderRadius.circular(12),
            borderSide: BorderSide.none),
        focusedBorder: OutlineInputBorder(
            borderRadius: BorderRadius.circular(12),
            borderSide: BorderSide(color: focusBorder, width: 1.5)),
      ),
    );
  }

  Widget _buildDatePickerField(TextEditingController controller, IconData icon,
      Color border, Color focusBorder) {
    return TextFormField(
      controller: controller,
      readOnly: true,
      onTap: () => _selectDate(context, controller),
      style: GoogleFonts.plusJakartaSans(
          fontSize: 14,
          color: AppColors.textPrimary,
          fontWeight: FontWeight.w600),
      decoration: InputDecoration(
        filled: true,
        fillColor: AppColors.grayLight,
        suffixIcon: Icon(icon, color: AppColors.textTertiary, size: 20),
        contentPadding:
            const EdgeInsets.symmetric(horizontal: 16, vertical: 16),
        border: OutlineInputBorder(
            borderRadius: BorderRadius.circular(12),
            borderSide: BorderSide.none),
      ),
    );
  }

  Future<void> _selectDate(
      BuildContext context, TextEditingController controller) async {
    DateTime initialDate =
        DateTime.now().subtract(const Duration(days: 365 * 20));
    if (controller.text.isNotEmpty) {
      try {
        initialDate = DateFormat('yyyy-MM-dd').parse(controller.text);
      } catch (_) {}
    }

    final DateTime? picked = await showDatePicker(
      context: context,
      initialDate: initialDate,
      firstDate: DateTime(1950),
      lastDate: DateTime.now(),
      builder: (context, child) {
        return Theme(
          data: Theme.of(context).copyWith(
            colorScheme: const ColorScheme.light(
              primary: AppColors.primaryRed,
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
        controller.text = DateFormat('yyyy-MM-dd').format(picked);
      });
    }
  }

  Widget _buildDropdownField(String? value, List<String> items, IconData icon,
      Color border, Color focusBorder, void Function(String?) onChanged) {
    return DropdownButtonFormField<String>(
      initialValue: (value != null && items.contains(value)) ? value : null,
      items: items
          .map((i) => DropdownMenuItem(
              value: i, child: Text(i, style: const TextStyle(fontSize: 14))))
          .toList(),
      onChanged: onChanged,
      style: const TextStyle(
          fontSize: 15,
          color: AppColors.textPrimary,
          fontWeight: FontWeight.w600),
      dropdownColor: Colors.white,
      decoration: InputDecoration(
        filled: true,
        fillColor: AppColors.grayLight,
        suffixIcon: Icon(icon, color: AppColors.textTertiary, size: 20),
        contentPadding:
            const EdgeInsets.symmetric(horizontal: 16, vertical: 16),
        border: OutlineInputBorder(
            borderRadius: BorderRadius.circular(12),
            borderSide: BorderSide.none),
      ),
    );
  }

  Widget _buildSectionCard(
      {required String title, required List<Widget> children}) {
    return Column(
      crossAxisAlignment: CrossAxisAlignment.start,
      children: [
        Padding(
          padding: const EdgeInsets.only(left: 4, bottom: 12),
          child: Text(title,
              style: const TextStyle(
                  fontSize: 11,
                  fontWeight: FontWeight.w900,
                  color: AppColors.textTertiary,
                  letterSpacing: 1.5)),
        ),
        Container(
          width: double.infinity,
          padding: const EdgeInsets.all(20),
          decoration: BoxDecoration(
            color: Colors.white,
            borderRadius: BorderRadius.circular(24),
            border: Border.all(color: AppColors.grayBorder),
            boxShadow: [
              BoxShadow(
                  color: Colors.black.withOpacity(0.02),
                  blurRadius: 10,
                  offset: const Offset(0, 4))
            ],
          ),
          child: Column(
            crossAxisAlignment: CrossAxisAlignment.start,
            children: children,
          ),
        ),
      ],
    );
  }
}
