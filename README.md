# dist-kv

분산 KV 스토어 토이 프로젝트

## 개요

etcd의 핵심 기능인 key-value 스토어를 모방한 간단한 gRPC 기반 KV 시스템입니다.

## 기능

- **gRPC 서버**: Put/Get API 제공
- **gRPC 클라이언트**: 서버와 통신하여 데이터 저장/조회
- **파일 기반 저장**: store.txt 파일에 영구 저장
- **동시성 지원**: 파일 액세스에 대한 뮤텍스 보호

## 빌드 및 실행

### 필요 조건
- Go 1.19+
- protobuf-compiler

### 빌드
```bash
go mod tidy
go build -o bin/server ./server/
go build -o bin/client ./client/
```

### 서버 실행
```bash
./bin/server
```

### 클라이언트 사용법
```bash
# 키-값 저장
./bin/client put <key> <value>

# 값 조회  
./bin/client get <key>
```

### 예시
```bash
# 데이터 저장
./bin/client put hello world
./bin/client put name john

# 데이터 조회
./bin/client get hello    # 출력: world
./bin/client get name     # 출력: john
./bin/client get missing  # 출력: Key not found
```

## 테스트

```bash
go test ./tests/
```

## 프로젝트 구조

```
dist-kv/
├── proto/
│   ├── kv.proto           # gRPC 서비스 정의
│   ├── kv.pb.go           # 생성된 protobuf 코드
│   └── kv_grpc.pb.go      # 생성된 gRPC 코드
├── server/
│   ├── main.go            # 서버 메인 프로그램
│   └── server.go          # 서버 구현
├── client/
│   └── main.go            # 클라이언트 프로그램
├── tests/
│   └── kv_test.go         # 테스트 코드
├── bin/                   # 빌드된 바이너리
├── store.txt              # 데이터 저장 파일 (런타임 생성)
├── go.mod
├── go.sum
├── README.md
├── REQUIREMENTS.md
└── Task.md
```

## API

### Put
- **요청**: key (string), value (string)
- **응답**: success (bool), message (string)
- **동작**: key-value 쌍을 store.txt 파일 끝에 추가

### Get  
- **요청**: key (string)
- **응답**: success (bool), value (string), message (string)
- **동작**: store.txt에서 해당 key의 최신 value 반환

## 특징

- **영구 저장**: 서버 재시작 후에도 데이터 유지
- **최신 값 반환**: 동일한 key에 여러 값이 저장된 경우 가장 최근 값 반환
- **동시성 안전**: 파일 읽기/쓰기에 대한 뮤텍스 보호
- **에러 처리**: 적절한 에러 메시지 제공
