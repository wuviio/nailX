// This file is a development stub.
// In production, run: flutterfire configure
// to generate this file with your actual Firebase project credentials.
import 'package:firebase_core/firebase_core.dart' show FirebaseOptions;
import 'package:flutter/foundation.dart'
    show defaultTargetPlatform, kIsWeb, TargetPlatform;

class DefaultFirebaseOptions {
  static FirebaseOptions get currentPlatform {
    if (kIsWeb) return web;
    switch (defaultTargetPlatform) {
      case TargetPlatform.android:
        return android;
      case TargetPlatform.iOS:
        return ios;
      default:
        return web;
    }
  }

  static const FirebaseOptions web = FirebaseOptions(
    apiKey: 'AIzaSyDyhp-D0Zd9jgkxvEQkkDQXCtsx8riTHGo',
    appId: '1:67384230476:web:bfcb5b864699ace2da6ae9',
    messagingSenderId: '67384230476',
    projectId: 'nailx-e45d9',
    authDomain: 'nailx-e45d9.firebaseapp.com',
    storageBucket: 'nailx-e45d9.firebasestorage.app',
    measurementId: 'G-TKEQHM9WGH',
  );

  // Replace these with values from google-services.json / GoogleService-Info.plist

  static const FirebaseOptions android = FirebaseOptions(
    apiKey: 'AIzaSyDGkrtejncSfmy3aYSBr__uVhoW6TVe69Q',
    appId: '1:67384230476:android:109e6a02744498b8da6ae9',
    messagingSenderId: '67384230476',
    projectId: 'nailx-e45d9',
    storageBucket: 'nailx-e45d9.firebasestorage.app',
  );

  static const FirebaseOptions ios = FirebaseOptions(
    apiKey: 'AIzaSyA8TZGRX-dKB4oCImSGMG15ST193asyllw',
    appId: '1:67384230476:ios:8aa69d856b3232fdda6ae9',
    messagingSenderId: '67384230476',
    projectId: 'nailx-e45d9',
    storageBucket: 'nailx-e45d9.firebasestorage.app',
    iosBundleId: 'com.example.frontend',
  );

}