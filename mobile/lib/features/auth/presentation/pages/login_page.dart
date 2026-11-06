import 'package:flutter/material.dart';
import 'package:flutter_bloc/flutter_bloc.dart';
import 'package:hris_app/features/auth/presentation/bloc/auth_bloc.dart';
import 'package:hris_app/features/auth/presentation/bloc/auth_event.dart';
import 'package:hris_app/features/auth/presentation/bloc/auth_state.dart';
import 'package:hris_app/features/auth/presentation/pages/forgot_password_page.dart';

import 'package:hris_app/core/theme/app_colors.dart';
import 'package:flutter_staggered_animations/flutter_staggered_animations.dart';
import 'package:package_info_plus/package_info_plus.dart';

class LoginPage extends StatefulWidget {
  const LoginPage({super.key});

  @override
  State<LoginPage> createState() => _LoginPageState();
}

class _LoginPageState extends State<LoginPage> {
  final _emailController = TextEditingController();
  final _passwordController = TextEditingController();
  final _formKey = GlobalKey<FormState>();
  bool _obscureText = true;
  bool _rememberMe = false;
  bool _biometricPrompted = false;
  String _appVersion = 'VERSION ... • KARSA HR';

  @override
  void initState() {
    super.initState();
    final authBloc = context.read<AuthBloc>();
    authBloc.add(CheckBiometricSupportRequested());
    authBloc.add(LoadRememberedCredentialsRequested());
    _loadAppVersion();
  }

  Future<void> _loadAppVersion() async {
    try {
      final packageInfo = await PackageInfo.fromPlatform();
      if (mounted) {
        setState(() {
          _appVersion = 'VERSION ${packageInfo.version} • KARSA HR';
        });
      }
    } catch (e) {
      // Fallback
    }
  }

  @override
  void dispose() {
    _emailController.dispose();
    _passwordController.dispose();
    super.dispose();
  }

  void _onLoginPressed() {
    if (_formKey.currentState!.validate()) {
      context.read<AuthBloc>().add(
        LoginRequested(
          _emailController.text.trim(),
          _passwordController.text,
          rememberMe: _rememberMe,
        ),
      );
    }
  }

  @override
  Widget build(BuildContext context) {
    return Scaffold(
      backgroundColor: Colors.white,
      body: BlocListener<AuthBloc, AuthState>(
        listener: (context, state) {
          if (state is Unauthenticated) {
            if (state.rememberedCredentials != null) {
              _emailController.text =
                  state.rememberedCredentials!['email'] ?? '';
              _passwordController.text =
                  state.rememberedCredentials!['password'] ?? '';
              setState(() {
                _rememberMe = true;
              });
            }

            if (state.isBiometricSupported &&
                state.isBiometricEnabled &&
                !_biometricPrompted) {
              _biometricPrompted = true;
              context.read<AuthBloc>().add(BiometricLoginRequested());
            }
          }
        },
        child: SafeArea(
          child: SingleChildScrollView(
            child: Padding(
              padding: const EdgeInsets.symmetric(horizontal: 28.0),
              child: AnimationLimiter(
                child: Form(
                  key: _formKey,
                  child: Column(
                    crossAxisAlignment: CrossAxisAlignment.stretch,
                    children: AnimationConfiguration.toStaggeredList(
                      duration: const Duration(milliseconds: 600),
                      childAnimationBuilder: (widget) => FadeInAnimation(
                        child: SlideAnimation(
                          verticalOffset: 30.0,
                          child: widget,
                        ),
                      ),
                      children: [
                        const SizedBox(height: 60),
                        // Branding Section
                        Center(
                          child: Column(
                            children: [
                              Container(
                                padding: const EdgeInsets.all(16),
                                child: Image.asset(
                                  'assets/brand.png',
                                  width: 120,
                                  height: 120,
                                ),
                              ),
                              const SizedBox(height: 32),
                              Text(
                                'Selamat Datang',
                                style: Theme.of(context).textTheme.displayLarge
                                    ?.copyWith(fontSize: 30),
                              ),
                              const SizedBox(height: 12),
                              Text(
                                'Aplikasi Operasional Karyawan\nPerusahaan Demo',
                                textAlign: TextAlign.center,
                                style: Theme.of(context).textTheme.bodyMedium
                                    ?.copyWith(
                                      fontSize: 14,
                                      height: 1.5,
                                      fontWeight: FontWeight.w500,
                                    ),
                              ),
                            ],
                          ),
                        ),
                        const SizedBox(height: 56),

                        // Email Field
                        _buildLabel('Email'),
                        const SizedBox(height: 12),
                        TextFormField(
                          controller: _emailController,
                          keyboardType: TextInputType.emailAddress,
                          style: const TextStyle(
                            fontWeight: FontWeight.w600,
                            fontSize: 15,
                          ),
                          decoration: const InputDecoration(
                            hintText: 'Contoh: email@anda.com',
                            prefixIcon: Icon(
                              Icons.person_outline_rounded,
                              size: 22,
                            ),
                          ),
                          validator: (value) => (value == null || value.isEmpty)
                              ? 'Masukkan NIK atau Email'
                              : null,
                        ),
                        const SizedBox(height: 28),

                        // Password Field
                        _buildLabel('Kata Sandi'),
                        const SizedBox(height: 12),
                        TextFormField(
                          controller: _passwordController,
                          obscureText: _obscureText,
                          style: const TextStyle(
                            fontWeight: FontWeight.w600,
                            fontSize: 15,
                          ),
                          decoration: InputDecoration(
                            hintText: '••••••••••••',
                            prefixIcon: const Icon(
                              Icons.lock_outline_rounded,
                              size: 22,
                            ),
                            suffixIcon: IconButton(
                              icon: Icon(
                                _obscureText
                                    ? Icons.visibility_off_outlined
                                    : Icons.visibility_outlined,
                                color: AppColors.textTertiary,
                                size: 20,
                              ),
                              onPressed: () =>
                                  setState(() => _obscureText = !_obscureText),
                            ),
                          ),
                          validator: (value) => (value == null || value.isEmpty)
                              ? 'Masukkan kata sandi'
                              : null,
                        ),

                        const SizedBox(height: 16),

                        Row(
                          mainAxisAlignment: MainAxisAlignment.spaceBetween,
                          crossAxisAlignment: CrossAxisAlignment.center,
                          children: [
                            InkWell(
                              onTap: () =>
                                  setState(() => _rememberMe = !_rememberMe),
                              child: Row(
                                mainAxisSize: MainAxisSize.min,
                                children: [
                                  SizedBox(
                                    height: 24,
                                    width: 24,
                                    child: Checkbox(
                                      value: _rememberMe,
                                      onChanged: (val) => setState(
                                        () => _rememberMe = val ?? false,
                                      ),
                                      activeColor: AppColors.primaryGreen,
                                      shape: RoundedRectangleBorder(
                                        borderRadius: BorderRadius.circular(6),
                                      ),
                                      side: const BorderSide(
                                        color: AppColors.grayBorder,
                                        width: 1.5,
                                      ),
                                    ),
                                  ),
                                  const SizedBox(width: 8),
                                  const Text(
                                    'Ingat Saya',
                                    style: TextStyle(
                                      fontSize: 13,
                                      fontWeight: FontWeight.w600,
                                      color: AppColors.textSecondary,
                                    ),
                                  ),
                                ],
                              ),
                            ),
                            TextButton(
                              onPressed: () {
                                Navigator.of(context).push(
                                  MaterialPageRoute(
                                    builder: (_) => const ForgotPasswordPage(),
                                  ),
                                );
                              },
                              style: TextButton.styleFrom(
                                foregroundColor: AppColors.primaryGreen,
                                padding: EdgeInsets.zero,
                                visualDensity: VisualDensity.compact,
                              ),
                              child: const Text(
                                'Lupa kata sandi?',
                                style: TextStyle(
                                  fontWeight: FontWeight.w700,
                                  fontSize: 13,
                                ),
                              ),
                            ),
                          ],
                        ),
                        const SizedBox(height: 48),

                        // Sign In Button
                        BlocBuilder<AuthBloc, AuthState>(
                          builder: (context, state) {
                            return Column(
                              crossAxisAlignment: CrossAxisAlignment.stretch,
                              children: [
                                Container(
                                  decoration: BoxDecoration(
                                    borderRadius: BorderRadius.circular(16),
                                    gradient: const LinearGradient(
                                      colors: AppColors.primaryGradient,
                                      begin: Alignment.topLeft,
                                      end: Alignment.bottomRight,
                                    ),
                                    boxShadow: [
                                      BoxShadow(
                                        color: AppColors.primaryGreen
                                            .withOpacity(0.3),
                                        blurRadius: 20,
                                        offset: const Offset(0, 10),
                                      ),
                                    ],
                                  ),
                                  child: ElevatedButton(
                                    onPressed: state is AuthLoading
                                        ? null
                                        : _onLoginPressed,
                                    style: ElevatedButton.styleFrom(
                                      backgroundColor: Colors.transparent,
                                      shadowColor: Colors.transparent,
                                    ),
                                    child: state is AuthLoading
                                        ? const SizedBox(
                                            height: 20,
                                            width: 20,
                                            child: CircularProgressIndicator(
                                              color: Colors.white,
                                              strokeWidth: 2,
                                            ),
                                          )
                                        : const Text('MASUK KE DASHBOARD'),
                                  ),
                                ),
                                if (state is! AuthLoading &&
                                    state is Unauthenticated &&
                                    state.isBiometricSupported &&
                                    state.isBiometricEnabled) ...[
                                  const SizedBox(height: 20),
                                  OutlinedButton.icon(
                                    onPressed: () => context
                                        .read<AuthBloc>()
                                        .add(BiometricLoginRequested()),
                                    icon: const Icon(
                                      Icons.fingerprint_rounded,
                                      size: 28,
                                    ),
                                    label: const Text(
                                      'MASUK DENGAN SIDIK JARI',
                                      style: TextStyle(
                                        fontWeight: FontWeight.w800,
                                      ),
                                    ),
                                    style: OutlinedButton.styleFrom(
                                      padding: const EdgeInsets.symmetric(
                                        vertical: 16,
                                      ),
                                      side: const BorderSide(
                                        color: AppColors.primaryGreen,
                                      ),
                                      foregroundColor: AppColors.primaryGreen,
                                      shape: RoundedRectangleBorder(
                                        borderRadius: BorderRadius.circular(16),
                                      ),
                                    ),
                                  ),
                                ],
                              ],
                            );
                          },
                        ),
                        const SizedBox(height: 48),

                        const SizedBox(height: 60),

                        // Footer
                        Center(
                          child: Column(
                            children: [
                              Container(
                                width: 40,
                                height: 2,
                                decoration: BoxDecoration(
                                  color: AppColors.grayBorder,
                                  borderRadius: BorderRadius.circular(1),
                                ),
                              ),
                              const SizedBox(height: 16),
                              Text(
                                _appVersion,
                                style: const TextStyle(
                                  color: AppColors.textTertiary,
                                  fontSize: 10,
                                  fontWeight: FontWeight.w800,
                                  letterSpacing: 2,
                                ),
                              ),
                            ],
                          ),
                        ),
                        const SizedBox(height: 32),
                      ],
                    ),
                  ),
                ),
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
      style: const TextStyle(
        fontSize: 13,
        fontWeight: FontWeight.w800,
        color: AppColors.textPrimary,
        letterSpacing: 0.5,
      ),
    );
  }
}
