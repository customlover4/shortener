package service

import (
	"errors"
	"shortener/internal/entities/redirect"
	"shortener/internal/entities/url"
	"shortener/internal/storage"
	"testing"
)

type redirectorMock struct {
	createF func(redirects []redirect.Redirect)
	getF    func(alias string) ([]redirect.Redirect, error)
	agrF    func(opts redirect.AgrigateOpts) (redirect.Agrigated, error)
}

func (rm *redirectorMock) CreateRedirects(redirects []redirect.Redirect) {
	rm.createF(redirects)
}

func (rm *redirectorMock) Redirects(alias string) ([]redirect.Redirect, error) {
	return rm.getF(alias)
}

func (rm *redirectorMock) AgrigatedRedirects(opts redirect.AgrigateOpts) (redirect.Agrigated, error) {
	return rm.agrF(opts)
}

func TestService_Redirects(t *testing.T) {
	type fields struct {
		redirector redirector
	}
	type args struct {
		alias string
	}
	tests := []struct {
		name   string
		fields fields
		args   args
		want   error
	}{
		{
			name: "good",
			fields: fields{
				redirector: &redirectorMock{
					getF: func(alias string) ([]redirect.Redirect, error) {
						return nil, nil
					},
				},
			},
			args: args{
				alias: "test",
			},
			want: nil,
		},
		{
			name: "unknown err",
			fields: fields{
				redirector: &redirectorMock{
					getF: func(alias string) ([]redirect.Redirect, error) {
						return nil, storage.ErrNotFound
					},
				},
			},
			args: args{
				alias: "test",
			},
			want: ErrNotFound,
		},
		{
			name: "good",
			fields: fields{
				redirector: &redirectorMock{
					getF: func(alias string) ([]redirect.Redirect, error) {
						return nil, errors.New("unknown")
					},
				},
			},
			args: args{
				alias: "test",
			},
			want: ErrStorageInternal,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := New(nil, tt.fields.redirector)
			_, err := s.Redirects(tt.args.alias)
			if !errors.Is(err, tt.want) {
				t.Errorf("Service.CreateRedirect() error = %v, wantErr %v", err, tt.want)
				return
			}
		})
	}
}

func TestService_AgrigatedRedirects(t *testing.T) {
	type fields struct {
		rs redirector
	}
	type args struct {
		opts redirect.AgrigateOpts
	}
	tests := []struct {
		name   string
		fields fields
		args   args
		want   error
	}{
		{
			name: "good",
			fields: fields{
				rs: &redirectorMock{
					agrF: func(opts redirect.AgrigateOpts) (redirect.Agrigated, error) {
						return redirect.Agrigated{}, nil
					},
				},
			},
			args: args{
				opts: redirect.AgrigateOpts{
					Alias: "test",
				},
			},
			want: nil,
		},
		{
			name: "not valid alias",
			fields: fields{
				rs: &redirectorMock{
					agrF: func(opts redirect.AgrigateOpts) (redirect.Agrigated, error) {
						return redirect.Agrigated{}, nil
					},
				},
			},
			args: args{
				opts: redirect.AgrigateOpts{
					Alias: "",
				},
			},
			want: ErrNotValidData,
		},
		{
			name: "not found",
			fields: fields{
				rs: &redirectorMock{
					agrF: func(opts redirect.AgrigateOpts) (redirect.Agrigated, error) {
						return redirect.Agrigated{}, storage.ErrNotFound
					},
				},
			},
			args: args{
				opts: redirect.AgrigateOpts{
					Alias: "asd",
				},
			},
			want: ErrNotFound,
		},
		{
			name: "unknown",
			fields: fields{
				rs: &redirectorMock{
					agrF: func(opts redirect.AgrigateOpts) (redirect.Agrigated, error) {
						return redirect.Agrigated{}, errors.New("unknown")
					},
				},
			},
			args: args{
				opts: redirect.AgrigateOpts{
					Alias: "asd",
				},
			},
			want: ErrStorageInternal,
		},
		{
			name: "wrong start date",
			fields: fields{
				rs: &redirectorMock{
					agrF: func(opts redirect.AgrigateOpts) (redirect.Agrigated, error) {
						return redirect.Agrigated{}, errors.New("unknown")
					},
				},
			},
			args: args{
				opts: redirect.AgrigateOpts{
					Alias:     "asd",
					StartDate: "kjsdhgkdhjkf",
				},
			},
			want: ErrNotValidData,
		},
		{
			name: "wrong end date",
			fields: fields{
				rs: &redirectorMock{
					agrF: func(opts redirect.AgrigateOpts) (redirect.Agrigated, error) {
						return redirect.Agrigated{}, errors.New("unknown")
					},
				},
			},
			args: args{
				opts: redirect.AgrigateOpts{
					Alias:   "asd",
					EndDate: "kjsdhgkdhjkf",
				},
			},
			want: ErrNotValidData,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := New(nil, tt.fields.rs)
			_, err := s.AgrigatedRedirects(tt.args.opts)
			if !errors.Is(err, tt.want) {
				t.Errorf("Service.AgrigatedRedirects() error = %v, wantErr %v", err, tt.want)
				return
			}
		})
	}
}

type UrlerMock struct {
	createF func(u url.URL) (string, error)
	getF    func(alias string) (string, error)
}

func RegenerationMock() func(u url.URL) (string, error) {
	counter := 0
	return func(u url.URL) (string, error) {
		if counter == 0 {
			counter++
			return "", storage.ErrNotUnique
		}

		return "", nil
	}
}

func (um *UrlerMock) CreateURL(u url.URL) (string, error) {
	return um.createF(u)
}

func (um *UrlerMock) URL(alias string) (string, error) {
	return um.getF(alias)
}

func TestService_CreateURL(t *testing.T) {
	type fields struct {
		urler urler
	}
	type args struct {
		u url.URL
	}
	tests := []struct {
		name   string
		fields fields
		args   args
		want   error
	}{
		{
			name: "good",
			fields: fields{
				urler: &UrlerMock{
					createF: func(u url.URL) (string, error) {
						return "test", nil
					},
				},
			},
			args: args{
				u: url.URL{
					Alias:    "test",
					Original: "http://google.com/test",
				},
			},
			want: nil,
		},
		{
			name: "good alias generating",
			fields: fields{
				urler: &UrlerMock{
					createF: func(u url.URL) (string, error) {
						return "", nil
					},
				},
			},
			args: args{
				u: url.URL{
					Original: "http://google.com/test",
				},
			},
			want: nil,
		},
		{
			name: "empty original",
			fields: fields{
				urler: &UrlerMock{
					createF: func(u url.URL) (string, error) {
						return "", nil
					},
				},
			},
			args: args{
				u: url.URL{},
			},
			want: ErrNotValidData,
		},
		{
			name: "good alias generating",
			fields: fields{
				urler: &UrlerMock{
					createF: func(u url.URL) (string, error) {
						return "", errors.New("unknown")
					},
				},
			},
			args: args{
				u: url.URL{
					Original: "http://google.com/test",
				},
			},
			want: ErrStorageInternal,
		},
		{
			name: "not unique",
			fields: fields{
				urler: &UrlerMock{
					createF: func(u url.URL) (string, error) {
						return "", storage.ErrNotUnique
					},
				},
			},
			args: args{
				u: url.URL{
					Alias:    "test",
					Original: "http://google.com/test",
				},
			},
			want: ErrNotUnique,
		},
		{
			name: "good alias generating",
			fields: fields{
				urler: &UrlerMock{
					createF: RegenerationMock(),
				},
			},
			args: args{
				u: url.URL{
					Original: "http://google.com/test",
				},
			},
			want: nil,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := New(tt.fields.urler, nil)
			_, err := s.CreateURL(tt.args.u)
			if !errors.Is(err, tt.want) {
				t.Errorf("Service.CreateURL() error = %v, wantErr %v", err, tt.want)
				return
			}
		})
	}
}

func TestService_URL(t *testing.T) {
	type fields struct {
		urler urler
	}
	type args struct {
		alias string
	}
	tests := []struct {
		name   string
		fields fields
		args   args
		want   error
	}{
		{
			name: "good",
			fields: fields{
				urler: &UrlerMock{
					getF: func(alias string) (string, error) {
						return "", nil
					},
				},
			},
			args: args{
				"test",
			},
			want: nil,
		},
		{
			name: "good",
			fields: fields{
				urler: &UrlerMock{
					getF: func(alias string) (string, error) {
						return "", nil
					},
				},
			},
			args: args{
				"",
			},
			want: ErrNotValidData,
		},
		{
			name: "good",
			fields: fields{
				urler: &UrlerMock{
					getF: func(alias string) (string, error) {
						return "", storage.ErrNotFound
					},
				},
			},
			args: args{
				"asdas",
			},
			want: ErrNotFound,
		},
		{
			name: "good",
			fields: fields{
				urler: &UrlerMock{
					getF: func(alias string) (string, error) {
						return "", errors.New("unknown")
					},
				},
			},
			args: args{
				"asdas",
			},
			want: ErrStorageInternal,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := New(tt.fields.urler, nil)
			_, err := s.URL(tt.args.alias)
			if !errors.Is(err, tt.want) {
				t.Errorf("Service.URL() error = %v, wantErr %v", err, tt)
				return
			}
		})
	}
}
