import 'package:flutter_test/flutter_test.dart';
import 'package:AISTD_DART_NAME/main.dart';

void main() {
  testWidgets('shows the project name', (tester) async {
    await tester.pumpWidget(const App());
    expect(find.text('AISTD_PROJECT_NAME'), findsOneWidget);
  });
}
