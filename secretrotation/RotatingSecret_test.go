package secretrotation

import "testing"

func TestRotatingSecret_Deserialize(t *testing.T) {
	type fields struct {
		Previous Secret
		Current  Secret
		Pending  Secret
	}
	type args struct {
		s string
	}
	tests := []struct {
		name           string
		fields         fields
		args           args
		wantErr        bool
		wantSerialized string
	}{
		{
			name:           "3 parts secret",
			args:           args{s: "s1,s2,s3"},
			wantSerialized: "s1,s2,s3",
		},
		{
			name:           "1 part secret",
			args:           args{s: "s1"},
			wantSerialized: "s1,s1,s1",
		},
		{
			name:           "empty secret",
			args:           args{s: ""},
			wantErr:        true,
			wantSerialized: ",,",
		},
		{
			name:           "secret with empty parts",
			args:           args{s: "s1,,s3"},
			wantErr:        true,
			wantSerialized: "s1,,s3",
		},
		{
			name:           "secret with too many parts",
			args:           args{s: "s1,s2,s3,s4"},
			wantErr:        true,
			wantSerialized: ",,",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			rs := &RotatingSecret{
				Previous: tt.fields.Previous,
				Current:  tt.fields.Current,
				Pending:  tt.fields.Pending,
			}
			if err := rs.Deserialize(tt.args.s); (err != nil) != tt.wantErr {
				t.Errorf("RotatingSecret.Deserialize() error = %v, wantErr %v", err, tt.wantErr)
			}
			if s := rs.Serialize(); s != tt.wantSerialized {
				t.Errorf("RotatingSecret.Deserialize() and serialize=%q, want=%q", s, tt.wantSerialized)
			}
		})
	}
}
