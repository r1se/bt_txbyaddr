package main

import (
	"net/http"
	"testing"
	"net/http/httptest"
	"strings"
	"net/http/httputil"
	"io/ioutil"
)

func Benchmark_commHandler(b *testing.B) {

	db = NewDB(Config.Database.Host,
		Config.Database.Port,
		Config.Database.Username,
		Config.Database.Password,
		Config.Database.DatabaseName)

	b.ReportAllocs()
	w := httptest.NewRecorder()
	r := 	httptest.NewRequest("POST",
			"http://localhost/gettransactions",
			strings.NewReader("1NDyJtNTjmwk5xPNhjgAMu4HDHigtobu1s"))

	for i := 0; i < b.N; i++ {
		commHandler(w, r)
	}
}



func Test_commHandler(t *testing.T) {

	db = NewDB(Config.Database.Host,
		Config.Database.Port,
		Config.Database.Username,
		Config.Database.Password,
		Config.Database.DatabaseName)

	type args struct {
		w http.ResponseWriter
		r *http.Request
	}
	tests := []struct {
		name string
		args args
		wantErr bool
	}{
		{
			"test rabbit log OK",
			args{
				httptest.NewRecorder(),
				httptest.NewRequest("POST",
					"http://localhost/gettransactions",
					strings.NewReader("1NDyJtNTjmwk5xPNhjgAMu4HDHigtobu1s")),
			},
			false,
		},

	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {

			//Dump request
			reqDump, _ := httputil.DumpRequest(tt.args.r, true)

			//Run function
			commHandler(tt.args.w, tt.args.r)

			//Cast responsewriter to response recorder
			tmp := tt.args.w.(*httptest.ResponseRecorder)
			if tmp.Code != 200 && !tt.wantErr {
				body, _ := ioutil.ReadAll(tmp.Result().Body)
				t.Errorf("validate() error = %s, request is %s ", body, reqDump)
			}

		})
	}

}
