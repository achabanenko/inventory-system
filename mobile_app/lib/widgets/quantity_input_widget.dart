import 'package:flutter/material.dart';
import 'package:flutter/services.dart';

class QuantityInputWidget extends StatefulWidget {
  final double initialQuantity;
  final double? maxQuantity;
  final double? minQuantity;
  final Function(double) onQuantityChanged;
  final String? label;
  final String? unit;
  final bool showQuickButtons;
  final List<double> quickButtonValues;
  final int decimalPlaces;
  final bool showSteppers;
  final double stepAmount;
  final Color? primaryColor;
  final bool isLarge;
  final String? helperText;
  final String? errorText;
  
  const QuantityInputWidget({
    super.key,
    required this.initialQuantity,
    required this.onQuantityChanged,
    this.maxQuantity,
    this.minQuantity = 0,
    this.label,
    this.unit,
    this.showQuickButtons = true,
    this.quickButtonValues = const [1, 5, 10, 25, 50, 100],
    this.decimalPlaces = 0,
    this.showSteppers = true,
    this.stepAmount = 1,
    this.primaryColor,
    this.isLarge = false,
    this.helperText,
    this.errorText,
  });

  @override
  State<QuantityInputWidget> createState() => _QuantityInputWidgetState();
}

class _QuantityInputWidgetState extends State<QuantityInputWidget> {
  late TextEditingController _controller;
  late double _currentQuantity;
  bool _hasError = false;
  
  @override
  void initState() {
    super.initState();
    _currentQuantity = widget.initialQuantity;
    _controller = TextEditingController(
      text: _formatQuantity(_currentQuantity),
    );
  }
  
  @override
  void dispose() {
    _controller.dispose();
    super.dispose();
  }
  
  String _formatQuantity(double quantity) {
    if (widget.decimalPlaces == 0) {
      return quantity.toInt().toString();
    }
    return quantity.toStringAsFixed(widget.decimalPlaces);
  }
  
  void _updateQuantity(double newQuantity) {
    // Validate quantity
    if (widget.minQuantity != null && newQuantity < widget.minQuantity!) {
      newQuantity = widget.minQuantity!;
    }
    if (widget.maxQuantity != null && newQuantity > widget.maxQuantity!) {
      newQuantity = widget.maxQuantity!;
    }
    
    setState(() {
      _currentQuantity = newQuantity;
      _controller.text = _formatQuantity(newQuantity);
      _hasError = false;
    });
    
    // Provide haptic feedback
    HapticFeedback.lightImpact();
    
    // Notify parent
    widget.onQuantityChanged(newQuantity);
  }
  
  void _incrementQuantity() {
    _updateQuantity(_currentQuantity + widget.stepAmount);
  }
  
  void _decrementQuantity() {
    _updateQuantity(_currentQuantity - widget.stepAmount);
  }
  
  void _onTextChanged(String value) {
    if (value.isEmpty) {
      setState(() {
        _hasError = true;
      });
      return;
    }
    
    final double? newQuantity = double.tryParse(value);
    if (newQuantity == null) {
      setState(() {
        _hasError = true;
      });
      return;
    }
    
    // Check bounds
    bool error = false;
    if (widget.minQuantity != null && newQuantity < widget.minQuantity!) {
      error = true;
    }
    if (widget.maxQuantity != null && newQuantity > widget.maxQuantity!) {
      error = true;
    }
    
    setState(() {
      _currentQuantity = newQuantity;
      _hasError = error;
    });
    
    if (!error) {
      widget.onQuantityChanged(newQuantity);
    }
  }
  
  Widget _buildStepperButton({
    required IconData icon,
    required VoidCallback onPressed,
    required bool isEnabled,
  }) {
    final color = widget.primaryColor ?? Theme.of(context).colorScheme.primary;
    final size = widget.isLarge ? 56.0 : 44.0;
    final iconSize = widget.isLarge ? 28.0 : 24.0;
    
    return Material(
      color: isEnabled ? color : color.withOpacity(0.3),
      borderRadius: BorderRadius.circular(8),
      child: InkWell(
        onTap: isEnabled ? onPressed : null,
        borderRadius: BorderRadius.circular(8),
        child: Container(
          width: size,
          height: size,
          alignment: Alignment.center,
          child: Icon(
            icon,
            size: iconSize,
            color: Colors.white,
          ),
        ),
      ),
    );
  }
  
  Widget _buildQuantityInput() {
    final inputHeight = widget.isLarge ? 60.0 : 44.0;
    final fontSize = widget.isLarge ? 24.0 : 18.0;
    
    return Container(
      height: inputHeight,
      constraints: BoxConstraints(
        minWidth: widget.isLarge ? 120 : 80,
        maxWidth: widget.isLarge ? 200 : 150,
      ),
      child: TextField(
        controller: _controller,
        keyboardType: TextInputType.numberWithOptions(
          decimal: widget.decimalPlaces > 0,
        ),
        textAlign: TextAlign.center,
        style: TextStyle(
          fontSize: fontSize,
          fontWeight: FontWeight.bold,
          color: _hasError ? Colors.red : null,
        ),
        decoration: InputDecoration(
          contentPadding: const EdgeInsets.symmetric(horizontal: 12),
          border: OutlineInputBorder(
            borderRadius: BorderRadius.circular(8),
            borderSide: BorderSide(
              color: _hasError ? Colors.red : Colors.grey,
            ),
          ),
          enabledBorder: OutlineInputBorder(
            borderRadius: BorderRadius.circular(8),
            borderSide: BorderSide(
              color: _hasError ? Colors.red : Colors.grey.shade300,
            ),
          ),
          focusedBorder: OutlineInputBorder(
            borderRadius: BorderRadius.circular(8),
            borderSide: BorderSide(
              color: _hasError 
                ? Colors.red 
                : (widget.primaryColor ?? Theme.of(context).colorScheme.primary),
              width: 2,
            ),
          ),
          errorBorder: OutlineInputBorder(
            borderRadius: BorderRadius.circular(8),
            borderSide: const BorderSide(
              color: Colors.red,
              width: 2,
            ),
          ),
          suffixText: widget.unit,
          suffixStyle: TextStyle(
            fontSize: fontSize * 0.75,
            color: Theme.of(context).colorScheme.onSurfaceVariant,
          ),
        ),
        onChanged: _onTextChanged,
        inputFormatters: [
          FilteringTextInputFormatter.allow(
            widget.decimalPlaces > 0 
              ? RegExp(r'^\d*\.?\d*$')
              : RegExp(r'^\d*$'),
          ),
        ],
      ),
    );
  }
  
  Widget _buildQuickButtons() {
    return Wrap(
      spacing: 8,
      runSpacing: 8,
      children: widget.quickButtonValues.map((value) {
        final isOverMax = widget.maxQuantity != null && 
                          (_currentQuantity + value) > widget.maxQuantity!;
        
        return Material(
          color: isOverMax 
            ? Colors.grey.shade300 
            : Theme.of(context).colorScheme.secondaryContainer,
          borderRadius: BorderRadius.circular(20),
          child: InkWell(
            onTap: isOverMax ? null : () {
              _updateQuantity(_currentQuantity + value);
            },
            borderRadius: BorderRadius.circular(20),
            child: Container(
              padding: const EdgeInsets.symmetric(horizontal: 16, vertical: 8),
              child: Text(
                '+${_formatQuantity(value)}',
                style: TextStyle(
                  fontWeight: FontWeight.bold,
                  color: isOverMax 
                    ? Colors.grey 
                    : Theme.of(context).colorScheme.onSecondaryContainer,
                ),
              ),
            ),
          ),
        );
      }).toList(),
    );
  }
  
  @override
  Widget build(BuildContext context) {
    final canDecrement = widget.minQuantity == null || 
                        _currentQuantity > widget.minQuantity!;
    final canIncrement = widget.maxQuantity == null || 
                        _currentQuantity < widget.maxQuantity!;
    
    return Column(
      crossAxisAlignment: CrossAxisAlignment.start,
      mainAxisSize: MainAxisSize.min,
      children: [
        // Label
        if (widget.label != null)
          Padding(
            padding: const EdgeInsets.only(bottom: 8),
            child: Text(
              widget.label!,
              style: Theme.of(context).textTheme.titleMedium?.copyWith(
                fontWeight: FontWeight.bold,
              ),
            ),
          ),
        
        // Main input row with steppers
        Row(
          mainAxisAlignment: MainAxisAlignment.center,
          children: [
            if (widget.showSteppers)
              _buildStepperButton(
                icon: Icons.remove,
                onPressed: _decrementQuantity,
                isEnabled: canDecrement,
              ),
            if (widget.showSteppers)
              const SizedBox(width: 12),
            
            _buildQuantityInput(),
            
            if (widget.showSteppers)
              const SizedBox(width: 12),
            if (widget.showSteppers)
              _buildStepperButton(
                icon: Icons.add,
                onPressed: _incrementQuantity,
                isEnabled: canIncrement,
              ),
          ],
        ),
        
        // Helper text or error text
        if (widget.helperText != null || widget.errorText != null || _hasError)
          Padding(
            padding: const EdgeInsets.only(top: 8),
            child: Text(
              widget.errorText ?? 
              (_hasError ? 'Invalid quantity' : widget.helperText ?? ''),
              style: Theme.of(context).textTheme.bodySmall?.copyWith(
                color: (widget.errorText != null || _hasError) 
                  ? Colors.red 
                  : Theme.of(context).colorScheme.onSurfaceVariant,
              ),
            ),
          ),
        
        // Quick buttons
        if (widget.showQuickButtons)
          Padding(
            padding: const EdgeInsets.only(top: 16),
            child: Column(
              crossAxisAlignment: CrossAxisAlignment.start,
              children: [
                Text(
                  'Quick add:',
                  style: Theme.of(context).textTheme.bodySmall?.copyWith(
                    color: Theme.of(context).colorScheme.onSurfaceVariant,
                  ),
                ),
                const SizedBox(height: 8),
                _buildQuickButtons(),
              ],
            ),
          ),
        
        // Max quantity indicator
        if (widget.maxQuantity != null)
          Padding(
            padding: const EdgeInsets.only(top: 8),
            child: LinearProgressIndicator(
              value: _currentQuantity / widget.maxQuantity!,
              backgroundColor: Colors.grey.shade300,
              valueColor: AlwaysStoppedAnimation<Color>(
                _currentQuantity > widget.maxQuantity! 
                  ? Colors.red 
                  : (widget.primaryColor ?? Theme.of(context).colorScheme.primary),
              ),
            ),
          ),
      ],
    );
  }
}