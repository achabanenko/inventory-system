import 'package:flutter/material.dart';
import 'package:provider/provider.dart';
import 'services/secure_auth_service.dart';
import 'services/api_service.dart';
import 'widgets/auth_wrapper.dart';
import 'theme/app_theme.dart';

void main() {
  runApp(const MyApp());
}

class MyApp extends StatelessWidget {
  const MyApp({super.key});

  @override
  Widget build(BuildContext context) {
    return MultiProvider(
      providers: [
        // API Service provider
        Provider(create: (context) => ApiService()),

        // Secure Authentication Service provider
        ChangeNotifierProxyProvider<ApiService, SecureAuthService>(
          create: (context) => SecureAuthService(),
          update: (context, apiService, authService) {
            authService!.setApiService(apiService);
            return authService;
          },
        ),
      ],
      child: MaterialApp(
        title: 'Inventory Manager',
        theme: AppTheme.lightTheme,
        darkTheme: AppTheme.darkTheme,
        themeMode: ThemeMode.system,
        // Use AuthWrapper to handle the authentication flow
        home: const AuthWrapper(),
        debugShowCheckedModeBanner: false,
      ),
    );
  }
}
