package main

import (
	"fmt"
	"log/slog"
	"net"
	"os"

	"github.com/gen2brain/malgo"
)

var (
	logger = slog.New(slog.NewTextHandler(os.Stderr, nil))
)

func recordTest() {
	ctx, err := malgo.InitContext(nil, malgo.ContextConfig{}, func(message string) {
		logger.Info("LOG", "msg", message)
	})
	if err != nil {
		logger.Warn("init malgo context failed", "error", err)
		os.Exit(1)
	}
	defer func() {
		_ = ctx.Uninit()
		ctx.Free()
	}()

	devCfg := malgo.DefaultDeviceConfig(malgo.Capture)
	devCfg.Capture.Channels = 1
	devCfg.Capture.Format = malgo.FormatS16
	devCfg.SampleRate = 48000
	devCfg.Alsa.NoMMap = 1

	var capturedSampleCount uint32 = 0
	captchedSamples := make([]byte, 0)
	sizeInBytes := uint32(malgo.SampleSizeInBytes(devCfg.Capture.Format))

	onRecvFrames := func(pSample2, pSample []byte, frameCount uint32) {
		sampleCount := frameCount * devCfg.Capture.Channels * sizeInBytes
		newCaptchedSampleCount := capturedSampleCount + sampleCount
		captchedSamples = append(captchedSamples, pSample...)
		capturedSampleCount = newCaptchedSampleCount
	}

	logger.Info("recording...")
	captureCallbacks := malgo.DeviceCallbacks{
		Data: onRecvFrames,
	}

	device, err := malgo.InitDevice(ctx.Context, devCfg, captureCallbacks)
	if err != nil {
		logger.Warn("init device failed", "error", err)
		os.Exit(1)
	}

	err = device.Start()
	if err != nil {
		logger.Warn("start device failed", "error", err)
		os.Exit(1)
	}
	logger.Info("press enter to stop recording...")
	fmt.Scanln()
	device.Uninit()

	logger.Info("captched", "samples", capturedSampleCount, "buf", len(captchedSamples))
}

/*
1. 连接到服务器
2. 发送登录请求 等待响应
3. 发送 开始会话 等待回复
4. 打开设备 开始录制 开始传输数据
5. 结束 发送结束请求
*/
func connectToHost(host, device, session string) error {
	rAddr, err := net.ResolveTCPAddr("tcp", host)
	if err != nil {
		return fmt.Errorf("resolve address failed %w", err)
	}
	dst, err := net.DialTCP("tcp", nil, rAddr)
	if err != nil {
		return fmt.Errorf("dial tcp failed %w", err)
	}

	devReq, err := NewLoginFrame(session, device)
	if err != nil {
		return fmt.Errorf("new login frame failed %w", err)
	}
	_, err = devReq.WriteTo(dst)
	if err != nil {
		return fmt.Errorf("write frame to peer failed %w", err)
	}

	loginRsp, err := fromReader(dst)
	if err != nil {
		return fmt.Errorf("receive login response failed")
	}
	if loginRsp.cmd != devLogin {
		return fmt.Errorf("frame cmd not match devLogin %d", loginRsp.cmd)
	}

	if loginRsp.body == nil || len(loginRsp.body) == 0 || loginRsp.body[0] != 0x00 {
		return fmt.Errorf("response failed")
	}

	// start stream
	ssReq, err := NewStartStreamFrame()
	if err != nil {
		return fmt.Errorf("new start stream failed %w", err)
	}
	_, err = ssReq.WriteTo(dst)
	if err != nil {
		return fmt.Errorf("write start stream failed %w", err)
	}
	ssRsp, err := fromReader(dst)
	if err != nil {
		return fmt.Errorf("receive start stream response failed %w", err)
	}

	return nil
}

func main() {
	err := connectToHost("127.0.0.1:12001", "DEVICE1", "SESSION1")
	if err != nil {
		logger.Warn("client error", "error", err)
		os.Exit(1)
	}

}
