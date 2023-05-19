package main

import (
	"errors"
	"io"
	"log"
	"net/http"
	"net/url"
	"strings"
)

type Mux struct{}

func (Mux) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	always, ok := getAlwaysActions()
	if ok {
		for i, v := range always {
			err := processAction(w, r, v)
			if err != nil {
				log.Printf("Failed to process always-action %d: %v", i, err)
				return
			}
		}
	}

	var actions []map[string]interface{}
	p := []string{"proxy", "paths", r.URL.Path}
	err := cf.GetToStruct(&actions, p...)
	if err != nil {
		p = []string{"proxy", "default"}
		err = cf.GetToStruct(&actions, p...)
	}
	if err != nil {
		log.Println("No defulat action found, dropping request:", err)
		return
	}
	processActions(w, r, actions)
}

func getAlwaysActions() ([]map[string]interface{}, bool) {
	var aa []map[string]interface{}
	err := cf.GetToStruct(&aa, "proxy", "always")
	if err != nil {
		log.Println(err)
		return nil, false
	}
	return aa, true
}

func processActions(w http.ResponseWriter, r *http.Request, actions []map[string]interface{}) error {
	for i, action := range actions {
		err := processAction(w, r, action)
		if err != nil {
			log.Printf("Failed to process always-action %d: %v", i, err)
			return err
		}
	}
	return nil
}

func processAction(w http.ResponseWriter, r *http.Request, action map[string]interface{}) error {
	t, ok := action["type"].(string)
	if !ok {
		log.Println("Got action that has no type")
		return errNoAction
	}
	ah := actionHandlers[t]
	if ah != nil {
		return ah(w, r, action)
	}
	return errNoAction
}

type actionHandler func(w http.ResponseWriter, r *http.Request, action map[string]interface{}) error

var (
	errNoAction         = errors.New("action not found")
	errWrongActionParam = errors.New("action parameter is wrong")
	errHijackFailed     = errors.New("webserver doesn't support hijacking")
	actionHandlers      = map[string]actionHandler{
		"drop": func(w http.ResponseWriter, _ *http.Request, _ map[string]interface{}) error {
			hj, ok := w.(http.Hijacker)
			if !ok {
				log.Println("webserver doesn't support hijacking?")
				return errHijackFailed
			}
			conn, _, err := hj.Hijack()
			if err != nil {
				return err
			}
			return conn.Close()
		},
		"log": func(_ http.ResponseWriter, r *http.Request, action map[string]interface{}) error {
			vars := []interface{}{
				r.Proto,
				r.Method,
				r.URL.Path,
				r.UserAgent(),
				r.ContentLength,
				r.Referer(),
			}
			if action["fmt"] == nil {
				log.Printf(strings.Repeat("%q ", len(vars)), vars...)
				return nil
			}
			fmtstr, ok := action["fmt"].(string)
			if !ok {
				return errors.New("fmt parameter is not string")
			}
			log.Printf(fmtstr, vars)
			return nil
		},
		"respond": func(w http.ResponseWriter, _ *http.Request, action map[string]interface{}) error {
			headers, ok := action["headers"].(map[string]interface{})
			if ok {
				for k, vv := range headers {
					v, ok := vv.(string)
					if ok {
						w.Header().Add(k, v)
					} else {
						log.Println("Header is wrong")
					}
				}
			}
			code, ok := action["code"].(float64)
			if !ok {
				code = 200
			}
			w.WriteHeader(int(code))
			data, ok := action["data"].(string)
			var err error
			if ok {
				_, err = w.Write([]byte(data))
			}
			return err
		},
		"serveFile": func(w http.ResponseWriter, r *http.Request, action map[string]interface{}) error {
			path, ok := action["path"].(string)
			if !ok {
				return errWrongActionParam
			}
			http.ServeFile(w, r, path)
			return nil
		},
		"headers": func(w http.ResponseWriter, _ *http.Request, action map[string]interface{}) error {
			op, ok := action["action"].(string)
			if !ok {
				return errWrongActionParam
			}
			switch op {
			case "add":
				headers, ok := action["headers"].(map[string]interface{})
				if !ok {
					return errWrongActionParam
				}
				for k, vv := range headers {
					v, ok := vv.(string)
					if !ok {
						return errWrongActionParam
					}
					w.Header().Add(k, v)
				}
			case "remove":
				headers, ok := action["headers"].([]string)
				if !ok {
					return errWrongActionParam
				}
				for _, vv := range headers {
					w.Header().Del(vv)
				}
			default:
				return errWrongActionParam
			}
			return nil
		},
		"pass": func(w http.ResponseWriter, r *http.Request, action map[string]interface{}) error {
			to, ok := action["dest"].(string)
			if !ok {
				return errWrongActionParam
			}
			toUrl, err := url.Parse(to)
			if err != nil {
				return err
			}
			err = r.ParseForm()
			if err != nil {
				return err
			}
			resp, err := http.DefaultClient.Do(&http.Request{
				Method:           r.Method,
				URL:              toUrl,
				Proto:            r.Proto,
				ProtoMajor:       r.ProtoMajor,
				ProtoMinor:       r.ProtoMinor,
				Header:           r.Header,
				Body:             r.Body,
				ContentLength:    r.ContentLength,
				TransferEncoding: r.TransferEncoding,
				Host:             r.Host,
				Form:             r.Form,
				PostForm:         r.PostForm,
				MultipartForm:    r.MultipartForm,
				Trailer:          r.Trailer,
			})
			if err != nil {
				return err
			}
			for k, vv := range resp.Header {
				for _, v := range vv {
					w.Header().Add(k, v)
				}
			}
			w.WriteHeader(resp.StatusCode)
			b, err := io.ReadAll(resp.Body)
			if err != nil {
				return err
			}
			_, err = w.Write(b)
			if err != nil {
				return err
			}
			return nil
		},
	}
)
