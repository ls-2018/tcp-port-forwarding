package main

import (
	"flag"
	"io"
	"log"
	"net"
	"os"
	"path"
	"reflect"
	"strings"
	"time"

	"github.com/spf13/viper"
)

// type
type Config struct {
	Peers []Peer
}

type Server struct {
}

var cfile = flag.String("c", "", "配置文件路径")

// 日志路径
var logpath = "./log/"
var ver = "20220817"
var loge = false
var logf = log.Printf
var curfile = ""
var size = 8192
var logFile *os.File
var TConf *Config = &Config{Peers: make([]Peer, 0)}

type Peer struct {
	// 监听名称
	Name string
	// 本地监听地址
	Listen string
	// 转发目标，支持多个地址，如果一个地址不够则自动切换到下个地址，前提是有足够的地址切换，期望支持AMQP
	Targets []string
	// 同步将数据转发到某个地址，支持AMQP
	Duplex string
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
	// Target   *net.Conn
	// Dup      *net.TCPConn
	Resolved []*net.TCPAddr
}

func start(p Peer) bool {
	local, err := net.ResolveTCPAddr("tcp", p.Listen)
	if err != nil {
		log.Printf("[ERR] [%s] 本地地址错误: %s \r\n", p.Name, err)
		return false
	}
	p.Resolved = make([]*net.TCPAddr, 0)
	for _, v := range p.Targets {
		if strings.HasPrefix(v, "amqp") {
			//TODO AMQP支持
		} else {
			r, e := net.ResolveTCPAddr("tcp", v)
			if e != nil {
				log.Printf("[ERR] [%s] 转发地址错误: %s \r\n", p.Name, v)
			}
			p.Resolved = append(p.Resolved, r)
		}
	}
	var DupAddr *net.TCPAddr
	if len(p.Duplex) > 0 {
		DupAddr, err = net.ResolveTCPAddr("tcp", p.Duplex)
		if err != nil {
			log.Printf("[ERR] [%s] 复制目标地址错误: %s \r\n", p.Name, err)
			return false
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
			conn, err := listenner.Accept() //程序会阻塞在这里，等待新的client连接进来
			if err != nil {
				log.Printf("[ERR] [%s] 新链接失败 : %s\r\n", p.Name, err)
				continue
			}
			//循环发起连接，如果成功则使用，否则使用Duplex链接
			var client *net.TCPConn
			for _, c := range p.Resolved {
				// net.Dia
				client, err = net.DialTCP("tcp", nil, c)
				if err != nil {
					log.Printf("[ERR] [%s] 目标打开失败 : %s \r\n", p.Name, c.String())
				} else {
					break
				}
				// log.Printf("新连接：%s\r\n", conn.RemoteAddr().String())
			}
			var dup *net.TCPConn = nil
			if DupAddr != nil {
				dup, err = net.DialTCP("tcp", nil, DupAddr)
				if err != nil {
					log.Printf("[ERR] [%s] 复制目标打开失败 : %s \r\n", p.Name, p.Duplex)
				}
			}
			if client == nil && dup == nil {
				conn.Close()
				log.Printf("[ERR] [%s] 无可用后端 \r\n", p.Name)
			} else {
				proxy(conn, client, dup, p)
			}
			log.Printf("新连接 [%s] %s > %s", p.Name, conn.RemoteAddr().String(), client.RemoteAddr().String())
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
			return
		}
	} else {
		viper.SetConfigName("proxy")
		e := viper.ReadInConfig()
		if e == nil {
			viper.Unmarshal(TConf)
			// fmt.Println(TConf)
			return
		}
	}
	return
}

func clog() {
	file := path.Join(logpath, time.Now().Local().Format("2006-01-02")+".log")
	if curfile != file {
		err := os.MkdirAll(logpath, 0766)
		log.Printf("使用日志文件：%s\r\n", file)
		if err != nil {
			log.Fatalln("日志文件错误" + err.Error())
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
	log.Printf("版本：%s ,技术支持：490523604@qq.com，请写明标题和内容\r\n", ver)
	// 解析命令行参数
	flag.Parse()
	// 启动日志文件的初始化处理
	clog()
	config()
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
func IsNil(i interface{}) bool {
	defer func() {
		recover()
	}()
	vi := reflect.ValueOf(i)
	return vi.IsNil()
}

// 转发
func proxy(source, target, dup net.Conn, p Peer) {
	if p.Log == "false" {
		io.Copy(source, target)
		io.Copy(target, source)
		return
	}
	go (func() {
		for {
			b := make([]byte, size)
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
				logf(">\t[%s]\t%s\t"+If(p.Log == "string", "%s", "%x")+"\r\n", p.Name, source.RemoteAddr().String(), b[:n])
			}
			if len(p.Duplex) > 5 && !IsNil(dup) {
				dup.Write(b[:n])
			}
		}
	})()
	go (func() {
		for {
			b := make([]byte, size)
			n, e := target.Read(b)
			if e != nil {
				source.Close()
				target.Close()
				return
			}
			source.Write(b[:n])
			if loge {
				logf("<\t[%s]\t%s\t"+If(p.Log == "string", "%s", "%x")+"\r\n", p.Name, source.RemoteAddr().String(), b[:n])
			}
		}
	})()
}

func If(b bool, t, f string) string {
	if b {
		return t
	}
	return f
}
