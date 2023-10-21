# xk6-tls-tcp
K6 extension to support TCP connection with TLS support


# build

## install xk6
```
go install go.k6.io/xk6/cmd/xk6@latest
```

## build k6 with this extension
```
xk6 build master --with github.com/kagu2023/xk6-tls-tcp
```

## run
Replace placeholders in `examples/pop3.js`
```
./k6.exe run examples/pop3.js
```