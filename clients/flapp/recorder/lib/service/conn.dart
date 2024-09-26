/*
var ws = WebSocketConn(wsUrl);
ws.open();
*/

import 'dart:io';
import 'dart:typed_data';

import 'frame.dart';

class WebSocketConn {
  var wsUrl = "ws://127.0.0.1:18000/websocket/link";
  var buf = Uint8List(0);
  WebSocket? sock;

  open() {
    WebSocket.connect(wsUrl).then((ws) {
      sock = ws;

      ws.pingInterval = const Duration(seconds: 30);
      // 处理ws事件
      ws.listen((data) {
        assert(data is Uint8List);
        Uint8List d = data;
        _onData(d);
      }, onDone: () {
        // 连接关闭
        _onDone();
      }, onError: (e) {
        // 处理错误
        _onError(e);
      }, cancelOnError: true);
      // 发送初始化数据
      _onWsOpen(ws);
    }, onError: (e) {
      // 连接失败
      _onConnectError(e);
    });
  }

  // 连接到了 ws
  _onWsOpen(WebSocket ws) {
    print('已连接到服务器');
    var req = Frame.newLoginFrame("MOBILE-SESSION", "DEVICE-01");
    ws.add(req.toBytes());
  }

  _onConnectError(dynamic e) {
    print('连接失败 $e');
  }

  _onError(dynamic e) {
    print('出错了 $e');
  }

  _onDone() {
    print('连接已关闭');
  }

  _onData(Uint8List d) {
    print('收到数据 ${d.length}');
    var origin = buf;
    buf = Uint8List(origin.length + d.length);
    buf.setRange(0, origin.length, origin);
    buf.setRange(origin.length, origin.length + d.length, d);
    // 处理完所有数据包
    while (true) {
      var unpacked = Frame.extractFrame(buf);
      // 如果已经没有数据了 则退出处理流程
      if (unpacked.$1 == null) {
        if (unpacked.$2 != 0) {
          throw Exception("extract frame, but response not match");
        }
        break;
      }
      var frame = unpacked.$1!;
      var used = unpacked.$2;
      buf = buf.sublist(used);
      _handleFrame(frame);
    }
  }

  _handleFrame(Frame f) {
    switch (f.cmd) {
      case cmdDevLogin:
        _handleLoginResponse(f.body);
        break;
      case cmdStartStream:
        _handleStartStreamResponse(f.body);
        break;
      default:
        throw Exception('暂不支持的数据格式');
    }
  }

  _handleLoginResponse(Uint8List body) {
    if(body[0] != 0x00) {
      throw Exception("认证失败");
    }
    print('认证完成');
    writeStartStream();
  }

  _handleStartStreamResponse(Uint8List body) {
    var resp = StartStreamResponse.parse(body);
    if(resp.code != 0) {
      throw Exception("开启失败");
    }
    print('开启会话完成');
  }

  writeFrame(Frame req) {
    if(sock == null) {
      throw Exception("未初始化或未连接");
    }
    sock!.add(req.toBytes());
  }

  writeStartStream() {
    var out = Frame.newStartStreamFrame(16, 1, 48000);
    writeFrame(out);
  }

  close() {
    if(sock != null) {
      sock!.close(0x1000, "close");
      sock = null;
    }
  }
}

var glws = WebSocketConn();
