import 'package:flutter/material.dart';
import 'package:flutter/services.dart';
import 'dart:async';
import '../models/item.dart';
import '../services/api_service.dart';
import '../widgets/barcode_scanner_overlay.dart';

enum SearchMode { text, barcode, description }

class UniversalItemSearch extends StatefulWidget {
  final Function(Item) onItemSelected;
  final String? initialQuery;
  final List<SearchMode> availableModes;
  final String? sourceLocation; // For filtering by location
  final bool showStockInfo;
  final String hintText;
  
  const UniversalItemSearch({
    super.key,
    required this.onItemSelected,
    this.initialQuery,
    this.availableModes = const [
      SearchMode.text,
      SearchMode.barcode,
      SearchMode.description
    ],
    this.sourceLocation,
    this.showStockInfo = false,
    this.hintText = 'Search by code, barcode, or description',
  });

  @override
  State<UniversalItemSearch> createState() => _UniversalItemSearchState();
}

class _UniversalItemSearchState extends State<UniversalItemSearch> {
  final TextEditingController _searchController = TextEditingController();
  final FocusNode _searchFocus = FocusNode();
  final ApiService _apiService = ApiService();
  
  SearchMode _currentMode = SearchMode.text;
  List<Item> _searchResults = [];
  List<String> _recentSearches = [];
  bool _isLoading = false;
  Timer? _debounceTimer;
  
  @override
  void initState() {
    super.initState();
    _searchController.text = widget.initialQuery ?? '';
    _loadRecentSearches();
    
    // Start search if initial query provided
    if (widget.initialQuery != null && widget.initialQuery!.isNotEmpty) {
      _performSearch(widget.initialQuery!);
    }
  }
  
  @override
  void dispose() {
    _searchController.dispose();
    _searchFocus.dispose();
    _debounceTimer?.cancel();
    super.dispose();
  }
  
  Future<void> _loadRecentSearches() async {
    // Load from SharedPreferences
    // For now, using mock data
    setState(() {
      _recentSearches = ['SKU001', 'Widget A', '1234567890'];
    });
  }
  
  void _onSearchChanged(String query) {
    // Cancel previous timer
    _debounceTimer?.cancel();
    
    if (query.isEmpty) {
      setState(() {
        _searchResults = [];
      });
      return;
    }
    
    // Debounce search by 300ms
    _debounceTimer = Timer(const Duration(milliseconds: 300), () {
      _performSearch(query);
    });
  }
  
  Future<void> _performSearch(String query) async {
    if (query.length < 2 && _currentMode != SearchMode.barcode) {
      return; // Require at least 2 characters for text/description search
    }
    
    setState(() {
      _isLoading = true;
    });
    
    try {
      List<Item> results = [];
      
      switch (_currentMode) {
        case SearchMode.text:
          // Search by item code (exact match)
          results = await _apiService.searchItems(query, 'code');
          break;
        case SearchMode.barcode:
          // Search by barcode (exact match)
          results = await _apiService.searchItems(query, 'barcode');
          break;
        case SearchMode.description:
          // Search by description (substring/fuzzy match)
          results = await _apiService.searchItems(query, 'description');
          break;
      }
      
      // Filter by source location if provided
      if (widget.sourceLocation != null && widget.showStockInfo) {
        // Filter items based on stock at source location
        // This would require additional API call to check stock levels
        // For now, showing all results with stock indicators
      }
      
      setState(() {
        _searchResults = results;
        _isLoading = false;
      });
      
      // Add to recent searches
      if (!_recentSearches.contains(query)) {
        _recentSearches.insert(0, query);
        if (_recentSearches.length > 10) {
          _recentSearches.removeLast();
        }
        // Save to SharedPreferences
      }
    } catch (e) {
      setState(() {
        _isLoading = false;
      });
      if (mounted) {
        ScaffoldMessenger.of(context).showSnackBar(
          SnackBar(content: Text('Search failed: ${e.toString()}')),
        );
      }
    }
  }
  
  void _openBarcodeScanner() async {
    final result = await showModalBottomSheet<String>(
      context: context,
      isScrollControlled: true,
      backgroundColor: Colors.transparent,
      builder: (context) => BarcodeScannerOverlay(
        onBarcodeScanned: (barcode) {
          Navigator.pop(context, barcode);
        },
        onCancel: () {
          Navigator.pop(context);
        },
      ),
    );
    
    if (result != null) {
      _searchController.text = result;
      _currentMode = SearchMode.barcode;
      _performSearch(result);
      
      // Provide haptic feedback
      HapticFeedback.mediumImpact();
    }
  }
  
  Widget _buildSearchBar() {
    return Container(
      padding: const EdgeInsets.all(16),
      decoration: BoxDecoration(
        color: Theme.of(context).colorScheme.surface,
        boxShadow: [
          BoxShadow(
            color: Colors.black.withOpacity(0.05),
            blurRadius: 10,
            offset: const Offset(0, 2),
          ),
        ],
      ),
      child: Column(
        children: [
          // Search input with mode toggle
          Row(
            children: [
              // Mode selector
              if (widget.availableModes.length > 1)
                Container(
                  decoration: BoxDecoration(
                    color: Theme.of(context).colorScheme.primaryContainer,
                    borderRadius: BorderRadius.circular(8),
                  ),
                  child: DropdownButton<SearchMode>(
                    value: _currentMode,
                    underline: const SizedBox(),
                    isDense: true,
                    padding: const EdgeInsets.symmetric(horizontal: 12, vertical: 8),
                    borderRadius: BorderRadius.circular(8),
                    items: widget.availableModes.map((mode) {
                      return DropdownMenuItem(
                        value: mode,
                        child: Text(
                          mode == SearchMode.text ? 'Code' :
                          mode == SearchMode.barcode ? 'Barcode' : 'Description',
                          style: const TextStyle(fontSize: 14),
                        ),
                      );
                    }).toList(),
                    onChanged: (mode) {
                      setState(() {
                        _currentMode = mode!;
                      });
                      if (_searchController.text.isNotEmpty) {
                        _performSearch(_searchController.text);
                      }
                    },
                  ),
                ),
              const SizedBox(width: 8),
              
              // Search input field
              Expanded(
                child: TextField(
                  controller: _searchController,
                  focusNode: _searchFocus,
                  decoration: InputDecoration(
                    hintText: widget.hintText,
                    prefixIcon: const Icon(Icons.search),
                    suffixIcon: Row(
                      mainAxisSize: MainAxisSize.min,
                      children: [
                        if (_searchController.text.isNotEmpty)
                          IconButton(
                            icon: const Icon(Icons.clear),
                            onPressed: () {
                              _searchController.clear();
                              setState(() {
                                _searchResults = [];
                              });
                            },
                          ),
                        IconButton(
                          icon: const Icon(Icons.qr_code_scanner),
                          onPressed: _openBarcodeScanner,
                          color: Theme.of(context).colorScheme.primary,
                        ),
                      ],
                    ),
                    border: OutlineInputBorder(
                      borderRadius: BorderRadius.circular(12),
                    ),
                    filled: true,
                    fillColor: Theme.of(context).colorScheme.surface,
                  ),
                  onChanged: _onSearchChanged,
                  textInputAction: TextInputAction.search,
                  onSubmitted: _performSearch,
                ),
              ),
            ],
          ),
          
          // Recent searches (shown when search is empty)
          if (_searchController.text.isEmpty && _recentSearches.isNotEmpty)
            Container(
              margin: const EdgeInsets.only(top: 8),
              height: 32,
              child: ListView.builder(
                scrollDirection: Axis.horizontal,
                itemCount: _recentSearches.length,
                itemBuilder: (context, index) {
                  return Container(
                    margin: const EdgeInsets.only(right: 8),
                    child: ActionChip(
                      label: Text(_recentSearches[index]),
                      onPressed: () {
                        _searchController.text = _recentSearches[index];
                        _performSearch(_recentSearches[index]);
                      },
                      backgroundColor: Theme.of(context).colorScheme.secondaryContainer,
                    ),
                  );
                },
              ),
            ),
        ],
      ),
    );
  }
  
  Widget _buildSearchResults() {
    if (_isLoading) {
      return const Center(
        child: Padding(
          padding: EdgeInsets.all(32),
          child: CircularProgressIndicator(),
        ),
      );
    }
    
    if (_searchController.text.isNotEmpty && _searchResults.isEmpty) {
      return Center(
        child: Padding(
          padding: const EdgeInsets.all(32),
          child: Column(
            mainAxisSize: MainAxisSize.min,
            children: [
              Icon(
                Icons.search_off,
                size: 64,
                color: Theme.of(context).colorScheme.onSurfaceVariant,
              ),
              const SizedBox(height: 16),
              Text(
                'No items found',
                style: Theme.of(context).textTheme.titleMedium,
              ),
              const SizedBox(height: 8),
              Text(
                'Try a different search term or scan a barcode',
                style: Theme.of(context).textTheme.bodyMedium?.copyWith(
                  color: Theme.of(context).colorScheme.onSurfaceVariant,
                ),
                textAlign: TextAlign.center,
              ),
            ],
          ),
        ),
      );
    }
    
    return ListView.builder(
      padding: const EdgeInsets.only(bottom: 80),
      itemCount: _searchResults.length,
      itemBuilder: (context, index) {
        final item = _searchResults[index];
        return _buildItemCard(item);
      },
    );
  }
  
  Widget _buildItemCard(Item item) {
    return Card(
      margin: const EdgeInsets.symmetric(horizontal: 16, vertical: 8),
      child: InkWell(
        onTap: () {
          HapticFeedback.lightImpact();
          widget.onItemSelected(item);
        },
        borderRadius: BorderRadius.circular(12),
        child: Padding(
          padding: const EdgeInsets.all(16),
          child: Column(
            crossAxisAlignment: CrossAxisAlignment.start,
            children: [
              // Item header
              Row(
                children: [
                  Expanded(
                    child: Column(
                      crossAxisAlignment: CrossAxisAlignment.start,
                      children: [
                        Text(
                          item.sku,
                          style: Theme.of(context).textTheme.titleMedium?.copyWith(
                            fontWeight: FontWeight.bold,
                          ),
                        ),
                        if (item.barcode != null && item.barcode!.isNotEmpty)
                          Text(
                            'Barcode: ${item.barcode}',
                            style: Theme.of(context).textTheme.bodySmall?.copyWith(
                              color: Theme.of(context).colorScheme.onSurfaceVariant,
                            ),
                          ),
                      ],
                    ),
                  ),
                  if (widget.showStockInfo && widget.sourceLocation != null)
                    Container(
                      padding: const EdgeInsets.symmetric(horizontal: 12, vertical: 4),
                      decoration: BoxDecoration(
                        color: Theme.of(context).colorScheme.primaryContainer,
                        borderRadius: BorderRadius.circular(16),
                      ),
                      child: Text(
                        'In Stock', // This would show actual stock level
                        style: Theme.of(context).textTheme.bodySmall,
                      ),
                    ),
                ],
              ),
              const SizedBox(height: 8),
              
              // Item description
              Text(
                item.name,
                style: Theme.of(context).textTheme.bodyLarge,
                maxLines: 2,
                overflow: TextOverflow.ellipsis,
              ),
              
              // Price info
              Padding(
                padding: const EdgeInsets.only(top: 8),
                child: Row(
                  children: [
                    Text(
                      '\$${item.price.toStringAsFixed(2)}',
                      style: Theme.of(context).textTheme.titleMedium?.copyWith(
                        color: Theme.of(context).colorScheme.primary,
                        fontWeight: FontWeight.bold,
                      ),
                    ),
                    const SizedBox(width: 16),
                    if (item.uom != null)
                      Text(
                        'per ${item.uom}',
                        style: Theme.of(context).textTheme.bodySmall?.copyWith(
                          color: Theme.of(context).colorScheme.onSurfaceVariant,
                        ),
                      ),
                    ],
                  ),
                ),
            ],
          ),
        ),
      ),
    );
  }
  
  @override
  Widget build(BuildContext context) {
    return Column(
      children: [
        _buildSearchBar(),
        Expanded(
          child: _buildSearchResults(),
        ),
      ],
    );
  }
}