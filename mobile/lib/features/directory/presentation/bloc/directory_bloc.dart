import 'package:equatable/equatable.dart';
import 'package:flutter_bloc/flutter_bloc.dart';
import 'package:hris_app/features/directory/data/models/directory_model.dart';
import 'package:hris_app/features/directory/data/repositories/directory_repository.dart';

// Events
abstract class DirectoryEvent extends Equatable {
  const DirectoryEvent();
  @override
  List<Object?> get props => [];
}

class FetchDirectoryRequested extends DirectoryEvent {
  final String? query;
  const FetchDirectoryRequested({this.query});
  @override
  List<Object?> get props => [query];
}

// States
abstract class DirectoryState extends Equatable {
  const DirectoryState();
  @override
  List<Object?> get props => [];
}

class DirectoryInitial extends DirectoryState {}

class DirectoryLoading extends DirectoryState {}

class DirectoryLoaded extends DirectoryState {
  final List<EmployeeDirectory> directory;
  const DirectoryLoaded(this.directory);
  @override
  List<Object?> get props => [directory];
}

class DirectoryFailure extends DirectoryState {
  final String message;
  const DirectoryFailure(this.message);
  @override
  List<Object?> get props => [message];
}

// BLoC
class DirectoryBloc extends Bloc<DirectoryEvent, DirectoryState> {
  final DirectoryRepository repository;

  DirectoryBloc({required this.repository}) : super(DirectoryInitial()) {
    on<FetchDirectoryRequested>((event, emit) async {
      emit(DirectoryLoading());
      final result = await repository.getDirectory(query: event.query);
      result.fold(
        (failure) => emit(DirectoryFailure(failure.message)),
        (lists) => emit(DirectoryLoaded(lists)),
      );
    });
  }
}
