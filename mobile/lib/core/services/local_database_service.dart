import 'dart:io' show Platform;
import 'package:flutter/foundation.dart' show kIsWeb;
import 'package:sqflite/sqflite.dart';
import 'package:path/path.dart';
import 'package:sqflite_common_ffi/sqflite_ffi.dart';

class LocalDatabaseService {
  static Database? _database;

  Future<Database> get database async {
    if (kIsWeb) {
      throw UnsupportedError('sqflite is not supported on web');
    }
    if (_database != null) return _database!;

    // Initialize FFI for desktop and tests
    if (Platform.isLinux || Platform.isWindows || Platform.isMacOS) {
      sqfliteFfiInit();
      databaseFactory = databaseFactoryFfi;
    }

    _database = await _initDatabase();
    return _database!;
  }

  Future<Database> _initDatabase() async {
    String path = join(await getDatabasesPath(), 'hris_offline.db');
    return await openDatabase(
      path,
      version: 1,
      onCreate: _onCreate,
    );
  }

  Future _onCreate(Database db, int version) async {
    await db.execute('''
      CREATE TABLE attendance_queue (
        id INTEGER PRIMARY KEY AUTOINCREMENT,
        type TEXT,
        photo_path TEXT,
        latitude REAL,
        longitude REAL,
        timestamp TEXT,
        face_embedding TEXT,
        is_synced INTEGER DEFAULT 0
      )
    ''');
  }

  Future<int> insertAttendance(Map<String, dynamic> data) async {
    if (kIsWeb) return 0;
    final db = await database;
    return await db.insert('attendance_queue', data);
  }

  Future<List<Map<String, dynamic>>> getUnsyncedAttendanceBars() async {
    if (kIsWeb) return [];
    final db = await database;
    return await db.query('attendance_queue', where: 'is_synced = 0');
  }

  Future<int> markAsSynced(int id) async {
    if (kIsWeb) return 0;
    final db = await database;
    return await db.update(
      'attendance_queue',
      {'is_synced': 1},
      where: 'id = ?',
      whereArgs: [id],
    );
  }

  Future<int> deleteSynced(int id) async {
    if (kIsWeb) return 0;
    final db = await database;
    return await db.delete(
      'attendance_queue',
      where: 'id = ?',
      whereArgs: [id],
    );
  }

  // New methods for build recovery
  Future<int> getQueueCount() async {
    if (kIsWeb) return 0;
    final db = await database;
    final result = await db.rawQuery(
        'SELECT COUNT(*) as count FROM attendance_queue WHERE is_synced = 0');
    return Sqflite.firstIntValue(result) ?? 0;
  }

  Future<List<Map<String, dynamic>>> getAttendanceQueue() async {
    if (kIsWeb) return [];
    final db = await database;
    return await db.query('attendance_queue', where: 'is_synced = 0');
  }

  Future<int> deleteAttendance(int id) async {
    if (kIsWeb) return 0;
    final db = await database;
    return await db.delete(
      'attendance_queue',
      where: 'id = ?',
      whereArgs: [id],
    );
  }
}
