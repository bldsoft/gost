package hls

import (
	"reflect"
	"testing"
)

func TestGetOneChildPlaylist(t *testing.T) {
	type args struct {
		master []byte
	}
	tests := []struct {
		name    string
		args    args
		want    []byte
		wantErr bool
	}{
		{
			name: "ok",
			args: args{
				master: []byte(`
#EXTM3U
## Generated with Rix 1.23.0.568 Rev.: develop 902eb0a7 -g (93...209,192...5-encoding_166)

#EXT-X-STREAM-INF:BANDWIDTH=2048000,RESOLUTION=1024x576,CODECS="avc1.4d401f,mp4a.40.2",FRAME-RATE=25.000
0_0_0_0/index/2/index_2560.m3u8`),
			},
			want:    []byte("0_0_0_0/index/2/index_2560.m3u8"),
			wantErr: false,
		},
		{
			name: "subtitles",
			args: args{
				master: []byte(`
				#EXT-X-MEDIA:TYPE=SUBTITLES,GROUP-ID="subtitles",NAME="caption_1",DEFAULT=YES,AUTOSELECT=YES,FORCED=NO,LANGUAGE="eng",URI="index_4_0.m3u8" 	
				`),
			},
			want:    nil,
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := GetOneChildPlaylist(tt.args.master)
			if (err != nil) != tt.wantErr {
				t.Errorf("GetOneChildPlaylist() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("GetOneChildPlaylist() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestCalculateM3U8Duration(t *testing.T) {
	type args struct {
		child []byte
	}
	tests := []struct {
		name    string
		args    args
		want    float64
		wantErr bool
	}{
		{
			name: "ok",
			args: args{
				child: []byte(`
				#EXT-X-VERSION:3
				#EXT-X-TARGETDURATION:7
				#EXT-X-MEDIA-SEQUENCE:8779957
				#EXTINF:6.0,
				index_1_8779957.ts?m=1566416212
				#EXTINF:6.00,
				index_1_8779958.ts?m=1566416212
				#EXTINF:5.000,
				index_1_8779959.ts?m=1566416212
				`),
			},
			want:    17,
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := CalculateM3U8Duration(tt.args.child)
			if (err != nil) != tt.wantErr {
				t.Errorf("CalculateM3U8Duration() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if got != tt.want {
				t.Errorf("CalculateM3U8Duration() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestGetOneSegment(t *testing.T) {
	type args struct {
		childPlaylist []byte
	}
	tests := []struct {
		name    string
		args    args
		want    []byte
		wantErr bool
	}{
		{
			name: "ok",
			args: args{
				childPlaylist: []byte(`
				#EXT-X-VERSION:3
				#EXT-X-TARGETDURATION:7
				#EXT-X-MEDIA-SEQUENCE:8779957
				#EXTINF:6.006,
				index_1_8779957.ts?m=1566416212
				#EXTINF:6.006,
				index_1_8779958.ts?m=1566416212
				#EXTINF:5.372,
				index_1_8779959.ts?m=1566416212
				`),
			},
			want:    []byte("index_1_8779957.ts"),
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := GetOneSegment(tt.args.childPlaylist)
			if (err != nil) != tt.wantErr {
				t.Errorf("GetOneSegment() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("GetOneSegment() = %v, want %v", got, tt.want)
			}
		})
	}
}
