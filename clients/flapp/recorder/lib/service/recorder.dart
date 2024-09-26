import 'package:record/record.dart';

class Recorder {
  final _record = AudioRecorder();

  startRecord() async {
    final hasPerm = await _record.hasPermission();
    if (!hasPerm) {
      print('no perm, ignored!');
      return;
    }
    final rs = await _record
        .startStream(const RecordConfig(encoder: AudioEncoder.pcm16bits));
    rs.listen((d) {
      print('RECORD: ${d.length}');
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
