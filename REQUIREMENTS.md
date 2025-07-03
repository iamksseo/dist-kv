# 분산 KV 스토어 프로젝트 요구사항 명세서

## 프로젝트 개요
etcd의 핵심 기능인 key-value 스토어를 모방한 간단한 분산 KV 시스템을 gRPC를 통해 구현합니다.

## 기능 요구사항

### 1. gRPC 서버 구현
- **목적**: etcd와 유사한 key-value 스토리지 서비스 제공
- **프로토콜**: gRPC
- **지원 API**: Put, Get (etcd gRPC API 기반)

### 2. gRPC 클라이언트 구현
- **목적**: 서버와 통신하여 key-value 데이터 저장/조회
- **프로토콜**: gRPC
- **지원 기능**: Put 요청, Get 요청

### 3. API 명세

#### 3.1 Put API
- **기능**: key-value 쌍을 저장
- **입력**: 
  - key (string)
  - value (string)
- **동작**: 
  - 서버는 요청받은 key-value를 store.txt 파일에 저장
  - 내부적으로 모든 데이터는 string 타입으로 처리
- **출력**: 성공/실패 응답

#### 3.2 Get API
- **기능**: 주어진 key에 대한 value 조회
- **입력**: 
  - key (string)
- **동작**: 
  - 서버는 store.txt 파일을 검색하여 해당 key의 value를 찾음
  - 최신 저장된 value를 반환 (같은 key가 여러 번 저장된 경우)
- **출력**: 
  - 성공 시: value (string)
  - 실패 시: key not found 메시지

### 4. 데이터 저장 방식
- **저장 위치**: store.txt 파일
- **저장 형식**: 각 라인에 key-value 쌍을 저장 (예: "key=value")
- **데이터 타입**: 모든 key, value는 string으로 처리
- **동작 방식**: 
  - Put 요청 시 store.txt 파일 끝에 append
  - Get 요청 시 store.txt 파일을 읽어 검색

## 기술 요구사항

### 1. 프로토콜 버퍼 정의
- etcd gRPC API를 참조하여 Put, Get 메시지 정의
- .proto 파일로 서비스 인터페이스 정의

### 2. 구현 언어
- 서버: Go
- 클라이언트: Go
- 테스트 코드: Go

### 3. 파일 I/O
- store.txt 파일에 대한 동시 접근 처리
- 파일 잠금 또는 동기화 메커니즘 고려

## 비기능 요구사항

### 1. 성능
- 단순한 구조로 높은 성능보다는 기능 구현에 집중

### 2. 확장성
- 향후 분산 기능 추가를 고려한 구조

### 3. 안정성
- 기본적인 에러 처리
- 파일 I/O 예외 처리

## 제약사항

### 1. 데이터 지속성
- 서버 재시작 시 store.txt 파일의 데이터는 유지되어야 함
- 메모리 캐싱은 구현하지 않음 (항상 파일에서 읽기)

### 2. 동시성
- 단일 서버 인스턴스만 지원
- 기본적인 파일 동시 접근 처리만 구현

### 3. 보안
- 인증/인가 기능 제외
- 네트워크 암호화 제외

## 성공 기준

1. gRPC 서버가 정상적으로 시작되고 클라이언트 연결을 수락
2. Put API를 통해 key-value 데이터가 store.txt에 저장됨
3. Get API를 통해 저장된 key의 value를 정확히 조회 가능
4. 동일한 key에 대한 여러 Put 요청 시 최신 value 반환
5. 존재하지 않는 key 조회 시 적절한 에러 메시지 반환

## 프로젝트 구조 (예상)

```
dist-kv/
├── proto/
│   └── kv.proto
├── server/
│   └── main.go
├── client/
│   └── main.go
├── tests/
│   └── kv_test.go
├── store.txt (런타임에 생성)
├── go.mod
├── go.sum
├── README.md
├── REQUIREMENTS.md
└── Task.md
```

## 다음 단계

1. Go 모듈 초기화 (go mod init)
2. 프로토콜 버퍼 파일 (.proto) 정의
3. gRPC 코드 생성 (protoc)
4. 서버 구현 (Go)
5. 클라이언트 구현 (Go)
6. 테스트 코드 작성 (Go)
7. 통합 테스트 및 검증
