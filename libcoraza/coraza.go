package main

/*
#ifndef _LIBCORAZA_H_
#define _LIBCORAZA_H_
#include <stdio.h>
#include <stdlib.h>
#include <stdint.h>

typedef struct coraza_intervention_t
{
	char *action;
    int status;
    int pause;
    int disruptive;
} coraza_intervention_t;

typedef struct coraza_config_t {
	void *pinner;
	void *inner;
} coraza_config_t;

typedef struct coraza_waf_t {
	void *pinner;
	void *inner;
} coraza_waf_t;

typedef struct coraza_transaction_t {
	void *pinner;
	void *inner;
} coraza_transaction_t;

typedef const char cchar_t;

typedef void (*coraza_log_cb) (cchar_t*);
void send_log_to_cb(coraza_log_cb cb, cchar_t *msg);
#endif
*/
import "C"
import (
	"io"
	"os"
	"runtime"
	"unsafe"

	"github.com/corazawaf/coraza/v3"

	"github.com/corazawaf/coraza/v3/types"
)

func freeC(p unsafe.Pointer) {
	C.free(p)
}

const sizeCorazaConfig = C.size_t(unsafe.Sizeof(C.coraza_config_t{}))

func allocCorazaConfig() *C.coraza_config_t {
	return (*C.coraza_config_t)(C.malloc(sizeCorazaConfig))
}

type corazaConfigInner struct {
	config coraza.WAFConfig
}

//export coraza_new_config
func coraza_new_config() *C.coraza_config_t {
	inner := &corazaConfigInner{
		config: coraza.NewWAFConfig(),
	}

	pinner := &runtime.Pinner{}
	cfg := allocCorazaConfig()
	cfg.pinner = unsafe.Pointer(pinner)
	cfg.inner = unsafe.Pointer(inner)

	pinner.Pin(pinner)
	pinner.Pin(inner)

	return cfg
}

//export coraza_free_config
func coraza_free_config(cfg *C.coraza_config_t) {
	(*runtime.Pinner)(cfg.pinner).Unpin()
	freeC(unsafe.Pointer(cfg))
}

const sizeWaf = C.size_t(unsafe.Sizeof(C.coraza_config_t{}))

func allocWaf() *C.coraza_waf_t {
	return (*C.coraza_waf_t)(C.malloc(sizeWaf))
}

type innerWaf struct {
	waf coraza.WAF
}

/**
 * Creates a new  WAF instance
 * @returns pointer to WAF instance
 */
//export coraza_new_waf
func coraza_new_waf(cfg *C.coraza_config_t, cerr **C.char) *C.coraza_waf_t {
	innerCfg := (*corazaConfigInner)(cfg.inner)

	waf, err := coraza.NewWAF(innerCfg.config)
	if err != nil {
		*cerr = C.CString(err.Error())
		return nil
	}

	inner := &innerWaf{
		waf: waf,
	}

	pinner := &runtime.Pinner{}
	wafWrapper := allocWaf()
	wafWrapper.pinner = unsafe.Pointer(pinner)
	wafWrapper.inner = unsafe.Pointer(inner)

	pinner.Pin(pinner)
	pinner.Pin(inner)

	return wafWrapper
}

const sizeTransaction = C.size_t(unsafe.Sizeof(C.coraza_transaction_t{}))

func allocTransaction() *C.coraza_transaction_t {
	return (*C.coraza_transaction_t)(C.malloc(sizeTransaction))
}

type innerTranscation struct {
	tx types.Transaction
}

/**
 * Creates a new transaction for a WAF instance
 * @param[in] pointer to valid WAF instance
 * @param[in] pointer to log callback, can be null
 * @returns pointer to transaction
 */
//export coraza_new_transaction
func coraza_new_transaction(waf *C.coraza_waf_t, logCb C.coraza_log_cb) *C.coraza_transaction_t {
	return coraza_new_transaction_with_id(waf, nil, logCb)
}

//export coraza_new_transaction_with_id
func coraza_new_transaction_with_id(wafWrapper *C.coraza_waf_t, id *C.char, logCb C.coraza_log_cb) *C.coraza_transaction_t {
	waf := (*innerWaf)(wafWrapper.inner).waf

	var tx types.Transaction

	if id == nil {
		tx = waf.NewTransaction()
	} else {
		tx = waf.NewTransactionWithID(C.GoString(id))
	}
	// TODO logCb

	inner := &innerTranscation{
		tx: tx,
	}

	pinner := &runtime.Pinner{}
	transaction := allocTransaction()
	transaction.pinner = unsafe.Pointer(pinner)
	transaction.inner = unsafe.Pointer(inner)

	pinner.Pin(pinner)
	pinner.Pin(inner)

	return transaction
}

const sizeInervention = C.size_t(unsafe.Sizeof(C.coraza_intervention_t{}))

func allocIntervention() *C.coraza_intervention_t {
	return (*C.coraza_intervention_t)(C.malloc(sizeInervention))
}

//export coraza_intervention
func coraza_intervention(wrapper *C.coraza_transaction_t) *C.coraza_intervention_t {
	tx := (*innerTranscation)(wrapper.inner).tx

	if tx.Interruption() == nil {
		return nil
	}

	mem := allocIntervention()
	mem.action = C.CString(tx.Interruption().Action)
	mem.status = C.int(tx.Interruption().Status)
	return mem
}

//export coraza_process_connection
func coraza_process_connection(t *C.coraza_transaction_t, sourceAddress *C.char, clientPort C.int, serverHost *C.char, serverPort C.int) C.int {
	tx := (*innerTranscation)(t.inner).tx
	srcAddr := C.GoString(sourceAddress)
	cp := int(clientPort)
	ch := C.GoString(serverHost)
	sp := int(serverPort)
	tx.ProcessConnection(srcAddr, cp, ch, sp)
	return 0
}

//export coraza_process_request_body
func coraza_process_request_body(t *C.coraza_transaction_t) C.int {
	tx := (*innerTranscation)(t.inner).tx
	if _, err := tx.ProcessRequestBody(); err != nil {
		return 1
	}
	return 0
}

//export coraza_update_status_code
func coraza_update_status_code(t *C.coraza_transaction_t, code C.int) C.int {
	// tx := ptrToTransaction(t)
	// c := strconv.Itoa(int(code))
	// tx.Variables.ResponseStatus.Set(c)
	return 1
}

// msr->t, r->unparsed_uri, r->method, r->protocol + offset
//
//export coraza_process_uri
func coraza_process_uri(t *C.coraza_transaction_t, uri *C.char, method *C.char, proto *C.char) C.int {
	tx := (*innerTranscation)(t.inner).tx

	tx.ProcessURI(C.GoString(uri), C.GoString(method), C.GoString(proto))
	return 0
}

//export coraza_add_request_header
func coraza_add_request_header(t *C.coraza_transaction_t, name *C.char, name_len C.int, value *C.char, value_len C.int) C.int {
	tx := (*innerTranscation)(t.inner).tx
	tx.AddRequestHeader(C.GoStringN(name, name_len), C.GoStringN(value, value_len))
	return 0
}

//export coraza_process_request_headers
func coraza_process_request_headers(t *C.coraza_transaction_t) C.int {
	tx := (*innerTranscation)(t.inner).tx
	tx.ProcessRequestHeaders()
	return 0
}

//export coraza_process_logging
func coraza_process_logging(t *C.coraza_transaction_t) C.int {
	tx := (*innerTranscation)(t.inner).tx
	tx.ProcessLogging()
	return 0
}

//export coraza_append_request_body
func coraza_append_request_body(t *C.coraza_transaction_t, data *C.uchar, length C.int) C.int {
	tx := (*innerTranscation)(t.inner).tx
	if _, _, err := tx.WriteRequestBody(C.GoBytes(unsafe.Pointer(data), length)); err != nil {
		return 1
	}
	return 0
}

//export coraza_add_get_args
func coraza_add_get_args(t *C.coraza_transaction_t, name *C.char, value *C.char) C.int {
	tx := (*innerTranscation)(t.inner).tx
	tx.AddGetRequestArgument(C.GoString(name), C.GoString(value))
	return 0
}

//export coraza_add_response_header
func coraza_add_response_header(t *C.coraza_transaction_t, name *C.char, name_len C.int, value *C.char, value_len C.int) C.int {
	tx := (*innerTranscation)(t.inner).tx
	tx.AddResponseHeader(C.GoStringN(name, name_len), C.GoStringN(value, value_len))
	return 0
}

//export coraza_append_response_body
func coraza_append_response_body(t *C.coraza_transaction_t, data *C.uchar, length C.int) C.int {
	tx := (*innerTranscation)(t.inner).tx
	if _, _, err := tx.WriteResponseBody(C.GoBytes(unsafe.Pointer(data), length)); err != nil {
		return 1
	}
	return 0
}

//export coraza_process_response_body
func coraza_process_response_body(t *C.coraza_transaction_t) C.int {
	tx := (*innerTranscation)(t.inner).tx
	if _, err := tx.ProcessResponseBody(); err != nil {
		return 1
	}
	return 0
}

//export coraza_process_response_headers
func coraza_process_response_headers(t *C.coraza_transaction_t, status C.int, proto *C.char) C.int {
	tx := (*innerTranscation)(t.inner).tx
	tx.ProcessResponseHeaders(int(status), C.GoString(proto))
	return 0
}

//export coraza_rules_add_file
func coraza_rules_add_file(wrapper *C.coraza_config_t, file *C.char) {
	inner := (*corazaConfigInner)(wrapper.inner)
	inner.config = inner.config.WithDirectivesFromFile(C.GoString(file))
}

//export coraza_rules_add
func coraza_rules_add(wrapper *C.coraza_config_t, directives *C.char) {
	inner := (*corazaConfigInner)(wrapper.inner)
	inner.config = inner.config.WithDirectives(C.GoString(directives))
}

//export coraza_rules_count
func coraza_rules_count(w *C.coraza_waf_t) C.int {
	return 1
}

//export coraza_free_transaction
func coraza_free_transaction(tx *C.coraza_transaction_t, cerr **C.char) C.int {
	err := (*innerTranscation)(tx.inner).tx.Close()
	(*runtime.Pinner)(tx.pinner).Unpin()
	freeC(unsafe.Pointer(tx))
	if err != nil {
		*cerr = C.CString(err.Error())
		return 1
	}
	return 0
}

//export coraza_free_intervention
func coraza_free_intervention(it *C.coraza_intervention_t) {
	if it.action != nil {
		freeC(unsafe.Pointer(it.action))
	}
	freeC(unsafe.Pointer(it))
}

//export coraza_rules_merge
func coraza_rules_merge(w1 *C.coraza_waf_t, w2 *C.coraza_waf_t, er **C.char) C.int {
	return 1
}

//export coraza_request_body_from_file
func coraza_request_body_from_file(t *C.coraza_transaction_t, file *C.char, cerr **C.char) C.int {
	tx := (*innerTranscation)(t.inner).tx
	f, err := os.Open(C.GoString(file))
	if err != nil {
		return 1
	}
	defer f.Close()
	_, _, err = tx.ReadRequestBodyFrom(f)
	if err != nil && err != io.EOF {
		*cerr = C.CString(err.Error())
		return 1
	}
	return 0
}

//export coraza_free_waf
func coraza_free_waf(t *C.coraza_waf_t) {
	// TODO waf close
	(*runtime.Pinner)(t.pinner).Unpin()
	freeC(unsafe.Pointer(t))
}

//export coraza_set_log_cb
func coraza_set_log_cb(waf *C.coraza_waf_t, cb C.coraza_log_cb) {
}

func main() {}
