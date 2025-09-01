import 'package:flutter/material.dart';

/// A high-performance virtual scrolling list view optimized for large datasets
class VirtualListView<T> extends StatefulWidget {
  final List<T> items;
  final Widget Function(BuildContext context, T item, int index) itemBuilder;
  final double itemHeight;
  final ScrollController? controller;
  final EdgeInsets? padding;
  final Widget? emptyWidget;
  final bool shrinkWrap;
  final ScrollPhysics? physics;
  final Widget? loadingWidget;
  final VoidCallback? onEndReached;
  final double endReachedThreshold;

  const VirtualListView({
    super.key,
    required this.items,
    required this.itemBuilder,
    required this.itemHeight,
    this.controller,
    this.padding,
    this.emptyWidget,
    this.shrinkWrap = false,
    this.physics,
    this.loadingWidget,
    this.onEndReached,
    this.endReachedThreshold = 0.8,
  });

  @override
  State<VirtualListView<T>> createState() => _VirtualListViewState<T>();
}

class _VirtualListViewState<T> extends State<VirtualListView<T>> {
  late ScrollController _scrollController;
  int _firstVisibleIndex = 0;
  int _lastVisibleIndex = 0;
  int _visibleItemCount = 0;

  @override
  void initState() {
    super.initState();
    _scrollController = widget.controller ?? ScrollController();
    _scrollController.addListener(_onScroll);
    
    WidgetsBinding.instance.addPostFrameCallback((_) {
      _calculateVisibleItems();
    });
  }

  @override
  void dispose() {
    if (widget.controller == null) {
      _scrollController.dispose();
    } else {
      _scrollController.removeListener(_onScroll);
    }
    super.dispose();
  }

  void _onScroll() {
    _calculateVisibleItems();
    _checkEndReached();
  }

  void _calculateVisibleItems() {
    if (!_scrollController.hasClients) return;

    final scrollOffset = _scrollController.offset;
    final viewportHeight = _scrollController.position.viewportDimension;
    
    // Calculate visible range with buffer
    const bufferItems = 5; // Render extra items for smooth scrolling
    _firstVisibleIndex = (scrollOffset / widget.itemHeight).floor() - bufferItems;
    _firstVisibleIndex = _firstVisibleIndex.clamp(0, widget.items.length);
    
    _visibleItemCount = (viewportHeight / widget.itemHeight).ceil() + (bufferItems * 2);
    _lastVisibleIndex = (_firstVisibleIndex + _visibleItemCount).clamp(0, widget.items.length);
    
    setState(() {});
  }

  void _checkEndReached() {
    if (widget.onEndReached == null) return;
    
    final maxScrollExtent = _scrollController.position.maxScrollExtent;
    final currentScrollPosition = _scrollController.offset;
    
    if (currentScrollPosition >= maxScrollExtent * widget.endReachedThreshold) {
      widget.onEndReached!();
    }
  }

  @override
  Widget build(BuildContext context) {
    if (widget.items.isEmpty) {
      return widget.emptyWidget ?? const Center(
        child: Text('No items found'),
      );
    }

    return ListView.builder(
      controller: _scrollController,
      padding: widget.padding,
      shrinkWrap: widget.shrinkWrap,
      physics: widget.physics,
      itemCount: widget.items.length,
      itemExtent: widget.itemHeight, // Fixed height for better performance
      itemBuilder: (context, index) {
        // Only build visible items + buffer
        if (index < _firstVisibleIndex || index >= _lastVisibleIndex) {
          return SizedBox(height: widget.itemHeight); // Placeholder
        }
        
        return widget.itemBuilder(context, widget.items[index], index);
      },
    );
  }
}

/// A high-performance grid view for item catalogs
class VirtualGridView<T> extends StatefulWidget {
  final List<T> items;
  final Widget Function(BuildContext context, T item, int index) itemBuilder;
  final int crossAxisCount;
  final double childAspectRatio;
  final double crossAxisSpacing;
  final double mainAxisSpacing;
  final EdgeInsets? padding;
  final ScrollController? controller;
  final Widget? emptyWidget;
  final VoidCallback? onEndReached;
  final double endReachedThreshold;

  const VirtualGridView({
    super.key,
    required this.items,
    required this.itemBuilder,
    this.crossAxisCount = 2,
    this.childAspectRatio = 1.0,
    this.crossAxisSpacing = 8.0,
    this.mainAxisSpacing = 8.0,
    this.padding,
    this.controller,
    this.emptyWidget,
    this.onEndReached,
    this.endReachedThreshold = 0.8,
  });

  @override
  State<VirtualGridView<T>> createState() => _VirtualGridViewState<T>();
}

class _VirtualGridViewState<T> extends State<VirtualGridView<T>> {
  late ScrollController _scrollController;

  @override
  void initState() {
    super.initState();
    _scrollController = widget.controller ?? ScrollController();
    if (widget.onEndReached != null) {
      _scrollController.addListener(_checkEndReached);
    }
  }

  @override
  void dispose() {
    if (widget.controller == null) {
      _scrollController.dispose();
    } else if (widget.onEndReached != null) {
      _scrollController.removeListener(_checkEndReached);
    }
    super.dispose();
  }

  void _checkEndReached() {
    if (!_scrollController.hasClients) return;
    
    final maxScrollExtent = _scrollController.position.maxScrollExtent;
    final currentScrollPosition = _scrollController.offset;
    
    if (currentScrollPosition >= maxScrollExtent * widget.endReachedThreshold) {
      widget.onEndReached!();
    }
  }

  @override
  Widget build(BuildContext context) {
    if (widget.items.isEmpty) {
      return widget.emptyWidget ?? const Center(
        child: Text('No items found'),
      );
    }

    return GridView.builder(
      controller: _scrollController,
      padding: widget.padding,
      gridDelegate: SliverGridDelegateWithFixedCrossAxisCount(
        crossAxisCount: widget.crossAxisCount,
        childAspectRatio: widget.childAspectRatio,
        crossAxisSpacing: widget.crossAxisSpacing,
        mainAxisSpacing: widget.mainAxisSpacing,
      ),
      itemCount: widget.items.length,
      itemBuilder: (context, index) {
        return widget.itemBuilder(context, widget.items[index], index);
      },
    );
  }
}

/// Performance-optimized item tile for catalogs
class OptimizedItemTile extends StatelessWidget {
  final String title;
  final String subtitle;
  final String? imageUrl;
  final Widget? leading;
  final VoidCallback? onTap;
  final Widget? trailing;
  final bool dense;

  const OptimizedItemTile({
    super.key,
    required this.title,
    required this.subtitle,
    this.imageUrl,
    this.leading,
    this.onTap,
    this.trailing,
    this.dense = false,
  });

  @override
  Widget build(BuildContext context) {
    return Material(
      color: Colors.transparent,
      child: InkWell(
        onTap: onTap,
        child: Container(
          padding: EdgeInsets.all(dense ? 8.0 : 12.0),
          child: Row(
            children: [
              // Leading widget (optimized image loading)
              if (leading != null)
                leading!
              else if (imageUrl != null)
                ClipRRect(
                  borderRadius: BorderRadius.circular(8),
                  child: Image.network(
                    imageUrl!,
                    width: dense ? 32 : 48,
                    height: dense ? 32 : 48,
                    fit: BoxFit.cover,
                    loadingBuilder: (context, child, loadingProgress) {
                      if (loadingProgress == null) return child;
                      return Container(
                        width: dense ? 32 : 48,
                        height: dense ? 32 : 48,
                        color: Colors.grey[200],
                        child: const Center(
                          child: SizedBox(
                            width: 16,
                            height: 16,
                            child: CircularProgressIndicator(strokeWidth: 2),
                          ),
                        ),
                      );
                    },
                    errorBuilder: (context, error, stackTrace) {
                      return Container(
                        width: dense ? 32 : 48,
                        height: dense ? 32 : 48,
                        decoration: BoxDecoration(
                          color: Colors.grey[200],
                          borderRadius: BorderRadius.circular(8),
                        ),
                        child: Icon(
                          Icons.image_not_supported,
                          color: Colors.grey[400],
                          size: dense ? 16 : 24,
                        ),
                      );
                    },
                  ),
                ),
              
              if (leading != null || imageUrl != null)
                SizedBox(width: dense ? 8 : 12),
              
              // Title and subtitle
              Expanded(
                child: Column(
                  crossAxisAlignment: CrossAxisAlignment.start,
                  mainAxisSize: MainAxisSize.min,
                  children: [
                    Text(
                      title,
                      style: TextStyle(
                        fontWeight: FontWeight.w600,
                        fontSize: dense ? 14 : 16,
                      ),
                      maxLines: 1,
                      overflow: TextOverflow.ellipsis,
                    ),
                    if (subtitle.isNotEmpty) ...[
                      SizedBox(height: dense ? 2 : 4),
                      Text(
                        subtitle,
                        style: TextStyle(
                          color: Colors.grey[600],
                          fontSize: dense ? 12 : 14,
                        ),
                        maxLines: dense ? 1 : 2,
                        overflow: TextOverflow.ellipsis,
                      ),
                    ],
                  ],
                ),
              ),
              
              // Trailing widget
              if (trailing != null) ...[
                SizedBox(width: dense ? 8 : 12),
                trailing!,
              ],
            ],
          ),
        ),
      ),
    );
  }
}

/// Pagination controller for large datasets
class PaginationController {
  final int pageSize;
  final Future<List<T>> Function<T>(int page, int limit) fetchItems;
  
  final ValueNotifier<List<dynamic>> _items = ValueNotifier([]);
  final ValueNotifier<bool> _isLoading = ValueNotifier(false);
  final ValueNotifier<bool> _hasMoreData = ValueNotifier(true);
  
  int _currentPage = 1;
  
  PaginationController({
    required this.pageSize,
    required this.fetchItems,
  });

  ValueListenable<List<dynamic>> get items => _items;
  ValueListenable<bool> get isLoading => _isLoading;
  ValueListenable<bool> get hasMoreData => _hasMoreData;

  Future<void> loadInitialData<T>() async {
    if (_isLoading.value) return;
    
    _isLoading.value = true;
    _currentPage = 1;
    
    try {
      final newItems = await fetchItems<T>(_currentPage, pageSize);
      _items.value = newItems;
      _hasMoreData.value = newItems.length == pageSize;
    } catch (e) {
      _items.value = [];
      _hasMoreData.value = false;
    } finally {
      _isLoading.value = false;
    }
  }

  Future<void> loadMoreData<T>() async {
    if (_isLoading.value || !_hasMoreData.value) return;
    
    _isLoading.value = true;
    _currentPage++;
    
    try {
      final newItems = await fetchItems<T>(_currentPage, pageSize);
      _items.value = [..._items.value, ...newItems];
      _hasMoreData.value = newItems.length == pageSize;
    } catch (e) {
      _currentPage--; // Revert page increment on error
      _hasMoreData.value = false;
    } finally {
      _isLoading.value = false;
    }
  }

  void reset() {
    _items.value = [];
    _isLoading.value = false;
    _hasMoreData.value = true;
    _currentPage = 1;
  }

  void dispose() {
    _items.dispose();
    _isLoading.dispose();
    _hasMoreData.dispose();
  }
}