import 'package:flutter/material.dart';
import 'package:provider/provider.dart';
import 'services/auth_service.dart';
import 'services/api_service.dart';
import 'screens/login_screen.dart';
import 'screens/home_screen.dart';
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
        Provider(create: (context) => ApiService()),
        ChangeNotifierProxyProvider<ApiService, AuthService>(
          create: (context) => AuthService(),
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
        home: Consumer<AuthService>(
          builder: (context, authService, child) {
            return authService.isAuthenticated 
                ? const HomeScreen() 
                : const LoginScreen();
          },
        ),
      ),
    );
  }
}
