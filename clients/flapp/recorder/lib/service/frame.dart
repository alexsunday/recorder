
import 'dart:convert';
import 'dart:typed_data';

const int cmdDevLogin = 0x01;
const int cmdStartStream = 0x02;
const int cmdStopStream = 0x03;
const int cmdAudioStream = 0x04;

class Frame {
  int cmd;
  int opt;
  Uint8List body;

  Frame(this.cmd, this.opt, this.body) {
    assert(cmd >= 0 && cmd <= 0xFF);
    assert(opt >= 0 && opt <= 0xFF);
  }

  // d without 2 bytes head
  static Frame fromBytes(Uint8List d) {
    var cmd = d[0];
    var opt = d[1];
    var body = d.sublist(2);
    return Frame(cmd, opt, body);
  }

  // 从一个足够大的缓冲区提取一个 Frame 返回一个Frame? 和 耗去的字符数量
  static (Frame?, int) extractFrame(Uint8List d) {
    var bView = ByteData.sublistView(d);
    // 先读取一个头部
    if(d.length < 4) {
      return (null, 0);
    }

    // 读两字节头
    var total = bView.getUint16(0, Endian.big);
    if(total < 4) {
      throw Exception("protocol error, head is too small $total");
    }
    if(d.length < total) {
      return (null, 0);
    }

    var sub1 = d.sublist(0, total);
    // fromBytes reveive a frame without header
    var rs = fromBytes(sub1.sublist(2));
    return (rs, total);
  }

  Uint8List toBytes() {
    var rs = Uint8List(2 + 2 + body.length);
    var bView = ByteData.sublistView(rs);
    bView.setUint16(0, 4 + body.length, Endian.big);
    rs[2] = cmd;
    rs[3] = opt;
    rs.setRange(4, rs.length, body);
    return rs;
  }

  static Frame newLoginFrame(String session, String device) {
    var req = {
      "session": session,
      "device": device,
    };
    var body = jsonEncode(req);
    return Frame(cmdDevLogin, 0x00, ascii.encode(body));
  }

  static Frame newStartStreamFrame(int bits, int channels, int sampleRate) {
    var body = jsonEncode({
      "bits": bits,
      "channels": channels,
      "sampleRate": sampleRate,
    });
    return Frame(cmdStartStream, 0x00, ascii.encode(body));
  }

  static newPcmFrame(Uint8List pcm) {
    return Frame(cmdAudioStream, 0x00, pcm);
  }
}

class StartStreamResponse {
  int code;
  int id;
  
  StartStreamResponse(this.code, this.id);

  static StartStreamResponse parse(Uint8List body) {
    var obj = jsonDecode(ascii.decode(body));
    if(obj is! Map) {
      throw Exception("json 解析后必须为 Map");
    }
    Map<dynamic, dynamic> mObj = obj;
    var code = mObj["code"];
    var id = mObj["id"];
    if(code is! int) {
      throw Exception("code 必须为int");
    }
    if(id is! int) {
      throw Exception("id 必须为 int");
    }
    return StartStreamResponse(code, id);
  }
}
