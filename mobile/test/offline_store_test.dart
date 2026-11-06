import 'package:flutter_test/flutter_test.dart';
import 'package:hris_app/features/sales/data/models/store_model.dart';
import 'package:hris_app/core/database/database.dart';
import 'package:drift/native.dart';
import 'package:drift/drift.dart' as drift;

void main() {
  group('Offline Store Unit Tests', () {
    test('StoreModel serialization with offline properties', () {
      final store = StoreModel(
        id: 'test-uuid-1234',
        name: 'Toko Sumber Barokah',
        ownerName: 'Pak Subur',
        phoneNumber: '081234567890',
        address: 'Jl. Merdeka No. 45, Medan',
        latitude: 3.595196,
        longitude: 98.672226,
        isNew: true,
        province: 'Sumatera Utara',
        regency: 'Kota Medan',
        district: 'Medan Kota',
        village: 'Pasar Baru',
        isOffline: true,
        syncStatus: 'pending',
        localPhotoPath: '/data/user/0/com.example.karsahrtrial/offline_stores/test-uuid-1234.jpg',
        accuracy: 12.5,
      );

      final json = store.toJson();
      expect(json['id'], 'test-uuid-1234');
      expect(json['name'], 'Toko Sumber Barokah');
      expect(json['is_offline'], true);
      expect(json['sync_status'], 'pending');
      expect(json['accuracy'], 12.5);
      expect(json['local_photo_path'], '/data/user/0/com.example.karsahrtrial/offline_stores/test-uuid-1234.jpg');

      final deserialized = StoreModel.fromJson(json);
      expect(deserialized.id, 'test-uuid-1234');
      expect(deserialized.name, 'Toko Sumber Barokah');
      expect(deserialized.isOffline, true);
      expect(deserialized.syncStatus, 'pending');
      expect(deserialized.accuracy, 12.5);
      expect(deserialized.localPhotoPath, '/data/user/0/com.example.karsahrtrial/offline_stores/test-uuid-1234.jpg');
    });

    test('Drift in-memory database saves and reads offline store', () async {
      final db = AppDatabase.forTesting(NativeDatabase.memory());

      final companion = StoresCompanion(
        id: const drift.Value('offline-store-uuid-001'),
        name: const drift.Value('Toko Maju Jaya'),
        address: const drift.Value('Jl. Gatot Subroto No. 10'),
        latitude: const drift.Value(3.5892),
        longitude: const drift.Value(98.6654),
        ownerName: const drift.Value('Ibu Siti'),
        phoneNumber: const drift.Value('081399887766'),
        localPhotoPath: const drift.Value('/path/to/photo.jpg'),
        isOffline: const drift.Value(true),
        syncStatus: const drift.Value('pending'),
        accuracy: const drift.Value(8.2),
        createdAt: drift.Value(DateTime.now()),
      );

      await db.into(db.stores).insertOnConflictUpdate(companion);

      final storesInDb = await (db.select(db.stores)
            ..where((t) => t.isOffline.equals(true)))
          .get();

      expect(storesInDb.length, 1);
      final saved = storesInDb.first;
      expect(saved.id, 'offline-store-uuid-001');
      expect(saved.name, 'Toko Maju Jaya');
      expect(saved.isOffline, true);
      expect(saved.syncStatus, 'pending');
      expect(saved.accuracy, 8.2);

      await db.close();
    });
  });
}
