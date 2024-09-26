import 'dart:typed_data';

import 'package:record/record.dart';

typedef cbType = void Function(Uint8List d);

class Recorder {
  final _record = AudioRecorder();
  cbType _cb = (d) {};

  setOnData(cbType f) {
    this._cb = f;
  }

  startRecord() async {
    final hasPerm = await _record.hasPermission();
    if (!hasPerm) {
      print('no perm, ignored!');
      return;
    }
    final rs = await _record
        .startStream(const RecordConfig(
          encoder: AudioEncoder.pcm16bits,
          numChannels: 1,
          sampleRate: 48000,
        ));
    rs.listen((d) {
      print('RECORD: ${d.length}');
      this._cb(d);
    }, onDone: () {
      print('RECORD DONE');
    }, onError: (e) {
      print("RECORD ERROR!");
    }, cancelOnError: true);
  }

  stopRecord() async {
    final rs = await _record.stop();
    print('RECORD stop $rs');
  }
}

final recorder = Recorder();
