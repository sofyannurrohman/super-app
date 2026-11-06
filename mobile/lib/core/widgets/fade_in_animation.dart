import 'package:flutter/material.dart';
import 'package:flutter_staggered_animations/flutter_staggered_animations.dart';

class FadeInAnimationWrapper extends StatelessWidget {
  final Widget child;
  final int index;

  const FadeInAnimationWrapper({
    super.key,
    required this.child,
    this.index = 0,
  });

  @override
  Widget build(BuildContext context) {
    return AnimationConfiguration.staggeredList(
      position: index,
      duration: const Duration(milliseconds: 600),
      child: FadeInAnimation(
        curve: Curves.easeOutCubic,
        child: SlideAnimation(
          verticalOffset: 30.0,
          curve: Curves.easeOutCubic,
          child: child,
        ),
      ),
    );
  }
}
