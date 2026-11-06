import 'package:cached_network_image/cached_network_image.dart';
import 'dart:io';
import 'package:flutter/foundation.dart';
import 'package:flutter/material.dart';
import 'package:flutter_bloc/flutter_bloc.dart';
import 'package:google_fonts/google_fonts.dart';
import 'package:image_picker/image_picker.dart';
import 'package:hris_app/features/overtime/domain/entities/overtime.dart';
import 'package:hris_app/features/overtime/presentation/bloc/overtime_event.dart';
import 'package:intl/intl.dart';
import '../bloc/overtime_bloc.dart';
import '../bloc/overtime_state.dart';
import 'package:hris_app/core/utils/snackbar_utils.dart';
import 'package:hris_app/core/theme/app_colors.dart';

class OvertimeRequestPage extends StatefulWidget {
  final Overtime? overtimeToEdit;
  const OvertimeRequestPage({super.key, this.overtimeToEdit});

  @override
  State<OvertimeRequestPage> createState() => _OvertimeRequestPageState();
}

class _OvertimeRequestPageState extends State<OvertimeRequestPage> {
  final List<Map<String, String>> _overtimeTypes = [
    {'label': 'Hari Kerja', 'value': 'working_day'},
    {'label': 'Hari Libur / Weekend', 'value': 'holiday'},
    {'label': 'Emergency / Call-out', 'value': 'emergency'},
  ];

  String _selectedTypeCode = 'working_day';
  final _reasonController = TextEditingController();
  DateTime _selectedDate = DateTime.now();
  TimeOfDay _startTime = const TimeOfDay(hour: 17, minute: 0);
  TimeOfDay _endTime = const TimeOfDay(hour: 20, minute: 0);
  XFile? _selectedAttachment;
  final ImagePicker _picker = ImagePicker();

  bool get _isEditing => widget.overtimeToEdit != null;

  @override
  void initState() {
    super.initState();
    if (_isEditing) {
      final ot = widget.overtimeToEdit!;
      _selectedDate = ot.date;
      _startTime = TimeOfDay.fromDateTime(ot.startTime);
      _endTime = TimeOfDay.fromDateTime(ot.endTime);
      _selectedTypeCode = ot.type;
      _reasonController.text = ot.reason;
    }
  }

  int get _durationMinutes {
    final start = DateTime(_selectedDate.year, _selectedDate.month,
        _selectedDate.day, _startTime.hour, _startTime.minute);
    var end = DateTime(_selectedDate.year, _selectedDate.month,
        _selectedDate.day, _endTime.hour, _endTime.minute);

    if (end.isBefore(start)) {
      end = end.add(const Duration(days: 1));
    }

    return end.difference(start).inMinutes;
  }

  Future<void> _selectDate(BuildContext context) async {
    final DateTime? picked = await showDatePicker(
      context: context,
      initialDate: _selectedDate,
      firstDate: DateTime.now().subtract(const Duration(days: 30)),
      lastDate: DateTime.now().add(const Duration(days: 365)),
      builder: (context, child) {
        return Theme(
          data: ThemeData.light().copyWith(
            colorScheme: const ColorScheme.light(
                primary: AppColors.primaryRed,
                onPrimary: Colors.white,
                onSurface: AppColors.textPrimary),
          ),
          child: child!,
        );
      },
    );
    if (picked != null) {
      setState(() {
        _selectedDate = picked;
      });
    }
  }

  Future<void> _selectTime(BuildContext context, bool isStart) async {
    final TimeOfDay? picked = await showTimePicker(
      context: context,
      initialTime: isStart ? _startTime : _endTime,
      builder: (context, child) {
        return Theme(
          data: ThemeData.light().copyWith(
            colorScheme: const ColorScheme.light(
                primary: AppColors.primaryRed,
                onPrimary: Colors.white,
                onSurface: AppColors.textPrimary),
          ),
          child: child!,
        );
      },
    );
    if (picked != null) {
      setState(() {
        if (isStart) {
          _startTime = picked;
        } else {
          _endTime = picked;
        }
      });
    }
  }

  Future<void> _pickImage(ImageSource source) async {
    try {
      final pickedFile =
          await _picker.pickImage(maxWidth: 1024, maxHeight: 1024, source: source, imageQuality: 70);
      if (pickedFile != null) {
        setState(() {
          _selectedAttachment = pickedFile;
        });
      }
    } catch (e) {
      SnackBarUtils.showError(context, 'Gagal mengambil gambar: $e');
    }
  }

  String _formatTime24h(TimeOfDay time) {
    return '${time.hour.toString().padLeft(2, '0')}:${time.minute.toString().padLeft(2, '0')}';
  }

  void _submitRequest() {
    if (_reasonController.text.trim().isEmpty) {
      SnackBarUtils.showError(context, 'Mohon isi detail pekerjaan lembur.');
      return;
    }

    final startDateTime = DateTime(_selectedDate.year, _selectedDate.month,
        _selectedDate.day, _startTime.hour, _startTime.minute);
    var endDateTime = DateTime(_selectedDate.year, _selectedDate.month,
        _selectedDate.day, _endTime.hour, _endTime.minute);

    if (endDateTime.isBefore(startDateTime)) {
      endDateTime = endDateTime.add(const Duration(days: 1));
    }

    if (_isEditing) {
      context.read<OvertimeBloc>().add(UpdateOvertimeRequested(
            id: widget.overtimeToEdit!.id,
            date: _selectedDate,
            startTime: startDateTime,
            endTime: endDateTime,
            durationMinutes: _durationMinutes,
            type: _selectedTypeCode,
            reason: _reasonController.text.trim(),
            attachment: _selectedAttachment,
          ));
    } else {
      context.read<OvertimeBloc>().add(SubmitOvertimeRequested(
            date: _selectedDate,
            startTime: startDateTime,
            endTime: endDateTime,
            durationMinutes: _durationMinutes,
            type: _selectedTypeCode,
            reason: _reasonController.text.trim(),
            attachment: _selectedAttachment,
          ));
    }
  }

  @override
  void dispose() {
    _reasonController.dispose();
    super.dispose();
  }

  @override
  Widget build(BuildContext context) {
    return BlocListener<OvertimeBloc, OvertimeState>(
      listener: (context, state) {
        if (state is OvertimeSuccess) {
          SnackBarUtils.showSuccess(context, state.message);
          Navigator.of(context).pop();
        } else if (state is OvertimeFailure) {
          SnackBarUtils.showError(context, state.errorMessage);
        }
      },
      child: Scaffold(
        backgroundColor: AppColors.backgroundAlt,
        appBar: AppBar(
          title: Text(_isEditing ? 'Ubah Pengajuan' : 'Pengajuan Lembur',
              style: GoogleFonts.plusJakartaSans(
                  fontSize: 16,
                  fontWeight: FontWeight.w800,
                  color: AppColors.textPrimary,
                  letterSpacing: 0.5)),
          centerTitle: true,
          backgroundColor: Colors.white,
          elevation: 0,
          leading: IconButton(
            icon: const Icon(Icons.arrow_back_ios_new,
                color: AppColors.textPrimary, size: 20),
            onPressed: () => Navigator.of(context).pop(),
          ),
        ),
        body: SingleChildScrollView(
          padding: const EdgeInsets.all(24),
          child: Column(
            crossAxisAlignment: CrossAxisAlignment.start,
            children: [
              _buildSectionHeader('WAKTU & TANGGAL'),
              const SizedBox(height: 12),
              _buildTimeDateCard(),
              const SizedBox(height: 24),
              _buildSectionHeader('TIPE LEMBUR'),
              const SizedBox(height: 12),
              _buildTypeSelector(),
              const SizedBox(height: 24),
              _buildSectionHeader('DETAIL PEKERJAAN'),
              const SizedBox(height: 12),
              _buildReasonCard(),
              const SizedBox(height: 24),
              _buildSectionHeader('LAMPIRAN BUKTI'),
              const SizedBox(height: 12),
              _buildAttachmentSection(),
              const SizedBox(height: 32),
              _buildSubmitButton(),
              const SizedBox(height: 40),
            ],
          ),
        ),
      ),
    );
  }

  Widget _buildSectionHeader(String title) {
    return Text(title,
        style: const TextStyle(
            fontSize: 11,
            fontWeight: FontWeight.w900,
            color: AppColors.textTertiary,
            letterSpacing: 1.5));
  }

  Widget _buildTimeDateCard() {
    return Container(
      padding: const EdgeInsets.all(20),
      decoration: BoxDecoration(
        color: Colors.white,
        borderRadius: BorderRadius.circular(24),
        border: Border.all(color: AppColors.grayBorder),
      ),
      child: Column(
        children: [
          _buildClickableField(
              'TANGGAL',
              DateFormat('dd MMMM yyyy').format(_selectedDate),
              Icons.calendar_today_rounded,
              () => _selectDate(context)),
          const Padding(
              padding: EdgeInsets.symmetric(vertical: 16),
              child: Divider(height: 1)),
          Row(
            children: [
              Expanded(
                  child: _buildClickableField(
                      'MULAI',
                      _formatTime24h(_startTime),
                      Icons.access_time_filled_rounded,
                      () => _selectTime(context, true))),
              Container(
                  width: 1,
                  height: 40,
                  color: AppColors.grayBorder,
                  margin: const EdgeInsets.symmetric(horizontal: 16)),
              Expanded(
                  child: _buildClickableField(
                      'SELESAI',
                      _formatTime24h(_endTime),
                      Icons.access_time_rounded,
                      () => _selectTime(context, false))),
            ],
          ),
        ],
      ),
    );
  }

  Widget _buildClickableField(
      String label, String value, IconData icon, VoidCallback onTap) {
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
              Text(value,
                  style: const TextStyle(
                      fontSize: 14,
                      fontWeight: FontWeight.w800,
                      color: AppColors.textPrimary)),
            ],
          ),
          Icon(icon, size: 20, color: AppColors.textTertiary),
        ],
      ),
    );
  }

  Widget _buildTypeSelector() {
    return Container(
      padding: const EdgeInsets.all(8),
      decoration: BoxDecoration(
          color: Colors.white,
          borderRadius: BorderRadius.circular(24),
          border: Border.all(color: AppColors.grayBorder)),
      child: Wrap(
        spacing: 8,
        runSpacing: 8,
        children: _overtimeTypes
            .map((type) => _buildTypeChip(type['label']!, type['value']!))
            .toList(),
      ),
    );
  }

  Widget _buildTypeChip(String label, String value) {
    final isSelected = _selectedTypeCode == value;
    return GestureDetector(
      onTap: () => setState(() => _selectedTypeCode = value),
      child: Container(
        padding: const EdgeInsets.symmetric(horizontal: 16, vertical: 10),
        decoration: BoxDecoration(
          color:
              isSelected ? AppColors.primaryRed.withOpacity(0.1) : Colors.white,
          border: Border.all(
              color: isSelected ? AppColors.primaryRed : AppColors.grayBorder,
              width: 1.5),
          borderRadius: BorderRadius.circular(16),
        ),
        child: Row(
          mainAxisSize: MainAxisSize.min,
          children: [
            if (isSelected) ...[
              const Icon(Icons.check_circle,
                  size: 16, color: AppColors.primaryRed),
              const SizedBox(width: 8),
            ],
            Text(label,
                style: TextStyle(
                    fontSize: 12,
                    fontWeight: FontWeight.w800,
                    color: isSelected
                        ? AppColors.primaryRed
                        : AppColors.textSecondary)),
          ],
        ),
      ),
    );
  }

  Widget _buildReasonCard() {
    return Container(
      decoration: BoxDecoration(
          color: Colors.white,
          borderRadius: BorderRadius.circular(24),
          border: Border.all(color: AppColors.grayBorder)),
      child: TextField(
        controller: _reasonController,
        maxLines: 5,
        style: const TextStyle(fontSize: 14, fontWeight: FontWeight.w600),
        decoration: InputDecoration(
          hintText:
              'Misal: Menyelesaikan laporan bulanan atau Maintenance Server...',
          hintStyle: TextStyle(
              color: AppColors.textTertiary.withOpacity(0.5), fontSize: 13),
          border: InputBorder.none,
          contentPadding: const EdgeInsets.all(20),
        ),
      ),
    );
  }

  Widget _buildAttachmentSection() {
    if (_selectedAttachment != null) {
      return Container(
        padding: const EdgeInsets.all(12),
        decoration: BoxDecoration(
          color: Colors.white,
          borderRadius: BorderRadius.circular(16),
          border: Border.all(color: AppColors.grayBorder),
        ),
        child: Row(
          children: [
            ClipRRect(
              borderRadius: BorderRadius.circular(8),
              child: kIsWeb
                  ? CachedNetworkImage(imageUrl: _selectedAttachment!.path, width: 60, height: 60, fit: BoxFit.cover, placeholder: (context, url) => const Center(child: CircularProgressIndicator()), errorWidget: (context, url, error) => const Icon(Icons.error), memCacheWidth: 800,)
                  : Image.file(File(_selectedAttachment!.path),
                      width: 60, height: 60, fit: BoxFit.cover),
            ),
            const SizedBox(width: 16),
            Expanded(
              child: Text(
                _selectedAttachment!.name,
                style: const TextStyle(
                    fontSize: 12,
                    fontWeight: FontWeight.w600,
                    color: AppColors.textPrimary),
                maxLines: 1,
                overflow: TextOverflow.ellipsis,
              ),
            ),
            IconButton(
              icon: const Icon(Icons.close, color: AppColors.primaryRed),
              onPressed: () => setState(() => _selectedAttachment = null),
            )
          ],
        ),
      );
    }

    return Row(
      children: [
        Expanded(
          child: ElevatedButton.icon(
            onPressed: () => _pickImage(ImageSource.camera),
            icon: const Icon(Icons.camera_alt,
                size: 18, color: AppColors.textSecondary),
            label: const Text('Kamera',
                style: TextStyle(color: AppColors.textSecondary)),
            style: ElevatedButton.styleFrom(
              backgroundColor: Colors.white,
              elevation: 0,
              shape: RoundedRectangleBorder(
                  borderRadius: BorderRadius.circular(12),
                  side: const BorderSide(color: AppColors.grayBorder)),
              padding: const EdgeInsets.symmetric(vertical: 12),
            ),
          ),
        ),
        const SizedBox(width: 12),
        Expanded(
          child: ElevatedButton.icon(
            onPressed: () => _pickImage(ImageSource.gallery),
            icon: const Icon(Icons.image,
                size: 18, color: AppColors.textSecondary),
            label: const Text('Galeri',
                style: TextStyle(color: AppColors.textSecondary)),
            style: ElevatedButton.styleFrom(
              backgroundColor: Colors.white,
              elevation: 0,
              shape: RoundedRectangleBorder(
                  borderRadius: BorderRadius.circular(12),
                  side: const BorderSide(color: AppColors.grayBorder)),
              padding: const EdgeInsets.symmetric(vertical: 12),
            ),
          ),
        ),
      ],
    );
  }

  Widget _buildSubmitButton() {
    return BlocBuilder<OvertimeBloc, OvertimeState>(
      builder: (context, state) {
        final isLoading = state is OvertimeLoading;
        return SizedBox(
          width: double.infinity,
          height: 60,
          child: ElevatedButton(
            onPressed: isLoading ? null : _submitRequest,
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
                : Text(_isEditing ? 'SIMPAN PERUBAHAN' : 'AJUKAN LEMBUR',
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
