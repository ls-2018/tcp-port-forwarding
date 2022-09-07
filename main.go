package main

import (
	"encoding/base64"
	"flag"
	"fmt"
	"log"
	"net"
	"os"
	"path"
	"proxy/lib"
	"strconv"
	"strings"
	"time"

	// "sync.aut"

	"github.com/spf13/viper"
)

// type
type Config struct {
	//名称，在AMQP模式下参与Queue的构建
	Name string
	//读取Buffer长度
	Size uint
	//日志路径地址
	Log string
	//多个监听地址配置
	Peers []Peer
}

type Server struct {
}

var cfile = flag.String("c", "", "配置文件路径")

// 日志路径
// var logpath = "./log/"
var ver = ""
var loge = false
var logf = log.Printf
var curfile = ""
var logFile *os.File
var TConf *Config = &Config{Peers: make([]Peer, 0)}

const (
	ConType_TCP  = 0
	ConType_AMQP = 1
	ConType_UDP  = 2
)

type Target struct {
	Type int
}

type Peer struct {
	// 监听名称
	Name string
	// 类型 UDP或TCP
	Type string
	// 本地监听地址
	Listen string
	// 转发目标，支持多个地址，如果一个地址不够则自动切换到下个地址，前提是有足够的地址切换，期望支持AMQP
	Targets []string
	// 同步将数据转发到某个地址，支持AMQP
	Duplex string
	// 复制参数类型
	DupType int
	// 连接超时时间
	Timeout uint
	// 日志文件路径，值存在时记录日志，否则不记录日志，自动按日期按名称切割文件，日志格式为 HH:mm:ss IP:Port > hex内容
	Log string
	// 开始包，必须以什么开始
	Start string
	// 结束包，必须以什么结束
	End string
	// 最小长度，低于该长度的数据不予处理
	MinLen int
	// 本地连接
	Listener *net.TCPListener
	// Target   *net.TCPConn
	// Dup      *net.TCPConn
	Resolved []*net.TCPAddr
	// 链接对
	TargetConnMap map[string]*net.TCPConn
	// 源链接对
	SourceConnMap map[string]*net.TCPConn

	Consumed bool
}

func start(p Peer) bool {
	// if p.Type == "UDP" {
	// 	ulocal, err := net.ResolveUDPAddr("udp", p.Listen)
	// 	if err != nil {
	// 		log.Printf("[ERR] [UDP] [%s] 本地地址错误: %s \r\n", p.Name, err)
	// 		return false
	// 	}
	// 	var Target *net.UDPAddr
	// 	for _, v := range p.Targets {
	// 		if strings.HasPrefix(v, "amqp") {
	// 			//TODO AMQP支持
	// 		} else {
	// 			r, e := net.ResolveUDPAddr("udp", v)
	// 			if e != nil {
	// 				log.Printf("[ERR] [UDP] [%s] 转发地址错误: %s \r\n", p.Name, v)
	// 			}
	// 			Target = r
	// 		}
	// 	}
	// 	var DupUDPAddr *net.UDPAddr
	// 	var DupUDPClient *net.UDPConn
	// 	if len(p.Duplex) > 0 {
	// 		DupUDPAddr, err = net.ResolveUDPAddr("udp", p.Duplex)
	// 		if err != nil {
	// 			log.Printf("[ERR] [UDP] [%s] 复制目标地址错误: %s \r\n", p.Name, err)
	// 			return false
	// 		}
	// 		cl, er := net.DialUDP("udp", nil, DupUDPAddr)
	// 		if er == nil {
	// 			DupUDPClient = cl
	// 		}
	// 	}
	// 	//2.监听服务器的地址
	// 	ulistenner, err := net.ListenUDP("tcp4", ulocal)
	// 	if err != nil {
	// 		log.Printf("[ERR] [UDP] [%s] 服务启动失败: %s \r\n", p.Name, err)
	// 		return false
	// 	}
	// 	log.Printf("服务启动: [UDP] [%s] %s", p.Name, p.Listen)

	// 	// ConnPair := make(map[string]string)
	// 	// go (func() {
	// 	// 	for {
	// 	// 		udata := make([]byte, 2048)
	// 	// 		n, client, e := ulistenner.ReadFromUDP(udata)
	// 	// 		if e != nil {
	// 	// 			continue
	// 	// 		}
	// 	// 		client.String()
	// 	// 		if !IsNil(DupUDPClient) {
	// 	// 			DupUDPClient.Write(udata[:n])
	// 	// 		}
	// 	// 		if len(Targets) > 0 {
	// 	// 			for _, v := range Targets {
	// 	// 				cl, er := net.DialUDP("udp", nil, v)
	// 	// 				if er == nil {
	// 	// 					cl.Write(udata[:n])
	// 	// 					//转发结束，准备开启接收，但是此时注意超时问题，
	// 	// 					go (func() {
	// 	// 						for {
	// 	// 							cdata := make([]byte, 2048)
	// 	// 							n, _, e := cl.ReadFromUDP(cdata)
	// 	// 							if e != nil {
	// 	// 								return
	// 	// 							}
	// 	// 							//写入到client
	// 	// 							// client.Write(cdata[:n])
	// 	// 							ulistenner.WriteToUDP(cdata[:n], client)
	// 	// 						}
	// 	// 					})()
	// 	// 					break
	// 	// 				}
	// 	// 			}
	// 	// 		}
	// 	// 	}
	// 	// })()
	// 	return true
	// }
	local, err := net.ResolveTCPAddr("tcp", p.Listen)
	if err != nil {
		log.Printf("[ERR] [%s] 本地地址错误: %s \r\n", p.Name, err)
		return false
	}
	p.TargetConnMap = make(map[string]*net.TCPConn)
	p.Resolved = make([]*net.TCPAddr, 0)
	p.Consumed = false
	p.SourceConnMap = make(map[string]*net.TCPConn)
	for _, v := range p.Targets {
		if strings.HasPrefix(v, "amqp") {
			//TODO AMQP支持
			// v.Type = ConType_AMQP
		} else {
			// v.Type = ConType_TCP
			r, e := net.ResolveTCPAddr("tcp", v)
			if e != nil {
				log.Printf("[ERR] [%s] 转发地址错误: %s \r\n", p.Name, v)
			}
			p.Resolved = append(p.Resolved, r)
		}
	}
	// var DupAddr *net.TCPAddr
	if len(p.Duplex) > 0 {
		if strings.HasPrefix(p.Duplex, "amqp://") {
			p.DupType = ConType_AMQP
		} else {
			p.DupType = ConType_TCP
			_, err = net.ResolveTCPAddr("tcp", p.Duplex)
			if err != nil {
				log.Printf("[ERR] [%s] 复制目标地址错误: %s \r\n", p.Name, err)
				return false
			}
		}
	}
	//2.监听服务器的地址
	listenner, err := net.ListenTCP("tcp4", local)
	if err != nil {
		log.Printf("[ERR] [%s] TCP服务启动失败: %s \r\n", p.Name, err)
		return false
	}
	log.Printf("服务启动: [%s] %s", p.Name, p.Listen)
	// defer listenner.Close()
	p.Listener = listenner

	go (func() {
		for {
			//程序会阻塞在这里，等待新的client连接进来
			conn, err := listenner.AcceptTCP()
			if err != nil {
				log.Printf("[ERR] [%s] 新链接失败 : %s\r\n", p.Name, err)
				continue
			}
			//循环发起连接，如果成功则使用，否则使用Duplex链接
			// var client *net.TCPConn
			// for _, c := range p.Resolved {
			// 	// net.Dia
			// 	_, err = net.DialTCP("tcp", nil, c)
			// 	if err != nil {
			// 		log.Printf("[ERR] [%s] 目标打开失败 : %s \r\n", p.Name, c.String())
			// 	} else {
			// 		break
			// 	}
			// 	// log.Printf("新连接：%s\r\n", conn.RemoteAddr().String())
			// }
			go proxyAMQP(*conn, p)
			// var dup *net.TCPConn = nil
			// if p.DupType == ConType_TCP && !lib.IsNil(DupAddr) {
			// 	dup, err = net.DialTCP("tcp", nil, DupAddr)
			// 	if err != nil {
			// 		log.Printf("[ERR] [%s] 复制目标打开失败 : %s \r\n", p.Name, p.Duplex)
			// 	}
			// } else if p.DupType == ConType_AMQP {
			// 	// lib.Publish(p.Duplex,p.Name,)
			// 	go (func() {

			// 	})()
			// 	return
			// }
			// if client == nil && dup == nil {
			// 	conn.Close()
			// 	log.Printf("[ERR] [%s] 无可用后端 \r\n", p.Name)
			// } else {
			// 	proxy(conn, client, dup, p)
			// }
			log.Printf("[%s] %s 新连接", p.Name, conn.RemoteAddr().String())
			//放到链接管理中，关闭不需要的连接
		}
	})()
	return true
}

func config() {
	viper.AddConfigPath(".")
	if len(*cfile) > 0 {
		// viper.SetConfigFile(*cfile)
		b, e := os.Open(*cfile)
		if e == nil {
			viper.ReadConfig(b)
			viper.Unmarshal(TConf)
			// return
		}
	} else {
		viper.SetConfigName("proxy")
		e := viper.ReadInConfig()
		if e == nil {
			viper.Unmarshal(TConf)
			// fmt.Println(TConf)
			// return
		}
	}
	if len(TConf.Name) == 0 {
		TConf.Name = "Go"
	}
	if len(TConf.Log) == 0 {
		TConf.Log = "./log/"
	}
	if TConf.Size < 100 {
		TConf.Size = 1024
	}
	for _, v := range TConf.Peers {
		if v.Timeout < 100 {
			v.Timeout = 1000
		}
	}
	// return
}

// 创建日志文件
func clog() {
	file := path.Join(TConf.Log, time.Now().Local().Format("2006-01-02")+".log")
	if curfile != file {
		err := os.MkdirAll(TConf.Log, 0766)
		log.Printf("使用日志文件：%s\r\n", file)
		if err != nil {
			log.Fatalln("日志文件错误: " + err.Error())
		}
		if logFile != nil {
			logFile.Close()
		}
		//按日期生成日志文件
		logFile, err = os.OpenFile(file, os.O_RDWR|os.O_CREATE|os.O_APPEND, 0766)
		if err != nil {
			log.Printf("[ERR] %s", err)
			os.Exit(0)
		}
		nlog := log.New(logFile, "", log.Ltime)
		logf = nlog.Printf
		loge = true
		curfile = file
	}
}

func main() {
	log.Printf("版本：%s \r\n技术支持：490523604@qq.com，请写明标题和内容\r\n项目开源地址：https://gitee.com/tansuyun/tcp-port-forwarding\r\n", ver)
	// 解析命令行参数
	flag.Parse()
	// 启动日志文件的初始化处理
	config()
	clog()
	if len(TConf.Peers) == 0 {
		log.Fatal("缺少监听配置信息")
	}
	go (func() {
		failed := make([]Peer, 0)
		for _, v := range TConf.Peers {
			if !start(v) {
				// log.Println("服务启动失败")
				// os.Exit(0)
				failed = append(failed, v)
			}
			// time.Sleep(1 * time.Second)
		}
		// while
		for len(failed) > 0 {
			time.Sleep(10 * time.Second)
			for k, v := range failed {
				if start(v) {
					failed = append(failed[:k], failed[k+1:]...)
				}
			}
		}
	})()
	for {
		time.Sleep(5 * time.Minute)
		clog()
	}
}

//关闭通道
func close(source net.TCPConn, p Peer) {
	// source.SetDeadline(time.Time{})
	remote := source.RemoteAddr().String()
	if !lib.IsNil(p.TargetConnMap[remote]) {
		p.TargetConnMap[remote].Close()
		delete(p.TargetConnMap, remote)
	}
	if !lib.IsNil(p.TargetConnMap[remote+"Dup"]) {
		p.TargetConnMap[remote+"Dup"].Close()
		delete(p.TargetConnMap, remote+"Dup")
	}
	if !lib.IsNil(p.SourceConnMap[remote]) {
		delete(p.SourceConnMap, remote)
	}
	source.Close()
	log.Printf("[%s] %s 连接关闭", p.Name, remote)
}

// 写入数据
func writeTo(p Peer, source net.TCPConn, remote string, b []byte) {
	remoteDup := remote + "Dup"
	target := false
	targetUrl := ""
	//需要存储来源链接=>目标链接对
	if !lib.IsNil(p.TargetConnMap[remote]) {
		//不是空的，此时可以直接发送
		_, e := p.TargetConnMap[remote].Write(b)
		if loge {
			logf("%s\t[%s]\t%s\t"+lib.If(p.Log == "string", "%s", "%x")+"\r\n", ">", p.Name, remote, b)
		}
		if e != nil {
			close(source, p)
			fmt.Printf("%s %s 转发失败，关闭连接\n", p.Name, remote)
			return
			// 目标失败的情况下自动重新选择并发起连接
		} else {
			targetUrl = p.TargetConnMap[remote].RemoteAddr().String()
			target = true
		}
	}
	if !target {
		for _, v := range p.Targets {
			if strings.HasPrefix(v, "amqp://") {
				if nil == lib.Publish(v, p.Name, strings.Join([]string{TConf.Name, remote, strconv.FormatInt(time.Now().Unix(), 10), base64.StdEncoding.EncodeToString(b)}, ",")) {
					if loge {
						logf("%s\t[%s]\t%s\t"+lib.If(p.Log == "string", "%s", "%x")+"\r\n", ">", p.Name, remote, b)
					}
					targetUrl = v
					target = true
					// 跳出循环，处理复制内容
					if !p.Consumed {
						lib.Subscribe(v, p.Name+"_"+TConf.Name, TConf.Name, func(s string) {
							spl := strings.Split(s, ",")
							if len(spl) == 4 && spl[0] == TConf.Name {
								//确认是发给本人的内容
								if !lib.IsNil(p.SourceConnMap[spl[1]]) {
									con := p.SourceConnMap[spl[1]]
									b, e := base64.StdEncoding.DecodeString(spl[3])
									if e == nil {
										con.Write(b)
										if loge {
											logf("%s\t[%s]\t%s\t"+lib.If(p.Log == "string", "%s", "%x")+"\r\n", "<", p.Name, spl[1], b)
										}
									}
								}
							}
						})
						p.Consumed = true
					}
					break
				}
			} else if strings.HasPrefix(v, "udp://") {
				// pkey =
			} else {
				// 处理成TCP转发，
				// 查询是否已经存在链接了，如果存在则取用，否则xx
				// addr, _ := net.ResolveTCPAddr("tcp", v)
				// con, e := net.DialTCP("tcp", nil, addr)
				con, e := net.DialTimeout("tcp", v, time.Duration(p.Timeout)*time.Microsecond)
				// net.TCP
				if e == nil {
					tcp := con.(*net.TCPConn)
					//建立链接成功
					p.TargetConnMap[remote] = tcp
					//开始读取并响应回复内容
					go copy(*tcp, source, p)
					if loge {
						logf("%s\t[%s]\t%s\t"+lib.If(p.Log == "string", "%s", "%x")+"\r\n", ">", p.Name, remote, b)
					}
					con.Write(b)
					target = true
					break
				}
			}
		}
		if !target {
			log.Printf("[%s] %s 找不到目标链接\r\n", p.Name, remote)
			close(source, p)
		}
	}

	if len(p.Duplex) > 5 {
		pass := false
		if !lib.IsNil(p.TargetConnMap[remoteDup]) {
			//不是空的，此时可以直接发送
			_, e := p.TargetConnMap[remoteDup].Write(b)
			if e != nil {
				// close(source, p)
				// 复制通道错误则尝试重连
				pass = false
			} else {
				pass = true
			}
		}

		if !pass {
			if strings.HasPrefix(p.Duplex, "amqp://") {
				if targetUrl != p.Duplex {
					if nil == lib.Publish(p.Duplex, p.Name, strings.Join([]string{TConf.Name, remote, strconv.FormatInt(time.Now().Unix(), 10), base64.StdEncoding.EncodeToString(b)}, ",")) {
						return
					}
				}
			} else if strings.HasPrefix(p.Duplex, "udp://") {
				// pkey =
				fmt.Println("暂不支持UDP")
			} else {
				//已经发送过了，不再发送
				if targetUrl == p.Duplex {
					return
				}
				// 处理成TCP转发，
				// 查询是否已经存在链接了，如果存在则取用，否则xx
				addr, _ := net.ResolveTCPAddr("tcp", p.Duplex)
				con, e := net.DialTCP("tcp", nil, addr)
				if e == nil {
					//建立链接成功
					p.TargetConnMap[remoteDup] = con
					//复制通道不处理读取，直接目标转发成功的情况下不处理读取
					if !target && lib.IsNil(p.TargetConnMap[remote]) {
						// 读取Dup的作为回复内容
						copy(*con, source, p)
					}
					con.Write(b)
				}
			}
		}
	}

}

// 复制读入流并写入日志文件
func copy(source, target net.TCPConn, p Peer) {
	for {
		b := make([]byte, TConf.Size)
		n, e := source.Read(b)
		if e != nil {
			source.Close()
			target.Close()
			return
		}
		if p.MinLen > 0 && n < p.MinLen {
			continue
		}
		//执行忽略逻辑
		// if len() == n*2 && *hex == fmt.Sprintf("%x", b[:n]) {
		// 	log.Println("忽略连接数据")
		// 	source.Close()
		// 	target.Close()
		// 	return
		// }
		target.Write(b[:n])
		if loge {
			logf("<\t[%s]\t%s\t"+lib.If(p.Log == "string", "%s", "%x")+"\r\n", p.Name, source.RemoteAddr().String(), b[:n])
		}
	}
}

//转发到AMQP
func proxyAMQP(source net.TCPConn, p Peer) {
	// go (func() {
	key := source.RemoteAddr().String()
	for {
		b := make([]byte, TConf.Size)
		source.SetReadDeadline(time.Now().Add(300 * time.Second))
		n, e := source.Read(b)
		if e != nil {
			// source.Close()
			close(source, p)
			return
		} else {
			p.SourceConnMap[key] = &source
		}
		if p.MinLen > 0 && n < p.MinLen {
			continue
		}
		writeTo(p, source, key, b[:n])
		// for _, v := range p.Targets {
		// 	if strings.HasPrefix(v, "amqp") {
		// 		if nil == lib.Publish(v, p.Name, strings.Join([]string{TConf.Name, source.RemoteAddr().String(), base64.StdEncoding.EncodeToString(b[:n])}, ",")) {
		// 			break
		// 		}
		// 	}
		// }
		// target.Write(b[:n])
		// if loge {
		// 	logf(">\t[%s]\t%s\t"+lib.If(p.Log == "string", "%s", "%x")+"\r\n", p.Name, source.RemoteAddr().String(), b[:n])
		// }
		// if !lib.IsNil(dup) {
		// 	dup.Write(b[:n])
		// }
	}
	// })()
}

// 转发
// func proxy(source, target, dup net.TCPConn, p Peer) {
// 	if p.Log == "false" {
// 		io.Copy(source, target)
// 		io.Copy(target, source)
// 		return
// 	}
// 	if !lib.IsNil(dup) {
// 		if dup.RemoteAddr().String() == target.RemoteAddr().String() {
// 			dup.Close()
// 			dup = nil
// 		}
// 	}
// 	go (func() {
// 		for {
// 			b := make([]byte, size)
// 			n, e := source.Read(b)
// 			if e != nil {
// 				source.Close()
// 				target.Close()
// 				return
// 			}
// 			if p.MinLen > 0 && n < p.MinLen {
// 				continue
// 			}
// 			//执行忽略逻辑
// 			// if len() == n*2 && *hex == fmt.Sprintf("%x", b[:n]) {
// 			// 	log.Println("忽略连接数据")
// 			// 	source.Close()
// 			// 	target.Close()
// 			// 	return
// 			// }
// 			target.Write(b[:n])
// 			if loge {
// 				logf(">\t[%s]\t%s\t"+lib.If(p.Log == "string", "%s", "%x")+"\r\n", p.Name, source.RemoteAddr().String(), b[:n])
// 			}
// 			if !lib.IsNil(dup) {
// 				dup.Write(b[:n])
// 			}
// 		}
// 	})()
// 	go (func() {
// 		for {
// 			b := make([]byte, size)
// 			n, e := target.Read(b)
// 			if e != nil {
// 				source.Close()
// 				target.Close()
// 				return
// 			}
// 			source.Write(b[:n])
// 			if loge {
// 				logf("<\t[%s]\t%s\t"+lib.If(p.Log == "string", "%s", "%x")+"\r\n", p.Name, source.RemoteAddr().String(), b[:n])
// 			}
// 		}
// 	})()
// }
