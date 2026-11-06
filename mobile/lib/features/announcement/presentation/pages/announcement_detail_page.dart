import 'dart:convert';
import 'dart:io';
import 'package:cached_network_image/cached_network_image.dart';
import 'package:dio/dio.dart';
import 'package:flutter/material.dart';
import 'package:flutter/services.dart';
import 'package:flutter_bloc/flutter_bloc.dart';
import 'package:google_fonts/google_fonts.dart';
import 'package:intl/intl.dart';
import 'package:path_provider/path_provider.dart';
import 'package:shared_preferences/shared_preferences.dart';
import 'package:share_plus/share_plus.dart';
import 'package:url_launcher/url_launcher.dart';
import 'package:hris_app/core/theme/app_colors.dart';
import 'package:hris_app/core/utils/constants.dart';
import 'package:hris_app/features/announcement/data/models/announcement_model.dart';
import 'package:hris_app/features/announcement/data/repositories/announcement_repository.dart';
import 'package:hris_app/features/announcement/presentation/bloc/announcement_bloc.dart';
import 'package:hris_app/injection.dart';

class AnnouncementDetailPage extends StatefulWidget {
  final Announcement announcement;

  const AnnouncementDetailPage({super.key, required this.announcement});

  @override
  State<AnnouncementDetailPage> createState() => _AnnouncementDetailPageState();
}

class _AnnouncementDetailPageState extends State<AnnouncementDetailPage> {
  late Announcement _currentAnnouncement;
  int _activeImageIndex = 0;
  final PageController _pageController = PageController();
  final ScrollController _scrollController = ScrollController();
  final TextEditingController _commentController = TextEditingController();
  final FocusNode _commentFocusNode = FocusNode();

  final AnnouncementRepository _repo = sl<AnnouncementRepository>();
  List<AnnouncementComment> _comments = [];
  bool _isLoadingComments = false;
  bool _isSubmittingComment = false;
  String _cachedMyName = '';
  String _cachedMyAvatar = '';

  @override
  void initState() {
    super.initState();
    _currentAnnouncement = widget.announcement;
    _loadLocalProfile();
    _loadComments();
  }

  void _loadLocalProfile() {
    try {
      final prefs = sl<SharedPreferences>();
      final cachedStr = prefs.getString('cached_employee_profile');
      if (cachedStr != null && cachedStr.isNotEmpty) {
        final Map<String, dynamic> data = jsonDecode(cachedStr);
        final firstName =
            data['first_name']?.toString() ??
            data['FirstName']?.toString() ??
            '';
        final lastName =
            data['last_name']?.toString() ?? data['LastName']?.toString() ?? '';
        final fullName = '$firstName $lastName'.trim();
        final avatar =
            data['face_reference_url']?.toString() ??
            data['FaceReferenceURL']?.toString() ??
            '';
        if (fullName.isNotEmpty) _cachedMyName = fullName;
        if (avatar.isNotEmpty) _cachedMyAvatar = avatar;
      }
    } catch (_) {}
  }

  @override
  void dispose() {
    _pageController.dispose();
    _scrollController.dispose();
    _commentController.dispose();
    _commentFocusNode.dispose();
    super.dispose();
  }

  Future<void> _loadComments() async {
    setState(() {
      _isLoadingComments = true;
    });

    final result = await _repo.getComments(_currentAnnouncement.id);
    if (!mounted) return;

    result.fold(
      (failure) {
        setState(() {
          _isLoadingComments = false;
        });
      },
      (commentsList) {
        setState(() {
          _comments = commentsList;
          _isLoadingComments = false;
          _currentAnnouncement = _currentAnnouncement.copyWith(
            commentsCount: commentsList.length,
          );
        });
      },
    );
  }

  Future<void> _submitComment() async {
    final text = _commentController.text.trim();
    if (text.isEmpty || _isSubmittingComment) return;

    setState(() {
      _isSubmittingComment = true;
    });

    final result = await _repo.addComment(_currentAnnouncement.id, text);
    if (!mounted) return;

    result.fold(
      (failure) {
        setState(() {
          _isSubmittingComment = false;
        });
        ScaffoldMessenger.of(context).showSnackBar(
          SnackBar(
            content: Text(failure.message),
            backgroundColor: AppColors.primaryRed,
          ),
        );
      },
      (newComment) {
        _commentController.clear();
        _commentFocusNode.unfocus();
        setState(() {
          _isSubmittingComment = false;
          _comments.add(newComment);
          _currentAnnouncement = _currentAnnouncement.copyWith(
            commentsCount: _comments.length,
          );
        });

        // Scroll to the bottom of the page
        Future.delayed(const Duration(milliseconds: 150), () {
          if (_scrollController.hasClients) {
            _scrollController.animateTo(
              _scrollController.position.maxScrollExtent,
              duration: const Duration(milliseconds: 300),
              curve: Curves.easeOut,
            );
          }
        });
      },
    );
  }

  Future<void> _deleteComment(AnnouncementComment comment) async {
    final confirm = await showDialog<bool>(
      context: context,
      builder: (ctx) => AlertDialog(
        shape: RoundedRectangleBorder(borderRadius: BorderRadius.circular(16)),
        title: Text(
          'Hapus Komentar?',
          style: GoogleFonts.plusJakartaSans(
            fontWeight: FontWeight.w800,
            fontSize: 16,
          ),
        ),
        content: Text(
          'Komentar ini akan dihapus permanen dari postingan.',
          style: GoogleFonts.plusJakartaSans(fontSize: 13),
        ),
        actions: [
          TextButton(
            onPressed: () => Navigator.pop(ctx, false),
            child: Text(
              'Batal',
              style: GoogleFonts.plusJakartaSans(
                color: AppColors.textSecondary,
                fontWeight: FontWeight.w700,
              ),
            ),
          ),
          ElevatedButton(
            onPressed: () => Navigator.pop(ctx, true),
            style: ElevatedButton.styleFrom(
              backgroundColor: AppColors.primaryRed,
              foregroundColor: Colors.white,
              elevation: 0,
              shape: RoundedRectangleBorder(
                borderRadius: BorderRadius.circular(8),
              ),
            ),
            child: Text(
              'Hapus',
              style: GoogleFonts.plusJakartaSans(fontWeight: FontWeight.w700),
            ),
          ),
        ],
      ),
    );

    if (confirm != true || !mounted) return;

    final result = await _repo.deleteComment(
      _currentAnnouncement.id,
      comment.id,
    );
    if (!mounted) return;

    result.fold(
      (failure) {
        ScaffoldMessenger.of(context).showSnackBar(
          SnackBar(
            content: Text(failure.message),
            backgroundColor: AppColors.primaryRed,
          ),
        );
      },
      (_) {
        setState(() {
          _comments.removeWhere((c) => c.id == comment.id);
          _currentAnnouncement = _currentAnnouncement.copyWith(
            commentsCount: _comments.length,
          );
        });
      },
    );
  }

  Future<void> _shareToWhatsApp() async {
    final title = _currentAnnouncement.title;
    final content = _currentAnnouncement.content;
    final extUrl = _currentAnnouncement.externalUrl ?? '';

    String message = "*$title*\n\n$content";
    if (extUrl.isNotEmpty) {
      message += "\n\n🔗 Link postingan: $extUrl";
    }

    final encoded = Uri.encodeComponent(message);
    final waUrl = Uri.parse("whatsapp://send?text=$encoded");

    try {
      if (await canLaunchUrl(waUrl)) {
        await launchUrl(waUrl, mode: LaunchMode.externalApplication);
      } else {
        await Share.share(message, subject: title);
      }
      if (_currentAnnouncement.trackShares) {
        _repo.recordActivity(
          _currentAnnouncement.id,
          'share',
          channel: 'whatsapp',
        );
      }
    } catch (_) {
      await Share.share(message, subject: title);
      if (_currentAnnouncement.trackShares) {
        _repo.recordActivity(
          _currentAnnouncement.id,
          'share',
          channel: 'whatsapp',
        );
      }
    }
  }

  Future<void> _shareGeneral() async {
    final title = _currentAnnouncement.title;
    final content = _currentAnnouncement.content;
    final extUrl = _currentAnnouncement.externalUrl ?? '';

    String shareText = "$title\n\n$content";
    if (extUrl.isNotEmpty) {
      shareText += "\n\n$extUrl";
    }

    await Share.share(shareText, subject: title);
    if (_currentAnnouncement.trackShares) {
      _repo.recordActivity(
        _currentAnnouncement.id,
        'share',
        channel: 'system_share',
      );
    }
  }

  bool _isDownloadingMedia = false;

  Future<void> _downloadMediaFile(String rawUrl) async {
    if (_isDownloadingMedia) return;

    if (_isGDriveUrl(rawUrl)) {
      final downloadUrl = _getGDriveDirectDownloadUrl(rawUrl);
      if (_currentAnnouncement.trackDownloads) {
        _repo.recordActivity(
          _currentAnnouncement.id,
          'download',
          channel: 'gdrive',
        );
      }
      _openExternalLink(downloadUrl);
      return;
    }

    final resolvedUrl = AppConstants.buildImageUrl(rawUrl);
    if (resolvedUrl.isEmpty) {
      ScaffoldMessenger.of(context).showSnackBar(
        const SnackBar(
          content: Text('URL media tidak valid untuk diunduh'),
          backgroundColor: AppColors.primaryRed,
        ),
      );
      return;
    }

    setState(() => _isDownloadingMedia = true);
    ScaffoldMessenger.of(context).showSnackBar(
      const SnackBar(
        content: Row(
          children: [
            SizedBox(
              width: 16,
              height: 16,
              child: CircularProgressIndicator(
                strokeWidth: 2,
                color: Colors.white,
              ),
            ),
            SizedBox(width: 12),
            Text('Sedang menyiapkan dan mengunduh berkas...'),
          ],
        ),
        duration: Duration(seconds: 2),
      ),
    );

    try {
      final dio = Dio();
      final uri = Uri.parse(resolvedUrl);
      String fileName = uri.pathSegments.isNotEmpty
          ? uri.pathSegments.last
          : 'media_file';
      if (!fileName.contains('.')) {
        if (resolvedUrl.toLowerCase().contains('.mp4') ||
            _currentAnnouncement.mediaType == 'video') {
          fileName += '.mp4';
        } else {
          fileName += '.jpg';
        }
      }

      final tempDir = await getTemporaryDirectory();
      final savePath = '${tempDir.path}/$fileName';

      await dio.download(resolvedUrl, savePath);

      final file = File(savePath);
      if (await file.exists()) {
        final isVideo =
            fileName.toLowerCase().endsWith('.mp4') ||
            fileName.toLowerCase().endsWith('.mov');
        final mimeType = isVideo ? 'video/mp4' : 'image/jpeg';
        final xFile = XFile(savePath, mimeType: mimeType, name: fileName);

        // ShareXFiles memicu sheet sistem Android/iOS untuk Simpan Berkas, Galeri, atau Forward ke WA
        await Share.shareXFiles([
          xFile,
        ], text: 'Media Pengumuman: ${_currentAnnouncement.title}');

        if (_currentAnnouncement.trackDownloads) {
          _repo.recordActivity(
            _currentAnnouncement.id,
            'download',
            channel: 'media_file',
          );
        }

        if (mounted) {
          ScaffoldMessenger.of(context).showSnackBar(
            SnackBar(
              content: Row(
                children: [
                  const Icon(
                    Icons.check_circle_rounded,
                    color: Colors.white,
                    size: 20,
                  ),
                  const SizedBox(width: 10),
                  Text(
                    'Berkas berhasil diunduh! Silakan simpan ke perangkat.',
                    style: GoogleFonts.plusJakartaSans(
                      fontWeight: FontWeight.w600,
                    ),
                  ),
                ],
              ),
              backgroundColor: const Color(0xFF10B981),
              behavior: SnackBarBehavior.floating,
              shape: RoundedRectangleBorder(
                borderRadius: BorderRadius.circular(12),
              ),
            ),
          );
        }
      }
    } catch (e) {
      if (mounted) {
        ScaffoldMessenger.of(context).showSnackBar(
          SnackBar(
            content: Text('Gagal mengunduh media: $e'),
            backgroundColor: AppColors.primaryRed,
          ),
        );
      }
    } finally {
      if (mounted) {
        setState(() => _isDownloadingMedia = false);
      }
    }
  }

  String _formatCommentTime(DateTime dateTime) {
    final now = DateTime.now();
    final localTime = dateTime.toLocal();
    final diff = now.difference(localTime);

    if (diff.isNegative || diff.inSeconds < 45) {
      return 'Baru saja';
    } else if (diff.inMinutes < 60) {
      return '${diff.inMinutes} mnt lalu';
    } else if (diff.inHours < 24 &&
        localTime.day == now.day &&
        localTime.month == now.month &&
        localTime.year == now.year) {
      final hour = localTime.hour.toString().padLeft(2, '0');
      final minute = localTime.minute.toString().padLeft(2, '0');
      return 'Hari ini $hour:$minute';
    } else if (diff.inDays < 2 ||
        (now.day - localTime.day == 1 && diff.inHours < 48)) {
      final hour = localTime.hour.toString().padLeft(2, '0');
      final minute = localTime.minute.toString().padLeft(2, '0');
      return 'Kemarin $hour:$minute';
    } else {
      return DateFormat('dd/MM HH:mm').format(localTime);
    }
  }

  bool _isGDriveUrl(String url) {
    return url.contains('drive.google.com') || url.contains('/file/d/');
  }

  String _extractGDriveId(String url) {
    final fileDMatch = RegExp(r'/file/d/([a-zA-Z0-9_-]+)').firstMatch(url);
    if (fileDMatch != null) return fileDMatch.group(1)!;
    final idMatch = RegExp(r'id=([a-zA-Z0-9_-]+)').firstMatch(url);
    if (idMatch != null) return idMatch.group(1)!;
    return '';
  }

  String _getGDrivePreviewUrl(String url) {
    final id = _extractGDriveId(url);
    if (id.isNotEmpty) {
      return 'https://drive.google.com/file/d/$id/preview';
    }
    return url;
  }

  String _getGDriveDirectDownloadUrl(String url) {
    final id = _extractGDriveId(url);
    if (id.isNotEmpty) {
      return 'https://drive.google.com/uc?export=download&id=$id';
    }
    return url;
  }

  Future<void> _openExternalLink(String url, {bool inApp = false}) async {
    final uri = Uri.parse(url);
    final mode = inApp
        ? LaunchMode.inAppBrowserView
        : LaunchMode.externalApplication;
    try {
      if (await canLaunchUrl(uri)) {
        await launchUrl(uri, mode: mode);
      } else {
        await launchUrl(uri, mode: LaunchMode.externalApplication);
      }
    } catch (_) {
      try {
        await launchUrl(uri, mode: LaunchMode.externalApplication);
      } catch (e) {
        if (mounted) {
          ScaffoldMessenger.of(context).showSnackBar(
            SnackBar(
              content: Text('Tidak dapat membuka link: $url'),
              backgroundColor: AppColors.primaryRed,
            ),
          );
        }
      }
    }
  }

  void _playVideo(String rawUrl) {
    if (rawUrl.isEmpty) return;
    if (_isGDriveUrl(rawUrl)) {
      final previewUrl = _getGDrivePreviewUrl(rawUrl);
      _openExternalLink(previewUrl, inApp: true);
    } else {
      final resolvedUrl = AppConstants.buildImageUrl(rawUrl);
      _openExternalLink(resolvedUrl, inApp: true);
    }
  }

  void _copyCaption() {
    Clipboard.setData(
      ClipboardData(
        text:
            "${_currentAnnouncement.title}\n\n${_currentAnnouncement.content}",
      ),
    );
    ScaffoldMessenger.of(context).showSnackBar(
      SnackBar(
        content: Row(
          children: [
            const Icon(
              Icons.check_circle_rounded,
              color: Colors.white,
              size: 20,
            ),
            const SizedBox(width: 10),
            Text(
              'Teks pengumuman berhasil disalin!',
              style: GoogleFonts.plusJakartaSans(fontWeight: FontWeight.w600),
            ),
          ],
        ),
        backgroundColor: const Color(0xFF1E293B),
        behavior: SnackBarBehavior.floating,
        shape: RoundedRectangleBorder(borderRadius: BorderRadius.circular(12)),
        duration: const Duration(seconds: 2),
      ),
    );
  }

  void _openFullscreenImage(BuildContext context, int initialIndex) {
    final mediaList = _currentAnnouncement.mediaUrls.isNotEmpty
        ? _currentAnnouncement.mediaUrls
        : (_currentAnnouncement.imageUrl != null
              ? [_currentAnnouncement.imageUrl!]
              : <String>[]);

    if (mediaList.isEmpty) return;
    int currentIndex = initialIndex;

    Navigator.of(context).push(
      MaterialPageRoute(
        builder: (_) => StatefulBuilder(
          builder: (context, setModalState) {
            return Scaffold(
              backgroundColor: Colors.black,
              appBar: AppBar(
                backgroundColor: Colors.black,
                elevation: 0,
                iconTheme: const IconThemeData(color: Colors.white),
                title: Text(
                  '${currentIndex + 1} / ${mediaList.length}',
                  style: GoogleFonts.plusJakartaSans(
                    color: Colors.white,
                    fontSize: 14,
                  ),
                ),
                actions: [
                  IconButton(
                    icon: _isDownloadingMedia
                        ? const SizedBox(
                            width: 18,
                            height: 18,
                            child: CircularProgressIndicator(
                              strokeWidth: 2,
                              color: Colors.white,
                            ),
                          )
                        : const Icon(
                            Icons.download_rounded,
                            color: Colors.white,
                          ),
                    tooltip: 'Unduh Media',
                    onPressed: _isDownloadingMedia
                        ? null
                        : () async {
                            final targetUrl = mediaList[currentIndex];
                            await _downloadMediaFile(targetUrl);
                          },
                  ),
                ],
              ),
              body: PageView.builder(
                controller: PageController(initialPage: initialIndex),
                itemCount: mediaList.length,
                onPageChanged: (idx) {
                  setModalState(() {
                    currentIndex = idx;
                  });
                },
                itemBuilder: (context, index) {
                  return Center(
                    child: InteractiveViewer(
                      panEnabled: true,
                      minScale: 0.5,
                      maxScale: 4.0,
                      child: CachedNetworkImage(
                        imageUrl: AppConstants.buildImageUrl(mediaList[index]),
                        fit: BoxFit.contain,
                        placeholder: (context, url) => const Center(
                          child: CircularProgressIndicator(color: Colors.white),
                        ),
                        errorWidget: (context, url, error) => const Icon(
                          Icons.broken_image_rounded,
                          color: Colors.white54,
                          size: 64,
                        ),
                      ),
                    ),
                  );
                },
              ),
            );
          },
        ),
      ),
    );
  }

  @override
  Widget build(BuildContext context) {
    return BlocListener<AnnouncementBloc, AnnouncementState>(
      listener: (context, state) {
        final found = state.announcements.firstWhere(
          (a) => a.id == _currentAnnouncement.id,
          orElse: () => _currentAnnouncement,
        );
        if (mounted && found != _currentAnnouncement) {
          setState(() {
            _currentAnnouncement = found;
          });
        }
      },
      child: Scaffold(
        backgroundColor: Colors.white,
        appBar: AppBar(
          title: Text(
            _getHeaderTitle(_currentAnnouncement.category),
            style: GoogleFonts.plusJakartaSans(
              fontSize: 13,
              fontWeight: FontWeight.w800,
              letterSpacing: 1.5,
              color: AppColors.textPrimary,
            ),
          ),
          centerTitle: true,
          backgroundColor: Colors.white,
          elevation: 0,
          leading: IconButton(
            icon: const Icon(
              Icons.arrow_back_ios_new_rounded,
              color: AppColors.textPrimary,
              size: 20,
            ),
            onPressed: () => Navigator.of(context).pop(),
          ),
          actions: [
            if (_currentAnnouncement.imageUrl != null &&
                    _currentAnnouncement.imageUrl!.isNotEmpty ||
                _currentAnnouncement.mediaUrls.isNotEmpty)
              IconButton(
                icon: _isDownloadingMedia
                    ? const SizedBox(
                        width: 18,
                        height: 18,
                        child: CircularProgressIndicator(
                          strokeWidth: 2,
                          color: AppColors.primaryRed,
                        ),
                      )
                    : const Icon(
                        Icons.download_rounded,
                        color: AppColors.textPrimary,
                      ),
                tooltip: 'Unduh Media',
                onPressed: _isDownloadingMedia
                    ? null
                    : () {
                        final rawUrl =
                            _currentAnnouncement.imageUrl != null &&
                                _currentAnnouncement.imageUrl!.isNotEmpty
                            ? _currentAnnouncement.imageUrl!
                            : (_currentAnnouncement.mediaUrls.isNotEmpty
                                  ? _currentAnnouncement.mediaUrls.first
                                  : '');
                        if (rawUrl.isNotEmpty) {
                          _downloadMediaFile(rawUrl);
                        }
                      },
              ),
            IconButton(
              icon: const Icon(
                Icons.share_outlined,
                color: AppColors.textPrimary,
              ),
              tooltip: 'Bagikan',
              onPressed: _shareGeneral,
            ),
          ],
        ),
        body: SingleChildScrollView(
          controller: _scrollController,
          child: SizedBox(
            width: double.infinity,
            child: Column(
              crossAxisAlignment: CrossAxisAlignment.start,
              children: [
                // Media Video / Carousel / Single Image
                if (_isVideoAnnouncement())
                  _buildVideoMediaCard()
                else if (_currentAnnouncement.mediaUrls.isNotEmpty)
                  _buildMediaCarousel()
                else if (_currentAnnouncement.imageUrl != null &&
                    _currentAnnouncement.imageUrl!.isNotEmpty)
                  _buildSingleImage(_currentAnnouncement.imageUrl!),

                // Title & Category Header
                Padding(
                  padding: const EdgeInsets.fromLTRB(20, 16, 20, 8),
                  child: Column(
                    crossAxisAlignment: CrossAxisAlignment.start,
                    children: [
                      Row(
                        mainAxisAlignment: MainAxisAlignment.spaceBetween,
                        crossAxisAlignment: CrossAxisAlignment.center,
                        children: [
                          _buildCategoryBadge(_currentAnnouncement.category),
                          Row(
                            children: [
                              const Icon(
                                Icons.calendar_month_rounded,
                                size: 14,
                                color: AppColors.textTertiary,
                              ),
                              const SizedBox(width: 6),
                              Text(
                                DateFormat('dd MMM yyyy, HH:mm').format(
                                  _currentAnnouncement.createdAt.toLocal(),
                                ),
                                style: GoogleFonts.plusJakartaSans(
                                  fontSize: 11,
                                  fontWeight: FontWeight.w700,
                                  color: AppColors.textTertiary,
                                ),
                              ),
                            ],
                          ),
                        ],
                      ),
                      const SizedBox(height: 14),
                      Text(
                        _currentAnnouncement.title,
                        style: GoogleFonts.plusJakartaSans(
                          fontSize: 20,
                          fontWeight: FontWeight.w900,
                          color: AppColors.textPrimary,
                          height: 1.3,
                        ),
                      ),
                      if (_currentAnnouncement.author != null &&
                          _currentAnnouncement.author!.isNotEmpty) ...[
                        const SizedBox(height: 8),
                        Row(
                          mainAxisSize: MainAxisSize.min,
                          children: [
                            Container(
                              padding: const EdgeInsets.symmetric(
                                horizontal: 8,
                                vertical: 3,
                              ),
                              decoration: BoxDecoration(
                                color: Colors.grey.shade100,
                                borderRadius: BorderRadius.circular(6),
                              ),
                              child: Row(
                                mainAxisSize: MainAxisSize.min,
                                children: [
                                  const Icon(
                                    Icons.person_outline_rounded,
                                    size: 13,
                                    color: AppColors.textSecondary,
                                  ),
                                  const SizedBox(width: 5),
                                  Text(
                                    'Oleh: ${_currentAnnouncement.author!}',
                                    style: GoogleFonts.plusJakartaSans(
                                      fontSize: 11,
                                      fontWeight: FontWeight.w700,
                                      color: AppColors.textSecondary,
                                    ),
                                  ),
                                ],
                              ),
                            ),
                          ],
                        ),
                      ],
                    ],
                  ),
                ),

                // Engagement Action Bar (Like, Share WA, Share Sosmed, Copy)
                _buildInteractionBar(),

                // External Media / Social Link Button
                if (_currentAnnouncement.externalUrl != null &&
                    _currentAnnouncement.externalUrl!.isNotEmpty)
                  _buildExternalLinkCard(_currentAnnouncement.externalUrl!),

                const Divider(
                  height: 24,
                  thickness: 1,
                  color: Color(0xFFF1F5F9),
                ),

                // Content / Caption
                Padding(
                  padding: const EdgeInsets.symmetric(horizontal: 20.0),
                  child: Column(
                    crossAxisAlignment: CrossAxisAlignment.start,
                    children: [
                      Row(
                        mainAxisAlignment: MainAxisAlignment.spaceBetween,
                        children: [
                          Text(
                            'DESKRIPSI & KETERANGAN',
                            style: GoogleFonts.plusJakartaSans(
                              fontSize: 11,
                              fontWeight: FontWeight.w800,
                              color: AppColors.primaryRed,
                              letterSpacing: 1.5,
                            ),
                          ),
                          InkWell(
                            onTap: _copyCaption,
                            borderRadius: BorderRadius.circular(8),
                            child: Padding(
                              padding: const EdgeInsets.symmetric(
                                horizontal: 6,
                                vertical: 4,
                              ),
                              child: Row(
                                mainAxisSize: MainAxisSize.min,
                                children: [
                                  const Icon(
                                    Icons.copy_rounded,
                                    size: 14,
                                    color: Color(0xFF0284C7),
                                  ),
                                  const SizedBox(width: 4),
                                  Text(
                                    'Salin Teks',
                                    style: GoogleFonts.plusJakartaSans(
                                      fontSize: 11,
                                      fontWeight: FontWeight.w700,
                                      color: const Color(0xFF0284C7),
                                    ),
                                  ),
                                ],
                              ),
                            ),
                          ),
                        ],
                      ),
                      const SizedBox(height: 12),
                      SelectionArea(
                        child: Text(
                          _currentAnnouncement.content,
                          style: GoogleFonts.plusJakartaSans(
                            fontSize: 14,
                            color: AppColors.textPrimary,
                            height: 1.7,
                            fontWeight: FontWeight.w500,
                          ),
                        ),
                      ),
                      const SizedBox(height: 32),

                      // Quick Share Promo Box
                      _buildSharePromoBox(),
                      const SizedBox(height: 32),

                      const Divider(
                        height: 1,
                        thickness: 1,
                        color: Color(0xFFF1F5F9),
                      ),
                      const SizedBox(height: 24),

                      // Comments & Discussion Section
                      _buildCommentsSection(),
                      const SizedBox(height: 32),
                    ],
                  ),
                ),
              ],
            ),
          ),
        ),
        bottomNavigationBar: _buildCommentInputBar(),
      ),
    );
  }

  bool _isVideoAnnouncement() {
    final mediaType = _currentAnnouncement.mediaType.toLowerCase();
    final url = (_currentAnnouncement.imageUrl ?? '').toLowerCase();
    final extUrl = (_currentAnnouncement.externalUrl ?? '').toLowerCase();
    return mediaType == 'video' ||
        mediaType == 'video_url' ||
        url.endsWith('.mp4') ||
        url.endsWith('.mov') ||
        url.endsWith('.webm') ||
        url.endsWith('.m4v') ||
        _isGDriveUrl(url) ||
        _isGDriveUrl(extUrl);
  }

  Widget _buildVideoMediaCard() {
    final rawUrl = _currentAnnouncement.imageUrl ?? '';
    final isGDrive =
        _isGDriveUrl(rawUrl) ||
        _isGDriveUrl(_currentAnnouncement.externalUrl ?? '');

    return Container(
      width: double.infinity,
      margin: const EdgeInsets.symmetric(horizontal: 16, vertical: 12),
      decoration: BoxDecoration(
        color: const Color(0xFF0F172A),
        borderRadius: BorderRadius.circular(20),
        boxShadow: [
          BoxShadow(
            color: Colors.black.withOpacity(0.12),
            blurRadius: 16,
            offset: const Offset(0, 6),
          ),
        ],
      ),
      child: Column(
        children: [
          // Video Banner / Player Launcher
          InkWell(
            onTap: () {
              final targetUrl = rawUrl.isNotEmpty
                  ? rawUrl
                  : (_currentAnnouncement.externalUrl ?? '');
              _playVideo(targetUrl);
            },
            borderRadius: const BorderRadius.vertical(top: Radius.circular(20)),
            child: Container(
              height: 180,
              width: double.infinity,
              decoration: BoxDecoration(
                borderRadius: const BorderRadius.vertical(
                  top: Radius.circular(20),
                ),
                gradient: LinearGradient(
                  colors: [Colors.indigo.shade900, const Color(0xFF1E1B4B)],
                  begin: Alignment.topLeft,
                  end: Alignment.bottomRight,
                ),
              ),
              child: Stack(
                alignment: Alignment.center,
                children: [
                  Positioned.fill(
                    child: Opacity(
                      opacity: 0.15,
                      child: Container(
                        decoration: const BoxDecoration(
                          image: DecorationImage(
                            image: AssetImage('assets/brand.png'),
                            fit: BoxFit.contain,
                          ),
                        ),
                      ),
                    ),
                  ),
                  Container(
                    width: 64,
                    height: 64,
                    decoration: BoxDecoration(
                      color: Colors.white.withOpacity(0.9),
                      shape: BoxShape.circle,
                      boxShadow: [
                        BoxShadow(
                          color: Colors.black.withOpacity(0.2),
                          blurRadius: 12,
                          offset: const Offset(0, 4),
                        ),
                      ],
                    ),
                    child: const Icon(
                      Icons.play_arrow_rounded,
                      color: Color(0xFF4338CA),
                      size: 40,
                    ),
                  ),
                  Positioned(
                    bottom: 12,
                    child: Container(
                      padding: const EdgeInsets.symmetric(
                        horizontal: 12,
                        vertical: 4,
                      ),
                      decoration: BoxDecoration(
                        color: Colors.black54,
                        borderRadius: BorderRadius.circular(12),
                      ),
                      child: Text(
                        isGDrive
                            ? 'Ketuk untuk Putar (Google Drive Player)'
                            : 'Ketuk untuk Putar Video',
                        style: GoogleFonts.plusJakartaSans(
                          color: Colors.white,
                          fontSize: 11,
                          fontWeight: FontWeight.w700,
                        ),
                      ),
                    ),
                  ),
                ],
              ),
            ),
          ),

          // Action Toolbar for Video
          Padding(
            padding: const EdgeInsets.symmetric(horizontal: 16, vertical: 12),
            child: Row(
              children: [
                const Icon(
                  Icons.videocam_rounded,
                  color: Colors.indigoAccent,
                  size: 20,
                ),
                const SizedBox(width: 8),
                Expanded(
                  child: Text(
                    isGDrive
                        ? 'Video Pengumuman (Google Drive)'
                        : 'Lampiran Video Pengumuman',
                    style: GoogleFonts.plusJakartaSans(
                      color: Colors.white,
                      fontSize: 12,
                      fontWeight: FontWeight.w700,
                    ),
                  ),
                ),
                Material(
                  color: Colors.transparent,
                  child: InkWell(
                    onTap: _isDownloadingMedia
                        ? null
                        : () => _downloadMediaFile(rawUrl),
                    borderRadius: BorderRadius.circular(10),
                    child: Container(
                      padding: const EdgeInsets.symmetric(
                        horizontal: 12,
                        vertical: 8,
                      ),
                      decoration: BoxDecoration(
                        color: _isDownloadingMedia
                            ? const Color(0xFF4F46E5).withOpacity(0.6)
                            : const Color(0xFF4F46E5),
                        borderRadius: BorderRadius.circular(10),
                      ),
                      child: Row(
                        mainAxisSize: MainAxisSize.min,
                        children: [
                          if (_isDownloadingMedia) ...[
                            const SizedBox(
                              width: 14,
                              height: 14,
                              child: CircularProgressIndicator(
                                strokeWidth: 2,
                                color: Colors.white,
                              ),
                            ),
                            const SizedBox(width: 6),
                          ] else ...[
                            const Icon(
                              Icons.download_rounded,
                              size: 16,
                              color: Colors.white,
                            ),
                            const SizedBox(width: 4),
                          ],
                          Text(
                            _isDownloadingMedia
                                ? 'Mengunduh...'
                                : 'Unduh Video',
                            style: GoogleFonts.plusJakartaSans(
                              fontSize: 11,
                              fontWeight: FontWeight.w800,
                              color: Colors.white,
                            ),
                          ),
                        ],
                      ),
                    ),
                  ),
                ),
              ],
            ),
          ),
        ],
      ),
    );
  }

  Widget _buildMediaCarousel() {
    final urls = _currentAnnouncement.mediaUrls;
    return Column(
      children: [
        SizedBox(
          height: 260,
          child: PageView.builder(
            controller: _pageController,
            itemCount: urls.length,
            onPageChanged: (index) {
              setState(() {
                _activeImageIndex = index;
              });
            },
            itemBuilder: (context, index) {
              return GestureDetector(
                onTap: () => _openFullscreenImage(context, index),
                child: Container(
                  margin: const EdgeInsets.symmetric(
                    horizontal: 16,
                    vertical: 8,
                  ),
                  decoration: BoxDecoration(
                    borderRadius: BorderRadius.circular(16),
                    boxShadow: [
                      BoxShadow(
                        color: Colors.black.withOpacity(0.08),
                        blurRadius: 10,
                        offset: const Offset(0, 4),
                      ),
                    ],
                  ),
                  child: ClipRRect(
                    borderRadius: BorderRadius.circular(16),
                    child: CachedNetworkImage(
                      imageUrl: urls[index],
                      fit: BoxFit.cover,
                      width: double.infinity,
                      placeholder: (context, url) => Container(
                        color: Colors.grey.shade100,
                        child: const Center(
                          child: CircularProgressIndicator(strokeWidth: 2),
                        ),
                      ),
                      errorWidget: (context, url, error) => Container(
                        color: Colors.grey.shade100,
                        child: const Center(
                          child: Icon(
                            Icons.image_not_supported_rounded,
                            color: Colors.grey,
                            size: 40,
                          ),
                        ),
                      ),
                    ),
                  ),
                ),
              );
            },
          ),
        ),
        if (urls.length > 1)
          Padding(
            padding: const EdgeInsets.only(bottom: 8),
            child: Row(
              mainAxisAlignment: MainAxisAlignment.center,
              children: List.generate(urls.length, (index) {
                return AnimatedContainer(
                  duration: const Duration(milliseconds: 200),
                  margin: const EdgeInsets.symmetric(horizontal: 3),
                  width: _activeImageIndex == index ? 20 : 6,
                  height: 6,
                  decoration: BoxDecoration(
                    color: _activeImageIndex == index
                        ? AppColors.primaryRed
                        : Colors.grey.shade300,
                    borderRadius: BorderRadius.circular(3),
                  ),
                );
              }),
            ),
          ),
      ],
    );
  }

  Widget _buildSingleImage(String url) {
    return GestureDetector(
      onTap: () => _openFullscreenImage(context, 0),
      child: Container(
        margin: const EdgeInsets.all(16),
        height: 240,
        width: double.infinity,
        decoration: BoxDecoration(
          borderRadius: BorderRadius.circular(16),
          boxShadow: [
            BoxShadow(
              color: Colors.black.withOpacity(0.08),
              blurRadius: 10,
              offset: const Offset(0, 4),
            ),
          ],
        ),
        child: ClipRRect(
          borderRadius: BorderRadius.circular(16),
          child: CachedNetworkImage(
            imageUrl: url,
            fit: BoxFit.cover,
            placeholder: (context, url) => Container(
              color: Colors.grey.shade100,
              child: const Center(
                child: CircularProgressIndicator(strokeWidth: 2),
              ),
            ),
            errorWidget: (context, url, error) => Container(
              color: Colors.grey.shade100,
              child: const Center(
                child: Icon(
                  Icons.image_not_supported_rounded,
                  color: Colors.grey,
                  size: 40,
                ),
              ),
            ),
          ),
        ),
      ),
    );
  }

  Widget _buildInteractionBar() {
    final isLiked = _currentAnnouncement.isLikedByMe;
    final likesCount = _currentAnnouncement.likesCount;
    final commentsCount = _comments.isNotEmpty
        ? _comments.length
        : _currentAnnouncement.commentsCount;

    return Container(
      margin: const EdgeInsets.symmetric(horizontal: 20, vertical: 8),
      child: Row(
        children: [
          // Like Button Pill
          Expanded(
            child: Material(
              color: isLiked
                  ? const Color(0xFFFEF2F2)
                  : const Color(0xFFF8FAFC),
              borderRadius: BorderRadius.circular(12),
              child: InkWell(
                onTap: () {
                  context.read<AnnouncementBloc>().add(
                    ToggleLikeAnnouncementEvent(_currentAnnouncement.id),
                  );
                },
                borderRadius: BorderRadius.circular(12),
                child: Container(
                  height: 44,
                  padding: const EdgeInsets.symmetric(horizontal: 12),
                  decoration: BoxDecoration(
                    borderRadius: BorderRadius.circular(12),
                    border: Border.all(
                      color: isLiked
                          ? const Color(0xFFFECACA)
                          : const Color(0xFFE2E8F0),
                    ),
                  ),
                  child: Row(
                    mainAxisAlignment: MainAxisAlignment.center,
                    children: [
                      AnimatedSwitcher(
                        duration: const Duration(milliseconds: 250),
                        transitionBuilder: (child, anim) =>
                            ScaleTransition(scale: anim, child: child),
                        child: Icon(
                          isLiked
                              ? Icons.favorite_rounded
                              : Icons.favorite_border_rounded,
                          key: ValueKey<bool>(isLiked),
                          color: isLiked
                              ? Colors.red.shade600
                              : AppColors.textSecondary,
                          size: 20,
                        ),
                      ),
                      const SizedBox(width: 8),
                      Text(
                        likesCount > 0 ? '$likesCount Suka' : 'Suka',
                        style: GoogleFonts.plusJakartaSans(
                          fontSize: 13,
                          fontWeight: FontWeight.w700,
                          color: isLiked
                              ? Colors.red.shade600
                              : AppColors.textPrimary,
                        ),
                      ),
                    ],
                  ),
                ),
              ),
            ),
          ),
          const SizedBox(width: 10),

          // Comment Button Pill
          Expanded(
            child: Material(
              color: const Color(0xFFF8FAFC),
              borderRadius: BorderRadius.circular(12),
              child: InkWell(
                onTap: () {
                  _commentFocusNode.requestFocus();
                },
                borderRadius: BorderRadius.circular(12),
                child: Container(
                  height: 44,
                  padding: const EdgeInsets.symmetric(horizontal: 12),
                  decoration: BoxDecoration(
                    borderRadius: BorderRadius.circular(12),
                    border: Border.all(color: const Color(0xFFE2E8F0)),
                  ),
                  child: Row(
                    mainAxisAlignment: MainAxisAlignment.center,
                    children: [
                      const Icon(
                        Icons.chat_bubble_outline_rounded,
                        color: Color(0xFF0284C7),
                        size: 19,
                      ),
                      const SizedBox(width: 8),
                      Text(
                        commentsCount > 0
                            ? '$commentsCount Komentar'
                            : 'Komentar',
                        style: GoogleFonts.plusJakartaSans(
                          fontSize: 13,
                          fontWeight: FontWeight.w700,
                          color: AppColors.textPrimary,
                        ),
                      ),
                    ],
                  ),
                ),
              ),
            ),
          ),
          const SizedBox(width: 10),

          // Quick Share Icon Button
          Material(
            color: const Color(0xFFF8FAFC),
            borderRadius: BorderRadius.circular(12),
            child: InkWell(
              onTap: _shareGeneral,
              borderRadius: BorderRadius.circular(12),
              child: Container(
                height: 44,
                width: 44,
                decoration: BoxDecoration(
                  borderRadius: BorderRadius.circular(12),
                  border: Border.all(color: const Color(0xFFE2E8F0)),
                ),
                child: const Icon(
                  Icons.share_outlined,
                  color: AppColors.textSecondary,
                  size: 20,
                ),
              ),
            ),
          ),
        ],
      ),
    );
  }

  Widget _buildCommentsSection() {
    return Column(
      crossAxisAlignment: CrossAxisAlignment.start,
      children: [
        Row(
          mainAxisAlignment: MainAxisAlignment.spaceBetween,
          children: [
            Row(
              mainAxisSize: MainAxisSize.min,
              children: [
                const Icon(
                  Icons.chat_bubble_rounded,
                  size: 16,
                  color: AppColors.primaryRed,
                ),
                const SizedBox(width: 8),
                Text(
                  'DISKUSI & KOMENTAR (${_comments.length})',
                  style: GoogleFonts.plusJakartaSans(
                    fontSize: 12,
                    fontWeight: FontWeight.w800,
                    color: AppColors.primaryRed,
                    letterSpacing: 1.2,
                  ),
                ),
              ],
            ),
            if (_isLoadingComments)
              const SizedBox(
                width: 14,
                height: 14,
                child: CircularProgressIndicator(
                  strokeWidth: 2,
                  color: AppColors.primaryRed,
                ),
              ),
          ],
        ),
        const SizedBox(height: 14),

        if (!_isLoadingComments && _comments.isEmpty)
          Container(
            width: double.infinity,
            padding: const EdgeInsets.symmetric(vertical: 24, horizontal: 16),
            decoration: BoxDecoration(
              color: const Color(0xFFF8FAFC),
              borderRadius: BorderRadius.circular(16),
              border: Border.all(color: const Color(0xFFE2E8F0)),
            ),
            child: Column(
              children: [
                Icon(
                  Icons.chat_bubble_outline_rounded,
                  size: 36,
                  color: Colors.grey.shade400,
                ),
                const SizedBox(height: 8),
                Text(
                  'Belum ada komentar',
                  style: GoogleFonts.plusJakartaSans(
                    fontWeight: FontWeight.w700,
                    fontSize: 13,
                    color: AppColors.textSecondary,
                  ),
                ),
                const SizedBox(height: 4),
                Text(
                  'Jadilah yang pertama memberikan tanggapan atau feedback!',
                  style: GoogleFonts.plusJakartaSans(
                    fontSize: 11,
                    color: AppColors.textTertiary,
                  ),
                  textAlign: TextAlign.center,
                ),
              ],
            ),
          )
        else
          ListView.separated(
            shrinkWrap: true,
            physics: const NeverScrollableScrollPhysics(),
            itemCount: _comments.length,
            separatorBuilder: (_, __) => const SizedBox(height: 12),
            itemBuilder: (context, index) {
              final comment = _comments[index];
              // Resolve author name with local fallback if it's the current user's comment
              String displayName = comment.userName.trim();
              if (comment.isMine &&
                  (displayName.isEmpty || displayName == 'Karyawan') &&
                  _cachedMyName.isNotEmpty) {
                displayName = _cachedMyName;
              } else if (displayName.isEmpty) {
                displayName = 'Karyawan';
              }

              // Resolve author avatar with local fallback
              String displayAvatar = comment.userAvatar.trim();
              if (comment.isMine &&
                  displayAvatar.isEmpty &&
                  _cachedMyAvatar.isNotEmpty) {
                displayAvatar = _cachedMyAvatar;
              }

              final initialLetter = displayName.isNotEmpty
                  ? displayName[0].toUpperCase()
                  : 'U';

              return Container(
                padding: const EdgeInsets.all(12),
                decoration: BoxDecoration(
                  color: const Color(0xFFF8FAFC),
                  borderRadius: BorderRadius.circular(14),
                  border: Border.all(color: const Color(0xFFE2E8F0)),
                ),
                child: Column(
                  crossAxisAlignment: CrossAxisAlignment.start,
                  children: [
                    Row(
                      crossAxisAlignment: CrossAxisAlignment.center,
                      children: [
                        ClipOval(
                          child: SizedBox(
                            width: 28,
                            height: 28,
                            child: displayAvatar.isNotEmpty
                                ? CachedNetworkImage(
                                    imageUrl: AppConstants.buildFaceImageUrl(
                                      displayAvatar,
                                    ),
                                    fit: BoxFit.cover,
                                    placeholder: (context, url) => Container(
                                      color: AppColors.primaryRed.withOpacity(
                                        0.12,
                                      ),
                                      alignment: Alignment.center,
                                      child: Text(
                                        initialLetter,
                                        style: GoogleFonts.plusJakartaSans(
                                          fontSize: 11,
                                          fontWeight: FontWeight.w800,
                                          color: AppColors.primaryRed,
                                        ),
                                      ),
                                    ),
                                    errorWidget: (context, url, error) =>
                                        Container(
                                          color: AppColors.primaryRed
                                              .withOpacity(0.12),
                                          alignment: Alignment.center,
                                          child: Text(
                                            initialLetter,
                                            style: GoogleFonts.plusJakartaSans(
                                              fontSize: 11,
                                              fontWeight: FontWeight.w800,
                                              color: AppColors.primaryRed,
                                            ),
                                          ),
                                        ),
                                  )
                                : Container(
                                    color: AppColors.primaryRed.withOpacity(
                                      0.12,
                                    ),
                                    alignment: Alignment.center,
                                    child: Text(
                                      initialLetter,
                                      style: GoogleFonts.plusJakartaSans(
                                        fontSize: 11,
                                        fontWeight: FontWeight.w800,
                                        color: AppColors.primaryRed,
                                      ),
                                    ),
                                  ),
                          ),
                        ),
                        const SizedBox(width: 8),
                        Expanded(
                          child: Column(
                            crossAxisAlignment: CrossAxisAlignment.start,
                            children: [
                              Text(
                                displayName,
                                style: GoogleFonts.plusJakartaSans(
                                  fontWeight: FontWeight.w800,
                                  fontSize: 12,
                                  color: AppColors.textPrimary,
                                ),
                              ),
                              if (comment.userRole.isNotEmpty)
                                Text(
                                  comment.userRole,
                                  style: GoogleFonts.plusJakartaSans(
                                    fontSize: 10,
                                    fontWeight: FontWeight.w600,
                                    color: AppColors.textTertiary,
                                  ),
                                ),
                            ],
                          ),
                        ),
                        Text(
                          _formatCommentTime(comment.createdAt),
                          style: GoogleFonts.plusJakartaSans(
                            fontSize: 10,
                            color: AppColors.textTertiary,
                          ),
                        ),
                        if (comment.isMine) ...[
                          const SizedBox(width: 6),
                          InkWell(
                            onTap: () => _deleteComment(comment),
                            borderRadius: BorderRadius.circular(6),
                            child: const Padding(
                              padding: EdgeInsets.all(4),
                              child: Icon(
                                Icons.delete_outline_rounded,
                                size: 16,
                                color: Colors.redAccent,
                              ),
                            ),
                          ),
                        ],
                      ],
                    ),
                    const SizedBox(height: 8),
                    Text(
                      comment.content,
                      style: GoogleFonts.plusJakartaSans(
                        fontSize: 13,
                        color: const Color(0xFF334155),
                        height: 1.4,
                      ),
                    ),
                  ],
                ),
              );
            },
          ),
      ],
    );
  }

  Widget _buildCommentInputBar() {
    return Container(
      padding: EdgeInsets.only(
        left: 16,
        right: 16,
        top: 10,
        bottom: MediaQuery.of(context).viewInsets.bottom > 0
            ? MediaQuery.of(context).viewInsets.bottom + 10
            : 16,
      ),
      decoration: BoxDecoration(
        color: Colors.white,
        boxShadow: [
          BoxShadow(
            color: Colors.black.withOpacity(0.06),
            blurRadius: 10,
            offset: const Offset(0, -3),
          ),
        ],
      ),
      child: SafeArea(
        top: false,
        child: Row(
          children: [
            Expanded(
              child: Container(
                decoration: BoxDecoration(
                  color: const Color(0xFFF1F5F9),
                  borderRadius: BorderRadius.circular(24),
                ),
                child: TextField(
                  controller: _commentController,
                  focusNode: _commentFocusNode,
                  maxLines: 4,
                  minLines: 1,
                  style: GoogleFonts.plusJakartaSans(fontSize: 13),
                  decoration: InputDecoration(
                    hintText: 'Tulis komentar atau tanggapan...',
                    hintStyle: GoogleFonts.plusJakartaSans(
                      fontSize: 12,
                      color: AppColors.textTertiary,
                    ),
                    contentPadding: const EdgeInsets.symmetric(
                      horizontal: 16,
                      vertical: 10,
                    ),
                    border: InputBorder.none,
                  ),
                ),
              ),
            ),
            const SizedBox(width: 8),
            InkWell(
              onTap: _isSubmittingComment ? null : _submitComment,
              borderRadius: BorderRadius.circular(24),
              child: Container(
                padding: const EdgeInsets.all(10),
                decoration: const BoxDecoration(
                  color: AppColors.primaryRed,
                  shape: BoxShape.circle,
                ),
                child: _isSubmittingComment
                    ? const SizedBox(
                        width: 18,
                        height: 18,
                        child: CircularProgressIndicator(
                          strokeWidth: 2,
                          color: Colors.white,
                        ),
                      )
                    : const Icon(
                        Icons.send_rounded,
                        color: Colors.white,
                        size: 18,
                      ),
              ),
            ),
          ],
        ),
      ),
    );
  }

  Widget _buildExternalLinkCard(String url) {
    IconData linkIcon = Icons.open_in_new_rounded;
    Color linkColor = const Color(0xFF4F46E5);
    String label = 'Buka Tautan Postingan Asli';

    if (url.contains('instagram.com')) {
      linkIcon = Icons.camera_alt_rounded;
      linkColor = const Color(0xFFE1306C);
      label = 'Tonton di Instagram';
    } else if (url.contains('tiktok.com')) {
      linkIcon = Icons.music_note_rounded;
      linkColor = Colors.black87;
      label = 'Tonton di TikTok';
    } else if (url.contains('youtube.com') || url.contains('youtu.be')) {
      linkIcon = Icons.play_circle_fill_rounded;
      linkColor = const Color(0xFFFF0000);
      label = 'Tonton Video di YouTube';
    }

    return Container(
      margin: const EdgeInsets.symmetric(horizontal: 20, vertical: 6),
      child: Material(
        color: Colors.transparent,
        child: InkWell(
          onTap: () => _openExternalLink(url),
          borderRadius: BorderRadius.circular(16),
          child: Container(
            padding: const EdgeInsets.symmetric(horizontal: 16, vertical: 14),
            decoration: BoxDecoration(
              color: linkColor.withValues(alpha: 0.08),
              borderRadius: BorderRadius.circular(16),
              border: Border.all(color: linkColor.withValues(alpha: 0.3)),
            ),
            child: Row(
              children: [
                Container(
                  padding: const EdgeInsets.all(10),
                  decoration: BoxDecoration(
                    color: linkColor,
                    shape: BoxShape.circle,
                  ),
                  child: Icon(linkIcon, color: Colors.white, size: 18),
                ),
                const SizedBox(width: 14),
                Expanded(
                  child: Column(
                    crossAxisAlignment: CrossAxisAlignment.start,
                    children: [
                      Text(
                        label,
                        style: GoogleFonts.plusJakartaSans(
                          fontSize: 13,
                          fontWeight: FontWeight.w800,
                          color: linkColor,
                        ),
                      ),
                      const SizedBox(height: 2),
                      Text(
                        url,
                        style: GoogleFonts.plusJakartaSans(
                          fontSize: 11,
                          color: AppColors.textTertiary,
                        ),
                        maxLines: 1,
                        overflow: TextOverflow.ellipsis,
                      ),
                    ],
                  ),
                ),
                Container(
                  padding: const EdgeInsets.all(6),
                  decoration: BoxDecoration(
                    color: linkColor.withValues(alpha: 0.12),
                    shape: BoxShape.circle,
                  ),
                  child: Icon(
                    Icons.arrow_forward_ios_rounded,
                    size: 12,
                    color: linkColor,
                  ),
                ),
              ],
            ),
          ),
        ),
      ),
    );
  }

  Widget _buildSharePromoBox() {
    return Container(
      padding: const EdgeInsets.all(18),
      decoration: BoxDecoration(
        gradient: const LinearGradient(
          colors: [Color(0xFFF0FDF4), Color(0xFFF8FAFC)],
          begin: Alignment.topLeft,
          end: Alignment.bottomRight,
        ),
        borderRadius: BorderRadius.circular(20),
        border: Border.all(color: const Color(0xFFBBF7D0)),
        boxShadow: [
          BoxShadow(
            color: const Color(0xFF22C55E).withValues(alpha: 0.08),
            blurRadius: 14,
            offset: const Offset(0, 4),
          ),
        ],
      ),
      child: Column(
        crossAxisAlignment: CrossAxisAlignment.start,
        children: [
          // Header info
          Row(
            children: [
              Container(
                padding: const EdgeInsets.all(9),
                decoration: BoxDecoration(
                  color: const Color(0xFF22C55E).withValues(alpha: 0.15),
                  shape: BoxShape.circle,
                ),
                child: const Icon(
                  Icons.campaign_rounded,
                  color: Color(0xFF15803D),
                  size: 22,
                ),
              ),
              const SizedBox(width: 12),
              Expanded(
                child: Column(
                  crossAxisAlignment: CrossAxisAlignment.start,
                  children: [
                    Text(
                      'SEBARKAN KE PELANGGAN & MEDSOS',
                      style: GoogleFonts.plusJakartaSans(
                        fontSize: 11,
                        fontWeight: FontWeight.w800,
                        color: const Color(0xFF15803D),
                        letterSpacing: 1.1,
                      ),
                    ),
                    const SizedBox(height: 2),
                    Text(
                      'Materi promosi siap kirim ke WhatsApp dan medsos.',
                      style: GoogleFonts.plusJakartaSans(
                        fontSize: 12,
                        color: AppColors.textSecondary,
                        fontWeight: FontWeight.w500,
                      ),
                    ),
                  ],
                ),
              ),
            ],
          ),
          const SizedBox(height: 16),

          // Primary Big WhatsApp Button (52px height)
          Material(
            color: Colors.transparent,
            child: InkWell(
              onTap: _shareToWhatsApp,
              borderRadius: BorderRadius.circular(14),
              child: Container(
                height: 52,
                width: double.infinity,
                padding: const EdgeInsets.symmetric(horizontal: 16),
                decoration: BoxDecoration(
                  gradient: const LinearGradient(
                    colors: [Color(0xFF25D366), Color(0xFF1EBE5D)],
                    begin: Alignment.centerLeft,
                    end: Alignment.centerRight,
                  ),
                  borderRadius: BorderRadius.circular(14),
                  boxShadow: [
                    BoxShadow(
                      color: const Color(0xFF25D366).withValues(alpha: 0.35),
                      blurRadius: 10,
                      offset: const Offset(0, 4),
                    ),
                  ],
                ),
                child: Row(
                  children: [
                    Container(
                      padding: const EdgeInsets.all(6),
                      decoration: BoxDecoration(
                        color: Colors.white.withValues(alpha: 0.2),
                        shape: BoxShape.circle,
                      ),
                      child: const Icon(
                        Icons.chat_rounded,
                        color: Colors.white,
                        size: 20,
                      ),
                    ),
                    const SizedBox(width: 12),
                    Expanded(
                      child: Column(
                        mainAxisAlignment: MainAxisAlignment.center,
                        crossAxisAlignment: CrossAxisAlignment.start,
                        children: [
                          Text(
                            'Bagikan ke WhatsApp',
                            style: GoogleFonts.plusJakartaSans(
                              fontSize: 14,
                              fontWeight: FontWeight.w800,
                              color: Colors.white,
                              letterSpacing: 0.2,
                            ),
                          ),
                          Text(
                            'Kirim langsung ke Chat / Status WA',
                            style: GoogleFonts.plusJakartaSans(
                              fontSize: 10,
                              fontWeight: FontWeight.w600,
                              color: Colors.white.withValues(alpha: 0.9),
                            ),
                          ),
                        ],
                      ),
                    ),
                    const Icon(
                      Icons.arrow_forward_rounded,
                      color: Colors.white,
                      size: 18,
                    ),
                  ],
                ),
              ),
            ),
          ),
          const SizedBox(height: 10),

          // Secondary Action Grid (44px height each)
          Row(
            children: [
              // Bagikan ke Medsos Lain
              Expanded(
                child: OutlinedButton.icon(
                  onPressed: _shareGeneral,
                  icon: const Icon(
                    Icons.share_rounded,
                    size: 16,
                    color: AppColors.textPrimary,
                  ),
                  label: Text(
                    'Bagikan Lainnya',
                    style: GoogleFonts.plusJakartaSans(
                      fontSize: 12,
                      fontWeight: FontWeight.w700,
                      color: AppColors.textPrimary,
                    ),
                  ),
                  style: OutlinedButton.styleFrom(
                    minimumSize: const Size(0, 44),
                    backgroundColor: Colors.white,
                    side: const BorderSide(color: Color(0xFFCBD5E1)),
                    shape: RoundedRectangleBorder(
                      borderRadius: BorderRadius.circular(12),
                    ),
                  ),
                ),
              ),
              const SizedBox(width: 8),

              // Salin Caption
              Expanded(
                child: OutlinedButton.icon(
                  onPressed: _copyCaption,
                  icon: const Icon(
                    Icons.copy_rounded,
                    size: 16,
                    color: Color(0xFF0284C7),
                  ),
                  label: Text(
                    'Salin Teks',
                    style: GoogleFonts.plusJakartaSans(
                      fontSize: 12,
                      fontWeight: FontWeight.w700,
                      color: const Color(0xFF0284C7),
                    ),
                  ),
                  style: OutlinedButton.styleFrom(
                    minimumSize: const Size(0, 44),
                    backgroundColor: Colors.white,
                    side: const BorderSide(color: Color(0xFFBAE6FD)),
                    shape: RoundedRectangleBorder(
                      borderRadius: BorderRadius.circular(12),
                    ),
                  ),
                ),
              ),
            ],
          ),
        ],
      ),
    );
  }

  String _getHeaderTitle(AnnouncementCategory category) {
    switch (category) {
      case AnnouncementCategory.socialMedia:
        return 'PENGUMUMAN & INFO';
      case AnnouncementCategory.promo:
        return 'PROMO & KATALOG';
      case AnnouncementCategory.operational:
        return 'INFO OPERASIONAL';
      case AnnouncementCategory.policy:
        return 'KEBIJAKAN PERUSAHAAN';
      default:
        return 'DETAIL PENGUMUMAN';
    }
  }

  Widget _buildCategoryBadge(AnnouncementCategory category) {
    Color color;
    String label;
    IconData icon;

    switch (category) {
      case AnnouncementCategory.policy:
        color = AppColors.primaryRed;
        label = 'KEBIJAKAN';
        icon = Icons.gavel_rounded;
        break;
      case AnnouncementCategory.birthday:
        color = const Color(0xFFEC4899);
        label = 'ULANG TAHUN';
        icon = Icons.cake_rounded;
        break;
      case AnnouncementCategory.alert:
      case AnnouncementCategory.urgentCall:
        color = AppColors.warning;
        label = 'PENTING';
        icon = Icons.priority_high_rounded;
        break;
      case AnnouncementCategory.socialMedia:
        color = const Color(0xFF8B5CF6);
        label = 'MEDIA SOSIAL';
        icon = Icons.photo_library_rounded;
        break;
      case AnnouncementCategory.promo:
        color = const Color(0xFF059669);
        label = 'PROMOSI';
        icon = Icons.local_offer_rounded;
        break;
      case AnnouncementCategory.operational:
        color = const Color(0xFFD97706);
        label = 'OPERASIONAL';
        icon = Icons.precision_manufacturing_rounded;
        break;
      default:
        color = const Color(0xFF6366F1);
        label = 'INFORMASI';
        icon = Icons.info_rounded;
    }

    return Container(
      padding: const EdgeInsets.symmetric(horizontal: 10, vertical: 6),
      decoration: BoxDecoration(
        color: color.withOpacity(0.1),
        borderRadius: BorderRadius.circular(10),
        border: Border.all(color: color.withOpacity(0.2)),
      ),
      child: Row(
        mainAxisSize: MainAxisSize.min,
        children: [
          Icon(icon, size: 12, color: color),
          const SizedBox(width: 6),
          Text(
            label,
            style: GoogleFonts.plusJakartaSans(
              fontSize: 10,
              fontWeight: FontWeight.w800,
              color: color,
              letterSpacing: 0.5,
            ),
          ),
        ],
      ),
    );
  }
}
