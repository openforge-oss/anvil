// Package detect identifies a project's ecosystem(s) from marker files:
// pubspec.yaml → Flutter; package.json with a react-native dep → React Native;
// build.gradle / settings.gradle → native Android; *.xcodeproj / *.xcworkspace
// / Podfile → native iOS; and more as drivers are added.
//
// Detectors run independently and return a confidence score; the pipeline
// resolves multiple hits (e.g. an RN app with native folders) by confidence and
// nesting. Implemented in the tool-build phase — see docs/ARCHITECTURE.md.
package detect
