module rush

go 1.14

require (
	github.com/PuerkitoBio/goquery v1.5.1
	github.com/adrg/xdg v0.2.1
	github.com/andybalholm/brotli v1.0.0
	github.com/avast/retry-go v3.0.0+incompatible
	github.com/chromedp/cdproto v0.0.0-20200209033844-7e00b02ea7d2
	github.com/chromedp/chromedp v0.5.3
	github.com/codahale/hdrhistogram v0.0.0-20161010025455-3a0bb77429bd // indirect
	github.com/davecgh/go-spew v1.1.1
	github.com/dgrijalva/jwt-go v3.2.0+incompatible
	github.com/elazarl/goproxy v0.0.0-20200310082302-296d8939dc5a
	github.com/fatih/structs v1.1.0
	github.com/getsentry/sentry-go v0.5.1
	github.com/go-delve/delve v1.5.0 // indirect
	github.com/go-redis/redis/v8 v8.0.0-beta.10
	github.com/go-zeromq/zmq4 v0.9.0
	github.com/gobwas/httphead v0.0.0-20180130184737-2c6c146eadee
	github.com/gobwas/pool v0.2.0
	github.com/gobwas/ws v1.0.3
	github.com/golang/protobuf v1.4.2
	github.com/google/go-cmp v0.5.1
	github.com/google/pprof v0.0.0-20201203190320-1bf35d6f28c2 // indirect
	github.com/google/uuid v1.1.1
	github.com/gorilla/handlers v1.4.2
	github.com/gorilla/websocket v1.4.1
	github.com/mailru/easyjson v0.7.1 // indirect
	github.com/mattn/go-sqlite3 v2.0.3+incompatible
	github.com/minio/minio-go/v7 v7.0.4
	github.com/opentracing-contrib/go-grpc v0.0.0-20191001143057-db30781987df
	github.com/opentracing/opentracing-go v1.1.0
	github.com/pkg/errors v0.9.1
	github.com/pmezard/go-difflib v1.0.0
	github.com/refraction-networking/utls v0.0.0-20200601200209-ada0bb9b38a0
	github.com/rustler47/SecureClient v0.0.0-20200821142845-582719d7def8
	github.com/sirupsen/logrus v1.6.0
	github.com/streadway/simpleuuid v0.0.0-20130420165545-6617b501e485
	github.com/stretchr/testify v1.6.1
	github.com/stripe/stripe-go/v71 v71.14.0
	github.com/tam7t/hpkp v0.0.0-20160821193359-2b70b4024ed5 // indirect
	github.com/uber/jaeger-client-go v2.22.1+incompatible
	github.com/uber/jaeger-lib v2.2.0+incompatible
	gitlab.com/yawning/utls.git v0.0.11-1 // indirect
	go.uber.org/atomic v1.5.0
	golang.org/x/crypto v0.0.0-20200820211705-5c72a883971a
	golang.org/x/net v0.0.0-20200707034311-ab3426394381
	golang.org/x/sys v0.0.0-20201207223542-d4d67f95c62d
	golang.org/x/text v0.3.3
	golang.org/x/tools v0.0.0-20200207183749-b753a1ba74fa
	google.golang.org/grpc v1.28.0
	gopkg.in/d4l3k/messagediff.v1 v1.2.1
	gopkg.in/ezzarghili/recaptcha-go.v4 v4.1.0
	gopkg.in/square/go-jose.v2 v2.5.1
	gopkg.in/src-d/go-vitess.v1 v1.8.0
	gopkg.in/yaml.v2 v2.3.0
	nhooyr.io/websocket v1.8.6
	syreclabs.com/go/faker v1.2.1
)

replace golang.org/x/net => ../golang.org/x/net

replace github.com/refraction-networking/utls => ../github.com/refraction-networking/utls

replace github.com/elazarl/goproxy => ../github.com/elazarl/goproxy
