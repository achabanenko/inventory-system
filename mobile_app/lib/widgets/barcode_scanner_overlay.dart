import 'package:flutter/material.dart';
import 'package:flutter/services.dart';
import 'package:mobile_scanner/mobile_scanner.dart';
import 'package:audioplayers/audioplayers.dart';

enum ScannerMode {
  single,     // Scan one item and close
  continuous, // Keep scanning multiple items
  batch,      // Scan multiple items with count
  showDetail, // Scan and show item detail screen
}

class BarcodeScannerOverlay extends StatefulWidget {
  final Function(String) onBarcodeScanned;
  final VoidCallback onCancel;
  final ScannerMode mode;
  final bool showManualEntry;
  final bool playSound;
  final String? headerText;
  final Widget? contextWidget;
  final String? documentType; // For showDetail mode
  final Map<String, dynamic>? contextData; // For showDetail mode

  const BarcodeScannerOverlay({
    super.key,
    required this.onBarcodeScanned,
    required this.onCancel,
    this.mode = ScannerMode.single,
    this.showManualEntry = true,
    this.playSound = true,
    this.headerText,
    this.contextWidget,
    this.documentType,
    this.contextData,
  });

  @override
  State<BarcodeScannerOverlay> createState() => _BarcodeScannerOverlayState();
}

class _BarcodeScannerOverlayState extends State<BarcodeScannerOverlay>
    with TickerProviderStateMixin {
  MobileScannerController? controller;
  bool hasScanned = false;
  late AnimationController _animationController;
  late Animation<double> _animation;
  final AudioPlayer _audioPlayer = AudioPlayer();
  final List<String> _scannedBarcodes = [];
  final TextEditingController _manualEntryController = TextEditingController();
  bool _showManualEntry = false;
  bool _torchEnabled = false;

  @override
  void initState() {
    super.initState();
    controller = MobileScannerController(
      formats: [
        BarcodeFormat.ean13,    // Standard UPC/EAN barcodes (most common)
        BarcodeFormat.ean8,     // Short EAN barcodes
        BarcodeFormat.upcA,     // UPC-A barcodes
        BarcodeFormat.upcE,     // UPC-E barcodes (compact)
        BarcodeFormat.code128,  // Code 128 (common for logistics)
        BarcodeFormat.code39,   // Code 39 (less common but still used)
      ],
      facing: CameraFacing.back,
      torchEnabled: false,
    );

    _animationController = AnimationController(
      duration: const Duration(milliseconds: 2000),
      vsync: this,
    );
    _animation = Tween<double>(begin: 0.0, end: 1.0).animate(
      CurvedAnimation(parent: _animationController, curve: Curves.easeInOut),
    );
    _animationController.repeat(reverse: true);
  }

  @override
  void dispose() {
    controller?.dispose();
    _animationController.dispose();
    super.dispose();
  }

  void _onDetect(BarcodeCapture capture) {
    if (widget.mode == ScannerMode.single && hasScanned) return;
    
    final List<Barcode> barcodes = capture.barcodes;
    if (barcodes.isNotEmpty) {
      final Barcode barcode = barcodes.first;
      final String? code = barcode.rawValue;
      
      if (code != null && code.isNotEmpty && _isValidProductBarcode(code, barcode.format)) {
        // Check if already scanned in batch/continuous mode
        if (widget.mode != ScannerMode.single && _scannedBarcodes.contains(code)) {
          _showDuplicateWarning(code);
          return;
        }
        
        setState(() {
          if (widget.mode == ScannerMode.single) {
            hasScanned = true;
          } else {
            _scannedBarcodes.add(code);
          }
        });
        
        // Provide feedback
        _provideFeedback();
        
        widget.onBarcodeScanned(code);
        
        // Note: Do NOT auto-close for single mode - let the parent handle navigation
        // The parent will navigate to ScannedItemScreen and we don't want to interfere
      }
    }
  }
  
  void _provideFeedback() async {
    // Haptic feedback
    HapticFeedback.mediumImpact();
    
    // Audio feedback
    if (widget.playSound) {
      try {
        await _audioPlayer.play(AssetSource('sounds/beep.mp3'));
      } catch (e) {
        // Fallback to system sound if custom sound fails
        SystemSound.play(SystemSoundType.click);
      }
    }
    
    // Visual feedback
    _showScanSuccess();
  }
  
  void _showScanSuccess() {
    ScaffoldMessenger.of(context).showSnackBar(
      SnackBar(
        content: Row(
          children: [
            const Icon(Icons.check_circle, color: Colors.white),
            const SizedBox(width: 8),
            Text(widget.mode == ScannerMode.single 
              ? 'Barcode scanned successfully!' 
              : 'Item added (${_scannedBarcodes.length} total)'),
          ],
        ),
        backgroundColor: Colors.green,
        duration: const Duration(seconds: 1),
        behavior: SnackBarBehavior.floating,
        margin: const EdgeInsets.all(16),
        shape: RoundedRectangleBorder(
          borderRadius: BorderRadius.circular(8),
        ),
      ),
    );
  }
  
  void _showDuplicateWarning(String code) {
    ScaffoldMessenger.of(context).showSnackBar(
      SnackBar(
        content: Row(
          children: [
            const Icon(Icons.warning, color: Colors.white),
            const SizedBox(width: 8),
            Expanded(child: Text('Item $code already scanned')),
          ],
        ),
        backgroundColor: Colors.orange,
        duration: const Duration(seconds: 2),
        behavior: SnackBarBehavior.floating,
        margin: const EdgeInsets.all(16),
        shape: RoundedRectangleBorder(
          borderRadius: BorderRadius.circular(8),
        ),
      ),
    );
    
    // Vibrate for warning
    HapticFeedback.heavyImpact();
  }

  /// Validate that the scanned code is a proper product barcode
  bool _isValidProductBarcode(String code, BarcodeFormat format) {
    // Remove any whitespace
    code = code.trim();
    
    // Reject obviously invalid codes
    if (code.isEmpty || code.length < 4) {
      return false;
    }
    
    // Check basic length requirements for different formats
    switch (format) {
      case BarcodeFormat.ean13:
        return code.length == 13 && _isNumeric(code);
      case BarcodeFormat.ean8:
        return code.length == 8 && _isNumeric(code);
      case BarcodeFormat.upcA:
        return code.length == 12 && _isNumeric(code);
      case BarcodeFormat.upcE:
        return code.length == 8 && _isNumeric(code);
      case BarcodeFormat.code128:
        // Code 128 can contain alphanumeric characters, typical length 6-20
        return code.length >= 6 && code.length <= 20 && _isAlphanumeric(code);
      case BarcodeFormat.code39:
        // Code 39 can contain alphanumeric + some symbols, typical length 6-20
        return code.length >= 6 && code.length <= 20 && _isValidCode39(code);
      default:
        return false;
    }
  }

  /// Check if string contains only numeric characters
  bool _isNumeric(String str) {
    return RegExp(r'^[0-9]+$').hasMatch(str);
  }

  /// Check if string contains only alphanumeric characters
  bool _isAlphanumeric(String str) {
    return RegExp(r'^[a-zA-Z0-9]+$').hasMatch(str);
  }

  /// Check if string is valid Code 39 format (alphanumeric + space, dash, period, etc.)
  bool _isValidCode39(String str) {
    return RegExp(r'^[A-Z0-9\-. $/+%]+$').hasMatch(str);
  }

  @override
  Widget build(BuildContext context) {
    return Scaffold(
      backgroundColor: Colors.black,
      body: Stack(
        children: [
          // Camera preview
          if (controller != null)
            MobileScanner(
              controller: controller!,
              onDetect: _onDetect,
            ),

          // Overlay with scanning frame
          CustomPaint(
            painter: ScannerOverlayPainter(),
            child: Container(),
          ),

          // Scanning line animation
          AnimatedBuilder(
            animation: _animation,
            builder: (context, child) {
              return Positioned(
                top: MediaQuery.of(context).size.height * 0.3 +
                    (_animation.value * 200),
                left: MediaQuery.of(context).size.width * 0.1,
                right: MediaQuery.of(context).size.width * 0.1,
                child: Container(
                  height: 2,
                  decoration: BoxDecoration(
                    gradient: LinearGradient(
                      colors: [
                        Colors.transparent,
                        Colors.blue,
                        Colors.transparent,
                      ],
                    ),
                    boxShadow: [
                      BoxShadow(
                        color: Colors.blue.withOpacity(0.5),
                        blurRadius: 10,
                        spreadRadius: 1,
                      ),
                    ],
                  ),
                ),
              );
            },
          ),

          // Top toolbar with context
          SafeArea(
            child: Column(
              children: [
                // Header toolbar
                Padding(
                  padding: const EdgeInsets.all(16.0),
                  child: Row(
                    children: [
                      IconButton(
                        onPressed: widget.onCancel,
                        icon: const Icon(Icons.arrow_back, color: Colors.white),
                      ),
                      Expanded(
                        child: Text(
                          widget.headerText ?? 'Scan Barcode',
                          textAlign: TextAlign.center,
                          style: const TextStyle(
                            color: Colors.white,
                            fontSize: 18,
                            fontWeight: FontWeight.bold,
                          ),
                        ),
                      ),
                      IconButton(
                        onPressed: () {
                          setState(() {
                            _torchEnabled = !_torchEnabled;
                            controller?.toggleTorch();
                          });
                        },
                        icon: Icon(
                          _torchEnabled ? Icons.flash_on : Icons.flash_off,
                          color: Colors.white,
                        ),
                      ),
                    ],
                  ),
                ),
                
                // Context widget if provided
                if (widget.contextWidget != null)
                  Container(
                    margin: const EdgeInsets.symmetric(horizontal: 16),
                    padding: const EdgeInsets.all(12),
                    decoration: BoxDecoration(
                      color: Colors.black54,
                      borderRadius: BorderRadius.circular(8),
                    ),
                    child: widget.contextWidget!,
                  ),
                
                // Scanned items counter for batch/continuous mode
                if (widget.mode != ScannerMode.single && _scannedBarcodes.isNotEmpty)
                  Container(
                    margin: const EdgeInsets.symmetric(horizontal: 16, vertical: 8),
                    padding: const EdgeInsets.symmetric(horizontal: 16, vertical: 8),
                    decoration: BoxDecoration(
                      color: Colors.green.withOpacity(0.8),
                      borderRadius: BorderRadius.circular(20),
                    ),
                    child: Row(
                      mainAxisSize: MainAxisSize.min,
                      children: [
                        const Icon(Icons.check_circle, color: Colors.white, size: 20),
                        const SizedBox(width: 8),
                        Text(
                          '${_scannedBarcodes.length} items scanned',
                          style: const TextStyle(
                            color: Colors.white,
                            fontWeight: FontWeight.bold,
                          ),
                        ),
                      ],
                    ),
                  ),
              ],
            ),
          ),

          // Bottom instruction text and manual entry
          Positioned(
            bottom: 20,
            left: 0,
            right: 0,
            child: Column(
              children: [
                // Manual entry section
                if (_showManualEntry)
                  Container(
                    margin: const EdgeInsets.symmetric(horizontal: 32),
                    padding: const EdgeInsets.all(16),
                    decoration: BoxDecoration(
                      color: Colors.white,
                      borderRadius: BorderRadius.circular(12),
                    ),
                    child: Column(
                      children: [
                        const Text(
                          'Enter Barcode Manually',
                          style: TextStyle(
                            fontSize: 16,
                            fontWeight: FontWeight.bold,
                          ),
                        ),
                        const SizedBox(height: 12),
                        TextField(
                          controller: _manualEntryController,
                          keyboardType: TextInputType.number,
                          decoration: InputDecoration(
                            hintText: 'Enter barcode number',
                            border: OutlineInputBorder(
                              borderRadius: BorderRadius.circular(8),
                            ),
                            suffixIcon: IconButton(
                              icon: const Icon(Icons.send),
                              onPressed: () {
                                final code = _manualEntryController.text.trim();
                                if (code.isNotEmpty) {
                                  widget.onBarcodeScanned(code);
                                  _manualEntryController.clear();
                                  setState(() {
                                    _showManualEntry = false;
                                  });
                                  if (widget.mode == ScannerMode.single) {
                                    Navigator.pop(context);
                                  }
                                }
                              },
                            ),
                          ),
                          onSubmitted: (code) {
                            if (code.trim().isNotEmpty) {
                              widget.onBarcodeScanned(code.trim());
                              _manualEntryController.clear();
                              setState(() {
                                _showManualEntry = false;
                              });
                              if (widget.mode == ScannerMode.single) {
                                Navigator.pop(context);
                              }
                            }
                          },
                          autofocus: true,
                        ),
                        const SizedBox(height: 8),
                        TextButton(
                          onPressed: () {
                            setState(() {
                              _showManualEntry = false;
                            });
                          },
                          child: const Text('Cancel'),
                        ),
                      ],
                    ),
                  )
                else
                  Container(
                    padding: const EdgeInsets.symmetric(horizontal: 32, vertical: 16),
                    decoration: BoxDecoration(
                      color: Colors.black54,
                      borderRadius: BorderRadius.circular(8),
                    ),
                    child: Column(
                      children: [
                        const Icon(
                          Icons.qr_code_scanner,
                          color: Colors.white,
                          size: 48,
                        ),
                        const SizedBox(height: 8),
                        const Text(
                          'Position the barcode within the frame',
                          textAlign: TextAlign.center,
                          style: TextStyle(
                            color: Colors.white,
                            fontSize: 16,
                          ),
                        ),
                        if (hasScanned)
                          const Padding(
                            padding: EdgeInsets.only(top: 8),
                            child: Text(
                              'Barcode detected! Processing...',
                              style: TextStyle(
                                color: Colors.green,
                                fontSize: 14,
                                fontWeight: FontWeight.bold,
                              ),
                            ),
                          ),
                        if (widget.showManualEntry)
                          Padding(
                            padding: const EdgeInsets.only(top: 12),
                            child: TextButton.icon(
                              onPressed: () {
                                setState(() {
                                  _showManualEntry = true;
                                });
                              },
                              icon: const Icon(Icons.keyboard, color: Colors.white),
                              label: const Text(
                                'Enter manually',
                                style: TextStyle(color: Colors.white),
                              ),
                              style: TextButton.styleFrom(
                                backgroundColor: Colors.white.withOpacity(0.2),
                                padding: const EdgeInsets.symmetric(
                                  horizontal: 16,
                                  vertical: 8,
                                ),
                                shape: RoundedRectangleBorder(
                                  borderRadius: BorderRadius.circular(20),
                                ),
                              ),
                            ),
                          ),
                      ],
                    ),
                  ),
              ],
            ),
          ),
        ],
      ),
    );
  }
}

class ScannerOverlayPainter extends CustomPainter {
  @override
  void paint(Canvas canvas, Size size) {
    final paint = Paint()
      ..color = Colors.black54
      ..style = PaintingStyle.fill;

    final framePaint = Paint()
      ..color = Colors.white
      ..style = PaintingStyle.stroke
      ..strokeWidth = 3;

    final cornerPaint = Paint()
      ..color = Colors.blue
      ..style = PaintingStyle.stroke
      ..strokeWidth = 4;

    // Calculate frame dimensions
    const frameWidth = 0.8;
    const frameHeight = 0.4;
    final frameLeft = size.width * (1 - frameWidth) / 2;
    final frameTop = size.height * 0.3;
    final frameRect = Rect.fromLTWH(
      frameLeft,
      frameTop,
      size.width * frameWidth,
      size.height * frameHeight * 0.5,
    );

    // Draw overlay background
    final path = Path()
      ..addRect(Rect.fromLTWH(0, 0, size.width, size.height))
      ..addRect(frameRect)
      ..fillType = PathFillType.evenOdd;

    canvas.drawPath(path, paint);

    // Draw frame border
    canvas.drawRect(frameRect, framePaint);

    // Draw corner indicators
    final cornerLength = 20.0;
    
    // Top-left corner
    canvas.drawLine(
      Offset(frameRect.left, frameRect.top),
      Offset(frameRect.left + cornerLength, frameRect.top),
      cornerPaint,
    );
    canvas.drawLine(
      Offset(frameRect.left, frameRect.top),
      Offset(frameRect.left, frameRect.top + cornerLength),
      cornerPaint,
    );

    // Top-right corner
    canvas.drawLine(
      Offset(frameRect.right, frameRect.top),
      Offset(frameRect.right - cornerLength, frameRect.top),
      cornerPaint,
    );
    canvas.drawLine(
      Offset(frameRect.right, frameRect.top),
      Offset(frameRect.right, frameRect.top + cornerLength),
      cornerPaint,
    );

    // Bottom-left corner
    canvas.drawLine(
      Offset(frameRect.left, frameRect.bottom),
      Offset(frameRect.left + cornerLength, frameRect.bottom),
      cornerPaint,
    );
    canvas.drawLine(
      Offset(frameRect.left, frameRect.bottom),
      Offset(frameRect.left, frameRect.bottom - cornerLength),
      cornerPaint,
    );

    // Bottom-right corner
    canvas.drawLine(
      Offset(frameRect.right, frameRect.bottom),
      Offset(frameRect.right - cornerLength, frameRect.bottom),
      cornerPaint,
    );
    canvas.drawLine(
      Offset(frameRect.right, frameRect.bottom),
      Offset(frameRect.right, frameRect.bottom - cornerLength),
      cornerPaint,
    );
  }

  @override
  bool shouldRepaint(covariant CustomPainter oldDelegate) => false;
}