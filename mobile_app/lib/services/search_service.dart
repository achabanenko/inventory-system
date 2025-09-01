import 'dart:async';
import 'package:flutter/material.dart';
import '../models/item.dart';

class SearchService {
  static const int _minSearchLength = 1; // Start suggestions after 1-2 characters as recommended

  /// Performs fuzzy search on items using multiple algorithms
  static List<Item> searchItems(List<Item> items, String query) {
    if (query.length < _minSearchLength || items.isEmpty) {
      return [];
    }

    final normalizedQuery = query.toLowerCase().trim();
    final searchResults = <SearchResult>[];

    for (final item in items) {
      final score = _calculateRelevanceScore(item, normalizedQuery);
      if (score > 0) {
        searchResults.add(SearchResult(item: item, score: score));
      }
    }

    // Sort by relevance score (highest first)
    searchResults.sort((a, b) => b.score.compareTo(a.score));
    
    return searchResults.map((result) => result.item).toList();
  }

  /// Calculate relevance score for an item based on the search query
  static double _calculateRelevanceScore(Item item, String query) {
    double score = 0.0;
    
    final itemName = item.name.toLowerCase();
    final itemSku = item.sku.toLowerCase();
    final itemBarcode = item.barcode?.toLowerCase() ?? '';
    final itemDescription = item.description?.toLowerCase() ?? '';

    // Exact matches get highest priority
    if (itemName == query || itemSku == query || itemBarcode == query) {
      return 100.0;
    }

    // Prefix matches get high priority
    if (itemName.startsWith(query)) {
      score += 90.0;
    }
    if (itemSku.startsWith(query)) {
      score += 85.0;
    }
    if (itemBarcode.startsWith(query)) {
      score += 80.0;
    }

    // Contains matches get medium priority
    if (itemName.contains(query)) {
      score += 70.0;
    }
    if (itemSku.contains(query)) {
      score += 65.0;
    }
    if (itemBarcode.contains(query)) {
      score += 60.0;
    }
    if (itemDescription.contains(query)) {
      score += 50.0;
    }

    // Fuzzy matching using Levenshtein distance
    final nameDistance = _calculateLevenshteinDistance(itemName, query);
    final skuDistance = _calculateLevenshteinDistance(itemSku, query);
    
    // Convert distance to score (lower distance = higher score)
    final maxLength = [itemName.length, itemSku.length, query.length].reduce((a, b) => a > b ? a : b);
    
    if (nameDistance <= maxLength * 0.3) { // Allow up to 30% character differences
      score += 40.0 * (1.0 - (nameDistance / maxLength));
    }
    
    if (skuDistance <= maxLength * 0.3) {
      score += 35.0 * (1.0 - (skuDistance / maxLength));
    }

    // N-gram matching for partial word matches
    final nameNGramScore = _calculateNGramScore(itemName, query);
    final skuNGramScore = _calculateNGramScore(itemSku, query);
    
    score += nameNGramScore * 30.0;
    score += skuNGramScore * 25.0;

    // Word boundary matching - individual words in name
    final nameWords = itemName.split(' ');
    for (final word in nameWords) {
      if (word.startsWith(query)) {
        score += 20.0;
      } else if (word.contains(query)) {
        score += 10.0;
      }
    }

    return score;
  }

  /// Calculate Levenshtein distance between two strings
  static int _calculateLevenshteinDistance(String a, String b) {
    if (a.isEmpty) return b.length;
    if (b.isEmpty) return a.length;

    final matrix = List.generate(
      a.length + 1,
      (i) => List.filled(b.length + 1, 0),
    );

    // Initialize first row and column
    for (int i = 0; i <= a.length; i++) {
      matrix[i][0] = i;
    }
    for (int j = 0; j <= b.length; j++) {
      matrix[0][j] = j;
    }

    // Fill the matrix
    for (int i = 1; i <= a.length; i++) {
      for (int j = 1; j <= b.length; j++) {
        final cost = a[i - 1] == b[j - 1] ? 0 : 1;
        matrix[i][j] = [
          matrix[i - 1][j] + 1,      // deletion
          matrix[i][j - 1] + 1,      // insertion
          matrix[i - 1][j - 1] + cost, // substitution
        ].reduce((a, b) => a < b ? a : b);
      }
    }

    return matrix[a.length][b.length];
  }

  /// Calculate N-gram similarity score
  static double _calculateNGramScore(String text, String query) {
    const n = 2; // Use bigrams
    
    final textNGrams = _generateNGrams(text, n);
    final queryNGrams = _generateNGrams(query, n);
    
    if (textNGrams.isEmpty || queryNGrams.isEmpty) {
      return 0.0;
    }

    final intersection = textNGrams.toSet().intersection(queryNGrams.toSet());
    final union = textNGrams.toSet().union(queryNGrams.toSet());
    
    return union.isEmpty ? 0.0 : intersection.length / union.length;
  }

  /// Generate N-grams from a string
  static List<String> _generateNGrams(String text, int n) {
    if (text.length < n) {
      return [text];
    }

    final nGrams = <String>[];
    for (int i = 0; i <= text.length - n; i++) {
      nGrams.add(text.substring(i, i + n));
    }
    return nGrams;
  }

  /// Filter items by category or other criteria
  static List<Item> filterItems(List<Item> items, {
    String? category,
    double? minPrice,
    double? maxPrice,
    bool? inStock,
  }) {
    return items.where((item) {
      if (category != null && item.category != category) {
        return false;
      }
      
      if (minPrice != null && item.price < minPrice) {
        return false;
      }
      
      if (maxPrice != null && item.price > maxPrice) {
        return false;
      }
      
      if (inStock != null && inStock) {
        // Assuming items have a stock field - adapt based on your model
        // This is a placeholder condition
      }
      
      return true;
    }).toList();
  }

  /// Get search suggestions based on partial input
  static List<String> getSearchSuggestions(List<Item> items, String query) {
    if (query.isEmpty) return [];

    final suggestions = <String>{};
    final normalizedQuery = query.toLowerCase();

    for (final item in items) {
      final name = item.name.toLowerCase();
      final sku = item.sku.toLowerCase();
      
      // Add name suggestions
      if (name.startsWith(normalizedQuery)) {
        suggestions.add(item.name);
      }
      
      // Add SKU suggestions
      if (sku.startsWith(normalizedQuery)) {
        suggestions.add(item.sku);
      }
      
      // Add word-based suggestions
      final words = name.split(' ');
      for (final word in words) {
        if (word.startsWith(normalizedQuery) && word.length > normalizedQuery.length) {
          suggestions.add(word);
        }
      }
    }

    return suggestions.take(10).toList(); // Limit to 10 suggestions
  }
}

class SearchResult {
  final Item item;
  final double score;

  SearchResult({required this.item, required this.score});
}

/// Search filters for advanced filtering
class SearchFilters {
  final String? category;
  final double? minPrice;
  final double? maxPrice;
  final bool? inStock;
  final String? supplier;

  const SearchFilters({
    this.category,
    this.minPrice,
    this.maxPrice,
    this.inStock,
    this.supplier,
  });

  SearchFilters copyWith({
    String? category,
    double? minPrice,
    double? maxPrice,
    bool? inStock,
    String? supplier,
  }) {
    return SearchFilters(
      category: category ?? this.category,
      minPrice: minPrice ?? this.minPrice,
      maxPrice: maxPrice ?? this.maxPrice,
      inStock: inStock ?? this.inStock,
      supplier: supplier ?? this.supplier,
    );
  }
}

/// Debounce utility for search input
class Debouncer {
  final Duration delay;
  Timer? _timer;

  Debouncer({required this.delay});

  void call(VoidCallback callback) {
    _timer?.cancel();
    _timer = Timer(delay, callback);
  }

  void cancel() {
    _timer?.cancel();
  }
}

