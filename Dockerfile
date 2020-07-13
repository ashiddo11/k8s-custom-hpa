FROM golang:latest

WORKDIR /go/src/github.com/ashiddo11/k8s-custom-hpa/

RUN go get gopkg.in/yaml.v2 \
    k8s.io/client-go/...  \
    google.golang.org/genproto/googleapis/monitoring/v3 \
    cloud.google.com/go/monitoring/apiv3 \
    github.com/PaesslerAG/gval

COPY .  .

RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o custom-hpa .

FROM scratch

COPY --from=0 /go/src/github.com/ashiddo11/k8s-custom-hpa/custom-hpa /

CMD ["/custom-hpa"]
