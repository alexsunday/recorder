import 'package:flutter/material.dart';
import 'package:recorder/service/conn.dart';
import 'package:recorder/service/recorder.dart';

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
                child: const Icon(Icons.link_outlined),
                onPressed: () {
                  _onOpenClicked();
                }),
            Container(
              width: 10,
            ),
            FloatingActionButton(
                child: const Icon(Icons.link_off),
                onPressed: () {
                  _onCloseClicked();
                }),
            Container(
              width: 10,
            ),
            FloatingActionButton(
              child: const Icon(
                Icons.circle,
                color: Colors.redAccent,
              ),
              onPressed: () {
                _onBeginRecord();
              },
            ),
            Container(
              width: 10,
            ),
            FloatingActionButton(
                child: const Icon(Icons.stop),
                onPressed: () {
                  _onStopRecord();
                }),
          ],
        )));
  }

  _onOpenClicked() {
    glws.open();
  }

  _onCloseClicked() {
    glws.close();
  }

  _onBeginRecord() {
    recorder.startRecord();
  }

  _onStopRecord() {
    recorder.stopRecord();
  }
}
