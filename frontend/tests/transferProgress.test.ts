// 收货进度展示逻辑测试：取消态不显示剩余待收，已发货显示真实剩余，收满保持已收货展示。
// 运行：npm exec --package=tsx -- tsx --test tests/transferProgress.test.ts
import { test } from 'node:test'
import assert from 'node:assert/strict'
import { transferProgress, transferRemaining } from '../src/utils/transferProgress'

test('已发货部分收货：显示真实剩余待收', () => {
  const p = transferProgress('shipped', 10, 4)
  assert.equal(p.received, 4)
  assert.equal(p.remaining, 6)
  assert.equal(p.showRemaining, true)
  assert.equal(p.showReturned, false)
})

test('已发货未收货：显示全部剩余', () => {
  const p = transferProgress('shipped', 10, 0)
  assert.equal(p.remaining, 10)
  assert.equal(p.showRemaining, true)
  assert.equal(p.showReturned, false)
})

test('已取消部分收货：不再显示剩余待收，提示未收已退回', () => {
  const p = transferProgress('cancelled', 10, 4)
  assert.equal(p.received, 4)
  assert.equal(p.remaining, 6)
  assert.equal(p.showRemaining, false)
  assert.equal(p.showReturned, true)
})

test('已取消未收货：不显示剩余也不显示退回（无法与未发货取消区分）', () => {
  const p = transferProgress('cancelled', 10, 0)
  assert.equal(p.showRemaining, false)
  assert.equal(p.showReturned, false)
})

test('已收满：保持已收货展示，无剩余无退回', () => {
  const p = transferProgress('received', 10, 10)
  assert.equal(p.remaining, 0)
  assert.equal(p.showRemaining, false)
  assert.equal(p.showReturned, false)
})

test('待确认/已确认：未进入收货阶段，不显示剩余', () => {
  for (const status of ['pending', 'confirmed'] as const) {
    const p = transferProgress(status, 10, 0)
    assert.equal(p.showRemaining, false, status)
    assert.equal(p.showReturned, false, status)
  }
})

test('transferRemaining：已收缺失按 0 处理，不为负', () => {
  assert.equal(transferRemaining(10, 0), 10)
  assert.equal(transferRemaining(10, undefined as unknown as number), 10)
  assert.equal(transferRemaining(10, 12), 0)
})
