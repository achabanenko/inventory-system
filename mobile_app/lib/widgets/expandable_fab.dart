import 'package:flutter/material.dart';
import 'dart:math' as math;

class ExpandableFab extends StatefulWidget {
  final List<FabAction> actions;
  final Duration animationDuration;
  final double distance;

  const ExpandableFab({
    super.key,
    required this.actions,
    this.animationDuration = const Duration(milliseconds: 250),
    this.distance = 112.0,
  });

  @override
  State<ExpandableFab> createState() => _ExpandableFabState();
}

class _ExpandableFabState extends State<ExpandableFab>
    with SingleTickerProviderStateMixin {
  late AnimationController _controller;
  late Animation<double> _expandAnimation;
  bool _open = false;

  @override
  void initState() {
    super.initState();
    _controller = AnimationController(
      duration: widget.animationDuration,
      vsync: this,
    );
    _expandAnimation = CurvedAnimation(
      parent: _controller,
      curve: Curves.fastOutSlowIn,
      reverseCurve: Curves.easeOutQuad,
    );
  }

  @override
  void dispose() {
    _controller.dispose();
    super.dispose();
  }

  void _toggle() {
    setState(() {
      _open = !_open;
      if (_open) {
        _controller.forward();
      } else {
        _controller.reverse();
      }
    });
  }

  void _close() {
    if (_open) {
      setState(() {
        _open = false;
        _controller.reverse();
      });
    }
  }

  @override
  Widget build(BuildContext context) {
    return SizedBox.expand(
      child: Stack(
        alignment: Alignment.bottomRight,
        clipBehavior: Clip.none,
        children: [
          // Background overlay
          if (_open)
            GestureDetector(
              onTap: _close,
              child: Container(
                color: Colors.black.withOpacity(0.3),
              ),
            ),
          
          // Action buttons
          ..._buildExpandingActionButtons(),
          
          // Main FAB
          _buildTapToCloseFab(),
        ],
      ),
    );
  }

  List<Widget> _buildExpandingActionButtons() {
    final children = <Widget>[];
    final count = widget.actions.length;
    final step = 90.0 / (count - 1);
    
    for (var i = 0, angleInDegrees = 0.0;
         i < count;
         i++, angleInDegrees += step) {
      children.add(
        _ExpandingActionButton(
          directionInDegrees: angleInDegrees,
          maxDistance: widget.distance,
          progress: _expandAnimation,
          action: widget.actions[i],
          onPressed: () {
            _close();
            widget.actions[i].onPressed?.call();
          },
        ),
      );
    }
    return children;
  }

  Widget _buildTapToCloseFab() {
    return AnimatedBuilder(
      animation: _expandAnimation,
      builder: (context, child) {
        return FloatingActionButton(
          onPressed: _toggle,
          child: AnimatedRotation(
            turns: _open ? 0.125 : 0.0, // 45 degrees rotation when open
            duration: widget.animationDuration,
            child: Icon(
              _open ? Icons.close : Icons.add,
              size: 28,
            ),
          ),
        );
      },
    );
  }
}

class _ExpandingActionButton extends StatelessWidget {
  const _ExpandingActionButton({
    required this.directionInDegrees,
    required this.maxDistance,
    required this.progress,
    required this.action,
    required this.onPressed,
  });

  final double directionInDegrees;
  final double maxDistance;
  final Animation<double> progress;
  final FabAction action;
  final VoidCallback onPressed;

  @override
  Widget build(BuildContext context) {
    return AnimatedBuilder(
      animation: progress,
      builder: (context, child) {
        final offset = Offset.fromDirection(
          directionInDegrees * (math.pi / 180.0),
          progress.value * maxDistance,
        );
        
        return Positioned(
          right: 4.0 + offset.dx,
          bottom: 4.0 + offset.dy,
          child: Transform.scale(
            scale: progress.value,
            child: FloatingActionButton.small(
              onPressed: progress.value == 1.0 ? onPressed : null,
              backgroundColor: action.backgroundColor ?? Theme.of(context).colorScheme.secondary,
              foregroundColor: action.foregroundColor ?? Colors.white,
              heroTag: action.heroTag,
              tooltip: action.tooltip,
              child: action.icon,
            ),
          ),
        );
      },
    );
  }
}

class FabAction {
  final Widget icon;
  final String? tooltip;
  final VoidCallback? onPressed;
  final Color? backgroundColor;
  final Color? foregroundColor;
  final String? heroTag;

  const FabAction({
    required this.icon,
    this.tooltip,
    this.onPressed,
    this.backgroundColor,
    this.foregroundColor,
    this.heroTag,
  });
}