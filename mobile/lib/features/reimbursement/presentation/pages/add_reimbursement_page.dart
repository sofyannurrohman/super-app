import 'package:cached_network_image/cached_network_image.dart';
import 'dart:typed_data';
import 'package:flutter/material.dart';
import 'package:flutter_bloc/flutter_bloc.dart';
import 'package:hris_app/features/reimbursement/domain/entities/reimbursement.dart';
import 'package:hris_app/features/reimbursement/presentation/bloc/reimbursement_bloc.dart';
import 'package:hris_app/features/reimbursement/presentation/bloc/reimbursement_event.dart';
import 'package:hris_app/features/reimbursement/presentation/bloc/reimbursement_state.dart';
import 'package:hris_app/injection.dart' as di;
import 'package:image_picker/image_picker.dart';
import 'package:hris_app/core/utils/snackbar_utils.dart';
import 'package:hris_app/core/network/api_client.dart';

class AddReimbursementPage extends StatefulWidget {
  final Reimbursement? reimbursementToEdit;
  final String? initialTitle;
  const AddReimbursementPage({super.key, this.reimbursementToEdit, this.initialTitle});

  @override
  State<AddReimbursementPage> createState() => _AddReimbursementPageState();
}

class _AddReimbursementPageState extends State<AddReimbursementPage> {
  final _titleController = TextEditingController();
  final _amountController = TextEditingController();
  final _descriptionController = TextEditingController();
  XFile? _attachment;
  Uint8List? _attachmentBytes;
  String? _attachmentName;
  String? _existingAttachmentUrl;
  final ImagePicker _picker = ImagePicker();

  bool get _isEditing => widget.reimbursementToEdit != null;

  @override
  void initState() {
    super.initState();
    if (_isEditing) {
      final r = widget.reimbursementToEdit!;
      _titleController.text = r.title;
      _amountController.text = r.amount.toString();
      _descriptionController.text = r.description ?? '';
      _existingAttachmentUrl = r.attachmentUrl;
    } else if (widget.initialTitle != null) {
      _titleController.text = widget.initialTitle!;
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
      final bytes = await image.readAsBytes();
      setState(() {
        _attachment = image;
        _attachmentBytes = bytes;
        _attachmentName = image.name;
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
              const Text(
                'Pilih Sumber Foto',
                style: TextStyle(fontSize: 18, fontWeight: FontWeight.bold),
              ),
              const SizedBox(height: 16),
              ListTile(
                leading: const Icon(Icons.camera_alt),
                title: const Text('Kamera'),
                onTap: () {
                  Navigator.pop(context);
                  _pickImage(ImageSource.camera);
                },
              ),
              ListTile(
                leading: const Icon(Icons.photo_library),
                title: const Text('Galeri'),
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

  @override
  Widget build(BuildContext context) {
    return BlocProvider(
      create: (context) => di.sl<ReimbursementBloc>(),
      child: Scaffold(
        appBar: AppBar(
          title:
              Text(_isEditing ? 'Ubah Reimbursement' : 'Ajukan Reimbursement'),
        ),
        body: BlocConsumer<ReimbursementBloc, ReimbursementState>(
          listener: (context, state) {
            if (state is ReimbursementActionSuccess) {
              SnackBarUtils.showSuccess(context, state.message);
              Navigator.pop(context, true);
            } else if (state is ReimbursementFailure) {
              SnackBarUtils.showError(context, state.message);
            }
          },
          builder: (context, state) {
            return SingleChildScrollView(
              padding: const EdgeInsets.all(16.0),
              child: Column(
                crossAxisAlignment: CrossAxisAlignment.stretch,
                children: [
                  TextField(
                    controller: _titleController,
                    decoration: const InputDecoration(
                      labelText: 'Judul (Contoh: Bensin Sales)',
                      border: OutlineInputBorder(),
                    ),
                  ),
                  const SizedBox(height: 16),
                  TextField(
                    controller: _amountController,
                    decoration: const InputDecoration(
                      labelText: 'Total Nominal (Rp)',
                      border: OutlineInputBorder(),
                      prefixText: 'Rp ',
                    ),
                    keyboardType: TextInputType.number,
                  ),
                  const SizedBox(height: 16),
                  TextField(
                    controller: _descriptionController,
                    decoration: const InputDecoration(
                      labelText: 'Keterangan',
                      border: OutlineInputBorder(),
                    ),
                    maxLines: 3,
                  ),
                  const SizedBox(height: 24),
                  const Row(
                    mainAxisAlignment: MainAxisAlignment.spaceBetween,
                    children: [
                      Text(
                        'Bukti Pembayaran / Nota',
                        style: TextStyle(fontWeight: FontWeight.bold),
                      ),
                      Text(
                        'Maks. 5MB',
                        style: TextStyle(fontSize: 12, color: Colors.grey),
                      ),
                    ],
                  ),
                  const SizedBox(height: 8),
                  InkWell(
                    onTap: () => _showImageSourceActionSheet(context),
                    child: Container(
                      height: 150,
                      decoration: BoxDecoration(
                        border: Border.all(color: Colors.grey[300]!),
                        borderRadius: BorderRadius.circular(8),
                        color: Colors.grey[50],
                      ),
                      child: (_attachmentBytes == null &&
                              _existingAttachmentUrl == null)
                          ? const Column(
                              mainAxisAlignment: MainAxisAlignment.center,
                              children: [
                                Icon(Icons.add_a_photo,
                                    size: 40, color: Colors.grey),
                                SizedBox(height: 8),
                                Text('Pilih Foto Nota',
                                    style: TextStyle(color: Colors.grey)),
                              ],
                            )
                          : Stack(
                              fit: StackFit.expand,
                              children: [
                                ClipRRect(
                                  borderRadius: BorderRadius.circular(8),
                                  child: _attachmentBytes != null
                                      ? Image.memory(_attachmentBytes!,
                                          fit: BoxFit.cover)
                                      : CachedNetworkImage(imageUrl: _existingAttachmentUrl!
                                                  .startsWith('http')
                                              ? _existingAttachmentUrl!
                                              : '${di.sl<ApiClient>().baseUrl.replaceAll('/api/v1/', '')}$_existingAttachmentUrl', fit: BoxFit.cover, placeholder: (context, url) => const Center(child: CircularProgressIndicator()), errorWidget: (context, url, error) => const Icon(Icons.error), memCacheWidth: 800,),
                                ),
                                Positioned(
                                  top: 8,
                                  right: 8,
                                  child: CircleAvatar(
                                    backgroundColor: Colors.red,
                                    radius: 12,
                                    child: IconButton(
                                      icon: const Icon(Icons.close,
                                          size: 12, color: Colors.white),
                                      onPressed: () => setState(() {
                                        _attachment = null;
                                        _attachmentBytes = null;
                                        _attachmentName = null;
                                        _existingAttachmentUrl = null;
                                      }),
                                      padding: EdgeInsets.zero,
                                    ),
                                  ),
                                ),
                              ],
                            ),
                    ),
                  ),
                  const SizedBox(height: 32),
                  ElevatedButton(
                    onPressed: state is ReimbursementLoading
                        ? null
                        : () {
                            if (_titleController.text.isEmpty ||
                                _amountController.text.isEmpty) {
                              SnackBarUtils.showError(
                                  context, 'Mohon lengkapi judul dan nominal.');
                              return;
                            }
                            if (_attachmentBytes == null &&
                                _existingAttachmentUrl == null) {
                              SnackBarUtils.showError(
                                  context, 'Mohon lampirkan nota.');
                              return;
                            }

                            final amount =
                                double.tryParse(_amountController.text) ?? 0;

                            if (_isEditing) {
                              context
                                  .read<ReimbursementBloc>()
                                  .add(UpdateReimbursementRequested(
                                    id: widget.reimbursementToEdit!.id,
                                    title: _titleController.text,
                                    description: _descriptionController.text,
                                    amount: amount,
                                    attachmentBytes: _attachmentBytes,
                                    attachmentName: _attachmentName,
                                  ));
                            } else {
                              context
                                  .read<ReimbursementBloc>()
                                  .add(SubmitReimbursementRequested(
                                    title: _titleController.text,
                                    description: _descriptionController.text,
                                    amount: amount,
                                    attachmentBytes: _attachmentBytes,
                                    attachmentName: _attachmentName,
                                  ));
                            }
                          },
                    child: state is ReimbursementLoading
                        ? const SizedBox(
                            height: 20,
                            width: 20,
                            child: CircularProgressIndicator(
                                strokeWidth: 2, color: Colors.white))
                        : Text(_isEditing
                            ? 'Simpan Perubahan'
                            : 'Kirim Pengajuan'),
                  ),
                ],
              ),
            );
          },
        ),
      ),
    );
  }
}
