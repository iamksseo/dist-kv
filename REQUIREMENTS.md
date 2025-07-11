# 분산 KV 스토어 프로젝트 요구사항 명세서

## 프로젝트 개요
etcd의 핵심 기능인 key-value 스토어를 모방한 간단한 분산 KV 시스템을 gRPC를 통해 구현합니다.

## 프로젝트 비전 및 목표

### 최종 목표
Kubernetes의 kube-apiserver와 통신하는 etcd를 대체하는 완전한 분산 KV 스토어 시스템 구축

### 아키텍처 설계
```
┌─────────────────┐    ┌─────────────────┐    ┌─────────────────┐
│   Kubernetes    │    │   dist-kv       │    │      TiKV       │
│  (kube-apiserver)│◄──►│  (etcd client   │◄──►│  (Storage       │
│                 │    │   V3 API)       │    │   Backend)      │
└─────────────────┘    └─────────────────┘    └─────────────────┘
```

#### Front-end: etcd Client V3 API 호환
- **목적**: Kubernetes에서 etcd를 완전히 대체
- **API 호환성**: etcd client V3 API 100% 호환
- **지원 기능**: KV, Watch, Txn, Lease 4가지 API 셋

#### Back-end: TiKV 연동
- **목적**: 수평확장, 부하분산, 장애 대응 기능 제공
- **연동 방식**: TiKV gRPC API 활용
- **장점**: 
  - 무제한 수평확장 
  - 자동 장애 복구
  - 분산 트랜잭션 지원

### 개발 단계별 로드맵

#### Phase 1: KV API (현재)
- **우선순위**: 1순위 (가장 간단)
- **기능**: 기본 key-value 저장/조회
- **상태**: 🔄 진행 중 (기본 구현 완료, TiKV 연동 필요)

#### Phase 2: Watch API
- **우선순위**: 2순위
- **기능**: key 변경 이벤트 감시
- **구현**: 이벤트 스트리밍, 실시간 알림

#### Phase 3: Txn API
- **우선순위**: 3순위  
- **기능**: 원자적 트랜잭션 처리
- **구현**: 다중 작업 원자성 보장

#### Phase 4: Lease API
- **우선순위**: 4순위 (가장 복잡)
- **기능**: TTL 기반 자동 만료
- **구현**: 임대 관리, 자동 갱신

## 기능 요구사항

### 현재 구현 (Phase 1: KV API)

#### 1. gRPC 서버 구현
- **목적**: etcd와 유사한 key-value 스토리지 서비스 제공
- **프로토콜**: gRPC
- **지원 API**: Put, Get (etcd gRPC API 기반)
- **현재 상태**: 파일 기반 저장소로 구현 완료

#### 2. gRPC 클라이언트 구현
- **목적**: 서버와 통신하여 key-value 데이터 저장/조회
- **프로토콜**: gRPC
- **지원 기능**: Put 요청, Get 요청
- **현재 상태**: CLI 인터페이스 구현 완료

#### 3. TiKV 백엔드 연동 (Phase 1 확장)
- **목적**: 파일 기반 저장소를 TiKV로 대체
- **연동 방식**: TiKV gRPC API 활용
- **장점**: 
  - 분산 저장소 확장성
  - 데이터 복제 및 일관성 보장
  - 자동 장애 복구
- **현재 상태**: 🔄 구현 필요

### 향후 구현 계획

#### Phase 2: Watch API 구현
- **목적**: key 변경 이벤트 실시간 감시
- **기능**: 
  - key 변경 감지
  - 이벤트 스트리밍
  - 필터링 및 조건부 감시
- **사용 사례**: Kubernetes 리소스 변경 감지

#### Phase 3: Txn API 구현  
- **목적**: 원자적 트랜잭션 처리
- **기능**:
  - 조건부 실행 (If-Then-Else)
  - 다중 작업 원자성 보장
  - 롤백 및 커밋 관리
- **사용 사례**: 복잡한 상태 변경 작업

#### Phase 4: Lease API 구현
- **목적**: TTL 기반 자동 만료 관리
- **기능**:
  - 임대 생성 및 관리
  - 자동 갱신 메커니즘
  - 만료 시 자동 정리
- **사용 사례**: 세션 관리, 분산 잠금

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

### 현재 구현 (Phase 1)

#### 1. 프로토콜 버퍼 정의
- etcd gRPC API를 참조하여 Put, Get 메시지 정의
- .proto 파일로 서비스 인터페이스 정의

#### 2. 구현 언어
- 서버: Go
- 클라이언트: Go
- 테스트 코드: Go

#### 3. 파일 I/O
- store.txt 파일에 대한 동시 접근 처리
- 파일 잠금 또는 동기화 메커니즘 고려

### 향후 기술 스택

#### Front-end: etcd Client V3 API 호환 레이어
- **언어**: Go
- **프로토콜**: gRPC (etcd V3 API 스펙 준수)
- **호환성**: 
  - etcd client V3와 100% API 호환
  - Kubernetes 무수정 연동 지원
  - 기존 etcd 클라이언트 라이브러리 지원

#### Back-end: TiKV 연동 레이어
- **스토리지**: TiKV (분산 키-값 데이터베이스)
- **연동 방식**: TiKV gRPC API 활용
- **장점**:
  - 수평확장 (Horizontal Scaling)
  - 자동 부하분산 (Auto Load Balancing)  
  - 장애 대응 (Fault Tolerance)
  - 분산 트랜잭션 지원 (Distributed ACID)

#### 시스템 아키텍처
```
┌─────────────────────────────────────────────────────────────┐
│                    Kubernetes Cluster                       │
│  ┌─────────────────┐  ┌─────────────────┐  ┌─────────────────┐ │
│  │  kube-apiserver │  │  kube-scheduler │  │ kube-controller │ │
│  │                 │  │                 │  │   -manager     │ │
│  └─────────┬───────┘  └─────────┬───────┘  └─────────┬───────┘ │
└────────────┼────────────────────┼────────────────────┼─────────┘
             │                    │                    │
             └────────────────────┼────────────────────┘
                                  │ etcd Client V3 API
         ┌────────────────────────┼────────────────────────┐
         │                 dist-kv System                  │
         │  ┌─────────────────────┼─────────────────────┐   │
         │  │        Front-end (API Gateway)           │   │
         │  │  ┌─────────────────────────────────────┐   │   │
         │  │  │     etcd Client V3 API Compatible   │   │   │
         │  │  │     - KV API                        │   │   │
         │  │  │     - Watch API                     │   │   │
         │  │  │     - Txn API                       │   │   │
         │  │  │     - Lease API                     │   │   │
         │  │  └─────────────────────────────────────┘   │   │
         │  └─────────────────────┬─────────────────────┘   │
         │                        │                         │
         │  ┌─────────────────────┼─────────────────────┐   │
         │  │           Back-end (Storage)             │   │
         │  │  ┌─────────────────────────────────────┐   │   │
         │  │  │              TiKV Cluster           │   │   │
         │  │  │  ┌─────────┐ ┌─────────┐ ┌─────────┐ │   │   │
         │  │  │  │  TiKV   │ │  TiKV   │ │  TiKV   │ │   │   │
         │  │  │  │  Node1  │ │  Node2  │ │  Node3  │ │   │   │
         │  │  │  └─────────┘ └─────────┘ └─────────┘ │   │   │
         │  │  └─────────────────────────────────────┘   │   │
         │  └─────────────────────────────────────────┘   │
         └────────────────────────────────────────────────┘
```

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

### Phase 1: KV API (🔄 진행 중)
1. ✅ gRPC 서버가 정상적으로 시작되고 클라이언트 연결을 수락
2. ✅ Put API를 통해 key-value 데이터가 store.txt에 저장됨
3. ✅ Get API를 통해 저장된 key의 value를 정확히 조회 가능
4. ✅ 동일한 key에 대한 여러 Put 요청 시 최신 value 반환
5. ✅ 존재하지 않는 key 조회 시 적절한 에러 메시지 반환
6. 🔄 TiKV 백엔드 연동 및 파일 저장소 대체
7. 🔄 분산 환경에서 데이터 일관성 검증

### Phase 2: Watch API
1. key 변경 이벤트 실시간 감지
2. 이벤트 스트리밍 안정성 확보
3. 다중 클라이언트 동시 감시 지원
4. 이벤트 필터링 및 조건부 감시

### Phase 3: Txn API  
1. 조건부 실행 (If-Then-Else) 정확한 동작
2. 다중 작업 원자성 보장
3. 트랜잭션 롤백 및 커밋 안정성
4. 동시 트랜잭션 충돌 해결

### Phase 4: Lease API
1. TTL 기반 자동 만료 정확한 동작
2. 임대 갱신 메커니즘 안정성
3. 만료 시 연관 데이터 자동 정리
4. 분산 환경에서 임대 일관성 유지

### 최종 목표
1. **Kubernetes 호환성**: kube-apiserver와 무수정 연동
2. **성능**: etcd 대비 동등 이상의 성능
3. **확장성**: TiKV 기반 무제한 수평확장
4. **안정성**: 24/7 운영 가능한 고가용성

## 프로젝트 구조 (확장 계획)

### 현재 구조 (Phase 1)
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

### 최종 구조 (All Phases)
```
dist-kv/
├── proto/                      # Protocol Buffers 정의
│   ├── kv.proto               # KV API 정의
│   ├── watch.proto            # Watch API 정의  
│   ├── txn.proto              # Transaction API 정의
│   ├── lease.proto            # Lease API 정의
│   └── generated/             # 생성된 gRPC 코드
├── frontend/                   # etcd Client V3 API 호환 레이어
│   ├── kv/                    # KV API 구현
│   ├── watch/                 # Watch API 구현
│   ├── txn/                   # Transaction API 구현
│   ├── lease/                 # Lease API 구현
│   └── gateway/               # API Gateway
├── backend/                    # TiKV 연동 레이어
│   ├── tikv/                  # TiKV 클라이언트
│   ├── storage/               # 스토리지 추상화
│   └── distributed/           # 분산 처리 로직
├── tests/                      # 테스트 코드
│   ├── integration/           # 통합 테스트
│   ├── performance/           # 성능 테스트
│   └── kubernetes/            # K8s 호환성 테스트
├── deployments/               # 배포 설정
│   ├── kubernetes/            # K8s 배포 매니페스트
│   └── docker/                # Docker 설정
├── docs/                      # 문서
│   ├── api/                   # API 문서
│   └── architecture/          # 아키텍처 문서
├── go.mod
├── go.sum
├── README.md
├── REQUIREMENTS.md
└── Task.md
```

## 개발 로드맵

### Phase 1: KV API (🔄 진행 중)
- **브랜치**: `main`  
- **기간**: 2025년 7월 1-2주차
- **구현**: 
  - ✅ 기본 Put/Get 기능 (파일 기반)
  - 🔄 TiKV 백엔드 연동

### Phase 2: Watch API (� 계획)
- **브랜치**: `etcd-Watch-API`
- **기간**: 2025년 7월 2주차
- **구현**: 이벤트 스트리밍, 실시간 감시

### Phase 3: Txn API (📋 계획)
- **브랜치**: `etcd-Txn-API`  
- **기간**: 2025년 7월 3주차
- **구현**: 원자적 트랜잭션 처리

### Phase 4: Lease API (📋 계획)
- **브랜치**: `etcd-Lease-API`
- **기간**: 2025년 7월 4주차  
- **구현**: TTL 기반 자동 만료

### Phase 5: K8s 호환성 (📋 계획)
- **브랜치**: `kubernetes-compatibility`
- **기간**: 2025년 8월 2-3주차
- **구현**: 완전한 etcd 대체

## 다음 단계

### Phase 1: KV API (🔄 진행 중)
1. ✅ Go 모듈 초기화 (go mod init)
2. ✅ 프로토콜 버퍼 파일 (.proto) 정의
3. ✅ gRPC 코드 생성 (protoc)
4. ✅ 서버 구현 (Go) - 파일 기반 저장소
5. ✅ 클라이언트 구현 (Go)
6. ✅ 테스트 코드 작성 (Go)
7. ✅ 통합 테스트 및 검증 - 파일 기반
8. 🔄 TiKV 클라이언트 라이브러리 연동
9. 🔄 파일 저장소를 TiKV로 마이그레이션
10. 🔄 TiKV 기반 테스트 및 검증

### Phase 2: Watch API (🔄 다음 단계)
1. Watch API 프로토콜 버퍼 정의
2. 이벤트 스트리밍 서버 구현
3. 클라이언트 감시 기능 구현
4. 이벤트 필터링 및 조건부 감시
5. 다중 클라이언트 동시 감시 지원
6. 테스트 및 성능 검증

### Phase 3: Txn API (📋 계획)
1. 트랜잭션 프로토콜 정의
2. 조건부 실행 로직 구현
3. 원자성 보장 메커니즘
4. 롤백 및 커밋 처리
5. 동시성 제어 구현

### Phase 4: Lease API (📋 계획)  
1. 임대 관리 시스템 설계
2. TTL 기반 자동 만료 구현
3. 임대 갱신 메커니즘
4. 분산 환경 임대 일관성
5. 성능 최적화

### Phase 5: Kubernetes 호환성 (📋 계획)
1. etcd Client V3 API 완전 호환
2. kube-apiserver 연동 테스트
3. 기존 etcd 대체 검증
4. 성능 및 안정성 테스트
5. 프로덕션 배포 가이드 작성
