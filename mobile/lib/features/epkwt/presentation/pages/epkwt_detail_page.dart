import 'package:cached_network_image/cached_network_image.dart';
import 'package:flutter/material.dart';
import 'package:flutter_bloc/flutter_bloc.dart';
import 'package:google_fonts/google_fonts.dart';
import 'package:intl/intl.dart';
import 'package:hris_app/core/theme/app_colors.dart';
import 'package:hris_app/core/utils/constants.dart';
import 'package:hris_app/features/epkwt/domain/entities/epkwt_request.dart';
import 'package:hris_app/features/epkwt/presentation/bloc/epkwt_bloc.dart';
import 'package:hris_app/features/epkwt/presentation/bloc/epkwt_event.dart';
import 'package:hris_app/features/epkwt/presentation/bloc/epkwt_state.dart';
import 'package:hris_app/features/epkwt/presentation/widgets/signature_canvas.dart';

class EpkwtDetailPage extends StatefulWidget {
  final EpkwtRequest request;
  final bool openSignatureInitially;

  const EpkwtDetailPage({
    super.key,
    required this.request,
    this.openSignatureInitially = false,
  });

  @override
  State<EpkwtDetailPage> createState() => _EpkwtDetailPageState();
}

class _EpkwtDetailPageState extends State<EpkwtDetailPage> {
  @override
  void initState() {
    super.initState();
    if (widget.openSignatureInitially) {
      WidgetsBinding.instance.addPostFrameCallback((_) {
        if (!mounted) return;
        final route = ModalRoute.of(context);
        if (route != null && route.animation != null) {
          void statusListener(AnimationStatus status) {
            if (status == AnimationStatus.completed) {
              route.animation!.removeStatusListener(statusListener);
              if (mounted) {
                _showSignatureBottomSheet();
              }
            }
          }

          if (route.animation!.status == AnimationStatus.completed) {
            _showSignatureBottomSheet();
          } else {
            route.animation!.addStatusListener(statusListener);
          }
        } else {
          _showSignatureBottomSheet();
        }
      });
    }
  }

  void _showSignatureBottomSheet() {
    showModalBottomSheet(
      context: context,
      isScrollControlled: true,
      backgroundColor: Colors.white,
      shape: const RoundedRectangleBorder(
        borderRadius: BorderRadius.vertical(top: Radius.circular(28)),
      ),
      builder: (sheetContext) {
        return Padding(
          padding: EdgeInsets.only(
            bottom: MediaQuery.of(sheetContext).viewInsets.bottom,
            top: 24,
            left: 24,
            right: 24,
          ),
          child: SafeArea(
            child: SignatureCanvas(
              onSave: (base64) {
                Navigator.pop(sheetContext);
                if (mounted) {
                  context.read<EpkwtBloc>().add(
                        SignEpkwtRequested(
                          id: widget.request.id,
                          base64Signature: base64,
                        ),
                      );
                }
              },
            ),
          ),
        );
      },
    );
  }

  @override
  Widget build(BuildContext context) {
    final currencyFormat = NumberFormat.currency(
      locale: 'id_ID',
      symbol: 'Rp ',
      decimalDigits: 0,
    );

    return Scaffold(
      backgroundColor: AppColors.backgroundAlt,
      appBar: AppBar(
        title: Text(
          'Detail Kontrak PKWT',
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
          child: Column(
            crossAxisAlignment: CrossAxisAlignment.start,
            children: [
              // Status Card
              _buildStatusCard(),
              const SizedBox(height: 20),

              // Candidate Details Card
              _buildSectionTitle('Informasi Pribadi'),
              const SizedBox(height: 10),
              _buildCard([
                _buildDetailRow('No. KTP / NIK', widget.request.identityNumber),
                _buildDetailRow('Nama Lengkap', widget.request.employeeName),
                _buildDetailRow(
                    'No. Telepon', widget.request.phoneNumber ?? '-'),
                _buildDetailRow('Jenis Kelamin', widget.request.gender ?? '-'),
                _buildDetailRow(
                  'Tempat, Tgl Lahir',
                  '${widget.request.birthPlace ?? '-'}, ${widget.request.birthDate != null ? DateFormat('dd MMMM yyyy', 'id_ID').format(widget.request.birthDate!) : '-'}',
                ),
                _buildDetailRow('Alamat KTP', widget.request.addressKtp ?? '-'),
                _buildDetailRow(
                    'Status Pernikahan', widget.request.maritalStatus ?? '-'),
              ]),
              const SizedBox(height: 24),

              // Job / Contract Details Card
              _buildSectionTitle('Detail Pekerjaan & Kontrak'),
              const SizedBox(height: 10),
              _buildCard([
                _buildDetailRow(
                  'Jabatan di Kontrak',
                  widget.request.jobPositionTitle ?? 'Menunggu review HR',
                ),
                _buildDetailRow(
                  'Departemen',
                  widget.request.departmentName ?? 'Menunggu review HR',
                ),
                _buildDetailRow(
                  'Gaji Pokok',
                  widget.request.salary > 0
                      ? currencyFormat.format(widget.request.salary)
                      : 'Menunggu penetapan HR',
                ),
                _buildDetailRow(
                  'Tanggal Mulai Kontrak',
                  widget.request.startDate != null
                      ? DateFormat('dd MMMM yyyy', 'id_ID')
                          .format(widget.request.startDate!)
                      : 'Menunggu review HR',
                ),
                _buildDetailRow(
                  'Tanggal Berakhir Kontrak',
                  widget.request.endDate != null
                      ? DateFormat('dd MMMM yyyy', 'id_ID')
                          .format(widget.request.endDate!)
                      : 'Menunggu review HR',
                ),
              ]),
              const SizedBox(height: 24),

              // Signatures Section
              if (widget.request.status == 'ACTIVE') ...[
                _buildSectionTitle('Tanda Tangan Digital'),
                const SizedBox(height: 10),
                _buildSignatureDisplay(),
              ],

              const SizedBox(height: 32),

              // Signature Prompt Button
              if (widget.request.status == 'APPROVED_DRAFT')
                SizedBox(
                  width: double.infinity,
                  height: 52,
                  child: ElevatedButton.icon(
                    onPressed: _showSignatureBottomSheet,
                    icon: const Icon(Icons.edit_document, color: Colors.white),
                    label: Text(
                      'TANDATANGANI DRAFT PKWT',
                      style: GoogleFonts.plusJakartaSans(
                        fontWeight: FontWeight.w800,
                        color: Colors.white,
                        fontSize: 13,
                        letterSpacing: 0.5,
                      ),
                    ),
                    style: ElevatedButton.styleFrom(
                      backgroundColor: AppColors.primaryGreen,
                      shape: RoundedRectangleBorder(
                        borderRadius: BorderRadius.circular(16),
                      ),
                      elevation: 2,
                    ),
                  ),
                ),
            ],
          ),
        ),
      ),
    );
  }

  Widget _buildSectionTitle(String title) {
    return Text(
      title,
      style: GoogleFonts.plusJakartaSans(
        fontWeight: FontWeight.w900,
        fontSize: 14,
        color: AppColors.textPrimary,
      ),
    );
  }

  Widget _buildCard(List<Widget> children) {
    return Container(
      width: double.infinity,
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

  Widget _buildDetailRow(String label, String value) {
    return Padding(
      padding: const EdgeInsets.only(bottom: 14.0),
      child: Row(
        crossAxisAlignment: CrossAxisAlignment.start,
        children: [
          Expanded(
            flex: 4,
            child: Text(
              label,
              style: GoogleFonts.plusJakartaSans(
                fontWeight: FontWeight.w600,
                color: AppColors.textTertiary,
                fontSize: 12,
              ),
            ),
          ),
          const SizedBox(width: 8),
          Expanded(
            flex: 6,
            child: Text(
              value,
              textAlign: TextAlign.right,
              style: GoogleFonts.plusJakartaSans(
                fontWeight: FontWeight.bold,
                color: AppColors.textPrimary,
                fontSize: 12,
              ),
            ),
          ),
        ],
      ),
    );
  }

  Widget _buildStatusCard() {
    Color statusColor = Colors.grey.shade600;
    String statusText = widget.request.status;
    String description = '';

    if (widget.request.status == 'PENDING_HR') {
      statusColor = Colors.amber.shade800;
      statusText = 'Menunggu Review HR';
      description =
          'Pengajuan personal data Anda telah terkirim. Staf HR sedang meninjau dan melengkapi informasi pekerjaan Anda.';
    } else if (widget.request.status == 'APPROVED_DRAFT') {
      statusColor = Colors.blue.shade700;
      statusText = 'Siap Ditandatangani';
      description =
          'HR telah menerbitkan draf kontrak PKWT Anda. Harap baca informasi kontrak dengan saksama sebelum membubuhkan tanda tangan digital.';
    } else if (widget.request.status == 'ACTIVE') {
      statusColor = Colors.green.shade700;
      statusText = 'Aktif / Selesai';
      description =
          'Kontrak kerja PKWT telah ditandatangani oleh kedua belah pihak dan sudah aktif secara legal di dalam sistem.';
    } else if (widget.request.status == 'REJECTED') {
      statusColor = Colors.red.shade700;
      statusText = 'Ditolak';
      description =
          'Pengajuan personal data ditolak oleh pihak HR. Silakan hubungi admin atau ajukan ulang.';
    }

    return Container(
      width: double.infinity,
      padding: const EdgeInsets.all(20),
      decoration: BoxDecoration(
        color: statusColor.withOpacity(0.08),
        borderRadius: BorderRadius.circular(20),
        border: Border.all(color: statusColor.withOpacity(0.2), width: 1.5),
      ),
      child: Column(
        crossAxisAlignment: CrossAxisAlignment.start,
        children: [
          Row(
            children: [
              Icon(Icons.info_outline_rounded, color: statusColor, size: 20),
              const SizedBox(width: 8),
              Text(
                statusText,
                style: GoogleFonts.plusJakartaSans(
                  fontWeight: FontWeight.w900,
                  fontSize: 14,
                  color: statusColor,
                ),
              ),
            ],
          ),
          const SizedBox(height: 8),
          Text(
            description,
            style: GoogleFonts.plusJakartaSans(
              fontSize: 12,
              fontWeight: FontWeight.w600,
              color: AppColors.textSecondary,
              height: 1.5,
            ),
          ),
        ],
      ),
    );
  }

  Widget _buildSignatureDisplay() {
    // Determine the base url to parse relative signature urls
    final baseUrl = Uri.parse(AppConstants.baseUrl).origin;
    final sigUrl = widget.request.employeeSignatureUrl != null
        ? (widget.request.employeeSignatureUrl!.startsWith('http')
            ? widget.request.employeeSignatureUrl!
            : '$baseUrl${widget.request.employeeSignatureUrl!.startsWith('/') ? widget.request.employeeSignatureUrl : '/${widget.request.employeeSignatureUrl}'}')
        : null;

    return Container(
      width: double.infinity,
      padding: const EdgeInsets.all(16),
      decoration: BoxDecoration(
        color: Colors.white,
        borderRadius: BorderRadius.circular(20),
        border: Border.all(color: Colors.grey.shade200),
      ),
      child: Column(
        children: [
          Text(
            'Tanda Tangan Karyawan',
            style: GoogleFonts.plusJakartaSans(
              fontSize: 11,
              fontWeight: FontWeight.w800,
              color: AppColors.textTertiary,
            ),
          ),
          const SizedBox(height: 12),
          if (sigUrl != null)
            CachedNetworkImage(imageUrl: sigUrl, height: 100,
              fit: BoxFit.contain, placeholder: (context, url) => const Center(child: CircularProgressIndicator()), errorWidget: (context, url, error) => const Icon(Icons.error), memCacheWidth: 800,)
          else
            const Icon(Icons.edit_square, size: 48, color: Colors.grey),
          const SizedBox(height: 8),
          Text(
            'Telah ditandatangani secara elektronik',
            style: GoogleFonts.plusJakartaSans(
              fontSize: 10,
              fontWeight: FontWeight.w700,
              color: AppColors.primaryGreen,
            ),
          ),
        ],
      ),
    );
  }
}
