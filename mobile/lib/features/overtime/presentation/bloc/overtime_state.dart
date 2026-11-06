import 'package:equatable/equatable.dart';
import 'package:hris_app/features/overtime/domain/entities/overtime.dart';

abstract class OvertimeState extends Equatable {
  const OvertimeState();

  @override
  List<Object?> get props => [];
}

class OvertimeInitial extends OvertimeState {}

class OvertimeLoading extends OvertimeState {}

class OvertimeSuccess extends OvertimeState {
  final String message;

  const OvertimeSuccess(this.message);

  @override
  List<Object?> get props => [message];
}

class OvertimeFailure extends OvertimeState {
  final String errorMessage;

  const OvertimeFailure(this.errorMessage);

  @override
  List<Object?> get props => [errorMessage];
}

class MyOvertimesLoaded extends OvertimeState {
  final List<Overtime> overtimes;

  const MyOvertimesLoaded(this.overtimes);

  @override
  List<Object?> get props => [overtimes];
}
