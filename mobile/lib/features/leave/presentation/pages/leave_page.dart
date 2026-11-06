import 'dart:typed_data';
import 'package:flutter/material.dart';
import 'package:flutter_bloc/flutter_bloc.dart';
import 'package:google_fonts/google_fonts.dart';
import 'package:intl/intl.dart';
import 'package:image_picker/image_picker.dart';
import 'package:hris_app/core/theme/app_colors.dart';
import "package:hris_app/core/utils/dialog_utils.dart";
import 'package:hris_app/core/utils/snackbar_utils.dart';
import 'package:hris_app/features/leave/presentation/bloc/leave_bloc.dart';
import 'package:hris_app/features/leave/presentation/bloc/leave_event.dart';
import 'package:hris_app/features/leave/presentation/bloc/leave_state.dart';
import 'package:hris_app/features/leave/domain/entities/leave_balance.dart';
import 'package:hris_app/features/leave/domain/entities/leave.dart';
import 'package:shimmer/shimmer.dart';

class LeavePage extends StatefulWidget {
  const LeavePage({super.key});

  @override
  State<LeavePage> createState() => _LeavePageState();
}

class _LeavePageState extends State<LeavePage> {
  @override
  void initState() {
    super.initState();
    _fetchData();
  }

  void _fetchData() {
    context.read<LeaveBloc>().add(const FetchLeaveBalancesRequested());
    context.read<LeaveBloc>().add(const FetchMyLeavesRequested());
  }

  @override
  Widget build(BuildContext context) {
    return DefaultTabController(
      length: 2,
      child: BlocListener<LeaveBloc, LeaveState>(
        listener: (context, state) {
          if (state.status == LeaveStatus.success &&
              state.actionMessage != null) {
            DialogUtils.showSuccess(
                context: context,
                title: "Berhasil",
                message: state.actionMessage!);
          } else if (state.status == LeaveStatus.failure &&
              state.actionMessage != null) {
            DialogUtils.showError(
                context: context,
                title: "Gagal",
                message: state.actionMessage!);
          }
        },
        child: Scaffold(
          backgroundColor: AppColors.backgroundAlt,
          body: NestedScrollView(
            headerSliverBuilder: (context, innerBoxIsScrolled) {
              return [
                _buildSliverAppBar(),
                SliverOverlapAbsorber(
                  handle:
                      NestedScrollView.sliverOverlapAbsorberHandleFor(context),
                  sliver: SliverPersistentHeader(
                    pinned: true,
                    delegate: _SliverTabDelegate(
                      TabBar(
                        labelColor: AppColors.primaryRed,
                        unselectedLabelColor: AppColors.textTertiary,
                        indicatorColor: AppColors.primaryRed,
                        indicatorWeight: 3,
                        labelStyle: GoogleFonts.plusJakartaSans(
                            fontWeight: FontWeight.w800,
                            fontSize: 13,
                            letterSpacing: 1),
                        tabs: const [
                          Tab(text: 'PENGAJUAN'),
                          Tab(text: 'RIWAYAT'),
                        ],
                      ),
                    ),
                  ),
                ),
              ];
            },
            body: const TabBarView(
              children: [
                LeaveFormTab(),
                LeaveHistoryTab(),
              ],
            ),
          ),
        ),
      ),
    );
  }

  Widget _buildSliverAppBar() {
    return SliverAppBar(
      pinned: true,
      backgroundColor: AppColors.primaryRed,
      elevation: 0,
      title: const Text('KELOLA CUTI',
          style: TextStyle(
              color: Colors.white,
              fontWeight: FontWeight.w800,
              fontSize: 16,
              letterSpacing: 1)),
      leading: IconButton(
        icon:
            const Icon(Icons.arrow_back_ios_new, color: Colors.white, size: 20),
        onPressed: () => Navigator.of(context).pop(),
      ),
    );
  }
}

class _SliverTabDelegate extends SliverPersistentHeaderDelegate {
  _SliverTabDelegate(this._tabBar);

  final TabBar _tabBar;

  @override
  double get minExtent => _tabBar.preferredSize.height;
  @override
  double get maxExtent => _tabBar.preferredSize.height;

  @override
  Widget build(
      BuildContext context, double shrinkOffset, bool overlapsContent) {
    return Container(
      color: Colors.white,
      child: _tabBar,
    );
  }

  @override
  bool shouldRebuild(_SliverTabDelegate oldDelegate) {
    return false;
  }
}

class LeaveFormTab extends StatefulWidget {
  final Leave? leave;
  const LeaveFormTab({super.key, this.leave});

  @override
  State<LeaveFormTab> createState() => _LeaveFormTabState();
}

class _LeaveFormTabState extends State<LeaveFormTab> {
  final _reasonController = TextEditingController();
  DateTime? _startDate;
  DateTime? _endDate;
  String? _selectedLeaveTypeId;
  final ImagePicker _picker = ImagePicker();
  XFile? _attachment;
  bool _isIzinMode = false;

  @override
  void initState() {
    super.initState();
    if (widget.leave != null) {
      _reasonController.text = widget.leave!.reason;
      _startDate = widget.leave!.startDate;
      _endDate = widget.leave!.endDate;
      _selectedLeaveTypeId = widget.leave!.leaveTypeId;
      // Determine mode based on initial leave type
      WidgetsBinding.instance.addPostFrameCallback((_) {
        final balances = context.read<LeaveBloc>().state.balances;
        if (balances.isNotEmpty) {
          final balance = balances.firstWhere(
              (b) => b.leaveTypeId == _selectedLeaveTypeId,
              orElse: () => balances.first);
          setState(() {
            _isIzinMode = !balance.requiresQuota;
          });
        }
      });
    }
  }

  @override
  void dispose() {
    _reasonController.dispose();
    super.dispose();
  }

  void _presentDatePicker(bool isStart) async {
    final now = DateTime.now();
    final today = DateTime(now.year, now.month, now.day);
    
    // Cek apakah tipe pengajuan adalah sakit atau musibah
    bool allowPastDate = false;
    if (_selectedLeaveTypeId != null) {
      final balances = context.read<LeaveBloc>().state.balances;
      try {
        final selectedType = balances.firstWhere((b) => b.leaveTypeId == _selectedLeaveTypeId);
        final typeName = selectedType.leaveTypeName.toLowerCase();
        if (typeName.contains('sakit') || typeName.contains('musibah')) {
          allowPastDate = true;
        }
      } catch (_) {}
    }

    // Batas minimum sesuai kebijakan
    final minDateForType = allowPastDate ? today.subtract(const Duration(days: 30)) : today;
    
    // Jika memilih tanggal selesai, tidak boleh sebelum tanggal mulai (jika sudah dipilih)
    final firstDate = isStart ? minDateForType : (_startDate ?? minDateForType);
    
    DateTime initial = isStart ? (_startDate ?? today) : (_endDate ?? _startDate ?? today);
    if (initial.isBefore(firstDate)) {
      initial = firstDate;
    }

    final pickedDate = await showDatePicker(
      context: context,
      initialDate: initial,
      firstDate: firstDate,
      lastDate: today.add(const Duration(days: 365)),
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
    if (pickedDate != null) {
      setState(() {
        if (isStart) {
          _startDate = pickedDate;
          // Reset endDate jika endDate yang sudah dipilih sebelumnya berada sebelum startDate yang baru
          if (_endDate != null && _endDate!.isBefore(_startDate!)) {
            _endDate = null;
          }
        } else {
          _endDate = pickedDate;
        }
      });
    }
  }

  void _pickImage(ImageSource source) async {
    final XFile? image = await _picker.pickImage(
      source: source,
      maxWidth: 1280,
      maxHeight: 1280,
      imageQuality: 70,
    );
    if (image != null) {
      setState(() {
        _attachment = image;
      });
    }
  }

  void _showImageSourceActionSheet(BuildContext context) {
    showModalBottomSheet(
      context: context,
      shape: const RoundedRectangleBorder(
        borderRadius: BorderRadius.vertical(top: Radius.circular(20)),
      ),
      builder: (BuildContext context) {
        return SafeArea(
          child: Column(
            mainAxisSize: MainAxisSize.min,
            children: [
              const SizedBox(height: 16),
              Text(
                'Pilih Sumber Lampiran',
                style: GoogleFonts.plusJakartaSans(
                    fontSize: 18, fontWeight: FontWeight.bold),
              ),
              const SizedBox(height: 16),
              ListTile(
                leading: const Icon(Icons.camera_alt_rounded,
                    color: AppColors.primaryRed),
                title: Text('Kamera',
                    style: GoogleFonts.plusJakartaSans(
                        fontWeight: FontWeight.w600)),
                onTap: () {
                  Navigator.pop(context);
                  _pickImage(ImageSource.camera);
                },
              ),
              ListTile(
                leading: const Icon(Icons.photo_library_rounded,
                    color: AppColors.primaryRed),
                title: Text('Galeri',
                    style: GoogleFonts.plusJakartaSans(
                        fontWeight: FontWeight.w600)),
                onTap: () {
                  Navigator.pop(context);
                  _pickImage(ImageSource.gallery);
                },
              ),
              const SizedBox(height: 16),
            ],
          ),
        );
      },
    );
  }

  void _submit() async {
    if (_startDate == null) {
      SnackBarUtils.showError(context, 'Pilih tanggal mulai.');
      return;
    }
    if (_endDate == null) {
      SnackBarUtils.showError(context, 'Pilih tanggal selesai.');
      return;
    }
    if (_selectedLeaveTypeId == null) {
      SnackBarUtils.showError(context, 'Pilih tipe cuti atau izin.');
      return;
    }
    if (_reasonController.text.trim().isEmpty) {
      SnackBarUtils.showError(context, 'Tuliskan alasan pengajuan anda.');
      return;
    }
    
    // Validasi double check untuk pengajuan masa lalu
    final now = DateTime.now();
    final today = DateTime(now.year, now.month, now.day);
    
    bool isSicknessOrDisaster = false;
    final balances = context.read<LeaveBloc>().state.balances;
    try {
      final selectedType = balances.firstWhere((b) => b.leaveTypeId == _selectedLeaveTypeId);
      final typeName = selectedType.leaveTypeName.toLowerCase();
      if (typeName.contains('sakit') || typeName.contains('musibah')) {
        isSicknessOrDisaster = true;
      }
    } catch (_) {}

    if (!isSicknessOrDisaster && _startDate!.isBefore(today)) {
      SnackBarUtils.showError(context, 'Tipe pengajuan ini tidak mengizinkan pengajuan mundur (ke masa lalu).');
      return;
    }

    Uint8List? attachmentBytes;
    String? attachmentName;
    if (_attachment != null) {
      attachmentBytes = await _attachment!.readAsBytes();
      attachmentName = _attachment!.name;
    }

    if (widget.leave != null) {
      context.read<LeaveBloc>().add(UpdateLeaveRequested(
            leaveId: widget.leave!.id,
            leaveTypeId: _selectedLeaveTypeId!,
            startDate: DateFormat('yyyy-MM-dd').format(_startDate!),
            endDate: DateFormat('yyyy-MM-dd').format(_endDate!),
            reason: _reasonController.text,
            attachmentBytes: attachmentBytes,
            attachmentName: attachmentName,
          ));
    } else {
      context.read<LeaveBloc>().add(SubmitLeaveRequested(
            leaveTypeId: _selectedLeaveTypeId!,
            startDate: DateFormat('yyyy-MM-dd').format(_startDate!),
            endDate: DateFormat('yyyy-MM-dd').format(_endDate!),
            reason: _reasonController.text,
            attachmentBytes: attachmentBytes,
            attachmentName: attachmentName,
          ));
    }
  }

  @override
  Widget build(BuildContext context) {
    return BlocListener<LeaveBloc, LeaveState>(
      listener: (context, state) {
        if (state.status == LeaveStatus.success &&
            state.actionMessage != null) {
          _reasonController.clear();
          setState(() {
            _startDate = null;
            _endDate = null;
            _selectedLeaveTypeId = null;
            _attachment = null;
          });
        }
      },
      child: Builder(builder: (context) {
        return CustomScrollView(
          slivers: [
            if (widget.leave == null)
              SliverOverlapInjector(
                handle:
                    NestedScrollView.sliverOverlapAbsorberHandleFor(context),
              ),
            SliverPadding(
              padding: const EdgeInsets.all(24),
              sliver: SliverToBoxAdapter(
                child: Column(
                  crossAxisAlignment: CrossAxisAlignment.start,
                  children: [
                    _buildModeSelector(),
                    const SizedBox(height: 24),
                    if (!_isIzinMode) ...[
                      _buildBalanceHeader(),
                      const SizedBox(height: 24),
                    ],
                    _buildSectionHeader('DETAIL LAYANAN'),
                    const SizedBox(height: 12),
                    _buildLeaveTypeSelection(),
                    const SizedBox(height: 24),
                    _buildSectionHeader('WAKTU PENGAJUAN'),
                    const SizedBox(height: 12),
                    _buildDateSelectors(),
                    const SizedBox(height: 24),
                    _buildSectionHeader('ALASAN & LAMPIRAN'),
                    const SizedBox(height: 12),
                    _buildReasonField(),
                    const SizedBox(height: 16),
                    _buildAttachmentPicker(),
                    const SizedBox(height: 48),
                    _buildSubmitButton(),
                    const SizedBox(height: 40),
                  ],
                ),
              ),
            ),
          ],
        );
      }),
    );
  }

  Widget _buildSectionHeader(String title) {
    return Text(
      title,
      style: GoogleFonts.plusJakartaSans(
          fontSize: 10,
          fontWeight: FontWeight.w900,
          color: AppColors.textTertiary,
          letterSpacing: 1.5),
    );
  }

  Widget _buildModeSelector() {
    return Container(
      padding: const EdgeInsets.all(6),
      decoration: BoxDecoration(
        color: AppColors.grayLight.withOpacity(0.5),
        borderRadius: BorderRadius.circular(20),
      ),
      child: Row(
        children: [
          _buildModeItem('Cuti', false, Icons.calendar_today_rounded,
              AppColors.primaryRed),
          _buildModeItem('Izin / Sakit', true, Icons.medical_services_rounded,
              Colors.orange),
        ],
      ),
    );
  }

  Widget _buildModeItem(String label, bool isIzin, IconData icon, Color color) {
    final isSelected = _isIzinMode == isIzin;
    return Expanded(
      child: GestureDetector(
        onTap: () {
          setState(() {
            _isIzinMode = isIzin;
            _selectedLeaveTypeId = null; // Reset selection when switching mode
          });
        },
        child: AnimatedContainer(
          duration: const Duration(milliseconds: 200),
          padding: const EdgeInsets.symmetric(vertical: 12),
          decoration: BoxDecoration(
            color: isSelected ? Colors.white : Colors.transparent,
            borderRadius: BorderRadius.circular(16),
            boxShadow: isSelected
                ? [
                    BoxShadow(
                        color: Colors.black.withOpacity(0.05),
                        blurRadius: 10,
                        offset: const Offset(0, 4))
                  ]
                : [],
          ),
          child: Row(
            mainAxisAlignment: MainAxisAlignment.center,
            children: [
              Icon(icon,
                  size: 16, color: isSelected ? color : AppColors.textTertiary),
              const SizedBox(width: 8),
              Text(
                label,
                style: GoogleFonts.plusJakartaSans(
                  fontSize: 13,
                  fontWeight: isSelected ? FontWeight.w800 : FontWeight.w700,
                  color: isSelected
                      ? AppColors.textPrimary
                      : AppColors.textTertiary,
                ),
              ),
            ],
          ),
        ),
      ),
    );
  }

  Widget _buildBalanceHeader() {
    return BlocBuilder<LeaveBloc, LeaveState>(
      builder: (context, state) {
        final annualLeave = state.balances.firstWhere(
          (b) =>
              b.leaveTypeName.toLowerCase().contains('cuti tahunan') ||
              b.leaveTypeName.toLowerCase().contains('annual'),
          orElse: () => state.balances.isNotEmpty
              ? state.balances.first
              : const LeaveBalance(
                  leaveTypeId: '',
                  leaveTypeName: 'Cuti',
                  isPaid: true,
                  requiresQuota: true,
                  total: 0,
                  used: 0,
                  remaining: 0),
        );

        return Container(
          width: double.infinity,
          padding: const EdgeInsets.all(24),
          decoration: BoxDecoration(
            gradient: LinearGradient(
              colors: [
                AppColors.primaryRed,
                AppColors.primaryRed.withOpacity(0.8)
              ],
              begin: Alignment.topLeft,
              end: Alignment.bottomRight,
            ),
            borderRadius: BorderRadius.circular(28),
            boxShadow: [
              BoxShadow(
                  color: AppColors.primaryRed.withOpacity(0.3),
                  blurRadius: 20,
                  offset: const Offset(0, 10))
            ],
          ),
          child: Stack(
            children: [
              Positioned(
                right: -20,
                bottom: -20,
                child: Icon(Icons.beach_access_rounded,
                    size: 100, color: Colors.white.withOpacity(0.1)),
              ),
              Column(
                crossAxisAlignment: CrossAxisAlignment.start,
                children: [
                  const Text('Saldo Cuti Tahunan',
                      style: TextStyle(
                          color: Colors.white70,
                          fontSize: 13,
                          fontWeight: FontWeight.w700)),
                  const SizedBox(height: 8),
                  Row(
                    crossAxisAlignment: CrossAxisAlignment.end,
                    children: [
                      Text('${annualLeave.remaining}',
                          style: const TextStyle(
                              color: Colors.white,
                              fontSize: 42,
                              fontWeight: FontWeight.w900,
                              height: 1)),
                      const SizedBox(width: 8),
                      const Padding(
                        padding: EdgeInsets.only(bottom: 8),
                        child: Text('HARI',
                            style: TextStyle(
                                color: Colors.white,
                                fontSize: 14,
                                fontWeight: FontWeight.w800,
                                letterSpacing: 1)),
                      ),
                    ],
                  ),
                  const SizedBox(height: 16),
                  Container(
                    padding:
                        const EdgeInsets.symmetric(horizontal: 12, vertical: 6),
                    decoration: BoxDecoration(
                        color: Colors.white.withOpacity(0.2),
                        borderRadius: BorderRadius.circular(10)),
                    child: Text('Total Jatah: ${annualLeave.total} Hari',
                        style: const TextStyle(
                            color: Colors.white,
                            fontSize: 11,
                            fontWeight: FontWeight.w700)),
                  ),
                ],
              ),
            ],
          ),
        );
      },
    );
  }

  Widget _buildLeaveTypeSelection() {
    return BlocBuilder<LeaveBloc, LeaveState>(
      builder: (context, state) {
        // Filter balances based on mode
        List<LeaveBalance> allBalances = state.balances;
        List<LeaveBalance> filteredBalances = allBalances.where((b) {
          final typeName = b.leaveTypeName.toLowerCase();
          final isIzinType = !b.requiresQuota ||
              typeName.contains('izin') ||
              typeName.contains('ijin') ||
              typeName.contains('sakit') ||
              typeName.contains('musibah') ||
              typeName.contains('dispensasi') ||
              typeName.contains('tugas') ||
              typeName.contains('dinas') ||
              typeName.contains('unpaid');
          return _isIzinMode ? isIzinType : !isIzinType;
        }).toList();

        // Ensure uniqueness to prevent Flutter Dropdown duplicate value crashes
        final uniqueMap = <String, LeaveBalance>{};
        for (var b in filteredBalances) {
          if (!uniqueMap.containsKey(b.leaveTypeId) &&
              b.leaveTypeId.isNotEmpty) {
            uniqueMap[b.leaveTypeId] = b;
          }
        }
        filteredBalances = uniqueMap.values.toList();

        if (state.status == LeaveStatus.loading && allBalances.isEmpty) {
          return Container(
            height: 60,
            decoration: BoxDecoration(
                color: Colors.white,
                borderRadius: BorderRadius.circular(24),
                border: Border.all(color: AppColors.grayBorder)),
            child: const Center(
                child: SizedBox(
                    width: 20,
                    height: 20,
                    child: CircularProgressIndicator(
                        strokeWidth: 2, color: AppColors.primaryRed))),
          );
        }

        if (state.status == LeaveStatus.failure && allBalances.isEmpty) {
          return InkWell(
            onTap: () => context
                .read<LeaveBloc>()
                .add(const FetchLeaveBalancesRequested()),
            child: Container(
              padding: const EdgeInsets.all(16),
              decoration: BoxDecoration(
                  color: AppColors.error.withOpacity(0.05),
                  borderRadius: BorderRadius.circular(20),
                  border: Border.all(color: AppColors.error.withOpacity(0.2))),
              child: Column(
                children: [
                  Text('Gagal memuat tipe pengajuan: ${state.message}',
                      style: const TextStyle(
                          fontSize: 12,
                          color: AppColors.error,
                          fontWeight: FontWeight.w600),
                      textAlign: TextAlign.center),
                  const SizedBox(height: 8),
                  const Text('Ketuk untuk coba lagi',
                      style: TextStyle(
                          fontSize: 10,
                          color: AppColors.error,
                          fontWeight: FontWeight.w800,
                          letterSpacing: 1)),
                ],
              ),
            ),
          );
        }

        if (filteredBalances.isEmpty) {
          final debugInfo = allBalances.isEmpty
              ? "List Kosong"
              : "${allBalances.length} tipe, 0 mode matches";
          return Container(
            padding: const EdgeInsets.all(20),
            decoration: BoxDecoration(
                color: AppColors.grayLight.withOpacity(0.3),
                borderRadius: BorderRadius.circular(24)),
            child: Column(
              children: [
                Icon(Icons.info_outline_rounded,
                    color: AppColors.textTertiary.withOpacity(0.5), size: 24),
                const SizedBox(height: 12),
                Text(
                  'Belum ada tipe ${_isIzinMode ? 'Izin' : 'Cuti'} tersedia.',
                  style: const TextStyle(
                      fontSize: 14,
                      fontWeight: FontWeight.w700,
                      color: AppColors.textPrimary),
                  textAlign: TextAlign.center,
                ),
                Text(
                  'Hubungi HRD untuk konfigurasi tipe pengajuan ($debugInfo).',
                  style: const TextStyle(
                      fontSize: 11, color: AppColors.textTertiary),
                ),
              ],
            ),
          );
        }

        // Auto-select if nothing selected or current selection is not in filtered list
        if (_selectedLeaveTypeId == null ||
            !filteredBalances
                .any((b) => b.leaveTypeId == _selectedLeaveTypeId)) {
          // Use a small delay to avoid setState during build
          Future.microtask(() {
            if (mounted && filteredBalances.isNotEmpty) {
              setState(() {
                _selectedLeaveTypeId = filteredBalances.first.leaveTypeId;
              });
            }
          });
        }

        return Container(
          decoration: BoxDecoration(
            color: Colors.white,
            borderRadius: BorderRadius.circular(24),
            border: Border.all(color: AppColors.grayBorder),
          ),
          child: Column(
            children: [
              Padding(
                padding:
                    const EdgeInsets.symmetric(horizontal: 20, vertical: 4),
                child: DropdownButtonFormField<String>(
                  initialValue: _selectedLeaveTypeId,
                  icon: const Icon(Icons.keyboard_arrow_down_rounded,
                      color: AppColors.textTertiary),
                  decoration: const InputDecoration(border: InputBorder.none),
                  hint: Text('Pilih Tipe ${_isIzinMode ? 'Izin' : 'Cuti'}',
                      style: GoogleFonts.plusJakartaSans(
                          color: AppColors.textTertiary,
                          fontWeight: FontWeight.w700,
                          fontSize: 14)),
                  items: filteredBalances
                      .map((b) => DropdownMenuItem(
                          value: b.leaveTypeId,
                          child: Text(b.leaveTypeName,
                              style: const TextStyle(
                                  fontSize: 14, fontWeight: FontWeight.w700))))
                      .toList(),
                  onChanged: (v) => setState(() => _selectedLeaveTypeId = v),
                ),
              ),
              if (_selectedLeaveTypeId != null &&
                  filteredBalances
                      .any((b) => b.leaveTypeId == _selectedLeaveTypeId)) ...[
                const Divider(height: 1),
                _buildPayInfo(filteredBalances
                    .firstWhere((b) => b.leaveTypeId == _selectedLeaveTypeId)),
              ],
            ],
          ),
        );
      },
    );
  }

  Widget _buildPayInfo(LeaveBalance balance) {
    final isPaid = balance.isPaid;
    final requiresQuota = balance.requiresQuota;

    String message = isPaid
        ? 'Gaji tetap dibayarkan untuk tipe ini.'
        : 'Terdapat pemotongan gaji pokok.';
    if (!requiresQuota) {
      message = isPaid
          ? 'Izin ini dibayar & tidak memotong jatah cuti.'
          : 'Izin ini tidak memotong jatah cuti (Unpaid).';
    }

    return Container(
      padding: const EdgeInsets.all(20),
      color: (isPaid ? Colors.blue : Colors.orange).withOpacity(0.03),
      child: Row(
        children: [
          Icon(isPaid ? Icons.check_circle_rounded : Icons.info_rounded,
              color: isPaid ? Colors.blue : Colors.orange, size: 18),
          const SizedBox(width: 12),
          Expanded(
            child: Text(
              message,
              style: TextStyle(
                  fontSize: 12,
                  fontWeight: FontWeight.w800,
                  color: isPaid ? Colors.blue : Colors.orange),
            ),
          ),
        ],
      ),
    );
  }

  Widget _buildDateSelectors() {
    return Container(
      padding: const EdgeInsets.all(20),
      decoration: BoxDecoration(
        color: Colors.white,
        borderRadius: BorderRadius.circular(24),
        border: Border.all(color: AppColors.grayBorder),
      ),
      child: Column(
        children: [
          _buildDateField('MULAI', _startDate, () => _presentDatePicker(true)),
          const Padding(
            padding: EdgeInsets.symmetric(vertical: 12),
            child: Divider(height: 1),
          ),
          _buildDateField('SELESAI', _endDate, () => _presentDatePicker(false)),
        ],
      ),
    );
  }

  Widget _buildDateField(String label, DateTime? date, VoidCallback onTap) {
    return InkWell(
      onTap: onTap,
      child: Row(
        mainAxisAlignment: MainAxisAlignment.spaceBetween,
        children: [
          Column(
            crossAxisAlignment: CrossAxisAlignment.start,
            children: [
              Text(label,
                  style: const TextStyle(
                      fontSize: 10,
                      fontWeight: FontWeight.w900,
                      color: AppColors.textTertiary,
                      letterSpacing: 1)),
              const SizedBox(height: 4),
              Text(
                  date != null
                      ? DateFormat('dd MMMM yyyy').format(date)
                      : 'Pilih Tanggal',
                  style: const TextStyle(
                      fontSize: 14,
                      fontWeight: FontWeight.w800,
                      color: AppColors.textPrimary)),
            ],
          ),
          const Icon(Icons.calendar_today_rounded,
              size: 20, color: AppColors.textTertiary),
        ],
      ),
    );
  }

  Widget _buildReasonField() {
    return Container(
      padding: const EdgeInsets.all(8),
      decoration: BoxDecoration(
        color: Colors.white,
        borderRadius: BorderRadius.circular(24),
        border: Border.all(color: AppColors.grayBorder),
      ),
      child: TextField(
        controller: _reasonController,
        maxLines: 4,
        style: const TextStyle(fontSize: 14, fontWeight: FontWeight.w600),
        decoration: InputDecoration(
          hintText: 'Tuliskan alasan pengajuan anda...',
          hintStyle: TextStyle(
              color: AppColors.textTertiary.withOpacity(0.5), fontSize: 13),
          border: InputBorder.none,
          contentPadding: const EdgeInsets.all(16),
        ),
      ),
    );
  }

  Widget _buildAttachmentPicker() {
    return InkWell(
      onTap: () => _showImageSourceActionSheet(context),
      child: Container(
        padding: const EdgeInsets.all(20),
        decoration: BoxDecoration(
          color: AppColors.grayLight.withOpacity(0.3),
          borderRadius: BorderRadius.circular(20),
          border:
              Border.all(color: AppColors.grayBorder, style: BorderStyle.solid),
        ),
        child: Row(
          children: [
            Container(
                padding: const EdgeInsets.all(8),
                decoration: BoxDecoration(
                    color: AppColors.primaryRed.withOpacity(0.1),
                    shape: BoxShape.circle),
                child: const Icon(Icons.attachment_rounded,
                    size: 18, color: AppColors.primaryRed)),
            const SizedBox(width: 16),
            Expanded(
                child: Text(
                    _attachment != null
                        ? _attachment!.name
                        : (widget.leave?.attachmentUrl != null
                            ? 'Lampiran: ${widget.leave!.attachmentUrl!.split('/').last}'
                            : 'Lampiran / Surat Dokter (Opsional)'),
                    style: TextStyle(
                        fontSize: 13,
                        fontWeight: FontWeight.w700,
                        color: (_attachment == null &&
                                widget.leave?.attachmentUrl == null)
                            ? AppColors.textSecondary
                            : AppColors.textPrimary),
                    overflow: TextOverflow.ellipsis)),
            if (_attachment != null)
              IconButton(
                  icon: const Icon(Icons.close_rounded, size: 18),
                  onPressed: () => setState(() => _attachment = null)),
          ],
        ),
      ),
    );
  }

  Widget _buildSubmitButton() {
    return BlocBuilder<LeaveBloc, LeaveState>(
      builder: (context, state) {
        final isLoading =
            state.status == LeaveStatus.loading && state.actionMessage == null;
        return SizedBox(
          width: double.infinity,
          height: 60,
          child: ElevatedButton(
            onPressed: isLoading ? null : _submit,
            style: ElevatedButton.styleFrom(
              backgroundColor: AppColors.primaryRed,
              shape: RoundedRectangleBorder(
                  borderRadius: BorderRadius.circular(20)),
              elevation: 4,
              shadowColor: AppColors.primaryRed.withOpacity(0.4),
            ),
            child: isLoading
                ? const SizedBox(
                    width: 24,
                    height: 24,
                    child: CircularProgressIndicator(
                        color: Colors.white, strokeWidth: 3))
                : Text(
                    widget.leave != null
                        ? 'PERBARUI PENGAJUAN'
                        : 'AJUKAN SEKARANG',
                    style: const TextStyle(
                        fontWeight: FontWeight.w900,
                        color: Colors.white,
                        letterSpacing: 1.5,
                        fontSize: 13)),
          ),
        );
      },
    );
  }
}

class LeaveHistoryTab extends StatelessWidget {
  const LeaveHistoryTab({super.key});

  @override
  Widget build(BuildContext context) {
    return BlocBuilder<LeaveBloc, LeaveState>(
      builder: (context, state) {
        List<Leave> leaves = state.leaves;

        if (state.status == LeaveStatus.loading && leaves.isEmpty) {
          return _buildLoadingShimmer();
        }

        if (leaves.isEmpty) {
          return Center(
            child: Column(
              mainAxisAlignment: MainAxisAlignment.center,
              children: [
                Icon(Icons.event_busy_rounded,
                    size: 64, color: AppColors.textTertiary.withOpacity(0.3)),
                const SizedBox(height: 16),
                const Text('Belum ada riwayat pengajuan',
                    style: TextStyle(
                        fontWeight: FontWeight.w800,
                        color: AppColors.textTertiary)),
              ],
            ),
          );
        }

        return Builder(builder: (context) {
          return RefreshIndicator(
            onRefresh: () async =>
                context.read<LeaveBloc>().add(const FetchMyLeavesRequested()),
            child: CustomScrollView(
              slivers: [
                SliverOverlapInjector(
                  handle:
                      NestedScrollView.sliverOverlapAbsorberHandleFor(context),
                ),
                SliverPadding(
                  padding: const EdgeInsets.all(24),
                  sliver: SliverList(
                    delegate: SliverChildBuilderDelegate(
                      (context, index) {
                        final leave = leaves[index];
                        return _buildLeaveRequestCard(context, leave);
                      },
                      childCount: leaves.length,
                    ),
                  ),
                ),
              ],
            ),
          );
        });
      },
    );
  }

  Widget _buildLeaveRequestCard(BuildContext context, Leave leave) {
    Color statusColor;
    IconData statusIcon;
    switch (leave.status.toLowerCase()) {
      case 'approved':
        statusColor = const Color(0xFF10B981);
        statusIcon = Icons.check_circle_rounded;
        break;
      case 'rejected':
        statusColor = AppColors.error;
        statusIcon = Icons.cancel_rounded;
        break;
      default:
        statusColor = Colors.orange;
        statusIcon = Icons.pending_rounded;
    }

    return Container(
      margin: const EdgeInsets.only(bottom: 16),
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
        children: [
          Padding(
            padding: const EdgeInsets.all(20),
            child: Row(
              crossAxisAlignment: CrossAxisAlignment.start,
              children: [
                Container(
                  padding: const EdgeInsets.all(10),
                  decoration: BoxDecoration(
                      color: statusColor.withOpacity(0.1),
                      borderRadius: BorderRadius.circular(14)),
                  child: Icon(statusIcon, color: statusColor, size: 24),
                ),
                const SizedBox(width: 16),
                Expanded(
                  child: Column(
                    crossAxisAlignment: CrossAxisAlignment.start,
                    children: [
                      Text(leave.leaveTypeName ?? 'Cuti',
                          style: const TextStyle(
                              fontWeight: FontWeight.w900,
                              fontSize: 16,
                              color: AppColors.textPrimary)),
                      const SizedBox(height: 4),
                      Text(
                          '${DateFormat('dd MMM').format(leave.startDate)} - ${DateFormat('dd MMM yyyy').format(leave.endDate)}',
                          style: const TextStyle(
                              fontSize: 12,
                              fontWeight: FontWeight.w700,
                              color: AppColors.textSecondary)),
                    ],
                  ),
                ),
                Container(
                  padding:
                      const EdgeInsets.symmetric(horizontal: 10, vertical: 4),
                  decoration: BoxDecoration(
                      color: statusColor.withOpacity(0.1),
                      borderRadius: BorderRadius.circular(8)),
                  child: Text(leave.status.toUpperCase(),
                      style: TextStyle(
                          fontSize: 10,
                          fontWeight: FontWeight.w900,
                          color: statusColor,
                          letterSpacing: 0.5)),
                ),
                if (leave.status.toLowerCase() == 'pending')
                  _buildActionMenu(context, leave),
              ],
            ),
          ),
          const Divider(height: 1),
          Container(
            width: double.infinity,
            padding: const EdgeInsets.symmetric(horizontal: 20, vertical: 12),
            color: AppColors.grayLight.withOpacity(0.2),
            child: Text(
              'Alasan Pengajuan: ${leave.reason}',
              style: const TextStyle(
                  fontSize: 12,
                  fontWeight: FontWeight.w600,
                  color: AppColors.textSecondary,
                  fontStyle: FontStyle.italic),
            ),
          ),
          if (leave.status.toLowerCase() == 'rejected') ...[
            const Divider(height: 1),
            Container(
              width: double.infinity,
              padding: const EdgeInsets.symmetric(horizontal: 20, vertical: 12),
              decoration: BoxDecoration(
                color: AppColors.error.withOpacity(0.06),
                borderRadius: const BorderRadius.vertical(bottom: Radius.circular(24)),
              ),
              child: Row(
                crossAxisAlignment: CrossAxisAlignment.start,
                children: [
                  const Icon(Icons.info_outline_rounded,
                      size: 16, color: AppColors.error),
                  const SizedBox(width: 8),
                  Expanded(
                    child: Column(
                      crossAxisAlignment: CrossAxisAlignment.start,
                      children: [
                        const Text(
                          'Alasan Penolakan HR:',
                          style: TextStyle(
                              fontSize: 11,
                              fontWeight: FontWeight.w800,
                              color: AppColors.error),
                        ),
                        const SizedBox(height: 2),
                        Text(
                          leave.rejectReason != null && leave.rejectReason!.isNotEmpty
                              ? leave.rejectReason!
                              : 'Tidak ada alasan penolakan yang dicantumkan.',
                          style: const TextStyle(
                              fontSize: 12,
                              fontWeight: FontWeight.w600,
                              color: AppColors.textPrimary),
                        ),
                        if (leave.approverName != null &&
                            leave.approverName!.isNotEmpty) ...[
                          const SizedBox(height: 4),
                          Text(
                            'Ditolak oleh: ${leave.approverName}',
                            style: TextStyle(
                                fontSize: 10,
                                fontWeight: FontWeight.w600,
                                color: AppColors.error.withOpacity(0.8)),
                          ),
                        ],
                      ],
                    ),
                  ),
                ],
              ),
            ),
          ] else if (leave.status.toLowerCase() == 'approved' &&
              leave.approverName != null &&
              leave.approverName!.isNotEmpty) ...[
            const Divider(height: 1),
            Container(
              width: double.infinity,
              padding: const EdgeInsets.symmetric(horizontal: 20, vertical: 8),
              decoration: BoxDecoration(
                color: const Color(0xFF10B981).withOpacity(0.06),
                borderRadius: const BorderRadius.vertical(bottom: Radius.circular(24)),
              ),
              child: Row(
                children: [
                  const Icon(Icons.check_circle_outline_rounded,
                      size: 14, color: Color(0xFF10B981)),
                  const SizedBox(width: 6),
                  Text(
                    'Disetujui oleh: ${leave.approverName}',
                    style: const TextStyle(
                        fontSize: 11,
                        fontWeight: FontWeight.w700,
                        color: Color(0xFF047857)),
                  ),
                ],
              ),
            ),
          ],
        ],
      ),
    );
  }

  Widget _buildActionMenu(BuildContext context, Leave leave) {
    return PopupMenuButton<String>(
      icon: const Icon(Icons.more_vert_rounded, color: AppColors.textTertiary),
      onSelected: (value) {
        if (value == 'edit') {
          Navigator.push(
            context,
            MaterialPageRoute(
              builder: (context) => EditLeavePage(leave: leave),
            ),
          );
        } else if (value == 'delete') {
          _showDeleteConfirmation(context, leave.id);
        }
      },
      itemBuilder: (context) => [
        const PopupMenuItem(
            value: 'edit',
            child: Row(children: [
              Icon(Icons.edit_rounded, size: 18),
              SizedBox(width: 8),
              Text('Edit',
                  style: TextStyle(fontSize: 13, fontWeight: FontWeight.w700))
            ])),
        const PopupMenuItem(
            value: 'delete',
            child: Row(children: [
              Icon(Icons.delete_rounded, size: 18, color: Colors.red),
              SizedBox(width: 8),
              Text('Batalkan',
                  style: TextStyle(
                      fontSize: 13,
                      fontWeight: FontWeight.w700,
                      color: Colors.red))
            ])),
      ],
    );
  }

  void _showDeleteConfirmation(BuildContext context, String leaveId) {
    showDialog(
      context: context,
      builder: (context) => AlertDialog(
        title: const Text('Batalkan Pengajuan',
            style: TextStyle(fontWeight: FontWeight.bold)),
        content: const Text(
            'Apakah anda yakin ingin membatalkan pengajuan cuti ini? Kuota cuti akan dikembalikan otomatis.'),
        actions: [
          TextButton(
              onPressed: () => Navigator.pop(context),
              child: const Text('Tidak')),
          TextButton(
              onPressed: () {
                context.read<LeaveBloc>().add(DeleteLeaveRequested(leaveId));
                Navigator.pop(context);
              },
              child: const Text('Ya, Batalkan',
                  style: TextStyle(color: Colors.red))),
        ],
      ),
    );
  }

  Widget _buildLoadingShimmer() {
    return Shimmer.fromColors(
      baseColor: AppColors.grayLight,
      highlightColor: Colors.white,
      child: ListView.builder(
        padding: const EdgeInsets.all(24),
        itemCount: 3,
        itemBuilder: (_, __) => Container(
          height: 120,
          margin: const EdgeInsets.only(bottom: 16),
          decoration: BoxDecoration(
              color: Colors.white, borderRadius: BorderRadius.circular(24)),
        ),
      ),
    );
  }
}

class EditLeavePage extends StatelessWidget {
  final Leave leave;
  const EditLeavePage({super.key, required this.leave});

  @override
  Widget build(BuildContext context) {
    return Scaffold(
      backgroundColor: AppColors.backgroundAlt,
      appBar: AppBar(
        title: Column(
          children: [
            Text('EDIT PENGAJUAN',
                style: GoogleFonts.plusJakartaSans(
                    color: AppColors.textPrimary,
                    fontWeight: FontWeight.w900,
                    fontSize: 13,
                    letterSpacing: 1.5)),
            const SizedBox(height: 2),
            Container(
              padding: const EdgeInsets.symmetric(horizontal: 8, vertical: 2),
              decoration: BoxDecoration(
                  color: Colors.orange.withOpacity(0.1),
                  borderRadius: BorderRadius.circular(6)),
              child: Text(leave.status.toUpperCase(),
                  style: GoogleFonts.plusJakartaSans(
                      fontSize: 9,
                      fontWeight: FontWeight.w900,
                      color: Colors.orange,
                      letterSpacing: 0.5)),
            ),
          ],
        ),
        centerTitle: true,
        backgroundColor: Colors.white,
        elevation: 0,
        leading: IconButton(
          icon: const Icon(Icons.arrow_back_ios_new_rounded,
              color: AppColors.textPrimary, size: 18),
          onPressed: () => Navigator.pop(context),
        ),
      ),
      body: BlocListener<LeaveBloc, LeaveState>(
        listener: (context, state) {
          if (state.status == LeaveStatus.success &&
              state.actionMessage != null) {
            Navigator.pop(context);
          }
        },
        child: LeaveFormTab(leave: leave),
      ),
    );
  }
}
