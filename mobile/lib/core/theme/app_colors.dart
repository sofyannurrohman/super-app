import 'package:flutter/material.dart';

class AppColors {
  // Primary Corporate Theme - Premium Elegant Emerald Green
  static const Color primaryGreen =
      Color(0xFF059669); // Premium Emerald Green (Primary Brand Color)
  static const Color primaryGreenLight =
      Color(0xFF34D399); // Soft Emerald Accent (Secondary Accent)
  static const Color primaryGreenDark =
      Color(0xFF064E3B); // Deep Forest Emerald (Deep Dark Accent)

  // Backwards compatibility for now (mapping red to corporate theme)
  static const Color primaryRed = primaryGreen;
  static const Color primaryRedLight = primaryGreenLight;

  // Premium Gradients for Headers and Cards
  static const List<Color> primaryGradient = [
    Color(0xFF059669),
    Color(0xFF064E3B)
  ]; // Emerald to Deep Forest Emerald
  static const List<Color> accentGradient = [
    Color(0xFF34D399),
    Color(0xFF059669)
  ]; // Soft Emerald to Medium Emerald

  // Backgrounds
  static const Color background = Color(0xFFFFFFFF);
  static const Color surface = Color(0xFFFFFFFF); // Pure white surface
  static const Color backgroundAlt = Color(0xFFFBFBFB); // Very light gray

  // Text - Slate palette for professional look
  static const Color textPrimary = Color(0xFF1E293B);
  static const Color textSecondary = Color(0xFF64748B);
  static const Color textTertiary = Color(0xFF94A3B8);

  // Greyscale
  static const Color grayBorder = Color(0xFFE2E8F0);
  static const Color grayLight = Color(0xFFF1F5F9);

  // Semantic
  static const Color success = Color(0xFF10B981);
  static const Color warning = Color(0xFFF59E0B);
  static const Color error = Color(0xFFEF4444);
  static const Color info = Color(0xFF0EA5E9);

  // Premium Elegant Blue Themes
  static const Color primaryBlue = Color(0xFF1D4ED8); // Royal Blue
  static const Color primaryBlueLight = Color(0xFF3B82F6); // Soft Blue
  static const Color primaryBlueDark = Color(0xFF1E3A8A); // Deep Navy Slate

  // Specific premium elegant dark blue variants for each operational role
  // 1. Warehouse (Logistics/Operations) - Premium Emerald Green
  static const Color blueWarehouse = Color(0xFF059669);
  static const Color blueWarehouseDark = Color(0xFF064E3B);
  static const Color blueWarehouseLight = Color(0xFF34D399);

  // 2. Delivery (Action/Logistics) - Rich Sapphire Navy
  static const Color blueDelivery = Color(0xFF1E40AF);
  static const Color blueDeliveryDark = Color(0xFF172554);
  static const Color blueDeliveryLight = Color(0xFF60A5FA);

  // 3. HR (Management/Corporate) - Premium Emerald Green
  static const Color blueHR = Color(0xFF059669);
  static const Color blueHRDark = Color(0xFF064E3B);
  static const Color blueHRLight = Color(0xFF34D399);

  // 4. Sales (Growth/Commercial) - Premium Emerald Green
  static const Color blueSales = Color(0xFF059669);
  static const Color blueSalesDark = Color(0xFF064E3B);
  static const Color blueSalesLight = Color(0xFF6EE7B7);
}
