import 'package:equatable/equatable.dart';

enum AnnouncementCategory {
  info,
  alert,
  birthday,
  policy,
  urgentCall,
  event,
  socialMedia,
  promo,
  operational,
}

class Announcement extends Equatable {
  final String id;
  final String title;
  final String content;
  final AnnouncementCategory category;
  final DateTime createdAt;
  final String? author;
  final String mediaType;
  final String? imageUrl;
  final List<String> mediaUrls;
  final String? externalUrl;
  final int likesCount;
  final int commentsCount;
  final bool isLikedByMe;
  final bool trackDownloads;
  final bool trackShares;

  const Announcement({
    required this.id,
    required this.title,
    required this.content,
    required this.category,
    required this.createdAt,
    this.author,
    this.mediaType = 'none',
    this.imageUrl,
    this.mediaUrls = const [],
    this.externalUrl,
    this.likesCount = 0,
    this.commentsCount = 0,
    this.isLikedByMe = false,
    this.trackDownloads = false,
    this.trackShares = false,
  });

  Announcement copyWith({
    String? id,
    String? title,
    String? content,
    AnnouncementCategory? category,
    DateTime? createdAt,
    String? author,
    String? mediaType,
    String? imageUrl,
    List<String>? mediaUrls,
    String? externalUrl,
    int? likesCount,
    int? commentsCount,
    bool? isLikedByMe,
    bool? trackDownloads,
    bool? trackShares,
  }) {
    return Announcement(
      id: id ?? this.id,
      title: title ?? this.title,
      content: content ?? this.content,
      category: category ?? this.category,
      createdAt: createdAt ?? this.createdAt,
      author: author ?? this.author,
      mediaType: mediaType ?? this.mediaType,
      imageUrl: imageUrl ?? this.imageUrl,
      mediaUrls: mediaUrls ?? this.mediaUrls,
      externalUrl: externalUrl ?? this.externalUrl,
      likesCount: likesCount ?? this.likesCount,
      commentsCount: commentsCount ?? this.commentsCount,
      isLikedByMe: isLikedByMe ?? this.isLikedByMe,
      trackDownloads: trackDownloads ?? this.trackDownloads,
      trackShares: trackShares ?? this.trackShares,
    );
  }

  @override
  List<Object?> get props => [
    id,
    title,
    content,
    category,
    createdAt,
    author,
    mediaType,
    imageUrl,
    mediaUrls,
    externalUrl,
    likesCount,
    commentsCount,
    isLikedByMe,
    trackDownloads,
    trackShares,
  ];

  factory Announcement.fromJson(Map<String, dynamic> json) {
    AnnouncementCategory category = AnnouncementCategory.info;
    final catStr = json['category']?.toString().toLowerCase() ?? '';
    switch (catStr) {
      case 'alert':
      case 'urgent_call':
        category = AnnouncementCategory.urgentCall;
        break;
      case 'event':
      case 'meeting':
        category = AnnouncementCategory.event;
        break;
      case 'birthday':
        category = AnnouncementCategory.birthday;
        break;
      case 'policy':
        category = AnnouncementCategory.policy;
        break;
      case 'social_media':
      case 'sosmed':
        category = AnnouncementCategory.socialMedia;
        break;
      case 'promo':
      case 'promosi':
        category = AnnouncementCategory.promo;
        break;
      case 'operational':
      case 'operasional':
        category = AnnouncementCategory.operational;
        break;
    }

    List<String> parsedMediaUrls = [];
    if (json['media_urls'] is List) {
      parsedMediaUrls = (json['media_urls'] as List)
          .map((e) => e.toString())
          .where((e) => e.isNotEmpty)
          .toList();
    }
    String? imgUrl = json['image_url']?.toString();
    if ((imgUrl == null || imgUrl.isEmpty) && parsedMediaUrls.isNotEmpty) {
      imgUrl = parsedMediaUrls.first;
    }
    if (parsedMediaUrls.isEmpty && imgUrl != null && imgUrl.isNotEmpty) {
      parsedMediaUrls = [imgUrl];
    }

    return Announcement(
      id: json['id'] ?? '',
      title: json['title'] ?? '',
      content: json['content'] ?? '',
      category: category,
      createdAt: DateTime.parse(
        json['created_at'] ?? DateTime.now().toIso8601String(),
      ).toLocal(),
      author: json['author'],
      mediaType: json['media_type']?.toString() ?? 'none',
      imageUrl: imgUrl,
      mediaUrls: parsedMediaUrls,
      externalUrl: json['external_url']?.toString(),
      likesCount: json['likes_count'] is int
          ? json['likes_count']
          : int.tryParse(json['likes_count']?.toString() ?? '0') ?? 0,
      commentsCount: json['comments_count'] is int
          ? json['comments_count']
          : int.tryParse(json['comments_count']?.toString() ?? '0') ?? 0,
      isLikedByMe: json['is_liked_by_me'] == true,
      trackDownloads: json['track_downloads'] == true,
      trackShares: json['track_shares'] == true,
    );
  }
}

class AnnouncementComment extends Equatable {
  final String id;
  final String announcementId;
  final String userId;
  final String content;
  final DateTime createdAt;
  final String userName;
  final String userAvatar;
  final String userRole;
  final bool isMine;

  const AnnouncementComment({
    required this.id,
    required this.announcementId,
    required this.userId,
    required this.content,
    required this.createdAt,
    this.userName = 'Karyawan',
    this.userAvatar = '',
    this.userRole = '',
    this.isMine = false,
  });

  factory AnnouncementComment.fromJson(Map<String, dynamic> json) {
    final rawName = json['user_name']?.toString().trim() ?? '';
    String resolvedName = rawName;
    if (resolvedName.isEmpty) {
      resolvedName = 'Karyawan';
    } else if (resolvedName.contains('@')) {
      // If server falls back to email (e.g. user without employee record),
      // Format the username prefix, e.g. "pegawai1" -> "Pegawai1".
      final parts = resolvedName.split('@').first.split('.');
      resolvedName = parts
          .map(
            (p) => p.isNotEmpty ? '${p[0].toUpperCase()}${p.substring(1)}' : '',
          )
          .join(' ')
          .trim();
      if (resolvedName.isEmpty) resolvedName = 'Karyawan';
    }

    return AnnouncementComment(
      id: json['id']?.toString() ?? '',
      announcementId: json['announcement_id']?.toString() ?? '',
      userId: json['user_id']?.toString() ?? '',
      content: json['content']?.toString() ?? '',
      createdAt: DateTime.parse(
        json['created_at'] ?? DateTime.now().toIso8601String(),
      ).toLocal(),
      userName: resolvedName,
      userAvatar: json['user_avatar']?.toString() ?? '',
      userRole: json['user_role']?.toString() ?? '',
      isMine: json['is_mine'] == true,
    );
  }

  @override
  List<Object?> get props => [
    id,
    announcementId,
    userId,
    content,
    createdAt,
    userName,
    userAvatar,
    userRole,
    isMine,
  ];
}
