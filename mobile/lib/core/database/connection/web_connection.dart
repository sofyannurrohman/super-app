import 'package:drift/drift.dart';
import 'package:drift/web.dart';

QueryExecutor constructDb() {
  return LazyDatabase(() async {
    final storage = await DriftWebStorage.indexedDbIfSupported('db');
    return WebDatabase.withStorage(storage);
  });
}
