import 'package:cached_network_image/cached_network_image.dart';
import 'package:flutter/foundation.dart';
import 'package:flutter/material.dart';
import 'package:hris_app/core/utils/constants.dart';

/// Universal network image widget that works smoothly on both Web and Mobile devices.
/// - Automatically resolves image URLs with CORS proxy on Web via AppConstants.buildImageUrl
/// - Provides memory caching on Android/iOS
/// - Handles error states and empty URLs gracefully with clean fallbacks
class AppNetworkImage extends StatelessWidget {
  final String imageUrl;
  final double? width;
  final double? height;
  final BoxFit fit;
  final Widget? placeholder;
  final Widget? errorWidget;
  final BorderRadius? borderRadius;
  final int? memCacheWidth;
  final int? memCacheHeight;

  const AppNetworkImage({
    super.key,
    required this.imageUrl,
    this.width,
    this.height,
    this.fit = BoxFit.cover,
    this.placeholder,
    this.errorWidget,
    this.borderRadius,
    this.memCacheWidth,
    this.memCacheHeight,
  });

  @override
  Widget build(BuildContext context) {
    final resolvedUrl = AppConstants.buildImageUrl(imageUrl);

    if (resolvedUrl.isEmpty) {
      return _buildFallback(context);
    }

    Widget imageContent;

    if (kIsWeb) {
      // On Flutter Web, use standard Image.network with error handling
      imageContent = Image.network(
        resolvedUrl,
        width: width,
        height: height,
        fit: fit,
        loadingBuilder: (context, child, loadingProgress) {
          if (loadingProgress == null) return child;
          return placeholder ?? _defaultPlaceholder();
        },
        errorBuilder: (context, error, stackTrace) {
          return errorWidget ?? _defaultErrorWidget();
        },
      );
    } else {
      // On Android / iOS / Desktop, use CachedNetworkImage for optimal disk & memory caching
      imageContent = CachedNetworkImage(
        imageUrl: resolvedUrl,
        width: width,
        height: height,
        fit: fit,
        memCacheWidth: memCacheWidth ?? (width != null ? (width! * 2).toInt() : 600),
        memCacheHeight: memCacheHeight ?? (height != null ? (height! * 2).toInt() : null),
        placeholder: (context, url) => placeholder ?? _defaultPlaceholder(),
        errorWidget: (context, url, error) => errorWidget ?? _defaultErrorWidget(),
      );
    }

    if (borderRadius != null) {
      return ClipRRect(
        borderRadius: borderRadius!,
        child: imageContent,
      );
    }

    return imageContent;
  }

  Widget _buildFallback(BuildContext context) {
    final fb = errorWidget ?? _defaultErrorWidget();
    if (borderRadius != null) {
      return ClipRRect(borderRadius: borderRadius!, child: fb);
    }
    return fb;
  }

  Widget _defaultPlaceholder() {
    return Container(
      width: width,
      height: height,
      color: Colors.grey[200],
      child: const Center(
        child: SizedBox(
          width: 20,
          height: 20,
          child: CircularProgressIndicator(
            strokeWidth: 2,
            color: Color(0xFF047857),
          ),
        ),
      ),
    );
  }

  Widget _defaultErrorWidget() {
    return Container(
      width: width,
      height: height,
      color: Colors.grey[200],
      child: Icon(
        Icons.store_mall_directory_rounded,
        size: width != null ? (width! * 0.5).clamp(16.0, 48.0) : 28,
        color: Colors.grey[400],
      ),
    );
  }
}
