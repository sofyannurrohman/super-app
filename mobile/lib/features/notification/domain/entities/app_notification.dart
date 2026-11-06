import 'package:equatable/equatable.dart';

enum AppNotificationType { announcement, leave, overtime, reimbursement }

class AppNotification extends Equatable {
  final String id;
  final String title;
  final String content;
  final DateTime timestamp;
  final AppNotificationType type;
  final String? status; // PENDING, APPROVED, REJECTED
  final dynamic originalData;

  const AppNotification({
    required this.id,
    required this.title,
    required this.content,
    required this.timestamp,
    required this.type,
    this.status,
    this.originalData,
  });

  @override
  List<Object?> get props => [id, title, content, timestamp, type, status];
}
