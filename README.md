# issue
redis는 single thread지만 컨슈머가 2개 이상일때 특정값을 증가 시키는 로직이 있을때
해당 로직은 atomic 하지 않아서 race condition이 발생 할 수 있다.

## 상황
> 1. 특정 상품의 재고는 10개이다.
> 2. 특정 상품 구매시 구매수가 1 이상이면 재고수를 1 감소시킨다.
> 3. 재고가 0일때 구매를 진행할 수 없다.

## 방법
### redis watch, transaction

redis watch 명령어로 optimistic locking을 하고 redis transaction을 해결할 수 있지만 제약조건이 있다.
redis cluster을 사용한다면 watch, transaction을 사용하지 못한다.

- redis cluster환경에서는 transaction을 지원하지 않습니다.
- 트랜잭션 도중에 해당 key값에 해당하는 value를 조회하려고 하면 항상 null이 return 되므로 value값에 따라 분기처리 불가능
- 복수의 컨슈머가 요청시 성공한요청 외에는 구매 실패 경우가 생긴다.

### lua script

redis는 2.6버전부터 내장된 lua script engine을 이용해 서버에서 lua script를 실행 할 수 있습니다.
이 기능을 이용해 상품 구매시 재고 확인에서 발생하는 race condition 문제를 해결 할 수 있습니다.
lua script에 작성한 로직을 실행하면 해당 연산이 redis 서버에서 원자적으로 처리되기 때문입니다.

```lua
local key = KEYS[1]
local cur = tonumber(redis.call("GET", key) or "-1")

if cur > 0 then
  return redis.call("INCRBY", key, -1)
else
  return -1
end
```
