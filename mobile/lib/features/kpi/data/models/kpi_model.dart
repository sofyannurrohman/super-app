import 'package:equatable/equatable.dart';

enum KPIType { sales, internal }

enum KPIStatus { draft, approved }

class KPIItem extends Equatable {
  final String label;
  final double score;
  final String? weight;

  const KPIItem({required this.label, required this.score, this.weight});

  @override
  List<Object?> get props => [label, score, weight];

  factory KPIItem.fromJson(Map<String, dynamic> json) {
    return KPIItem(
      label: json['label'] ?? '',
      score: (json['score'] as num?)?.toDouble() ?? 0.0,
      weight: json['weight'],
    );
  }
}

class KPIModel extends Equatable {
  final KPIType type;
  final String period;
  final double achievementPercentage; // For Sales
  final double finalScore; // Overall Score
  final KPIStatus status;
  final String? approvedBy;
  final String? approvedAt;
  final List<KPIItem> items; // For Detailed breakdown

  const KPIModel({
    required this.type,
    required this.period,
    this.achievementPercentage = 0.0,
    this.finalScore = 0.0,
    required this.status,
    this.approvedBy,
    this.approvedAt,
    this.items = const [],
  });

  @override
  List<Object?> get props => [
        type,
        period,
        achievementPercentage,
        finalScore,
        status,
        approvedBy,
        approvedAt,
        items
      ];

  factory KPIModel.fromJson(Map<String, dynamic> json) {
    return KPIModel(
      type: json['type'] == 'SALES' ? KPIType.sales : KPIType.internal,
      period: json['period'] ?? '',
      achievementPercentage:
          (json['achievement_percentage'] as num?)?.toDouble() ?? 0.0,
      finalScore: (json['final_score'] as num?)?.toDouble() ?? 0.0,
      status:
          json['status'] == 'APPROVED' ? KPIStatus.approved : KPIStatus.draft,
      approvedBy: json['approved_by'],
      approvedAt: json['approved_at'],
      items:
          (json['items'] as List?)?.map((i) => KPIItem.fromJson(i)).toList() ??
              [],
    );
  }
}
