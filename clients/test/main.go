package main

import (
	"fmt"
	"io"
	"log/slog"
	"net"
	"os"

	"github.com/gen2brain/malgo"
)

var (
	logger = slog.New(slog.NewTextHandler(os.Stderr, nil))
)

type recordOpt struct {
	Channels    int
	SampleRate  int
	SampleDepth int
}

// type DataProc func(pOutputSample, pInputSamples []byte, framecount uint32)

func startRecord(out io.ReadWriteCloser, opt *recordOpt) error {
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

	var audioFmt malgo.FormatType
	if opt.SampleDepth == 16 {
		audioFmt = malgo.FormatS16
	} else if opt.SampleDepth == 8 {
		audioFmt = malgo.FormatU8
	} else if opt.SampleDepth == 24 {
		audioFmt = malgo.FormatS24
	} else if opt.SampleDepth == 32 {
		audioFmt = malgo.FormatS32
	}

	devCfg := malgo.DefaultDeviceConfig(malgo.Capture)
	devCfg.Capture.Channels = uint32(opt.Channels)
	devCfg.Capture.Format = audioFmt
	devCfg.SampleRate = uint32(opt.SampleRate)
	devCfg.Alsa.NoMMap = 1

	var capturedSampleCount uint32 = 0
	captchedSamples := make([]byte, 0)
	// sizeInBytes := uint32(malgo.SampleSizeInBytes(devCfg.Capture.Format))

	onRecvFrames := func(pOutputSample, pInputSamples []byte, frameCount uint32) {
		// sampleCount := frameCount * devCfg.Capture.Channels * sizeInBytes
		// newCaptchedSampleCount := capturedSampleCount + sampleCount
		// captchedSamples = append(captchedSamples, pInputSamples...)
		// capturedSampleCount = newCaptchedSampleCount
		// pInputSamples 就是录制的内容
		_, err = NewFrame(audioStream, 0, pInputSamples).WriteTo(out)
		if err != nil {
			logger.Warn("输出音频流失败", "error", err)
		}
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
	return nil
}

/*
1. 连接到服务器
2. 发送登录请求 等待响应
3. 发送 开始会话 等待回复
4. 打开设备 开始录制 开始传输数据
5. 结束 发送结束请求
*/
func connectToHost(host, device, session string) (io.ReadWriteCloser, error) {
	rAddr, err := net.ResolveTCPAddr("tcp", host)
	if err != nil {
		return nil, fmt.Errorf("resolve address failed %w", err)
	}
	dst, err := net.DialTCP("tcp", nil, rAddr)
	if err != nil {
		return nil, fmt.Errorf("dial tcp failed %w", err)
	}

	// login request
	devReq, err := NewLoginFrame(session, device)
	if err != nil {
		return nil, fmt.Errorf("new login frame failed %w", err)
	}
	_, err = devReq.WriteTo(dst)
	if err != nil {
		return nil, fmt.Errorf("write frame to peer failed %w", err)
	}

	loginRsp, err := fromReader(dst)
	if err != nil {
		return nil, fmt.Errorf("receive login response failed")
	}
	if loginRsp.cmd != devLogin {
		return nil, fmt.Errorf("frame cmd not match devLogin %d", loginRsp.cmd)
	}

	if loginRsp.body == nil || len(loginRsp.body) == 0 || loginRsp.body[0] != 0x00 {
		return nil, fmt.Errorf("response failed")
	}

	return dst, nil
}

func sendStartStreamRequest(p io.ReadWriteCloser, opt *recordOpt) error {
	// start stream request
	ssReq, err := NewStartStreamFrame(opt.SampleDepth, opt.Channels, opt.SampleRate)
	if err != nil {
		return fmt.Errorf("new start stream failed %w", err)
	}
	_, err = ssReq.WriteTo(p)
	if err != nil {
		return fmt.Errorf("write start stream failed %w", err)
	}
	ssRsp, err := fromReader(p)
	if err != nil {
		return fmt.Errorf("receive start stream response failed %w", err)
	}
	rsp, err := parseStartStreamResponse(ssRsp)
	if err != nil {
		return fmt.Errorf("parse start stream response failed %w", err)
	}
	if rsp.Code != 0 {
		return fmt.Errorf("start stream request failed %d", rsp.Code)
	}
	return nil
}

func main() {
	conn, err := connectToHost("127.0.0.1:12001", "DEVICE1", "SESSION1")
	if err != nil {
		logger.Warn("client error", "error", err)
		os.Exit(1)
	}

	var opt = &recordOpt{
		SampleRate:  48000,
		Channels:    1,
		SampleDepth: 16,
	}
	err = sendStartStreamRequest(conn, opt)
	if err != nil {
		logger.Warn("send start stream request failed", "error", err)
		os.Exit(1)
	}

	err = startRecord(conn, opt)
	if err != nil {
		logger.Warn("start record failed", "error", err)
		os.Exit(1)
	}
}
