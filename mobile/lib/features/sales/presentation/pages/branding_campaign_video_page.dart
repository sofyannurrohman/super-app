import 'package:dio/dio.dart';
import 'package:file_picker/file_picker.dart';
import 'package:flutter/material.dart';
import 'package:hris_app/core/network/api_client.dart';
import 'package:hris_app/injection.dart' as di;
import 'package:url_launcher/url_launcher.dart';

class BrandingCampaignVideoPage extends StatefulWidget {
  const BrandingCampaignVideoPage({super.key});

  @override
  State<BrandingCampaignVideoPage> createState() =>
      _BrandingCampaignVideoPageState();
}

class _BrandingCampaignVideoPageState extends State<BrandingCampaignVideoPage> {
  final _api = di.sl<ApiClient>();
  List<Map<String, dynamic>> _videos = [];
  bool _loading = true;

  @override
  void initState() {
    super.initState();
    _load();
  }

  Future<void> _load() async {
    setState(() => _loading = true);
    try {
      final response = await _api.client.get('branding/campaign-videos');
      final data = response.data['data'] as List<dynamic>? ?? [];
      if (mounted) {
        setState(() => _videos = data
            .map((item) => Map<String, dynamic>.from(item as Map))
            .toList());
      }
    } catch (_) {
      _message('Gagal memuat video kampanye');
    } finally {
      if (mounted) setState(() => _loading = false);
    }
  }

  void _message(String value) {
    if (mounted) {
      ScaffoldMessenger.of(context)
          .showSnackBar(SnackBar(content: Text(value)));
    }
  }

  Future<void> _openLink(String raw) async {
    final uri = Uri.tryParse(raw);
    if (uri == null ||
        uri.scheme != 'https' ||
        !await launchUrl(uri, mode: LaunchMode.externalApplication)) {
      _message('Tautan tidak dapat dibuka');
    }
  }

  Future<void> _edit([Map<String, dynamic>? video]) async {
    final saved = await showModalBottomSheet<bool>(
      context: context,
      isScrollControlled: true,
      useSafeArea: true,
      builder: (_) => _CampaignVideoForm(api: _api, video: video),
    );
    if (saved == true) _load();
  }

  Future<void> _delete(Map<String, dynamic> video) async {
    final confirmed = await showDialog<bool>(
      context: context,
      builder: (dialogContext) => AlertDialog(
        title: const Text('Hapus video kampanye?'),
        content: Text('“${video['title']}” akan dihapus dari daftar kampanye.'),
        actions: [
          TextButton(
              onPressed: () => Navigator.pop(dialogContext, false),
              child: const Text('Batal')),
          TextButton(
              onPressed: () => Navigator.pop(dialogContext, true),
              child: const Text('Hapus')),
        ],
      ),
    );
    if (confirmed != true) return;
    try {
      await _api.client.delete('branding/campaign-videos/${video['id']}');
      _message('Video dihapus dari daftar');
      _load();
    } catch (_) {
      _message('Gagal menghapus video');
    }
  }

  @override
  Widget build(BuildContext context) {
    return Scaffold(
      appBar: AppBar(title: const Text('Video Kampanye'), actions: [
        IconButton(
            onPressed: _load,
            icon: const Icon(Icons.refresh_rounded),
            tooltip: 'Muat ulang'),
      ]),
      floatingActionButton: FloatingActionButton.extended(
        onPressed: () => _edit(),
        icon: const Icon(Icons.add),
        label: const Text('Tambah video'),
      ),
      body: _loading
          ? const Center(child: CircularProgressIndicator())
          : RefreshIndicator(
              onRefresh: _load,
              child: ListView(
                padding: const EdgeInsets.all(16),
                children: [
                  const Card(
                      child: Padding(
                    padding: EdgeInsets.all(16),
                    child: Text(
                        'Video disimpan di Google Drive. Tag produk atau keranjang di TikTok, Instagram, dan Facebook ditambahkan saat memublikasikan video melalui fitur toko masing-masing platform.'),
                  )),
                  const SizedBox(height: 8),
                  if (_videos.isEmpty)
                    const Padding(
                        padding: EdgeInsets.all(40),
                        child:
                            Center(child: Text('Belum ada video kampanye.'))),
                  for (final video in _videos)
                    Card(
                      margin: const EdgeInsets.only(bottom: 12),
                      child: Padding(
                        padding: const EdgeInsets.all(14),
                        child: Column(
                            crossAxisAlignment: CrossAxisAlignment.start,
                            children: [
                              Row(children: [
                                const Icon(Icons.video_library_rounded),
                                const SizedBox(width: 8),
                                Expanded(
                                    child: Text('${video['title'] ?? ''}',
                                        style: const TextStyle(
                                            fontWeight: FontWeight.w800,
                                            fontSize: 16))),
                                Chip(
                                    label: Text(video['is_active'] == true
                                        ? 'Tayang'
                                        : 'Draft')),
                              ]),
                              if ('${video['description'] ?? ''}'.isNotEmpty)
                                Padding(
                                    padding: const EdgeInsets.only(top: 6),
                                    child: Text('${video['description']}')),
                              const SizedBox(height: 8),
                              Text('Produk: ${video['product_name'] ?? ''}',
                                  style: const TextStyle(
                                      fontWeight: FontWeight.w600)),
                              Wrap(spacing: 8, children: [
                                TextButton.icon(
                                    onPressed: () => _openLink(
                                        '${video['video_url'] ?? ''}'),
                                    icon: const Icon(Icons.play_circle_outline),
                                    label: const Text('Buka video')),
                                TextButton.icon(
                                    onPressed: () => _openLink(
                                        '${video['purchase_url'] ?? ''}'),
                                    icon: const Icon(Icons.open_in_new),
                                    label: const Text('Tautan produk')),
                                TextButton.icon(
                                    onPressed: () => _edit(video),
                                    icon: const Icon(Icons.edit_outlined),
                                    label: const Text('Edit')),
                                TextButton.icon(
                                    onPressed: () => _delete(video),
                                    icon: const Icon(Icons.delete_outline),
                                    label: const Text('Hapus')),
                              ]),
                            ]),
                      ),
                    ),
                  const SizedBox(height: 72),
                ],
              ),
            ),
    );
  }
}

class _CampaignVideoForm extends StatefulWidget {
  final ApiClient api;
  final Map<String, dynamic>? video;
  const _CampaignVideoForm({required this.api, this.video});

  @override
  State<_CampaignVideoForm> createState() => _CampaignVideoFormState();
}

class _CampaignVideoFormState extends State<_CampaignVideoForm> {
  final _formKey = GlobalKey<FormState>();
  late final TextEditingController _title;
  late final TextEditingController _description;
  late final TextEditingController _videoUrl;
  late final TextEditingController _productName;
  late final TextEditingController _purchaseUrl;
  late bool _active;
  bool _saving = false;
  bool _uploading = false;

  @override
  void initState() {
    super.initState();
    final video = widget.video ?? {};
    _title = TextEditingController(text: '${video['title'] ?? ''}');
    _description = TextEditingController(text: '${video['description'] ?? ''}');
    _videoUrl = TextEditingController(text: '${video['video_url'] ?? ''}');
    _productName =
        TextEditingController(text: '${video['product_name'] ?? ''}');
    _purchaseUrl =
        TextEditingController(text: '${video['purchase_url'] ?? ''}');
    _active = video['is_active'] != false;
  }

  @override
  void dispose() {
    _title.dispose();
    _description.dispose();
    _videoUrl.dispose();
    _productName.dispose();
    _purchaseUrl.dispose();
    super.dispose();
  }

  String? _validateDrive(String? value) {
    final uri = Uri.tryParse(value?.trim() ?? '');
    if (uri?.scheme != 'https' ||
        uri?.host != 'drive.google.com' ||
        !(uri!.path.contains('/file/d/') ||
            (uri.path == '/open' && uri.queryParameters.containsKey('id')))) {
      return 'Masukkan tautan file Google Drive';
    }
    return null;
  }

  String? _validatePurchase(String? value) {
    final uri = Uri.tryParse(value?.trim() ?? '');
    return uri?.scheme == 'https' && (uri?.host.isNotEmpty ?? false)
        ? null
        : 'Masukkan tautan HTTPS produk';
  }

  void _message(String value) {
    if (mounted) {
      ScaffoldMessenger.of(context)
          .showSnackBar(SnackBar(content: Text(value)));
    }
  }

  Future<void> _upload() async {
    final picked = await FilePicker.platform.pickFiles(type: FileType.video);
    if (picked == null || picked.files.isEmpty) return;
    final file = picked.files.first;
    if (file.size > 95 * 1024 * 1024 || file.path == null) {
      _message('Pilih video maksimal 95 MB dari penyimpanan perangkat');
      return;
    }
    setState(() => _uploading = true);
    try {
      final payload = FormData.fromMap({
        'video': await MultipartFile.fromFile(file.path!, filename: file.name)
      });
      final response = await widget.api.client.post(
        'branding/campaign-videos/upload',
        data: payload,
        options: Options(
            sendTimeout: const Duration(minutes: 10),
            receiveTimeout: const Duration(minutes: 10)),
      );
      _videoUrl.text = '${response.data['video_url'] ?? ''}';
      _message('Video berhasil diunggah ke Google Drive');
    } on DioException catch (e) {
      _message('${e.response?.data?['error'] ?? 'Gagal mengunggah video'}');
    } catch (_) {
      _message('Gagal mengunggah video');
    } finally {
      if (mounted) setState(() => _uploading = false);
    }
  }

  Future<void> _save() async {
    if (!_formKey.currentState!.validate() || _uploading) return;
    setState(() => _saving = true);
    final payload = {
      'title': _title.text.trim(),
      'description': _description.text.trim(),
      'video_url': _videoUrl.text.trim(),
      'product_name': _productName.text.trim(),
      'purchase_url': _purchaseUrl.text.trim(),
      'is_active': _active,
    };
    try {
      if (widget.video == null) {
        await widget.api.client.post('branding/campaign-videos', data: payload);
      } else {
        await widget.api.client.put(
            'branding/campaign-videos/${widget.video!['id']}',
            data: payload);
      }
      if (mounted) Navigator.pop(context, true);
    } on DioException catch (e) {
      _message('${e.response?.data?['error'] ?? 'Gagal menyimpan video'}');
    } catch (_) {
      _message('Gagal menyimpan video');
    } finally {
      if (mounted) setState(() => _saving = false);
    }
  }

  @override
  Widget build(BuildContext context) {
    return Padding(
      padding: EdgeInsets.only(bottom: MediaQuery.viewInsetsOf(context).bottom),
      child: SizedBox(
        height: MediaQuery.sizeOf(context).height * .88,
        child: ListView(padding: const EdgeInsets.all(20), children: [
          Text(
              widget.video == null
                  ? 'Tambah Video Kampanye'
                  : 'Edit Video Kampanye',
              style: Theme.of(context).textTheme.titleLarge),
          const SizedBox(height: 16),
          Form(
              key: _formKey,
              child: Column(children: [
                TextFormField(
                    controller: _title,
                    maxLength: 200,
                    decoration: const InputDecoration(labelText: 'Judul video'),
                    validator: (v) => (v?.trim().isEmpty ?? true)
                        ? 'Judul wajib diisi'
                        : null),
                TextFormField(
                    controller: _description,
                    maxLines: 2,
                    decoration: const InputDecoration(labelText: 'Deskripsi')),
                const SizedBox(height: 8),
                OutlinedButton.icon(
                    onPressed: _uploading ? null : _upload,
                    icon: const Icon(Icons.cloud_upload_outlined),
                    label: Text(_uploading
                        ? 'Mengunggah...'
                        : 'Unggah ke Google Drive (maks. 95 MB)')),
                TextFormField(
                    controller: _videoUrl,
                    keyboardType: TextInputType.url,
                    decoration: const InputDecoration(
                        labelText: 'Atau tautan video Google Drive'),
                    validator: _validateDrive),
                TextFormField(
                    controller: _productName,
                    maxLength: 200,
                    decoration: const InputDecoration(labelText: 'Nama produk'),
                    validator: (v) => (v?.trim().isEmpty ?? true)
                        ? 'Nama produk wajib diisi'
                        : null),
                TextFormField(
                    controller: _purchaseUrl,
                    keyboardType: TextInputType.url,
                    decoration:
                        const InputDecoration(labelText: 'Tautan produk HTTPS'),
                    validator: _validatePurchase),
                SwitchListTile(
                    value: _active,
                    onChanged: (value) => setState(() => _active = value),
                    title: const Text('Tayang')),
                const Text(
                    'Tautan produk ini tidak otomatis menjadi keranjang pada postingan media sosial.',
                    style: TextStyle(fontSize: 12, color: Colors.black54)),
                const SizedBox(height: 16),
                FilledButton(
                    onPressed: _saving || _uploading ? null : _save,
                    child: Text(_saving ? 'Menyimpan...' : 'Simpan video')),
              ])),
        ]),
      ),
    );
  }
}
