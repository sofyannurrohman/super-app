import 'package:equatable/equatable.dart';
import 'package:hris_app/features/epkwt/domain/entities/epkwt_request.dart';

abstract class EpkwtState extends Equatable {
  const EpkwtState();

  @override
  List<Object?> get props => [];
}

class EpkwtInitial extends EpkwtState {}

class EpkwtLoading extends EpkwtState {}

class EpkwtListLoaded extends EpkwtState {
  final List<EpkwtRequest> epkwts;

  const EpkwtListLoaded(this.epkwts);

  @override
  List<Object?> get props => [epkwts];
}

class EpkwtSuccess extends EpkwtState {
  final String message;

  const EpkwtSuccess(this.message);

  @override
  List<Object?> get props => [message];
}

class EpkwtFailure extends EpkwtState {
  final String message;

  const EpkwtFailure(this.message);

  @override
  List<Object?> get props => [message];
}
