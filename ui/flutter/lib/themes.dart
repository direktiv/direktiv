import 'dart:ui';

import 'package:flutter/material.dart';

const List<Color> lightBackground = [
  Color(0xffe9ecef),
  Color(0xffe9ecef),
  Color(0xffe0e0e0),
  Color(0xffa1a1a1)
];

const List<Color> darkBackground = [
  Color(0xffe9ecef),
  Color(0xffe9ecef),
  Color(0xff0e0e0),
Color(0xffa1a1a1)
];

const Map<String, Color> statusColors = {
  "complete": Color(0xFF28a745),
  "pending":  Color(0xffe5b208),
  "failed":   Color(0xFFFF4040),
};

const Color unknownStatus = Colors.grey;

final colors = UITheme();

class UITheme {
  bool _isListTheme;

  UITheme() {
    this._isListTheme = true;
  }

  Color background(int index) {
    if (_isListTheme) {
      return lightBackground[index];
    }
    return darkBackground[index];
  }

  Color status(String status) {
    debugPrint(status);
    if (statusColors.containsKey(status)) {
      return statusColors[status];
    }
    return unknownStatus;
  }
}
