import 'package:flutter/material.dart';
import 'package:flutter_bloc/flutter_bloc.dart';
import 'package:hris_app/features/auth/presentation/bloc/auth_bloc.dart';
import 'package:hris_app/features/auth/presentation/bloc/auth_event.dart';
import 'package:hris_app/features/auth/presentation/bloc/auth_state.dart';
import 'package:google_fonts/google_fonts.dart';
import 'package:hris_app/features/attendance/presentation/pages/face_verification_page.dart';
import 'package:flutter/foundation.dart' show kIsWeb;
import 'dart:io' show File;
import 'package:hris_app/core/utils/snackbar_utils.dart';
import 'package:flutter_staggered_animations/flutter_staggered_animations.dart';
import 'package:hris_app/injection.dart' as di;
import 'package:hris_app/features/auth/domain/repositories/auth_repository.dart';
import 'package:hris_app/core/theme/app_colors.dart';
import 'package:hris_app/core/utils/dialog_utils.dart';

class RegisterPage extends StatefulWidget {
  const RegisterPage({super.key});

  @override
  State<RegisterPage> createState() => _RegisterPageState();
}

class _RegisterPageState extends State<RegisterPage> {
  final _nameController = TextEditingController();
  final _emailController = TextEditingController();
  final _passwordController = TextEditingController();

  bool _isPasswordVisible = false;
  bool _agreedToTerms = false;
  final _formKey = GlobalKey<FormState>();

  List<double>? _faceEmbedding;
  String? _selfiePath;
  String? _selectedJobPositionId;
  String? _selectedBranchId;

  List<Map<String, dynamic>> _jobPositions = [];
  List<Map<String, dynamic>> _branches = [];
  List<Map<String, dynamic>> _shifts = [];
  String? _selectedShiftId;
  bool _isLoadingData = true;

  @override
  void initState() {
    super.initState();
    _fetchData();
  }

  Future<void> _fetchData() async {
    final repo = di.sl<AuthRepository>();

    final branchesResult = await repo.getBranches();
    final jobsResult = await repo.getJobPositions();

    if (mounted) {
      branchesResult.fold(
        (failure) => SnackBarUtils.showError(context, failure.message),
        (data) => setState(() => _branches = data),
      );

      jobsResult.fold(
        (failure) => SnackBarUtils.showError(context, failure.message),
        (data) => setState(() => _jobPositions = data),
      );

      final shiftsResult = await repo.getShifts();
      shiftsResult.fold(
        (failure) => SnackBarUtils.showError(context, failure.message),
        (data) => setState(() => _shifts = data),
      );

      setState(() => _isLoadingData = false);
    }
  }

  @override
  void dispose() {
    _nameController.dispose();
    _emailController.dispose();
    _passwordController.dispose();
    super.dispose();
  }

  void _onRegisterPressed() {
    if (_formKey.currentState!.validate() && _agreedToTerms) {
      if (_selfiePath == null) {
        SnackBarUtils.showError(
          context,
          'Harap ambil selfie verifikasi terlebih dahulu',
        );
        return;
      }
      if (_selectedBranchId == null) {
        SnackBarUtils.showError(context, 'Harap pilih cabang');
        return;
      }
      if (_selectedJobPositionId == null) {
        SnackBarUtils.showError(context, 'Harap pilih posisi pekerjaan');
        return;
      }

      final systemGeneratedNik = 'REG-${DateTime.now().millisecondsSinceEpoch}';

      context.read<AuthBloc>().add(
        RegisterRequested(
          _nameController.text.trim(),
          _emailController.text.trim(),
          systemGeneratedNik, // satisfying the 'required' tag on backend
          _passwordController.text,
          _selectedJobPositionId!,
          _selectedBranchId!,
          _selectedShiftId,
          embedding: _faceEmbedding,
          selfiePath: _selfiePath,
        ),
      );
    } else if (!_agreedToTerms) {
      SnackBarUtils.showError(
        context,
        'Harap setuju dengan syarat dan ketentuan',
      );
    }
  }

  @override
  Widget build(BuildContext context) {
    return Scaffold(
      backgroundColor: const Color(0xFFF8FAFC),
      body: BlocListener<AuthBloc, AuthState>(
        listener: (context, state) {
          if (state is RegisterSuccess) {
            DialogUtils.showSuccess(
              context: context,
              title: 'Pendaftaran Berhasil',
              message:
                  'Akun anda telah berhasil dibuat. Silahkan login menggunakan email dan kata sandi anda.',
              onConfirm: () => Navigator.of(context).pop(),
            );
          }
        },
        child: SafeArea(
          child: SingleChildScrollView(
            child: Column(
              children: [
                _buildHeader(),
                Padding(
                  padding: const EdgeInsets.fromLTRB(24, 0, 24, 32),
                  child: AnimationLimiter(
                    child: Form(
                      key: _formKey,
                      child: Column(
                        children: AnimationConfiguration.toStaggeredList(
                          duration: const Duration(milliseconds: 600),
                          childAnimationBuilder: (widget) => SlideAnimation(
                            verticalOffset: 20.0,
                            child: FadeInAnimation(child: widget),
                          ),
                          children: [
                            _buildSection(
                              title: 'INFORMASI PRIBADI',
                              icon: Icons.person_outline_rounded,
                              children: [
                                _buildTextField(
                                  label: 'Nama Lengkap',
                                  controller: _nameController,
                                  hintText: 'Sesuai KTP',
                                  icon: Icons.badge_outlined,
                                  validator: (v) =>
                                      v!.isEmpty ? 'Nama wajib diisi' : null,
                                ),
                                const SizedBox(height: 20),
                                _buildTextField(
                                  label: 'Alamat Email',
                                  controller: _emailController,
                                  hintText: 'email@karyawan.com',
                                  icon: Icons.email_outlined,
                                  keyboardType: TextInputType.emailAddress,
                                  validator: (v) => !v!.contains('@')
                                      ? 'Email tidak valid'
                                      : null,
                                ),
                                const SizedBox(height: 12),
                                _buildSystemGeneratedBadge(),
                              ],
                            ),
                            const SizedBox(height: 24),
                            _buildSection(
                              title: 'PENEMPATAN KERJA',
                              icon: Icons.business_outlined,
                              children: [
                                _isLoadingData
                                    ? const Center(
                                        child: Padding(
                                          padding: EdgeInsets.all(16),
                                          child: CircularProgressIndicator(),
                                        ),
                                      )
                                    : Column(
                                        children: [
                                          _buildDropdown(
                                            label: 'Cabang / Unit',
                                            value: _selectedBranchId,
                                            hint: 'Pilih cabang lokasi bekerja',
                                            icon: Icons.location_on_outlined,
                                            items: _branches
                                                .where(
                                                  (branch) =>
                                                      (branch['id'] ??
                                                          branch['ID']) !=
                                                      null,
                                                )
                                                .map(
                                                  (
                                                    branch,
                                                  ) => DropdownMenuItem<String>(
                                                    value:
                                                        (branch['id'] ??
                                                                branch['ID'])
                                                            .toString(),
                                                    child: Text(
                                                      (branch['name'] ??
                                                              branch['Name'] ??
                                                              'Unit unknown')
                                                          .toString(),
                                                      overflow:
                                                          TextOverflow.ellipsis,
                                                      maxLines: 1,
                                                    ),
                                                  ),
                                                )
                                                .toList(),
                                            onChanged: (v) => setState(
                                              () => _selectedBranchId = v,
                                            ),
                                            validator: (v) => v == null
                                                ? 'Pilih cabang'
                                                : null,
                                          ),
                                          const SizedBox(height: 20),
                                          _buildDropdown(
                                            label: 'Posisi Pekerjaan',
                                            value: _selectedJobPositionId,
                                            hint: 'Pilih jabatan anda',
                                            icon: Icons.work_outline_rounded,
                                            items: _jobPositions
                                                .where(
                                                  (pos) =>
                                                      (pos['id'] ??
                                                          pos['ID']) !=
                                                      null,
                                                )
                                                .map(
                                                  (
                                                    pos,
                                                  ) => DropdownMenuItem<String>(
                                                    value:
                                                        (pos['id'] ?? pos['ID'])
                                                            .toString(),
                                                    child: Text(
                                                      (pos['title'] ??
                                                              pos['Title'] ??
                                                              pos['name'] ??
                                                              pos['Name'] ??
                                                              'Posisi')
                                                          .toString(),
                                                      overflow:
                                                          TextOverflow.ellipsis,
                                                      maxLines: 1,
                                                    ),
                                                  ),
                                                )
                                                .toList(),
                                            onChanged: (v) => setState(
                                              () => _selectedJobPositionId = v,
                                            ),
                                            validator: (v) => v == null
                                                ? 'Pilih posisi'
                                                : null,
                                          ),
                                          const SizedBox(height: 20),
                                          _buildDropdown(
                                            label: 'Shift Kerja / Jam Kerja',
                                            value: _selectedShiftId,
                                            hint: 'Pilih jam kerja anda',
                                            icon: Icons.schedule_outlined,
                                            items: _shifts.where((s) => (s['id'] ?? s['ID']) != null).map((
                                              s,
                                            ) {
                                              final startTime =
                                                  (s['start_time'] ??
                                                          s['StartTime'] ??
                                                          '')
                                                      .toString();
                                              final endTime =
                                                  (s['end_time'] ??
                                                          s['EndTime'] ??
                                                          '')
                                                      .toString();

                                              // Robust format function: Use "Wall Time" (direct string extract)
                                              // to avoid unwanted timezone shifts for shift hours.
                                              String format(String t) {
                                                if (t.isEmpty) return '--.--';
                                                try {
                                                  // Handle ISO strings like 0000-01-01T09:00:12...
                                                  if (t.contains('T')) {
                                                    final timePart = t
                                                        .split('T')
                                                        .last;
                                                    final parts = timePart
                                                        .split(':');
                                                    if (parts.length >= 2) {
                                                      return '${parts[0]}.${parts[1]}';
                                                    }
                                                  }
                                                  // Handle HH:mm:ss strings
                                                  final parts = t.split(':');
                                                  if (parts.length >= 2) {
                                                    String hour = parts[0];
                                                    if (hour.contains(' '))
                                                      hour = hour
                                                          .split(' ')
                                                          .last;
                                                    return '$hour.${parts[1]}';
                                                  }
                                                } catch (e) {
                                                  // Fallback
                                                }
                                                return t;
                                              }

                                              final timeRange =
                                                  '${format(startTime)} sampai ${format(endTime)}';

                                              return DropdownMenuItem<String>(
                                                value: (s['id'] ?? s['ID'])
                                                    .toString(),
                                                child: Text(
                                                  '${s['name'] ?? s['Name'] ?? 'Shift'} ($timeRange)',
                                                  overflow:
                                                      TextOverflow.ellipsis,
                                                  maxLines: 1,
                                                ),
                                              );
                                            }).toList(),
                                            onChanged: (v) => setState(
                                              () => _selectedShiftId = v,
                                            ),
                                            validator: (v) => v == null
                                                ? 'Pilih shift kerja'
                                                : null,
                                          ),
                                        ],
                                      ),
                              ],
                            ),
                            const SizedBox(height: 24),
                            _buildSection(
                              title: 'KEAMANAN & VERIFIKASI',
                              icon: Icons.security_rounded,
                              children: [
                                _buildTextField(
                                  label: 'Kata Sandi',
                                  controller: _passwordController,
                                  hintText: 'Minimal 6 karakter',
                                  icon: Icons.lock_outline_rounded,
                                  obscureText: !_isPasswordVisible,
                                  suffixIcon: IconButton(
                                    icon: Icon(
                                      _isPasswordVisible
                                          ? Icons.visibility_off_outlined
                                          : Icons.visibility_outlined,
                                      color: AppColors.textTertiary,
                                      size: 20,
                                    ),
                                    onPressed: () => setState(
                                      () => _isPasswordVisible =
                                          !_isPasswordVisible,
                                    ),
                                  ),
                                  validator: (v) =>
                                      v!.length < 6 ? 'Min 6 karakter' : null,
                                ),
                                const SizedBox(height: 24),
                                _buildFaceVerificationCard(),
                              ],
                            ),
                            const SizedBox(height: 32),
                            _buildTermsCheckbox(),
                            const SizedBox(height: 32),
                            _buildSubmitButton(),
                            const SizedBox(height: 32),
                            _buildLoginRedirect(),
                          ],
                        ),
                      ),
                    ),
                  ),
                ),
              ],
            ),
          ),
        ),
      ),
    );
  }

  Widget _buildHeader() {
    return Container(
      width: double.infinity,
      padding: const EdgeInsets.symmetric(vertical: 48, horizontal: 24),
      decoration: const BoxDecoration(
        color: Colors.white,
        borderRadius: BorderRadius.only(
          bottomLeft: Radius.circular(40),
          bottomRight: Radius.circular(40),
        ),
      ),
      child: Column(
        children: [
          Hero(
            tag: 'logo',
            child: Image.asset('assets/brand.png', width: 80, height: 80),
          ),
          const SizedBox(height: 24),
          Text(
            'Pendaftaran Akun',
            style: GoogleFonts.plusJakartaSans(
              fontSize: 28,
              fontWeight: FontWeight.w800,
              color: AppColors.textPrimary,
              letterSpacing: -0.5,
            ),
          ),
          const SizedBox(height: 8),
          Text(
            'Silahkan lengkapi data anda untuk\nmengakses Karsa HR.',
            textAlign: TextAlign.center,
            style: GoogleFonts.plusJakartaSans(
              fontSize: 14,
              color: AppColors.textSecondary,
              height: 1.5,
              fontWeight: FontWeight.w500,
            ),
          ),
        ],
      ),
    );
  }

  Widget _buildSection({
    required String title,
    required IconData icon,
    required List<Widget> children,
  }) {
    return Column(
      crossAxisAlignment: CrossAxisAlignment.start,
      children: [
        Padding(
          padding: const EdgeInsets.only(left: 4, bottom: 12),
          child: Row(
            children: [
              Icon(icon, size: 18, color: AppColors.primaryRed),
              const SizedBox(width: 8),
              Text(
                title,
                style: GoogleFonts.plusJakartaSans(
                  fontSize: 12,
                  fontWeight: FontWeight.w800,
                  color: AppColors.textTertiary,
                  letterSpacing: 1.5,
                ),
              ),
            ],
          ),
        ),
        Container(
          padding: const EdgeInsets.all(20),
          decoration: BoxDecoration(
            color: Colors.white,
            borderRadius: BorderRadius.circular(24),
            boxShadow: [
              BoxShadow(
                color: Colors.black.withOpacity(0.03),
                blurRadius: 20,
                offset: const Offset(0, 10),
              ),
            ],
          ),
          child: Column(
            crossAxisAlignment: CrossAxisAlignment.stretch,
            children: children,
          ),
        ),
      ],
    );
  }

  Widget _buildTextField({
    required String label,
    required TextEditingController controller,
    required String hintText,
    required IconData icon,
    bool obscureText = false,
    TextInputType? keyboardType,
    Widget? suffixIcon,
    String? Function(String?)? validator,
  }) {
    return Column(
      crossAxisAlignment: CrossAxisAlignment.start,
      children: [
        Text(
          label,
          style: GoogleFonts.plusJakartaSans(
            fontSize: 13,
            fontWeight: FontWeight.w700,
            color: AppColors.textPrimary,
          ),
        ),
        const SizedBox(height: 8),
        TextFormField(
          controller: controller,
          obscureText: obscureText,
          keyboardType: keyboardType,
          style: GoogleFonts.plusJakartaSans(
            fontSize: 15,
            fontWeight: FontWeight.w600,
          ),
          decoration: InputDecoration(
            hintText: hintText,
            prefixIcon: Icon(icon, size: 20),
            suffixIcon: suffixIcon,
            filled: true,
            fillColor: const Color(0xFFF9FAFB),
            contentPadding: const EdgeInsets.symmetric(
              horizontal: 16,
              vertical: 16,
            ),
            border: OutlineInputBorder(
              borderRadius: BorderRadius.circular(14),
              borderSide: BorderSide.none,
            ),
            enabledBorder: OutlineInputBorder(
              borderRadius: BorderRadius.circular(14),
              borderSide: BorderSide.none,
            ),
            focusedBorder: OutlineInputBorder(
              borderRadius: BorderRadius.circular(14),
              borderSide: const BorderSide(
                color: AppColors.primaryRed,
                width: 1.5,
              ),
            ),
            errorBorder: OutlineInputBorder(
              borderRadius: BorderRadius.circular(14),
              borderSide: const BorderSide(color: Colors.red, width: 1),
            ),
          ),
          validator: validator,
        ),
      ],
    );
  }

  Widget _buildDropdown({
    required String label,
    required String? value,
    required String hint,
    required IconData icon,
    required List<DropdownMenuItem<String>> items,
    required void Function(String?) onChanged,
    String? Function(String?)? validator,
  }) {
    return Column(
      crossAxisAlignment: CrossAxisAlignment.start,
      children: [
        Text(
          label,
          style: GoogleFonts.plusJakartaSans(
            fontSize: 13,
            fontWeight: FontWeight.w700,
            color: AppColors.textPrimary,
          ),
        ),
        const SizedBox(height: 8),
        DropdownButtonFormField<String>(
          initialValue: value,
          items: items,
          onChanged: onChanged,
          validator: validator,
          isExpanded: true,
          style: GoogleFonts.plusJakartaSans(
            fontSize: 15,
            fontWeight: FontWeight.w600,
            color: AppColors.textPrimary,
          ),
          decoration: InputDecoration(
            hintText: hint,
            prefixIcon: Icon(icon, size: 20),
            filled: true,
            fillColor: const Color(0xFFF9FAFB),
            contentPadding: const EdgeInsets.symmetric(
              horizontal: 16,
              vertical: 16,
            ),
            border: OutlineInputBorder(
              borderRadius: BorderRadius.circular(14),
              borderSide: BorderSide.none,
            ),
            enabledBorder: OutlineInputBorder(
              borderRadius: BorderRadius.circular(14),
              borderSide: BorderSide.none,
            ),
            focusedBorder: OutlineInputBorder(
              borderRadius: BorderRadius.circular(14),
              borderSide: const BorderSide(
                color: AppColors.primaryRed,
                width: 1.5,
              ),
            ),
          ),
        ),
      ],
    );
  }

  Widget _buildSystemGeneratedBadge() {
    return Container(
      padding: const EdgeInsets.symmetric(horizontal: 12, vertical: 12),
      decoration: BoxDecoration(
        color: AppColors.primaryRed.withOpacity(0.05),
        borderRadius: BorderRadius.circular(14),
        border: Border.all(color: AppColors.primaryRed.withOpacity(0.1)),
      ),
      child: Row(
        children: [
          const Icon(
            Icons.auto_awesome_rounded,
            size: 16,
            color: AppColors.primaryRed,
          ),
          const SizedBox(width: 12),
          Expanded(
            child: Text(
              'NIK Karyawan akan dibuat otomatis oleh sistem untuk konsistensi data.',
              style: GoogleFonts.plusJakartaSans(
                fontSize: 12,
                fontWeight: FontWeight.w700,
                color: AppColors.primaryRed,
              ),
            ),
          ),
        ],
      ),
    );
  }

  Widget _buildFaceVerificationCard() {
    return InkWell(
      onTap: () async {
        final result = await Navigator.of(context).push<FaceVerificationResult>(
          MaterialPageRoute(
            builder: (_) => const FaceVerificationPage(isRegistration: true),
          ),
        );
        if (result != null) {
          setState(() {
            _faceEmbedding = result.embedding;
            _selfiePath = result.imagePath;
          });
        }
      },
      child: Container(
        padding: const EdgeInsets.all(16),
        decoration: BoxDecoration(
          color: _selfiePath != null
              ? AppColors.primaryRed.withOpacity(0.04)
              : const Color(0xFFF9FAFB),
          borderRadius: BorderRadius.circular(16),
          border: Border.all(
            color: _selfiePath != null
                ? AppColors.primaryRed.withOpacity(0.2)
                : Colors.transparent,
          ),
        ),
        child: Row(
          children: [
            Container(
              width: 56,
              height: 56,
              decoration: BoxDecoration(
                color: Colors.white,
                borderRadius: BorderRadius.circular(14),
                boxShadow: [
                  if (_selfiePath != null)
                    BoxShadow(
                      color: AppColors.primaryRed.withOpacity(0.1),
                      blurRadius: 10,
                    ),
                ],
                image: _selfiePath != null
                    ? DecorationImage(
                        image: kIsWeb
                            ? NetworkImage(_selfiePath!) as ImageProvider
                            : FileImage(File(_selfiePath!)),
                        fit: BoxFit.cover,
                      )
                    : null,
              ),
              child: _selfiePath == null
                  ? const Icon(
                      Icons.face_retouching_natural_rounded,
                      color: AppColors.textTertiary,
                    )
                  : null,
            ),
            const SizedBox(width: 16),
            Expanded(
              child: Column(
                crossAxisAlignment: CrossAxisAlignment.start,
                children: [
                  Text(
                    _selfiePath == null
                        ? 'Verifikasi Wajah'
                        : 'Wajah Terdaftar',
                    style: GoogleFonts.plusJakartaSans(
                      fontSize: 15,
                      fontWeight: FontWeight.w800,
                      color: _selfiePath == null
                          ? AppColors.textPrimary
                          : AppColors.primaryRed,
                    ),
                  ),
                  const SizedBox(height: 2),
                  Text(
                    _selfiePath == null
                        ? 'Wajib untuk fitur absensi'
                        : 'Identitas wajah tersimpan',
                    style: GoogleFonts.plusJakartaSans(
                      fontSize: 12,
                      color: AppColors.textSecondary,
                      fontWeight: FontWeight.w500,
                    ),
                  ),
                ],
              ),
            ),
            Icon(
              _selfiePath == null
                  ? Icons.arrow_forward_ios_rounded
                  : Icons.check_circle_rounded,
              size: 16,
              color: _selfiePath == null
                  ? AppColors.textTertiary
                  : Colors.green,
            ),
          ],
        ),
      ),
    );
  }

  Widget _buildTermsCheckbox() {
    return Row(
      crossAxisAlignment: CrossAxisAlignment.start,
      children: [
        SizedBox(
          width: 24,
          height: 24,
          child: Checkbox(
            value: _agreedToTerms,
            onChanged: (v) => setState(() => _agreedToTerms = v ?? false),
            activeColor: AppColors.primaryRed,
            shape: RoundedRectangleBorder(
              borderRadius: BorderRadius.circular(6),
            ),
          ),
        ),
        const SizedBox(width: 12),
        Expanded(
          child: RichText(
            text: TextSpan(
              style: GoogleFonts.plusJakartaSans(
                color: AppColors.textSecondary,
                fontSize: 13,
                height: 1.5,
                fontWeight: FontWeight.w500,
              ),
              children: const [
                TextSpan(text: 'Saya menyetujui '),
                TextSpan(
                  text: 'Syarat & Ketentuan',
                  style: TextStyle(
                    color: AppColors.primaryRed,
                    fontWeight: FontWeight.w700,
                  ),
                ),
                TextSpan(text: ' serta '),
                TextSpan(
                  text: 'Kebijakan Privasi',
                  style: TextStyle(
                    color: AppColors.primaryRed,
                    fontWeight: FontWeight.w700,
                  ),
                ),
                TextSpan(text: ' yang berlaku.'),
              ],
            ),
          ),
        ),
      ],
    );
  }

  Widget _buildSubmitButton() {
    return BlocBuilder<AuthBloc, AuthState>(
      builder: (context, state) {
        bool isLoading = state is AuthLoading;
        return Container(
          width: double.infinity,
          height: 58,
          decoration: BoxDecoration(
            borderRadius: BorderRadius.circular(18),
            gradient: const LinearGradient(
              colors: AppColors.primaryGradient,
              begin: Alignment.topLeft,
              end: Alignment.bottomRight,
            ),
            boxShadow: [
              BoxShadow(
                color: AppColors.primaryRed.withOpacity(0.35),
                blurRadius: 20,
                offset: const Offset(0, 10),
              ),
            ],
          ),
          child: ElevatedButton(
            onPressed: isLoading ? null : _onRegisterPressed,
            style: ElevatedButton.styleFrom(
              backgroundColor: Colors.transparent,
              shadowColor: Colors.transparent,
              shape: RoundedRectangleBorder(
                borderRadius: BorderRadius.circular(18),
              ),
            ),
            child: isLoading
                ? const SizedBox(
                    width: 24,
                    height: 24,
                    child: CircularProgressIndicator(
                      color: Colors.white,
                      strokeWidth: 3,
                    ),
                  )
                : Text(
                    'DAFTAR SEKARANG',
                    style: GoogleFonts.plusJakartaSans(
                      fontSize: 16,
                      fontWeight: FontWeight.w800,
                      color: Colors.white,
                      letterSpacing: 1.5,
                    ),
                  ),
          ),
        );
      },
    );
  }

  Widget _buildLoginRedirect() {
    return Row(
      mainAxisAlignment: MainAxisAlignment.center,
      children: [
        Text(
          'Sudah punya akun? ',
          style: GoogleFonts.plusJakartaSans(
            color: AppColors.textSecondary,
            fontSize: 14,
            fontWeight: FontWeight.w500,
          ),
        ),
        GestureDetector(
          onTap: () => Navigator.of(context).pop(),
          child: Text(
            'Masuk di sini',
            style: GoogleFonts.plusJakartaSans(
              color: AppColors.primaryRed,
              fontSize: 14,
              fontWeight: FontWeight.w800,
            ),
          ),
        ),
      ],
    );
  }
}
