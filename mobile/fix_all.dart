import 'dart:io';

void main() async {
  final dir = Directory('lib');
  if (!dir.existsSync()) {
    print('lib directory not found');
    return;
  }

  final files = dir.listSync(recursive: true).whereType<File>().where((f) => f.path.endsWith('.dart'));

  for (final file in files) {
    String content = await file.readAsString();
    bool changed = false;

    // 1. Replace Image.network
    if (content.contains('Image.network(')) {
      if (!content.contains('package:cached_network_image/cached_network_image.dart')) {
        // add import at top
        final importIdx = content.indexOf('import');
        if (importIdx != -1) {
          content = "import 'package:cached_network_image/cached_network_image.dart';\n" + content;
        } else {
          content = "import 'package:cached_network_image/cached_network_image.dart';\n" + content;
        }
      }

      // We need to carefully replace Image.network(url, ...) with CachedNetworkImage(imageUrl: url, errorWidget: ...)
      // Let's use a regex to capture the URL argument and the rest of the arguments.
      // This is tricky because dart code can be multi-line.
      final regExp = RegExp(r'Image\.network\s*\(\s*([^,]+?)(,|\))');
      content = content.replaceAllMapped(regExp, (match) {
        final url = match.group(1);
        final terminator = match.group(2);
        if (terminator == ')') {
           return 'CachedNetworkImage(imageUrl: $url, errorWidget: (context, url, error) => const Icon(Icons.error), memCacheWidth: 800)';
        } else {
           return 'CachedNetworkImage(imageUrl: $url, errorWidget: (context, url, error) => const Icon(Icons.error), memCacheWidth: 800,';
        }
      });
      
      // Also we need to replace fit: BoxFit.cover in Image.network. But wait, CachedNetworkImage accepts the same arguments like fit, width, height!
      // So replacing the class name and the first positional arg to named arg `imageUrl:` works perfectly!
      
      changed = true;
    }

    // 2. Replace final dio = Dio(); with GetIt.I<ApiClient>().client
    if (content.contains('Dio()') && !file.path.contains('api_client.dart')) {
       // Make sure GetIt and ApiClient are imported
       if (!content.contains("package:get_it/get_it.dart")) {
         content = "import 'package:get_it/get_it.dart';\n" + content;
       }
       if (!content.contains("package:hris_app/core/network/api_client.dart")) {
         content = "import 'package:hris_app/core/network/api_client.dart';\n" + content;
       }
       content = content.replaceAll('Dio()', 'GetIt.I<ApiClient>().client');
       changed = true;
    }

    if (changed) {
      await file.writeAsString(content);
      print('Updated ${file.path}');
    }
  }
}
