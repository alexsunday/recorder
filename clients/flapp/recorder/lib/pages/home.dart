import 'package:flutter/material.dart';
import 'package:recorder/service/conn.dart';

class HomePage extends StatefulWidget {
  const HomePage({super.key});

  @override
  State<StatefulWidget> createState() {
    return _HomePage();
  }
}

class _HomePage extends State<HomePage> {
  @override
  Widget build(BuildContext context) {
    return Scaffold(
        backgroundColor: Colors.greenAccent,
        body: Center(
            child: Row(
              mainAxisAlignment: MainAxisAlignment.center,
          children: [
            FloatingActionButton(
                child: const Icon(Icons.play_arrow),
                onPressed: () {
                  onOpenClicked();
                }),
                Container(
                  width: 10,
                ),
            FloatingActionButton(
                child: const Icon(Icons.stop),
                onPressed: () {
                  onCloseClicked();
                }),
          ],
        )));
  }

  onOpenClicked() {
    glws.open();
  }

  onCloseClicked() {
    glws.close();
  }
}
