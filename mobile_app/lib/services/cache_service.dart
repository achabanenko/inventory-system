import 'dart:convert';
import 'package:shared_preferences/shared_preferences.dart';
import '../models/item.dart';

/// Three-tier caching service for optimal performance:
/// 1. Memory cache (100 items) - immediate access
/// 2. Local storage (1,000 items) - offline functionality  
/// 3. Server-side - complete catalogs
class CacheService {
  static CacheService? _instance;
  static CacheService get instance => _instance ??= CacheService._();
  
  CacheService._();

  // Memory cache - fastest access
  final Map<String, Item> _memoryCache = {};
  final Map<String, DateTime> _memoryCacheTimestamps = {};
  final Map<String, List<Item>> _searchCache = {};
  
  // Configuration
  static const int maxMemoryCacheSize = 100;
  static const int maxLocalCacheSize = 1000;
  static const Duration cacheExpiration = Duration(minutes: 30);
  static const Duration searchCacheExpiration = Duration(minutes: 5);
  
  // Keys for SharedPreferences
  static const String _itemCacheKey = 'cached_items';
  static const String _cacheTimestampKey = 'cache_timestamps';
  static const String _lastSyncKey = 'last_sync_timestamp';

  /// Get item from cache (checks memory first, then local storage)
  Future<Item?> getItem(String itemId) async {
    // Check memory cache first
    if (_memoryCache.containsKey(itemId)) {
      final timestamp = _memoryCacheTimestamps[itemId];
      if (timestamp != null && _isValidCache(timestamp)) {
        return _memoryCache[itemId];
      } else {
        // Remove expired item from memory
        _memoryCache.remove(itemId);
        _memoryCacheTimestamps.remove(itemId);
      }
    }

    // Check local storage
    final item = await _getItemFromLocalStorage(itemId);
    if (item != null) {
      // Add to memory cache for faster next access
      _addToMemoryCache(itemId, item);
      return item;
    }

    return null;
  }

  /// Cache item in both memory and local storage
  Future<void> cacheItem(Item item) async {
    // Add to memory cache
    _addToMemoryCache(item.id, item);
    
    // Add to local storage
    await _addToLocalStorage(item);
  }

  /// Cache multiple items efficiently
  Future<void> cacheItems(List<Item> items) async {
    for (final item in items) {
      _addToMemoryCache(item.id, item);
    }
    await _addItemsToLocalStorage(items);
  }

  /// Get cached search results
  List<Item>? getCachedSearchResults(String query) {
    final normalizedQuery = query.toLowerCase().trim();
    if (_searchCache.containsKey(normalizedQuery)) {
      return _searchCache[normalizedQuery];
    }
    return null;
  }

  /// Cache search results
  void cacheSearchResults(String query, List<Item> results) {
    final normalizedQuery = query.toLowerCase().trim();
    _searchCache[normalizedQuery] = results;
    
    // Clean up old search cache entries
    if (_searchCache.length > 50) {
      final oldestKey = _searchCache.keys.first;
      _searchCache.remove(oldestKey);
    }
  }

  /// Get frequently accessed items from memory cache
  List<Item> getFrequentItems({int limit = 20}) {
    return _memoryCache.values.take(limit).toList();
  }

  /// Get cached items with optional search filtering
  Future<List<Item>> getCachedItems({String? search}) async {
    final List<Item> items = [];
    
    // Get all items from memory cache first
    items.addAll(_memoryCache.values);
    
    // Get items from local storage
    try {
      final prefs = await SharedPreferences.getInstance();
      final cachedItems = prefs.getString(_itemCacheKey);
      final timestamps = prefs.getString(_cacheTimestampKey);
      
      if (cachedItems != null && timestamps != null) {
        final itemsMap = Map<String, dynamic>.from(jsonDecode(cachedItems));
        final timestampsMap = Map<String, int>.from(jsonDecode(timestamps));
        
        for (final entry in itemsMap.entries) {
          final timestamp = timestampsMap[entry.key];
          if (timestamp != null && _isValidCache(DateTime.fromMillisecondsSinceEpoch(timestamp))) {
            // Only add if not already in memory cache
            if (!_memoryCache.containsKey(entry.key)) {
              try {
                final item = Item.fromJson(entry.value);
                items.add(item);
              } catch (e) {
                print('Error parsing cached item ${entry.key}: $e');
              }
            }
          }
        }
      }
    } catch (e) {
      print('Error reading cached items: $e');
    }
    
    // Apply search filter if provided
    if (search != null && search.isNotEmpty) {
      final normalizedSearch = search.toLowerCase();
      return items.where((item) => 
        item.name.toLowerCase().contains(normalizedSearch) ||
        item.sku.toLowerCase().contains(normalizedSearch) ||
        (item.barcode?.toLowerCase().contains(normalizedSearch) ?? false)
      ).toList();
    }
    
    return items;
  }

  /// Preload items for better performance
  Future<void> preloadItems(List<String> itemIds) async {
    final uncachedIds = <String>[];
    
    for (final id in itemIds) {
      if (!_memoryCache.containsKey(id)) {
        uncachedIds.add(id);
      }
    }
    
    if (uncachedIds.isNotEmpty) {
      // This would typically fetch from API - placeholder for now
      // final items = await apiService.getItemsByIds(uncachedIds);
      // await cacheItems(items);
    }
  }

  /// Clear expired cache entries
  Future<void> clearExpiredCache() async {
    // Clear expired memory cache
    final expiredKeys = <String>[];
    _memoryCacheTimestamps.forEach((key, timestamp) {
      if (!_isValidCache(timestamp)) {
        expiredKeys.add(key);
      }
    });
    
    for (final key in expiredKeys) {
      _memoryCache.remove(key);
      _memoryCacheTimestamps.remove(key);
    }

    // Clear expired local storage cache
    await _clearExpiredLocalStorage();
  }

  /// Get cache statistics
  CacheStats getCacheStats() {
    return CacheStats(
      memoryCacheSize: _memoryCache.length,
      searchCacheSize: _searchCache.length,
      maxMemorySize: maxMemoryCacheSize,
      maxLocalSize: maxLocalCacheSize,
    );
  }

  /// Clear all cache
  Future<void> clearAllCache() async {
    _memoryCache.clear();
    _memoryCacheTimestamps.clear();
    _searchCache.clear();
    
    final prefs = await SharedPreferences.getInstance();
    await prefs.remove(_itemCacheKey);
    await prefs.remove(_cacheTimestampKey);
  }

  /// Get last sync timestamp
  Future<DateTime?> getLastSyncTimestamp() async {
    final prefs = await SharedPreferences.getInstance();
    final timestamp = prefs.getInt(_lastSyncKey);
    return timestamp != null ? DateTime.fromMillisecondsSinceEpoch(timestamp) : null;
  }

  /// Update last sync timestamp
  Future<void> updateLastSyncTimestamp() async {
    final prefs = await SharedPreferences.getInstance();
    await prefs.setInt(_lastSyncKey, DateTime.now().millisecondsSinceEpoch);
  }

  // Private methods

  void _addToMemoryCache(String itemId, Item item) {
    // Remove oldest item if cache is full
    if (_memoryCache.length >= maxMemoryCacheSize) {
      final oldestKey = _memoryCacheTimestamps.entries
          .reduce((a, b) => a.value.isBefore(b.value) ? a : b)
          .key;
      _memoryCache.remove(oldestKey);
      _memoryCacheTimestamps.remove(oldestKey);
    }
    
    _memoryCache[itemId] = item;
    _memoryCacheTimestamps[itemId] = DateTime.now();
  }

  Future<Item?> _getItemFromLocalStorage(String itemId) async {
    try {
      final prefs = await SharedPreferences.getInstance();
      final cachedItems = prefs.getString(_itemCacheKey);
      final timestamps = prefs.getString(_cacheTimestampKey);
      
      if (cachedItems == null || timestamps == null) return null;
      
      final itemsMap = Map<String, dynamic>.from(jsonDecode(cachedItems));
      final timestampsMap = Map<String, int>.from(jsonDecode(timestamps));
      
      if (!itemsMap.containsKey(itemId)) return null;
      
      final timestamp = timestampsMap[itemId];
      if (timestamp == null || !_isValidCache(DateTime.fromMillisecondsSinceEpoch(timestamp))) {
        return null;
      }
      
      return Item.fromJson(itemsMap[itemId]);
    } catch (e) {
      print('Error reading from local storage: $e');
      return null;
    }
  }

  Future<void> _addToLocalStorage(Item item) async {
    await _addItemsToLocalStorage([item]);
  }

  Future<void> _addItemsToLocalStorage(List<Item> items) async {
    try {
      final prefs = await SharedPreferences.getInstance();
      
      // Load existing cache
      Map<String, dynamic> cachedItems = {};
      Map<String, int> timestamps = {};
      
      final existingItems = prefs.getString(_itemCacheKey);
      final existingTimestamps = prefs.getString(_cacheTimestampKey);
      
      if (existingItems != null && existingTimestamps != null) {
        cachedItems = Map<String, dynamic>.from(jsonDecode(existingItems));
        timestamps = Map<String, int>.from(jsonDecode(existingTimestamps));
      }
      
      // Add new items
      final now = DateTime.now().millisecondsSinceEpoch;
      for (final item in items) {
        cachedItems[item.id] = item.toJson();
        timestamps[item.id] = now;
      }
      
      // Remove excess items if over limit
      if (cachedItems.length > maxLocalCacheSize) {
        final sortedEntries = timestamps.entries.toList()
          ..sort((a, b) => a.value.compareTo(b.value));
        
        final itemsToRemove = cachedItems.length - maxLocalCacheSize;
        for (int i = 0; i < itemsToRemove; i++) {
          final keyToRemove = sortedEntries[i].key;
          cachedItems.remove(keyToRemove);
          timestamps.remove(keyToRemove);
        }
      }
      
      // Save updated cache
      await prefs.setString(_itemCacheKey, jsonEncode(cachedItems));
      await prefs.setString(_cacheTimestampKey, jsonEncode(timestamps));
    } catch (e) {
      print('Error saving to local storage: $e');
    }
  }

  Future<void> _clearExpiredLocalStorage() async {
    try {
      final prefs = await SharedPreferences.getInstance();
      final cachedItems = prefs.getString(_itemCacheKey);
      final timestampsData = prefs.getString(_cacheTimestampKey);
      
      if (cachedItems == null || timestampsData == null) return;
      
      final itemsMap = Map<String, dynamic>.from(jsonDecode(cachedItems));
      final timestamps = Map<String, int>.from(jsonDecode(timestampsData));
      
      final validItems = <String, dynamic>{};
      final validTimestamps = <String, int>{};
      
      timestamps.forEach((key, timestamp) {
        if (_isValidCache(DateTime.fromMillisecondsSinceEpoch(timestamp))) {
          validItems[key] = itemsMap[key];
          validTimestamps[key] = timestamp;
        }
      });
      
      await prefs.setString(_itemCacheKey, jsonEncode(validItems));
      await prefs.setString(_cacheTimestampKey, jsonEncode(validTimestamps));
    } catch (e) {
      print('Error cleaning local storage: $e');
    }
  }

  bool _isValidCache(DateTime timestamp) {
    return DateTime.now().difference(timestamp) < cacheExpiration;
  }
}

/// Cache statistics for monitoring performance
class CacheStats {
  final int memoryCacheSize;
  final int searchCacheSize;
  final int maxMemorySize;
  final int maxLocalSize;

  const CacheStats({
    required this.memoryCacheSize,
    required this.searchCacheSize,
    required this.maxMemorySize,
    required this.maxLocalSize,
  });

  double get memoryCacheUsage => memoryCacheSize / maxMemorySize;
  
  Map<String, dynamic> toJson() {
    return {
      'memoryCacheSize': memoryCacheSize,
      'searchCacheSize': searchCacheSize,
      'maxMemorySize': maxMemorySize,
      'maxLocalSize': maxLocalSize,
      'memoryCacheUsage': memoryCacheUsage,
    };
  }
}

/// Cache-aware API service wrapper
mixin CacheAwareMixin {
  
  /// Get items with caching
  Future<List<Item>> getCachedItems({
    String? search,
    int page = 1,
    int limit = 20,
    bool useCache = true,
  }) async {
    if (!useCache) {
      return _fetchItemsFromAPI(search: search, page: page, limit: limit);
    }

    // Check cache for search results
    if (search != null && search.isNotEmpty) {
      final cachedResults = CacheService.instance.getCachedSearchResults(search);
      if (cachedResults != null) {
        return cachedResults;
      }
    }

    // Fetch from API and cache results
    final items = await _fetchItemsFromAPI(search: search, page: page, limit: limit);
    
    // Cache the results
    await CacheService.instance.cacheItems(items);
    
    if (search != null && search.isNotEmpty) {
      CacheService.instance.cacheSearchResults(search, items);
    }
    
    return items;
  }

  /// Get single item with caching
  Future<Item?> getCachedItem(String itemId, {bool useCache = true}) async {
    if (!useCache) {
      return _fetchItemFromAPI(itemId);
    }

    // Check cache first
    final cachedItem = await CacheService.instance.getItem(itemId);
    if (cachedItem != null) {
      return cachedItem;
    }

    // Fetch from API and cache
    final item = await _fetchItemFromAPI(itemId);
    if (item != null) {
      await CacheService.instance.cacheItem(item);
    }
    
    return item;
  }

  // Abstract methods to be implemented by API service
  Future<List<Item>> _fetchItemsFromAPI({String? search, int page = 1, int limit = 20});
  Future<Item?> _fetchItemFromAPI(String itemId);
}