import 'package:firebase_auth/firebase_auth.dart';
import 'package:flutter/material.dart';
import 'package:go_router/go_router.dart';
import 'package:riverpod_annotation/riverpod_annotation.dart';

import '../../features/ar/presentation/ar_screen.dart';
import '../../features/auction/presentation/auction_detail_screen.dart';
import '../../features/auction/presentation/auction_new_screen.dart';
import '../../features/auth/presentation/login_screen.dart';
import '../../features/auth/presentation/register_screen.dart';
import '../../features/booking/presentation/booking_detail_screen.dart';
import '../../features/designs/presentation/design_detail_screen.dart';
import '../../features/designs/presentation/design_new_screen.dart';
import '../../features/designs/presentation/home_screen.dart';
import '../../features/profile/presentation/profile_edit_screen.dart';
import '../../features/auction/presentation/bid_form_screen.dart';
import '../../features/auction/presentation/salon_dashboard_screen.dart';
import '../../features/profile/presentation/ip_wallet_screen.dart';
import '../../features/profile/presentation/notification_screen.dart';
import '../../features/profile/presentation/settings_screen.dart';

part 'router.g.dart';

class AppRoutes {
  static const splash = '/';
  static const login = '/login';
  static const register = '/register';
  static const home = '/home';
  static const designDetail = '/designs/:id';
  static const designNew = '/designs/new';
  static const ar = '/ar';
  static const profile = '/profile';
  static const profileEdit = '/profile/edit';
  static const salonDetail = '/salons/:id';
  static const auctionNew = '/auctions/new';
  static const auctionDetail = '/auctions/:id';
  static const bookings = '/bookings';
  static const bookingDetail = '/bookings/:id';
  static const notifications = '/notifications';
  static const settings = '/settings';
  static const ipWallet = '/wallet';
  static const salonDashboard = '/salon/dashboard';
  static const bidForm = '/salon/bid/:request_id';
}

@riverpod
GoRouter router(RouterRef ref) {
  return GoRouter(
    initialLocation: AppRoutes.splash,
    redirect: (context, state) {
      final user = FirebaseAuth.instance.currentUser;
      final isAuth = user != null;
      final path = state.uri.path;

      // Splash → route based on auth state
      if (path == AppRoutes.splash) return isAuth ? AppRoutes.home : AppRoutes.login;

      // Unauthenticated users can only see login and register
      // (register shows a "please sign in first" message when accessed without auth)
      if (!isAuth && path != AppRoutes.login && path != AppRoutes.register) {
        return AppRoutes.login;
      }

      // Authenticated users on login page → home
      // Register page is intentionally kept accessible for new-user profile completion
      if (isAuth && path == AppRoutes.login) return AppRoutes.home;

      return null;
    },
    routes: [
      GoRoute(
        path: AppRoutes.splash,
        builder: (_, __) => const _LoadingPage(),
      ),
      GoRoute(
        path: AppRoutes.login,
        builder: (_, __) => const LoginScreen(),
      ),
      GoRoute(
        path: AppRoutes.register,
        builder: (_, __) => const RegisterScreen(),
      ),
      GoRoute(
        path: AppRoutes.home,
        builder: (_, __) => const HomeScreen(),
      ),
      // designs/new must be registered before designs/:id to avoid path conflicts
      GoRoute(
        path: AppRoutes.designNew,
        builder: (_, __) => const DesignNewScreen(),
      ),
      GoRoute(
        path: AppRoutes.designDetail,
        builder: (_, state) => DesignDetailScreen(id: state.pathParameters['id']!),
      ),
      GoRoute(
        path: AppRoutes.ar,
        builder: (_, __) => const ARScreen(),
      ),
      GoRoute(
        path: AppRoutes.profileEdit,
        builder: (_, __) => const ProfileEditScreen(),
      ),
      GoRoute(
        path: AppRoutes.salonDetail,
        builder: (_, state) => _Placeholder('サロン詳細 ${state.pathParameters['id']}'),
      ),
      // auctions/new before auctions/:id
      GoRoute(
        path: AppRoutes.auctionNew,
        builder: (_, state) => AuctionNewScreen(
          designIpId: state.uri.queryParameters['design_ip_id'],
        ),
      ),
      GoRoute(
        path: AppRoutes.auctionDetail,
        builder: (_, state) => AuctionDetailScreen(id: state.pathParameters['id']!),
      ),
      GoRoute(
        path: AppRoutes.bookings,
        builder: (_, __) => const _Placeholder('マイ予約一覧'),
      ),
      GoRoute(
        path: AppRoutes.bookingDetail,
        builder: (_, state) => BookingDetailScreen(id: state.pathParameters['id']!),
      ),
      GoRoute(
        path: AppRoutes.notifications,
        builder: (_, __) => const NotificationScreen(),
      ),
      GoRoute(
        path: AppRoutes.settings,
        builder: (_, __) => const SettingsScreen(),
      ),
      GoRoute(
        path: AppRoutes.ipWallet,
        builder: (_, __) => const IPWalletScreen(),
      ),
      GoRoute(
        path: AppRoutes.salonDashboard,
        builder: (_, __) => const SalonDashboardScreen(),
      ),
      GoRoute(
        path: AppRoutes.bidForm,
        builder: (_, state) => BidFormScreen(requestId: state.pathParameters['request_id']!),
      ),
    ],
    errorBuilder: (_, state) => Scaffold(
      body: Center(child: Text('ページが見つかりません: ${state.uri}')),
    ),
  );
}

class _LoadingPage extends StatelessWidget {
  const _LoadingPage();

  @override
  Widget build(BuildContext context) =>
      const Scaffold(body: Center(child: CircularProgressIndicator()));
}

class _Placeholder extends StatelessWidget {
  final String name;
  const _Placeholder(this.name);

  @override
  Widget build(BuildContext context) => Scaffold(
        appBar: AppBar(title: Text(name)),
        body: Center(child: Text('$name\n（実装予定）', textAlign: TextAlign.center)),
      );
}
