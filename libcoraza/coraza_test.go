package main

import (
	"runtime"
	"sync/atomic"
	"testing"
	"unsafe"

	"github.com/corazawaf/coraza/v3"
)

func TestWafInitialization(t *testing.T) {
	var cerr *_Ctype_char = nil
	defer func() {
		if cerr != nil {
			freeC(unsafe.Pointer(cerr))
		}
	}()

	cfg := coraza_new_config()
	defer coraza_free_config(cfg)

	waf := coraza_new_waf(cfg, &cerr)
	if waf == nil {
		t.Fatal(cerr)
	}
	defer coraza_free_waf(waf)
}

/*
	func TestCoraza_add_get_args(t *testing.T) {
		waf := coraza_new_waf()
		tt := coraza_new_transaction(waf, nil)
		coraza_add_get_args(tt, stringToC("aa"), stringToC("bb"))
		tx := ptrToTransaction(tt)
		txi := tx.(plugintypes.TransactionState)
		argsGet := txi.Variables().ArgsGet()
		value := argsGet.Get("aa")
		if len(value) != 1 && value[0] != "bb" {
			t.Fatal("coraza_add_get_args can't add args")
		}
		coraza_add_get_args(tt, stringToC("dd"), stringToC("ee"))
		value = argsGet.Get("dd")
		if len(value) != 1 && value[0] != "ee" {
			t.Fatal("coraza_add_get_args can't add args with another key")
		}
		coraza_add_get_args(tt, stringToC("aa"), stringToC("cc"))
		value = argsGet.Get("aa")
		if len(value) != 2 && value[0] != "bb" && value[1] != "cc" {
			t.Fatal("coraza_add_get_args can't add args with same key more than once")
		}
	}

	func TestTransactionInitialization(t *testing.T) {
		waf := coraza_new_waf()
		tt := coraza_new_transaction(waf, nil)
		if tt == 0 {
			t.Fatal("Transaction initialization failed")
		}
		t2 := coraza_new_transaction(waf, nil)
		if t2 == tt {
			t.Fatal("Transactions are duplicated")
		}
		tx := ptrToTransaction(tt)
		tx.ProcessConnection("127.0.0.1", 8080, "127.0.0.1", 80)
	}

	func TestTxCleaning(t *testing.T) {
		waf := coraza_new_waf()
		txPtr := coraza_new_transaction(waf, nil)
		coraza_free_transaction(txPtr)
		if _, ok := txMap[uint64(txPtr)]; ok {
			t.Fatal("Transaction was not removed from the map")
		}
	}

	func BenchmarkTransactionCreation(b *testing.B) {
		waf := coraza_new_waf()
		for i := 0; i < b.N; i++ {
			coraza_new_transaction(waf, nil)
		}
	}
*/

func TestAlloc(t *testing.T) {
	var counter atomic.Int64

	cfg := coraza_new_config()
	runtime.SetFinalizer((*corazaConfigInner)(cfg.inner), func(inner *corazaConfigInner) {
		counter.Add(1)
	})
	if counter.Load() != 0 {
		t.Error("unexpected counter ", counter.Load(), " != 0: unwanted GC coraza_config_t")
	}

	runtime.GC()
	runtime.GC()

	if counter.Load() != 0 {
		t.Error("unexpected counter ", counter.Load(), " != 0: unwanted GC coraza_config_t")
	}

	coraza_free_config(cfg)
	runtime.GC()
	runtime.GC()

	if counter.Load() != 1 {
		t.Error("unexpected counter ", counter.Load(), " != 1: GC did not clean coraza_config_t")
	}
}

func TestAllocWaf(t *testing.T) {
	var counter atomic.Int64

	cfg := coraza_new_config()
	defer coraza_free_config(cfg)

	waf := coraza_new_waf(cfg, nil)
	if waf == nil {
		t.Fatal("could not initialize waf")
	}

	runtime.SetFinalizer((*innerWaf)(waf.inner), func(inner *innerWaf) {
		counter.Add(1)
	})
	if counter.Load() != 0 {
		t.Error("unexpected counter ", counter.Load(), " != 0: unwanted GC coraza_waf_t")
	}

	runtime.GC()
	runtime.GC()

	if counter.Load() != 0 {
		t.Error("unexpected counter ", counter.Load(), " != 0: unwanted GC coraza_waf_t")
	}

	coraza_free_waf(waf)
	runtime.GC()
	runtime.GC()

	if counter.Load() != 1 {
		t.Error("unexpected counter ", counter.Load(), " != 1: GC did not clean coraza_waf_t")
	}
}

func TestAllocTransaction(t *testing.T) {
	var counter atomic.Int64

	cfg := coraza_new_config()
	defer coraza_free_config(cfg)

	waf := coraza_new_waf(cfg, nil)
	if waf == nil {
		t.Fatal("could not initialize waf")
	}
	defer coraza_free_waf(waf)

	tx := coraza_new_transaction(waf, nil)

	runtime.SetFinalizer((*innerTranscation)(tx.inner), func(inner *innerTranscation) {
		counter.Add(1)
	})
	if counter.Load() != 0 {
		t.Error("unexpected counter ", counter.Load(), " != 0: unwanted GC coraza_transaction_t")
	}

	runtime.GC()
	runtime.GC()

	if counter.Load() != 0 {
		t.Error("unexpected counter ", counter.Load(), " != 0: unwanted GC coraza_transaction_t")
	}

	coraza_free_transaction(tx, nil)

	runtime.GC()
	runtime.GC()

	if counter.Load() != 1 {
		t.Error("unexpected counter ", counter.Load(), " != 1: GC did not clean coraza_transaction_t")
	}
}

const benchmarkConfig = `SecRule UNIQUE_ID "" "id:1"`

func BenchmarkTransactionProcessing(b *testing.B) {
	var cerr *_Ctype_char = nil
	defer func() {
		if cerr != nil {
			freeC(unsafe.Pointer(cerr))
		}
	}()

	rawConfig := _Cfunc_CString(benchmarkConfig)
	defer freeC(unsafe.Pointer(rawConfig))

	cfg := coraza_new_config()
	coraza_rules_add(cfg, rawConfig)
	waf := coraza_new_waf(cfg, &cerr)
	if waf == nil {
		b.Fatal(_Cfunc_GoString(cerr))
	}

	for i := 0; i < b.N; i++ {
		txPtr := coraza_new_transaction(waf, nil)
		tx := (*innerTranscation)(txPtr.inner).tx

		tx.ProcessConnection("127.0.0.1", 55555, "127.0.0.1", 80)
		tx.ProcessURI("https://www.example.com/some?params=123", "GET", "HTTP/1.1")
		tx.AddRequestHeader("Host", "www.example.com")
		tx.ProcessRequestHeaders()
		tx.ProcessRequestBody()
		tx.AddResponseHeader("Content-Type", "text/html")
		tx.ProcessResponseHeaders(200, "OK")
		tx.ProcessResponseBody()
		tx.ProcessLogging()

		if coraza_free_transaction(txPtr, &cerr) != 0 {
			b.Error(_Cfunc_GoString(cerr))
		}
	}
}

func BenchmarkTransactionCompare(b *testing.B) {
	cfg := coraza.NewWAFConfig().WithDirectives(benchmarkConfig)
	waf, err := coraza.NewWAF(cfg)
	if err != nil {
		b.Fatal(err)
	}

	for i := 0; i < b.N; i++ {
		tx := waf.NewTransaction()
		tx.ProcessConnection("127.0.0.1", 55555, "127.0.0.1", 80)
		tx.ProcessURI("https://www.example.com/some?params=123", "GET", "HTTP/1.1")
		tx.AddRequestHeader("Host", "www.example.com")
		tx.ProcessRequestHeaders()
		tx.ProcessRequestBody()
		tx.AddResponseHeader("Content-Type", "text/html")
		tx.ProcessResponseHeaders(200, "OK")
		tx.ProcessResponseBody()
		tx.ProcessLogging()
	}
}
