
import "dart:typed_data";

import "package:recorder/service/frame.dart";
import "package:test/test.dart";

void main() {
  test('new frame', () {
    var f1 = Frame(1, 0, Uint8List.fromList([]));
    expect(f1.cmd, 1);
    expect(f1.opt, 0);
    expect(f1.body.length, 0);
  });

  test('from bytes', () {
    var b1 = Uint8List.fromList([0x03, 0x01, 0x00, 0x00, 0x01]);
    var f1 = Frame.fromBytes(b1);
    expect(f1.cmd, 0x03);
    expect(f1.opt, 0x01);
    expect(f1.body.length, 0x03);
    expect(f1.body[2], 0x01);
  });

  test('to bytes', () {
    var b1 = Uint8List.fromList([0x03, 0x01, 0x00, 0x00, 0x01]);
    var f1 = Frame(0x01, 0x12, b1);
    var out = f1.toBytes();
    expect(out[0], 0);
    expect(out[1], b1.length + 4);
    expect(out[2], 0x01);
    expect(out[3], 0x12);
    expect(out[4], 0x03);
  });

  test('extract frame', () {
    var b1 = Uint8List.fromList([0x03, 0x01, 0x00, 0x00, 0x01]);
    var f1 = Frame(0x01, 0x12, b1);
    var f2 = Frame(0x02, 0x00, Uint8List.fromList([]));
    var f3 = Frame(0x03, 0x01, b1.sublist(1));

    var d1 = f1.toBytes();
    var d2 = f2.toBytes();
    var d3 = f3.toBytes();
    var data = Uint8List(d1.length + d2.length + d3.length + 0x13);
    data.setRange(0, d1.length, d1);
    data.setRange(d1.length, d1.length + d2.length, d2);
    data.setRange(d1.length + d2.length, d1.length + d2.length + d3.length, d3);

    var p1 = Frame.extractFrame(data);
    expect(p1.$1!.cmd, 0x01);
    expect(p1.$2, d1.length);
    
    data = data.sublist(p1.$2);
    var p2 = Frame.extractFrame(data);
    expect(p2.$1!.cmd, 0x02);
    expect(p2.$2, d2.length);

    data = data.sublist(p2.$2);
    var p3 = Frame.extractFrame(data);
    expect(p3.$1!.cmd, 0x03);
    expect(p3.$2, d3.length);

    data = data.sublist(p3.$2);
    expect(data.length, 0x13);
  });
}
