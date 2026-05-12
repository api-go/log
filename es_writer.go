package log

import (
	"bytes"
	"fmt"
	"io"
	"log"
	"net/http"
	"net/url"
	"strings"
	"sync"
	"time"

	"apigo.cc/go/cast"
)

type esWriter struct {
	config   *Config
	url      string
	user     string
	password string
	group    string
	lock     sync.Mutex
	queue    []string
	last     int64
	client   *http.Client
	prefix   string
}

func newESWriter(conf *Config) Writer {
	w := &esWriter{
		config: conf,
		queue:  make([]string, 0),
		client: &http.Client{},
	}

	esUrl, err := url.Parse(conf.File)
	if err != nil {
		if DefaultLogger != nil {
			// 这里不能直接调用 DefaultLogger 的方法，防止死循环
			log.Printf("invalid es url: %s, error: %v", conf.File, err)
		}
		return nil
	}

	if esUrl.User != nil {
		w.user = esUrl.User.Username()
		w.password, _ = esUrl.User.Password()
		esUrl.User = nil
	}

	if esUrl.Scheme == "ess" {
		esUrl.Scheme = "https"
	} else {
		esUrl.Scheme = "http"
	}

	if timeout := esUrl.Query().Get("timeout"); timeout != "" {
		w.client.Timeout = cast.Duration(timeout)
	}

	if len(esUrl.Path) > 1 {
		w.group = strings.ReplaceAll(esUrl.Path[1:], "/", ".")
	}

	esUrl.Path = "_bulk"
	esUrl.RawQuery = ""
	w.url = esUrl.String()

	if w.group != "" {
		w.prefix = fmt.Sprintf("{\"index\":{\"_index\":\"%s.%s\"}}", w.group, w.config.Name)
	} else {
		w.prefix = fmt.Sprintf("{\"index\":{\"_index\":\"%s\"}}", w.config.Name)
	}

	return w
}

func (w *esWriter) Log(entry LogEntry, data []byte) {
	objBytes, err := cast.ToJSONBytes(entry)
	if err != nil || len(objBytes) == 0 {
		return
	}
	dataString := string(objBytes)

	w.lock.Lock()
	w.queue = append(w.queue, w.prefix, dataString)
	w.lock.Unlock()
}

var responseOkBytes = []byte("\"errors\":false")

func (w *esWriter) Run() {
	now := time.Now().Unix()
	w.lock.Lock()
	queueLen := len(w.queue)
	w.lock.Unlock()

	// 超过100条数据 或 过了1秒 发送数据
	if queueLen > 100 || (queueLen > 0 && (now > w.last || !LoggerService.Running.Load())) {
		w.lock.Lock()
		sendings := w.queue
		w.queue = make([]string, 0)
		w.lock.Unlock()

		var body bytes.Buffer
		for _, s := range sendings {
			body.WriteString(s)
			body.WriteByte('\n')
		}

		req, err := http.NewRequest("POST", w.url, &body)
		if err != nil {
			log.Println("es request creation failed", err)
			return
		}

		req.Header.Set("Content-Type", "application/json")
		if w.user != "" {
			req.SetBasicAuth(w.user, w.password)
		}

		res, err := w.client.Do(req)
		if err != nil {
			log.Println("es sent failed", err)
			return
		}
		defer res.Body.Close()

		result, err := io.ReadAll(res.Body)
		if err != nil {
			log.Println("es result read failed", err)
			return
		}

		if !bytes.Contains(result, responseOkBytes) {
			log.Printf("es sent errors: %s, data: %s", string(result), body.String())
		}
		w.last = now
	}
}
